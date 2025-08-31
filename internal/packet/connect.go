package packet

import (
	"encoding/binary"
	"fmt"

	"github/go/gobromq/util"
)

type ConnectPacket struct {
	ProtocolName  string
	ProtocolLevel byte
	ConnectFlags  byte
	KeepAlive     uint16
	ClientID      string
	WillTopic     string
	WillMessage   string
	Username      string
	Password      string

	UserNameFlag bool
	PasswordFlag bool
	WillRetain   bool
	WillQoS      byte
	WillFlag     bool
	CleanSession bool
}

func ParseConnect(data []byte) (*ConnectPacket, error) {
	if len(data) < 10 {
		return nil, fmt.Errorf("CONNECT packet too short")
	}

	offset := 0

	protocolName, newOffset, err := util.ReadMQTTString(data, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to read protocol name: %w", err)
	}

	offset = newOffset

	if protocolName != "MQTT" && protocolName != "MQIsdp" {
		return nil, fmt.Errorf("unsupported protocol name: %s", protocolName)
	}

	if len(data) < offset+1 {
		return nil, fmt.Errorf("insufficient data for protocol level")
	}

	protocolLevel := data[offset]
	offset++

	if protocolLevel != 4 && protocolLevel != 3 {
		return nil, fmt.Errorf("unsupported protocol level: %d", protocolLevel)
	}

	if len(data) < offset+1 {
		return nil, fmt.Errorf("insufficient data for connect flags")
	}

	connectFlags := data[offset]
	offset++

	userNameFlag := (connectFlags & 0x80) != 0
	passwordFlag := (connectFlags & 0x40) != 0
	willRetain := (connectFlags & 0x20) != 0
	willQoS := (connectFlags & 0x18) >> 3
	willFlag := (connectFlags & 0x04) != 0
	cleanSession := (connectFlags & 0x02) != 0

	if (connectFlags & 0x01) != 0 {
		return nil, fmt.Errorf("reserved flag must be 0")
	}

	if len(data) < offset+2 {
		return nil, fmt.Errorf("insufficient data for keep alive")
	}
	keepAlive := binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2

	clientID, newOffset, err := util.ReadMQTTString(data, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to read client ID: %w", err)
	}
	offset = newOffset

	if len(clientID) == 0 {
		return nil, fmt.Errorf("client ID cannot be empty when clean session is false")
	}

	connect := &ConnectPacket{
		ProtocolName:  protocolName,
		ProtocolLevel: protocolLevel,
		ConnectFlags:  connectFlags,
		KeepAlive:     keepAlive,
		ClientID:      clientID,
		UserNameFlag:  userNameFlag,
		PasswordFlag:  passwordFlag,
		WillRetain:    willRetain,
		WillQoS:       willQoS,
		WillFlag:      willFlag,
		CleanSession:  cleanSession,
	}

	if willFlag {
		connect.WillTopic, offset, err = util.ReadMQTTString(data, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to read will topic: %w", err)
		}

		connect.WillMessage, offset, err = util.ReadMQTTString(data, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to read will message: %w", err)
		}
	}

	if userNameFlag {
		connect.Username, offset, err = util.ReadMQTTString(data, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to read username: %w", err)
		}
	}

	if passwordFlag {
		connect.Password, _, err = util.ReadMQTTString(data, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to read password: %w", err)
		}
	}

	return connect, nil
}

func CreateConnAck(sessionPresent bool, returnCode byte) []byte {
	flags := byte(0)
	if sessionPresent {
		flags = 0x01
	}

	return []byte{flags, returnCode}
}
