package entities

import (
	"net"
	"sync"
	"time"
)

type Client struct {
	ID            string
	Conn          net.Conn
	UserName      string
	Authenticated bool
	KeepAlive     time.Duration
	LastSeen      time.Time
	Subscriptions map[string]byte
	OutgoingChan  chan *Delivery
	MuRW          sync.RWMutex
}

type Message struct {
	Topic     string
	Payload   []byte
	QoS       byte
	Retain    bool
	From      *string
	Timestamp time.Time
}

type Delivery struct {
	ClientID  string
	Message   *Message
	QoS       byte
	PacketID  uint16
	Timestamp time.Time
}

type PublishRequest struct {
	Topic   string
	Payload []byte
	QoS     byte
	Retain  bool
	From    *string
}

type SubscribeRequest struct {
	ClientID      string
	Subscriptions []TopicQoS
	PacketID      uint16
}

type TopicQoS struct {
	Topic string
	QoS   byte
}
