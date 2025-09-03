package connection

import (
	"fmt"
	"net"
	"time"

	"github/go/gobromq/internal/entities"
	"github/go/gobromq/internal/packet"
)

type Handler struct {
	conn   net.Conn
	parser *packet.Parser
	client *entities.Client
	broker BrokerInterface
}

type BrokerInterface interface {
	AddClient(client *entities.Client) error
	RemoveClient(clientID string, graceful bool)
	HandlePublish(req *entities.PublishRequest) (uint16, error)
	HandleSubscribe(req *entities.SubscribeRequest) []byte
	GetStats() map[string]interface{}
	AuthenticateClient(clientID string, username, password string) error
	CreateClientSession(clientID string, cleanSession bool, willMessage *entities.WillMessage)
	HandlePubAck(clientID string, packetID uint16) error
	HandlePubRec(clientID string, packetID uint16) error
	HandlePubComp(clientID string, packetID uint16) error
	GetPendingPubRel(clientID string) []uint16
}

func NewHandler(conn net.Conn, broker BrokerInterface) *Handler {
	return &Handler{
		conn:   conn,
		broker: broker,
		parser: packet.NewParser(conn),
	}
}

func (h *Handler) Start() error {
	defer h.cleanup()

	remoteAddr := h.conn.RemoteAddr().String()
	fmt.Printf("🔌 Starting handler for %s\n", remoteAddr)
	h.conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	firstPacket, err := h.parser.ReadPacket()
	if err != nil {
		fmt.Printf("❌ Error reading first packet from %s: %v\n", remoteAddr, err)
		return err
	}

	if firstPacket.Type != entities.CONNECT {
		fmt.Printf("❌ First packet must be CONNECT, got %s from %s\n",
			entities.GetPacketTypeName(firstPacket.Type), remoteAddr)
	}

	connectPacket, err := packet.ParseConnect(firstPacket.Data)
	if err != nil {
		fmt.Printf("❌ Error parsing CONNECT from %s: %v\n", remoteAddr, err)

		connectData := packet.CreateConnAck(false, entities.ConnectRefusedProtocolVersion)
		h.parser.WriteResponse(entities.CONNACK, connectData)
		return err
	}

	if len(connectPacket.ClientID) == 0 {
		connectPacket.ClientID = fmt.Sprintf("auto-%d", time.Now().UnixNano())
	}

	fmt.Printf("✅ CONNECT parsed successfully from %s:\n", remoteAddr)
	fmt.Printf("   - Client ID: %s\n", connectPacket.ClientID)
	fmt.Printf("   - Protocol: %s v%d\n", connectPacket.ProtocolName, connectPacket.ProtocolLevel)
	fmt.Printf("   - Keep Alive: %d seconds\n", connectPacket.KeepAlive)
	fmt.Printf("   - Clean Session: %t\n", connectPacket.CleanSession)
	fmt.Printf("   - Username: %s\n", connectPacket.Username)
	fmt.Printf("   - Has Password: %t\n", connectPacket.PasswordFlag)

	h.client = &entities.Client{
		ID:            string(connectPacket.ClientID),
		Conn:          h.conn,
		UserName:      connectPacket.Username,
		Authenticated: true,
		KeepAlive:     time.Duration(connectPacket.KeepAlive) * time.Second,
		LastSeen:      time.Now(),
	}

	var authError error
	var returnCode byte = entities.ConnectAccepted

	if connectPacket.Username != "" || connectPacket.Password != "" {
		authError = h.broker.AuthenticateClient(h.client.ID, connectPacket.Username, connectPacket.Password)
		if authError == nil {
			h.client.Authenticated = true
			fmt.Printf("🔐 Client %s authenticated successfully\n", h.client.ID)
		} else {
			fmt.Printf("🔒 Authentication failed for client %s: %v\n", h.client.ID, authError)
			returnCode = entities.ConnectRefusedBadCredentials
		}
	} else {
		fmt.Printf("🔓 Client %s attempting anonymous connection\n", h.client.ID)
	}

	var willMessage *entities.WillMessage

	if connectPacket.WillFlag {
		willMessage = &entities.WillMessage{
			Topic:   string(connectPacket.WillTopic),
			Payload: []byte(connectPacket.WillMessage),
			QoS:     entities.QoSLevel(connectPacket.WillQoS),
			Retain:  connectPacket.WillRetain,
		}
		fmt.Printf("📄 Will message configured for client %s: topic=%s, QoS=%d\n",
			h.client.ID, willMessage.Topic, willMessage.QoS)
	}

	h.broker.CreateClientSession(h.client.ID, connectPacket.CleanSession, willMessage)

	if err := h.broker.AddClient(h.client); err != nil {
		fmt.Printf("❌ Error adding client to broker: %v\n", err)

		if authError != nil {
			returnCode = entities.ConnectRefusedBadCredentials
		} else {
			returnCode = entities.ConnectRefusedServerUnavailable
		}
		connackData := packet.CreateConnAck(false, returnCode)
		h.parser.WriteResponse(entities.CONNACK, connackData)
		return err
	}

	connackData := packet.CreateConnAck(false, entities.ConnectAccepted)
	if err := h.parser.WriteResponse(entities.CONNACK, connackData); err != nil {
		fmt.Printf("❌ Error sending CONNACK to %s: %v\n", remoteAddr, err)
		return err
	}

	if returnCode == entities.ConnectAccepted {
		fmt.Printf("✅ CONNACK sent to %s (client: %s) - Connection accepted\n", remoteAddr, h.client.ID)
	} else {
		fmt.Printf("❌ CONNACK sent to %s (client: %s) - Connection rejected (code: %d)\n",
			remoteAddr, h.client.ID, returnCode)
		return fmt.Errorf("connection rejected")
	}

	go h.outgoingWorker()
	go h.pubrelWorker()

	return h.handlePackets()
}

