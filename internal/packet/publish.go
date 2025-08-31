package packet

import (
	"fmt"

	"github/go/gobromq/util"
)

type PublishPacket struct {
	Topic     string
	PacketID  uint16
	Payload   []byte
	QoS       byte
	Retain    bool
	Duplicate bool
}

func ParsePublish(data []byte, flags byte) (*PublishPacket, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("PUBLISH packet too short")
	}

	offset := 0

	duplicate := (flags & 0x08) != 0
	qos := (flags & 0x06) >> 1
	retain := (flags & 0x01) != 0

	topic, newOffset, err := util.ReadMQTTString(data, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to read topic: %v", err)
	}

	offset = newOffset

	if len(topic) == 0 {
		return nil, fmt.Errorf("topic name is empty")
	}

	var PacketID uint16
	if qos > 0 {
		if len(data) < offset+2 {
			return nil, fmt.Errorf("insufficient data for packet ID")
		}

		PacketID, offset, err = util.ReadUint16(data, offset)
		if err != nil {
			return nil, fmt.Errorf("error reading packet ID: %w", err)
		}
	}

	var payload []byte
	if offset < len(data) {
		payload = data[offset:]
	}

	return &PublishPacket{
		Topic:     topic,
		PacketID:  PacketID,
		Payload:   payload,
		QoS:       qos,
		Retain:    retain,
		Duplicate: duplicate,
	}, nil
}

func CreatePubAck(packetID uint16) []byte {
	return util.WriteUint16(packetID)
}

func CreatePublish(topic string, payload []byte, qos byte, retain bool, packetID uint16) []byte {
	var data []byte

	data = append(data, util.WriteMQTTString(topic)...)
	if qos > 0 {
		data = append(data, util.WriteUint16(packetID)...)
	}
	data = append(data, payload...)
	return data
}
