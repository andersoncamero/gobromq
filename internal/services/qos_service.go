package services

import (
	"fmt"
	"sync"
	"time"

	"github/go/gobromq/internal/entities"
)

// generatePendingKey crea una clave compuesta para el mapa global de pending messages
func generatePendingKey(clientID string, packetID uint16) string {
	return fmt.Sprintf("%s:%d", clientID, packetID)
}

type QoSService struct {
	sessions        map[string]*entities.SessionState
	pendingMessages map[string]*entities.PendingMessage // Cambiamos de uint16 a string para usar clave compuesta
	retryInterval   time.Duration
	maxRetries      int
	mu              sync.RWMutex
	stopCh          chan struct{}
}

func NewQoSService() *QoSService {
	service := &QoSService{
		sessions:        make(map[string]*entities.SessionState),
		pendingMessages: make(map[string]*entities.PendingMessage),
		retryInterval:   30 * time.Second,
		maxRetries:      3,
		stopCh:          make(chan struct{}),
	}

	go service.retryWorker()

	return service
}

func (qs *QoSService) CreateSession(clientID string, cleanSession bool) *entities.SessionState {
	qs.mu.Lock()
	defer qs.mu.Unlock()

	if cleanSession {
		delete(qs.sessions, clientID)
	}

	session, exists := qs.sessions[clientID]

	if !exists {
		session = &entities.SessionState{
			ClientID:        clientID,
			CleanSession:    cleanSession,
			Subscriptions:   make(map[string]entities.QoSLevel),
			PendingMessages: make(map[uint16]*entities.PendingMessage),
			Connected:       true,
			LastPacketID:    0,
		}
		qs.sessions[clientID] = session
	} else {
		session.Connected = true
		session.CleanSession = cleanSession
	}

	return session
}

func (qs *QoSService) RemoveSession(clientID string) {
	qs.mu.Lock()
	defer qs.mu.Unlock()

	session, exists := qs.sessions[clientID]

	if !exists {
		return
	}

	session.Connected = false

	if session.CleanSession {
		for packetID := range session.PendingMessages {
			key := generatePendingKey(clientID, packetID)
			delete(qs.pendingMessages, key)
		}
		delete(qs.sessions, clientID)
	}
}

func (qs *QoSService) AddPendingMessage(clientID string, msg entities.Message, qos entities.QoSLevel) uint16 {
	qs.mu.Lock()
	defer qs.mu.Unlock()

	session, exists := qs.sessions[clientID]

	if !exists {
		return 0
	}

	packetID := session.GetNextPacketID()

	pendingMsg := &entities.PendingMessage{
		PacketID:  packetID,
		Message:   &msg,
		ClientID:  clientID,
		QoS:       qos,
		Timestamp: time.Now(),
		Retries:   0,
		State:     entities.StatePublished,
	}

	session.PendingMessages[packetID] = pendingMsg
	key := generatePendingKey(clientID, packetID)
	qs.pendingMessages[key] = pendingMsg

	fmt.Printf("📋 Added pending message: client=%s, packetID=%d, QoS=%s\n",
		clientID, packetID, qos.String())
	return packetID
}

func (qs *QoSService) HandlePubAck(clientID string, packetID uint16) error {
	qs.mu.Lock()
	defer qs.mu.Unlock()

	session, exists := qs.sessions[clientID]
	if !exists {
		return fmt.Errorf("no session found for client %s", clientID)
	}

	pendingMsg, exists := session.PendingMessages[packetID]
	if !exists {
		return fmt.Errorf("no pending message found for packet ID %d", packetID)
	}

	if pendingMsg.QoS != entities.QoSAtLeastOnce {
		return fmt.Errorf("PUBACK received for non-QoS1 message")
	}

	pendingMsg.State = entities.StateCompleted

	delete(session.PendingMessages, packetID)
	key := generatePendingKey(clientID, packetID)
	delete(qs.pendingMessages, key)

	fmt.Printf("✅ PUBACK processed: client=%s, packetID=%d\n", clientID, packetID)

	return nil
}

func (qs *QoSService) HandlePubRec(clientID string, packetID uint16) error {
	qs.mu.Lock()
	defer qs.mu.Unlock()

	session, exists := qs.sessions[clientID]
	if !exists {
		return fmt.Errorf("no session found for client %s", clientID)
	}

	pendingMsg, exists := session.PendingMessages[packetID]
	if !exists {
		return fmt.Errorf("no pending message found for packet ID %d", packetID)
	}

	if pendingMsg.QoS != entities.QoSExactlyOnce {
		return fmt.Errorf("PUBREC received for non-QoS2 message")
	}

	pendingMsg.State = entities.StateReceived

	fmt.Printf("📨 PUBREC processed: client=%s, packetID=%d, state=%s\n",
		clientID, packetID, pendingMsg.State.String())

	return nil
}

