package packet

import (
	"fmt"

	"github/go/gobromq/util"
)

const (
	SubAckMaxQoS0 = 0x00
	SubAckMaxQoS1 = 0x01
	SubAckMaxQoS2 = 0x02
	SubAckFailure = 0x80
)

type SubscribePacket struct {
	PacketID      uint16
	Subscriptions []Subscription
}

type Subscription struct {
	Topic string
	QoS   byte
}

func ParseSubscribe(data []byte) (*SubscribePacket, error) {
	if len(data) < 3 {
		return nil, fmt.Errorf("SUBSCRIBE packet too short")
	}

	offset := 0

	packetID, newOffset, err := util.ReadUint16(data, offset)
	if err != nil {
		return nil, fmt.Errorf("error reading packet ID: %w", err)
	}

	offset = newOffset

	var subscriptions []Subscription

	for offset < len(data) {
		topic, newOffset, err := util.ReadMQTTString(data, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to read topic: %v", err)
		}
		offset = newOffset

		if offset >= len(data) {
			return nil, fmt.Errorf("missing QoS for subscription")
		}
		qos := data[offset]
		offset++

		if qos > 2 {
			return nil, fmt.Errorf("invalid QoS level: %d", qos)
		}

		subscriptions = append(subscriptions, Subscription{
			Topic: topic,
			QoS:   qos,
		})
	}

	if len(subscriptions) == 0 {
		return nil, fmt.Errorf("nSUBSCRIBE must contain at least one subscription")
	}

	return &SubscribePacket{
		PacketID:      packetID,
		Subscriptions: subscriptions,
	}, nil
}

func CreateSubAck(packetID uint16, returnCodes []byte) []byte {
	data := util.WriteUint16(packetID)
	data = append(data, returnCodes...)
	return data
}