func (h *Handler) handlePackets() error {
	remoteAddr := h.conn.RemoteAddr().String()
	for {
		if h.client.KeepAlive > 0 {
			timeout := h.client.KeepAlive + (h.client.KeepAlive / 2)
			h.conn.SetReadDeadline(time.Now().Add(timeout))
		}

		pkt, err := h.parser.ReadPacket()

		if err != nil {
			fmt.Printf("❌ Connection closed for %s: %v\n", remoteAddr, err)
			return err
		}

		fmt.Printf("📦 Received %s packet from %s (length: %d)\n",
			entities.GetPacketTypeName(pkt.Type), remoteAddr, pkt.Length)

		if err := h.handlePacket(pkt); err != nil {
			fmt.Printf("Failed to handle packet %s from %s: %v\n",
				entities.GetPacketTypeName(pkt.Type), remoteAddr, err)
		}

		h.client.LastSeen = time.Now()
	}
}

func (h *Handler) handlePacket(pkt *entities.Packet) error {
	switch pkt.Type {
	case entities.PUBLISH:
		return h.handlePublish(pkt)
	case entities.SUBSCRIBE:
		return h.handleSubscribe(pkt)
	case entities.PINGREQ:
		return h.handlePingReq()
	case entities.DISCONNECT:
		return h.handleDisconnect()
	case entities.UNSUBSCRIBE:
		return h.handleUnsubscribe()
	case entities.PUBACK:
		return h.handlePubAck(pkt)
	case entities.PUBREC:
		return h.handlePubRec(pkt)
	case entities.PUBREL:
		return h.handlePubRel(pkt)
	case entities.PUBCOMP:
		return h.handlePubComp(pkt)
	default:
		return fmt.Errorf("unknown packet type: %s", entities.GetPacketTypeName(pkt.Type))
	}
}

func (h *Handler) handlePingReq() error {
	fmt.Printf("PINGREQ from client %s\n", h.client.ID)
	return h.parser.WriteResponse(entities.PINGRESP, nil)
}

func (h *Handler) handleDisconnect() error {
	fmt.Printf("👋 DISCONNECT from client %s\n", h.client.ID)
	return fmt.Errorf("client disconnected")
}

