package packet

import (
	"bufio"
	"fmt"
	"io"
	"net"

	"github/go/gobromq/internal/entities"
	"github/go/gobromq/util"
)

type Parser struct {
	reader *bufio.Reader
	conn   net.Conn
}

func NewParser(conn net.Conn) *Parser {
	return &Parser{
		reader: bufio.NewReader(conn),
		conn:   conn,
	}
}

func (p *Parser) ReadPacket() (*entities.Packet, error) {

	firstByte, err := p.reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("error reading fixed header: %w", err)
	}

	packetType := firstByte & 0xF0
	flags := firstByte & 0x0F

	remainingLength, err := util.DecodeRemainingLength(p.reader)
	if err != nil {
		return nil, fmt.Errorf("error reading remaining length: %w", err)
	}

	var data []byte
	if remainingLength > 0 {
		data = make([]byte, remainingLength)
		_, err = io.ReadFull(p.reader, data)
		if err != nil {
			return nil, fmt.Errorf("error reading packet data: %w", err)
		}
	}

	packet := &entities.Packet{
		Type:   packetType,
		Flags:  flags,
		Length: remainingLength,
		Data:   data,
	}

	return packet, nil
}

func (p *Parser) WritePacket(packet *entities.Packet) error {
	firstByte := packet.Type | packet.Flags
	if err := p.writeByte(firstByte); err != nil {
		return fmt.Errorf("failed to write first byte: %w", err)
	}

	lengthBytes := util.EncodeRemainingLength(packet.Length)
	if err := p.writeBytes(lengthBytes); err != nil {
		return fmt.Errorf("failed to write remaining length: %w", err)
	}

	if len(packet.Data) > 0 {
		if err := p.writeBytes(packet.Data); err != nil {
			return fmt.Errorf("failed to write packet data: %w", err)
		}
	}
	return nil
}

func (p *Parser) WriteResponse(packetType byte, data []byte) error {

	flags := byte(0)
	if packetType == 0x60 {
		flags = 0x02
	}

	packet := &entities.Packet{
		Type:   packetType,
		Flags:  flags,
		Length: len(data),
		Data:   data,
	}
	return p.WritePacket(packet)
}

func (p *Parser) writeByte(b byte) error {
	_, err := p.conn.Write([]byte{b})
	return err
}

func (p *Parser) writeBytes(data []byte) error {
	_, err := p.conn.Write(data)
	return err
}

func (p *Parser) ValidatePacket(packet *entities.Packet) error {
	switch packet.Type {
	case entities.CONNECT:
		return p.validateConnecte(packet)
	case entities.PUBLISH:
		return p.validatePublish(packet)
	case entities.SUBSCRIBE:
		return p.validateSubscribe(packet)
	case entities.PINGREQ, entities.DISCONNECT:
		if packet.Length != 0 {
			return fmt.Errorf("%s packet must have zero length", entities.GetPacketTypeName(packet.Type))
		}
		return nil
	default:
		return nil
	}
}

func (p *Parser) validateConnecte(packet *entities.Packet) error {
	if packet.Length < 10 {
		return fmt.Errorf("CONNECT packet too short")
	}
	return nil
}

func (p *Parser) validatePublish(packet *entities.Packet) error {
	if packet.Length < 2 {
		return fmt.Errorf("PUBLISH packet too short")
	}
	return nil
}

func (p *Parser) validateSubscribe(packet *entities.Packet) error {
	if packet.Length < 3 {
		return fmt.Errorf("SUBSCRIBE packet too short")
	}
	return nil
}