func (qs *QoSService) HandlePubComp(clientID string, packetID uint16) error {
	qs.mu.Lock()
	defer qs.mu.Unlock()

	session, exists := qs.sessions[clientID]
	if !exists {
		return fmt.Errorf("no session found for client %s", clientID)
	}

	pendingMsg, exists := session.PendingMessages[packetID]
	if !exists {
		return fmt.Errorf("no pending message found for packet ID %d", packetID)
	}

	if pendingMsg.QoS != entities.QoSExactlyOnce || pendingMsg.State != entities.StateReleased {
		return fmt.Errorf("PUBCOMP received in invalid state")
	}

	pendingMsg.State = entities.StateCompleted

	delete(session.PendingMessages, packetID)
	key := generatePendingKey(clientID, packetID)
	delete(qs.pendingMessages, key)

	fmt.Printf("✅ PUBCOMP processed: client=%s, packetID=%d, QoS 2 flow completed\n",
		clientID, packetID)

	return nil
}

func (qs *QoSService) GetPendingPubRel(clientID string) []uint16 {
	qs.mu.RLock()
	defer qs.mu.RUnlock()

	var pubrels []uint16

	session, exists := qs.sessions[clientID]
	if !exists {
		return pubrels
	}

	for packetID, pendingMsg := range session.PendingMessages {
		if pendingMsg.QoS == entities.QoSExactlyOnce && pendingMsg.State == entities.StateReceived {
			pubrels = append(pubrels, packetID)

			pendingMsg.State = entities.StateReleased
		}
	}
	return pubrels
}

func (qs *QoSService) SetWillMessage(clientID string, will *entities.WillMessage) {
	qs.mu.Lock()
	defer qs.mu.Unlock()

	session, exists := qs.sessions[clientID]
	if exists {
		session.WillMessage = will
		fmt.Printf("📄 Will message set for client %s: topic=%s\n", clientID, will.Topic)
	}
}

func (qs *QoSService) TriggerWillMessage(clientID string) *entities.WillMessage {
	qs.mu.Lock()
	defer qs.mu.Unlock()

	session, exists := qs.sessions[clientID]
	if !exists || session.WillMessage == nil {
		return nil
	}

	will := session.WillMessage
	session.WillMessage = nil

	fmt.Printf("⚡ Will message triggered for client %s: topic=%s\n", clientID, will.Topic)
	return will
}

func (qs *QoSService) retryWorker() {
	ticker := time.NewTicker(qs.retryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-qs.stopCh:
			return
		case <-ticker.C:
			qs.processRetries()
		}
	}
}

func (qs *QoSService) processRetries() {
	qs.mu.Lock()
	defer qs.mu.Unlock()

	now := time.Now()

	for key, pendingMsg := range qs.pendingMessages {
		if now.Sub(pendingMsg.Timestamp) > qs.retryInterval {
			session, exists := qs.sessions[pendingMsg.ClientID]
			if !exists || !session.Connected {
				continue
			}

			pendingMsg.Retries++
			pendingMsg.Timestamp = now

			if pendingMsg.Retries > qs.maxRetries {
				fmt.Printf("❌ Message retry limit exceeded: client=%s, packetID=%d, removing\n",
					pendingMsg.ClientID, pendingMsg.PacketID)

				delete(session.PendingMessages, pendingMsg.PacketID)
				delete(qs.pendingMessages, key)
			} else {
				fmt.Printf("🔄 Retrying message: client=%s, packetID=%d, attempt=%d\n",
					pendingMsg.ClientID, pendingMsg.PacketID, pendingMsg.Retries)
			}
		}
	}
}

func (qs *QoSService) Stop() {
	close(qs.stopCh)
}

func (qs *QoSService) GetStats() map[string]interface{} {
	qs.mu.RLock()
	defer qs.mu.RUnlock()

	qosStats := make(map[string]int)
	stateStats := make(map[string]int)

	for _, pendingMsg := range qs.pendingMessages {
		qosStats[pendingMsg.QoS.String()]++
		stateStats[pendingMsg.State.String()]++
	}

	return map[string]interface{}{
		"active_sessions":    len(qs.sessions),
		"pending_messages":   len(qs.pendingMessages),
		"qos_distribution":   qosStats,
		"state_distribution": stateStats,
		"retry_interval":     qs.retryInterval.String(),
		"max_retries":        qs.maxRetries,
	}
}