func (h *Handler) handlePublish(pkt *entities.Packet) error {
	publishPacket, err := packet.ParsePublish(pkt.Data, pkt.Flags)
	if err != nil {
		fmt.Printf("❌ Error parsing PUBLISH from client %s: %v\n", h.client.ID, err)
		return err
	}

	fmt.Printf("📝 PUBLISH from %s: topic='%s', qos=%d, retain=%t, payload=%d bytes\n",
		h.client.ID, publishPacket.Topic, publishPacket.QoS, publishPacket.Retain, len(publishPacket.Payload))

	pubReq := &entities.PublishRequest{
		Topic:   string(publishPacket.Topic),
		Payload: publishPacket.Payload,
		QoS:     publishPacket.QoS,
		Retain:  publishPacket.Retain,
		From:    &h.client.ID,
	}

	_, err = h.broker.HandlePublish(pubReq)

	if err != nil {
		fmt.Printf("❌ Error handling publish from client %s: %v\n", h.client.ID, err)
		return err
	}

	switch publishPacket.QoS {
	case 1:
		pubackData := packet.CreatePubAck(publishPacket.PacketID)
		if err := h.parser.WriteResponse(entities.PUBACK, pubackData); err != nil {
			fmt.Printf("❌ Error sending PUBACK to client %s: %v\n", h.client.ID, err)
			return err
		}
		fmt.Printf("📤 PUBACK sent to client %s (packetID: %d)\n", h.client.ID, publishPacket.PacketID)
	case 2:
		pubrecData := packet.CreatePubRec(publishPacket.PacketID)
		if err := h.parser.WriteResponse(entities.PUBREC, pubrecData); err != nil {
			fmt.Printf("❌ Error sending PUBREC to client %s: %v\n", h.client.ID, err)
			return err
		}
		fmt.Printf("📤 PUBREC sent to client %s (packetID: %d)\n", h.client.ID, publishPacket.PacketID)
	}

	return nil
}

func (h *Handler) handleSubscribe(pkt *entities.Packet) error {
	subscribePacket, err := packet.ParseSubscribe(pkt.Data)

	if err != nil {
		fmt.Printf("❌ Error parsing SUBSCRIBE from client %s: %v\n", h.client.ID, err)
		return err
	}

	fmt.Printf("📋 SUBSCRIBE from %s: %d subscriptions\n",
		h.client.ID, len(subscribePacket.Subscriptions))

	var topicQoS []entities.TopicQoS

	for _, sub := range subscribePacket.Subscriptions {
		topicQoS = append(topicQoS, entities.TopicQoS{
			Topic: sub.Topic,
			QoS:   sub.QoS,
		})
		fmt.Printf("   - Topic: %s (QoS %d)\n", sub.Topic, sub.QoS)
	}

	subReq := &entities.SubscribeRequest{
		ClientID:      h.client.ID,
		Subscriptions: topicQoS,
		PacketID:      subscribePacket.PacketID,
	}

	returnCodes := h.broker.HandleSubscribe(subReq)

	subackData := packet.CreateSubAck(subscribePacket.PacketID, returnCodes)
	if err := h.parser.WriteResponse(entities.SUBACK, subackData); err != nil {
		fmt.Printf("❌ Error sending SUBACK to client %s: %v\n", h.client.ID, err)
		return err
	}

	fmt.Printf("📤 SUBACK sent to client %s\n", h.client.ID)
	return nil
}

func (h *Handler) handleUnsubscribe() error {
	fmt.Printf("📤 UNSUBACK sent to client %s (not implemented)\n", h.client.ID)
	return h.parser.WriteResponse(entities.UNSUBACK, []byte{0x00, 0x00})
}

func (h *Handler) handlePubAck(pkt *entities.Packet) error {
	packetID, err := packet.ParseAckPacket(pkt.Data)
	if err != nil {
		return fmt.Errorf("error parsing PUBACK: %w", err)
	}

	fmt.Printf("📨 PUBACK received from client %s (packetID: %d)\n", h.client.ID, packetID)

	return h.broker.HandlePubAck(h.client.ID, packetID)
}

