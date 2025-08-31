package entities

import "fmt"

const (
	CONNECT     = 0x10
	CONNACK     = 0x20
	PUBLISH     = 0x30
	PUBACK      = 0x40
	PUBREC      = 0x50
	PUBREL      = 0x60
	PUBCOMP     = 0x70
	SUBSCRIBE   = 0x80
	SUBACK      = 0x90
	UNSUBSCRIBE = 0xA0
	UNSUBACK    = 0xB0
	PINGREQ     = 0xC0
	PINGRESP    = 0xD0
	DISCONNECT  = 0xE0
)
const (
	QoS0 = 0x00
	QoS1 = 0x01
	QoS2 = 0x02
)

const (
	ConnectAccepted                  = 0x00
	ConnectRefusedProtocolVersion    = 0x01
	ConnectRefusedIdentifierRejected = 0x02
	ConnectRefusedServerUnavailable  = 0x03
	ConnectRefusedBadCredentials     = 0x04
	ConnectRefusedNotAuthorized      = 0x05
)

type Packet struct {
	Type   byte
	Flags  byte
	Length int
	Data   []byte
}

func GetPacketTypeName(packetType byte) string {
	packetType = packetType & 0xF0
	switch packetType {
	case CONNECT:
		return "CONNECT"
	case CONNACK:
		return "CONNACK"
	case PUBLISH:
		return "PUBLISH"
	case PUBACK:
		return "PUBACK"
	case PUBREC:
		return "PUBREC"
	case PUBREL:
		return "PUBREL"
	case PUBCOMP:
		return "PUBCOMP"
	case SUBSCRIBE:
		return "SUBSCRIBE"
	case SUBACK:
		return "SUBACK"
	case UNSUBSCRIBE:
		return "UNSUBSCRIBE"
	case UNSUBACK:
		return "UNSUBACK"
	case PINGREQ:
		return "PINGREQ"
	case PINGRESP:
		return "PINGRESP"
	case DISCONNECT:
		return "DISCONNECT"
	default:
		return fmt.Sprintf("UNKNOWN(0x%02X)", packetType)
	}
}
