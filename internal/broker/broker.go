package broker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github/go/gobromq/config"
	"github/go/gobromq/internal/connection"
	"github/go/gobromq/internal/entities"
)

type Broker struct {
	config     *config.Config
	listener   *connection.Listener
	clients    map[string]*entities.Client
	pubsub     *PubSubEngine
	deliveryCh chan *entities.Delivery
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	running    bool
	mu         sync.RWMutex
}

func NewBroker(cfg *config.Config) *Broker {
	ctx, cancel := context.WithCancel(context.Background())

	return &Broker{
		config:     cfg,
		listener:   connection.NewListener(cfg.Server.Host, cfg.Server.Port),
		clients:    make(map[string]*entities.Client),
		pubsub:     NewPubSubEngine(),
		deliveryCh: make(chan *entities.Delivery, 10000),
		ctx:        ctx,
		cancel:     cancel,
		running:    false,
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

func (b *Broker) AddClient(client *entities.Client) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.clients) >= b.config.Server.MaxClients {
		return fmt.Errorf("maximum clients reached")
	}

	client.Subscriptions = make(map[string]byte)
	client.OutgoingChan = make(chan *entities.Delivery, 100)

	b.clients[client.ID] = client

	fmt.Printf("Client added: %s\n", client.ID)
	return nil
}

func (b *Broker) RemoveClient(clientID string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if client, exists := b.clients[clientID]; exists {
		client.Conn.Close()
		close(client.OutgoingChan)
		b.pubsub.UnsubscribeAll(clientID)
		delete(b.clients, clientID)
		fmt.Printf("Client removed: %s\n", clientID)
	}
}

func (b *Broker) HandlePublish(req *entities.PublishRequest) error {

	msg := &entities.Message{
		Topic:     req.Topic,
		Payload:   req.Payload,
		QoS:       req.QoS,
		Retain:    req.Retain,
		From:      req.From,
		Timestamp: time.Now(),
	}

	deliveries := b.pubsub.Publish(msg)
	for _, delivery := range deliveries {
		select {
		case b.deliveryCh <- &delivery:
		default:
			fmt.Printf("⚠️ Delivery queue full, dropping message to %s\n", delivery.ClientID)
		}
	}

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
		if err := b.pubsub.Subscribe(req.ClientID, sub.Topic, sub.QoS); err != nil {
			returnCodes = append(returnCodes, 0x80)
			continue
		}

		client.MuRW.Lock()
		client.Subscriptions[sub.Topic] = sub.QoS
		client.MuRW.Unlock()

		retainedMsg := b.pubsub.GetRetainedMessages(sub.Topic)

		for _, msg := range retainedMsg {
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
