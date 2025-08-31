package util

import (
	"bufio"
	"encoding/binary"
	"fmt"
)

func DecodeRemainingLength(reader *bufio.Reader) (int, error) {
	mutiplier := 1
	value := 0

	for {
		encodedByte, err := reader.ReadByte()

		if err != nil {
			return 0, fmt.Errorf("failed to read remaining length: %w", err)
		}

		value += int(encodedByte&0x7F) * mutiplier

		if (encodedByte & 0x80) == 0 {
			break
		}

		mutiplier *= 128

		if mutiplier > 128*128*128 {
			return 0, fmt.Errorf("malformed remaining length")
		}
	}

	return value, nil
}

func EncodeRemainingLength(length int) []byte {
	if length == 0 {
		return []byte{0}
	}

	var encoded []byte

	for length > 0 {
		encodedByte := byte(length % 128)
		length /= 128

		if length > 0 {
			encodedByte |= 0x80
		}

		encoded = append(encoded, encodedByte)
	}
	return encoded
}

func ReadMQTTString(data []byte, offset int) (string, int, error) {
	if len(data) < offset+2 {
		return "", offset, fmt.Errorf("insufficient data for string length")
	}

	stringLen := binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2

	if len(data) < offset+int(stringLen) {
		return "", offset, fmt.Errorf("insufficient data for string content")
	}

	str := string(data[offset : offset+int(stringLen)])
	offset += int(stringLen)

	return str, offset, nil
}

func WriteMQTTString(str string) []byte {
	strBytes := []byte(str)
	length := uint16(len(strBytes))

	result := make([]byte, 2+len(strBytes))
	binary.BigEndian.PutUint16(result[0:2], length)
	copy(result[2:], strBytes)

	return result
}

func ReadUint16(data []byte, offset int) (uint16, int, error) {
	if len(data) < offset+2 {
		return 0, offset, fmt.Errorf("insufficient data for uint16")
	}

	value := binary.BigEndian.Uint16(data[offset : offset+2])
	return value, offset + 2, nil
}

func WriteUint16(value uint16) []byte {
	result := make([]byte, 2)
	binary.BigEndian.PutUint16(result, value)
	return result
}
