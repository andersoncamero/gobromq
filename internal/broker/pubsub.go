package broker

import (
	"fmt"
	"strings"
	"sync"

	"github/go/gobromq/internal/entities"
)

type PubSubEngine struct {
	subscriptions map[string]map[string]byte
	retainedMsgs  map[string]*entities.Message
	mu            sync.RWMutex
}

func NewPubSubEngine() *PubSubEngine {
	return &PubSubEngine{
		subscriptions: make(map[string]map[string]byte),
		retainedMsgs:  make(map[string]*entities.Message),
	}
}

func (ps *PubSubEngine) Subscribe(clientID string, topic string, qos byte) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if err := ps.validateSubscriptionTopic(topic); err != nil {
		return err
	}

	if ps.subscriptions[topic] == nil {
		ps.subscriptions[topic] = make(map[string]byte)
	}

	ps.subscriptions[topic][clientID] = qos
	fmt.Printf("Client %s subscribed to topic %s with QoS %d\n", clientID, topic, qos)
	return nil
}

func (ps *PubSubEngine) Unsubscribe(clientID string, topic string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if subs, exists := ps.subscriptions[topic]; exists {
		delete(subs, clientID)
		if len(subs) == 0 {
			delete(ps.subscriptions, topic)
		}
	}

	fmt.Printf("Client %s unsubscribed from topic %s\n", clientID, topic)
}

func (ps *PubSubEngine) UnsubscribeAll(clientID string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	for topic, subs := range ps.subscriptions {
		if _, exists := subs[clientID]; exists {
			delete(subs, clientID)
			if len(subs) == 0 {
				delete(ps.subscriptions, topic)
			}
		}
	}

	fmt.Printf("Client %s unsubscribed from all topics\n", clientID)
}

func (ps *PubSubEngine) Publish(msg *entities.Message) []entities.Delivery {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if msg.Retain {
		if len(msg.Payload) == 0 {
			delete(ps.retainedMsgs, msg.Topic)
		} else {
			ps.retainedMsgs[msg.Topic] = msg
		}
	}

	var deliveries []entities.Delivery

	for subTopic, subscribers := range ps.subscriptions {
		if ps.topicMatches(subTopic, string(msg.Topic)) {
			for clientID, subQoS := range subscribers {
				deliveryQoS := msg.QoS
				if subQoS < msg.QoS {
					deliveryQoS = subQoS
				}

				delivery := entities.Delivery{
					ClientID: clientID,
					Message:  msg,
					QoS:      deliveryQoS,
				}

				deliveries = append(deliveries, delivery)
			}
		}
	}

	fmt.Printf("📤 Publishing to %s: %d deliveries\n", msg.Topic, len(deliveries))
	return deliveries
}

func (ps *PubSubEngine) GetRetainedMessages(topicFilter string) []*entities.Message {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	var messages []*entities.Message

	for topic, msg := range ps.retainedMsgs {
		if ps.topicMatches(topicFilter, topic) {
			messages = append(messages, msg)
		}
	}
	return messages
}

func (ps *PubSubEngine) topicMatches(filter, topic string) bool {
	return ps.matchTopicLevels(strings.Split(filter, "/"), strings.Split(topic, "/"))
}

func (ps *PubSubEngine) matchTopicLevels(filterLevels, topicLevels []string) bool {
	filterIndex, topicIndex := 0, 0

	for filterIndex < len(filterLevels) && topicIndex < len(topicLevels) {
		filterLevel := filterLevels[filterIndex]
		topicLevel := topicLevels[topicIndex]

		if filterLevel == "#" {
			return true
		}

		if filterLevel != "+" {
			filterIndex++
			topicIndex++
		}

		if filterLevel != topicLevel {
			return false
		}

		filterIndex++
		topicIndex++
	}

	if filterIndex < len(filterLevels) {
		return len(filterLevels) == filterIndex+1 && filterLevels[filterIndex] == "#"
	}

	return filterIndex == len(filterLevels) && topicIndex == len(topicLevels)
}

func (ps *PubSubEngine) validateSubscriptionTopic(topic string) error {
	if len(topic) == 0 {
		return fmt.Errorf("topic cannot be empty")
	}

	if strings.HasPrefix(topic, "$") {
		return fmt.Errorf("cannot subscribe to system topic: %s", topic)
	}
	return nil
}

func (ps *PubSubEngine) GetSubscriptions() int {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	count := 0

	for _, subs := range ps.subscriptions {
		count += len(subs)
	}

	return count

}
