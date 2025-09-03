package entities

import "time"

type QoSLevel byte

const (
	QoSAtMostOnce  QoSLevel = 0
	QoSAtLeastOnce QoSLevel = 1
	QoSExactlyOnce QoSLevel = 2
)

func (q QoSLevel) String() string {
	switch q {
	case QoSAtMostOnce:
		return "AT_MOST_ONCE"
	case QoSAtLeastOnce:
		return "AT_LEAST_ONCE"
	case QoSExactlyOnce:
		return "EXACTLY_ONCE"
	default:
		return "UNKNOWN"
	}
}

type PendingMessage struct {
	PacketID  uint16
	Message   *Message
	ClientID  string
	QoS       QoSLevel
	Timestamp time.Time
	Retries   int
	State     MessageState
}

type MessageState int

const (
	StatePublished MessageState = iota
	StateReceived
	StateReleased
	StateCompleted
)

func (s MessageState) String() string {
	switch s {
	case StatePublished:
		return "PUBLISHED"
	case StateReceived:
		return "RECEIVED"
	case StateReleased:
		return "RELEASED"
	case StateCompleted:
		return "COMPLETED"
	default:
		return "UNKNOWN"
	}
}

type WillMessage struct {
	Topic   string
	Payload []byte
	QoS     QoSLevel
	Retain  bool
}

type SessionState struct {
	ClientID        string
	CleanSession    bool
	Subscriptions   map[string]QoSLevel
	PendingMessages map[uint16]*PendingMessage
	WillMessage     *WillMessage
	Connected       bool
	LastPacketID    uint16
}

func (ss *SessionState) GetNextPacketID() uint16 {
	ss.LastPacketID++

	if ss.LastPacketID == 0 {
		ss.LastPacketID = 1
	}

	return ss.LastPacketID
}
