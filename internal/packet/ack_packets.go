package packet

import (
	"fmt"

	"github/go/gobromq/util"
)

func CreatePubRec(packetID uint16) []byte {
	return util.WriteUint16(packetID)
}

func CreatePubRel(packetID uint16) []byte {
	return util.WriteUint16(packetID)
}

func CreatePubComp(packetID uint16) []byte {
	return util.WriteUint16(packetID)
}

func ParseAckPacket(data []byte) (uint16, error) {
	if len(data) < 2 {
		return 0, fmt.Errorf("ACK packet too short")
	}

	packetID, _, err := util.ReadUint16(data, 0)
	return packetID, err
}