func (h *Handler) handlePubRec(pkt *entities.Packet) error {
	packetID, err := packet.ParseAckPacket(pkt.Data)
	if err != nil {
		return fmt.Errorf("error parsing PUBREC: %w", err)
	}

	fmt.Printf("📨 PUBREC received from client %s (packetID: %d)\n", h.client.ID, packetID)

	if err := h.broker.HandlePubRec(h.client.ID, packetID); err != nil {
		return err
	}

	pubrelData := packet.CreatePubRel(packetID)
	if err := h.parser.WriteResponse(entities.PUBREL, pubrelData); err != nil {
		return fmt.Errorf("error sending PUBREL: %w", err)
	}

	fmt.Printf("📤 PUBREL sent to client %s (packetID: %d)\n", h.client.ID, packetID)
	return nil
}

func (h *Handler) handlePubRel(pkt *entities.Packet) error {
	packetID, err := packet.ParseAckPacket(pkt.Data)
	if err != nil {
		return fmt.Errorf("error parsing PUBREL: %w", err)
	}

	fmt.Printf("📨 PUBREL received from client %s (packetID: %d)\n", h.client.ID, packetID)

	pubcompData := packet.CreatePubComp(packetID)
	if err := h.parser.WriteResponse(entities.PUBCOMP, pubcompData); err != nil {
		return fmt.Errorf("error sending PUBCOMP: %w", err)
	}

	fmt.Printf("📤 PUBCOMP sent to client %s (packetID: %d)\n", h.client.ID, packetID)
	return nil
}

func (h *Handler) handlePubComp(pkt *entities.Packet) error {
	packetID, err := packet.ParseAckPacket(pkt.Data)
	if err != nil {
		return fmt.Errorf("error parsing PUBCOMP: %w", err)
	}

	fmt.Printf("📨 PUBCOMP received from client %s (packetID: %d)\n", h.client.ID, packetID)

	return h.broker.HandlePubComp(h.client.ID, packetID)
}

func (h *Handler) pubrelWorker() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		pendingPubRel := h.broker.GetPendingPubRel(h.client.ID)

		for _, packetID := range pendingPubRel {
			pubrelData := packet.CreatePubRel(packetID)
			if err := h.parser.WriteResponse(entities.PUBREL, pubrelData); err != nil {
				fmt.Printf("❌ Error sending PUBREL to client %s: %v\n", h.client.ID, err)
			} else {
				fmt.Printf("🔄 PUBREL resent to client %s (packetID: %d)\n", h.client.ID, packetID)
			}
		}
	}
}

func (h *Handler) outgoingWorker() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("⚠️ Outgoing worker panic for client %s: %v\n", h.client.ID, r)
		}
	}()

	for delivery := range h.client.OutgoingChan {
		if err := h.sendMessage(delivery); err != nil {
			fmt.Printf("❌ Error sending message to client %s: %v\n", h.client.ID, err)
			return
		}
	}
	fmt.Printf("📤 Outgoing worker stopped for client %s\n", h.client.ID)
}

func (h *Handler) sendMessage(delivery *entities.Delivery) error {
	msg := delivery.Message

	fmt.Printf("📤 Sending message to client %s: topic='%s', payload=%d bytes\n",
		h.client.ID, msg.Topic, len(msg.Payload))

	publishData := packet.CreatePublish(string(msg.Topic), msg.Payload, delivery.QoS, msg.Retain, delivery.PacketID)

	flags := byte(0)
	if msg.Retain {
		flags |= 0x01
	}

	flags |= (delivery.QoS << 1)

	publishPacket := &entities.Packet{
		Type:   entities.PUBLISH,
		Flags:  flags,
		Length: len(publishData),
		Data:   publishData,
	}
	return h.parser.WritePacket(publishPacket)
}

func (h *Handler) cleanup() {
	if h.client != nil {
		fmt.Printf("🧹 Cleaning up client %s\n", h.client.ID)
		h.broker.RemoveClient(h.client.ID, false)
	}

	if h.conn != nil {
		h.conn.Close()
	}
}

func (h *Handler) Close() error {
	return h.conn.Close()
}
