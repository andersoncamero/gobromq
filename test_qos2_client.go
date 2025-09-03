package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	// Conectar al broker
	conn, err := net.Dial("tcp", "127.0.0.1:1884")
	if err != nil {
		fmt.Printf("Error connecting: %v\n", err)
		return
	}
	defer conn.Close()

	fmt.Println("Connected to broker")

	// Enviar CONNECT packet
	connectPacket := []byte{
		0x10, 0x26, // CONNECT packet type and remaining length
		0x00, 0x04, 'M', 'Q', 'T', 'T', // Protocol name
		0x04,       // Protocol level
		0xC2,       // Connect flags (username, password, clean session)
		0x00, 0x3C, // Keep alive (60 seconds)
		0x00, 0x0B, 't', 'e', 's', 't', '_', 'q', 'o', 's', '2', '_', 'p', 'u', 'b', // Client ID
		0x00, 0x05, 'a', 'd', 'm', 'i', 'n', // Username
		0x00, 0x05, 'a', 'd', 'm', 'i', 'n', // Password
	}

	if _, err := conn.Write(connectPacket); err != nil {
		fmt.Printf("Error sending CONNECT: %v\n", err)
		return
	}

	// Leer CONNACK
	response := make([]byte, 4)
	if _, err := conn.Read(response); err != nil {
		fmt.Printf("Error reading CONNACK: %v\n", err)
		return
	}

	if response[3] != 0x00 {
		fmt.Printf("Connection rejected: %v\n", response[3])
		return
	}

	fmt.Println("Connected successfully")

	// Enviar PUBLISH QoS 2
	publishPacket := []byte{
		0x34, 0x19, // PUBLISH packet type with QoS 2 and remaining length
		0x00, 0x0A, 't', 'e', 's', 't', '/', 'm', 'u', 'l', 't', 'i', // Topic
		0x00, 0x01, // Packet ID
		'H', 'e', 'l', 'l', 'o', ' ', 'Q', 'o', 'S', ' ', '2', // Payload
	}

	if _, err := conn.Write(publishPacket); err != nil {
		fmt.Printf("Error sending PUBLISH: %v\n", err)
		return
	}

	fmt.Println("PUBLISH QoS 2 sent, waiting for PUBREC...")

	// Leer PUBREC
	pubrec := make([]byte, 4)
	if _, err := conn.Read(pubrec); err != nil {
		fmt.Printf("Error reading PUBREC: %v\n", err)
		return
	}

	if pubrec[0] != 0x50 {
		fmt.Printf("Expected PUBREC, got: %x\n", pubrec[0])
		return
	}

	fmt.Println("PUBREC received, sending PUBREL...")

	// Enviar PUBREL
	pubrelPacket := []byte{
		0x62, 0x02, // PUBREL packet type and remaining length
		0x00, 0x01, // Packet ID
	}

	if _, err := conn.Write(pubrelPacket); err != nil {
		fmt.Printf("Error sending PUBREL: %v\n", err)
		return
	}

	// Leer PUBCOMP
	pubcomp := make([]byte, 4)
	if _, err := conn.Read(pubcomp); err != nil {
		fmt.Printf("Error reading PUBCOMP: %v\n", err)
		return
	}

	if pubcomp[0] != 0x70 {
		fmt.Printf("Expected PUBCOMP, got: %x\n", pubcomp[0])
		return
	}

	fmt.Println("PUBCOMP received! QoS 2 flow completed successfully!")

	// Mantener conexión por un momento para ver los logs del broker
	time.Sleep(2 * time.Second)

	// Enviar DISCONNECT
	disconnectPacket := []byte{0xE0, 0x00}
	conn.Write(disconnectPacket)

	fmt.Println("Test completed successfully!")
}
