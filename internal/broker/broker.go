package broker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github/go/gobromq/config"
	"github/go/gobromq/internal/auth"
	"github/go/gobromq/internal/connection"
	"github/go/gobromq/internal/entities"
	"github/go/gobromq/internal/service"
)

type Broker struct {
	config      *config.Config
	listener    *connection.Listener
	clients     map[string]*entities.Client
	pubsub      *PubSubEngine
	authService *service.AuthService
	deliveryCh  chan *entities.Delivery
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	running     bool
	mu          sync.RWMutex
}

func NewBroker(cfg *config.Config) *Broker {
	ctx, cancel := context.WithCancel(context.Background())

	authenticator := auth.NewSimpleAuth()

	return &Broker{
		config:      cfg,
		listener:    connection.NewListener(cfg.Server.Host, cfg.Server.Port),
		clients:     make(map[string]*entities.Client),
		pubsub:      NewPubSubEngine(),
		authService: service.NewAuthService(authenticator),
		deliveryCh:  make(chan *entities.Delivery, 10000),
		ctx:         ctx,
		cancel:      cancel,
		running:     false,
	}
}

func (b *Broker) Start() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.running {
		return fmt.Errorf("broker is already running")
	}

	if err := b.listener.Start(); err != nil {
		return fmt.Errorf("failed to start listener: %w", err)
	}

	b.running = true

	b.wg.Add(2)
	go b.connectionWorker()
	go b.deliveryWorker()

	fmt.Println("GoBroMq broker started successfully")
	fmt.Printf("🔐 Authentication: %s\n",
		map[bool]string{true: "ENABLED", false: "DISABLED"}[b.config.Auth.Enabled])
	return nil
}

func (b *Broker) Stop() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.running {
		return fmt.Errorf("broker is not running")
	}

	fmt.Println("Stopping GoBroMq broker...")
	b.cancel()

	b.listener.Stop()
	close(b.deliveryCh)

	for _, client := range b.clients {
		client.Conn.Close()
	}

	b.wg.Wait()

	b.running = false
	fmt.Println("GoBroMq broker stopped")

	return nil
}

func (b *Broker) AuthenticateClient(clientID, username, password string) error {
	if !b.config.Auth.Enabled {
		return nil
	}

	_, err := b.authService.AuthenticateClient(clientID, username, password)

	if err != nil {
		fmt.Printf("❌ Authentication failed for client %s (user: %s): %v\n",
			clientID, username, err)
		return err
	}

	fmt.Printf("✅ Client %s authenticated as user %s\n", clientID, username)
	return nil
}

func (b *Broker) AddClient(client *entities.Client) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.clients) >= b.config.Server.MaxClients {
		return fmt.Errorf("maximum clients reached")
	}

	if b.config.Auth.Enabled && b.config.Auth.RequireAuth {
		if !b.authService.IsAuthenticated(client.ID) {
			return fmt.Errorf("client must be authenticated")
		}
	}

	client.Subscriptions = make(map[string]byte)
	client.OutgoingChan = make(chan *entities.Delivery, 100)

	b.clients[client.ID] = client

	if session, exists := b.authService.GetSession(client.ID); exists {
		fmt.Printf("➕ Client %s added (user: %s, roles: %v)\n",
			client.ID, session.Username, session.Roles)
	} else {
		fmt.Printf("➕ Client %s added (unauthenticated)\n", client.ID)
	}

	return nil
}

func (b *Broker) RemoveClient(clientID string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if client, exists := b.clients[clientID]; exists {
		client.Conn.Close()
		close(client.OutgoingChan)
		b.pubsub.UnsubscribeAll(clientID)
		b.authService.RemoveSession(clientID)
		delete(b.clients, clientID)
		fmt.Printf("➖ Client %s removed from broker\n", clientID)
	}
}

func (b *Broker) HandlePublish(req *entities.PublishRequest) error {

	if b.config.Auth.Enabled {
		if !b.authService.CanPublish(*req.From, string(req.Topic)) {
			return fmt.Errorf("client %s not authorized to publish to topic %s",
				*req.From, req.Topic)
		}
	}

	msg := &entities.Message{
		Topic:     req.Topic,
		Payload:   req.Payload,
		QoS:       req.QoS,
		Retain:    req.Retain,
		From:      req.From,
		Timestamp: time.Now(),
	}

	deliveries := b.pubsub.Publish(msg)
	authorizedDeliveries := b.filterAuthorizedDeliveries(deliveries)

	for _, delivery := range authorizedDeliveries {
		select {
		case b.deliveryCh <- &delivery:
		default:
			fmt.Printf("⚠️ Delivery queue full, dropping message to %s\n", delivery.ClientID)
		}
	}

	fmt.Printf("📤 Published to %s: %d/%d authorized deliveries\n",
		req.Topic, len(authorizedDeliveries), len(deliveries))

	return nil
}

func (b *Broker) HandleSubscribe(req *entities.SubscribeRequest) []byte {
	var returnCodes []byte

	b.mu.RLock()
	client, exists := b.clients[req.ClientID]
	defer b.mu.RUnlock()

	if !exists {
		for range req.Subscriptions {
			returnCodes = append(returnCodes, 0x80)
		}
		return returnCodes
	}

	for _, sub := range req.Subscriptions {
		if b.config.Auth.Enabled {
			if !b.authService.CanSubscribe(req.ClientID, sub.Topic) {
				returnCodes = append(returnCodes, 0x80)
				continue
			}
		}

		if err := b.pubsub.Subscribe(req.ClientID, sub.Topic, sub.QoS); err != nil {
			returnCodes = append(returnCodes, 0x80)
			continue
		}

		client.MuRW.Lock()
		client.Subscriptions[sub.Topic] = sub.QoS
		client.MuRW.Unlock()

		retainedMsg := b.pubsub.GetRetainedMessages(sub.Topic)

		for _, msg := range retainedMsg {
			if b.config.Auth.Enabled && !b.authService.CanSubscribe(req.ClientID, string(msg.Topic)) {
				continue
			}

			delivery := entities.Delivery{
				ClientID: req.ClientID,
				Message:  msg,
				QoS:      sub.QoS,
			}

			select {
			case b.deliveryCh <- &delivery:
			default:
			}
		}

		returnCodes = append(returnCodes, sub.QoS)
	}
	return returnCodes
}

func (b *Broker) filterAuthorizedDeliveries(deliveries []entities.Delivery) []entities.Delivery {
	if !b.config.Auth.Enabled {
		return deliveries
	}

	var authorized []entities.Delivery

	for _, delivery := range deliveries {
		if b.authService.CanSubscribe(delivery.ClientID, string(delivery.Message.Topic)) {
			authorized = append(authorized, delivery)
		} else {
			fmt.Printf("🚫 Delivery blocked: client %s cannot read topic %s\n",
				delivery.ClientID, delivery.Message.Topic)
		}
	}
	return authorized
}

func (b *Broker) IsRunning() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.running
}

func (b *Broker) connectionWorker() {
	defer b.wg.Done()

	for {
		select {
		case <-b.ctx.Done():
			return
		case conn, ok := <-b.listener.GetConnections():
			if !ok {
				return
			}

			handler := connection.NewHandler(conn, b)

			go func() {
				if err := handler.Start(); err != nil {
					fmt.Printf("⚠️ Handler error: %v\n", err)
				}
			}()
		}
	}
}

func (b *Broker) deliveryWorker() {
	defer b.wg.Done()
	for {
		select {
		case <-b.ctx.Done():
			return
		case delivery, ok := <-b.deliveryCh:
			if !ok {
				return
			}
			b.deliveryMessage(delivery)
		}
	}
}

func (b *Broker) deliveryMessage(delivery *entities.Delivery) {
	b.mu.RLock()
	client, exists := b.clients[delivery.ClientID]
	b.mu.RUnlock()

	if !exists {
		return
	}

	select {
	case client.OutgoingChan <- delivery:
	default:
		fmt.Printf("⚠️ Client %s is slow, dropping message\n", delivery.ClientID)
	}
}

func (b *Broker) GetStats() map[string]interface{} {
	b.mu.RLock()
	defer b.mu.Unlock()

	return map[string]interface{}{
		"clients":  len(b.clients),
		"messages": b.pubsub.GetSubscriptions(),
	}
}
