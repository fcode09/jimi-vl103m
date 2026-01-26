package encoder

import (
	"bytes"
	"fmt"
	"unicode/utf16"

	"github.com/intelcon-group/jimi-vl103m/internal/validator"
	"github.com/intelcon-group/jimi-vl103m/pkg/jimi/protocol"
)

// AddressResponseParams contains parameters for creating address response packets
type AddressResponseParams struct {
	// ServerFlag is a 4-byte marker (usually zeros)
	ServerFlag [4]byte

	// AlarmSMS is the alarm identifier (typically "ALARMSMS")
	AlarmSMS string

	// Address is the parsed address string
	Address string

	// PhoneNumber is the destination phone number (21 bytes)
	// For alarm packets uploaded to server, use "000000000000000000000"
	PhoneNumber string

	// SerialNumber is the packet serial number (should match the alarm packet)
	SerialNumber uint16

	// Language determines which protocol to use
	// - LanguageChinese (0x01) → Protocol 0x17 (short packet, UNICODE)
	// - LanguageEnglish (0x02) → Protocol 0x97 (long packet, ASCII/UTF-8)
	Language protocol.Language
}

// ChineseAddressResponse creates a Chinese address response packet (0x17)
//
// Packet structure:
// - Start Bit: 0x7878 (short packet)
// - Packet Length: 1 byte
// - Protocol Number: 0x17
// - Content Length: 1 byte
// - Server Flag: 4 bytes
// - ALARMSMS: 8 bytes (ASCII)
// - "&&": 2 bytes
// - Address: N bytes (UTF-16 BE)
// - "&&": 2 bytes
// - Phone Number: 21 bytes (ASCII)
// - "##": 2 bytes
// - Serial Number: 2 bytes
// - CRC: 2 bytes
// - Stop Bit: 0x0D0A
func ChineseAddressResponse(params AddressResponseParams) ([]byte, error) {
	// Validate parameters
	if err := validateAddressParams(params); err != nil {
		return nil, err
	}

	// Encode address to UTF-16 BE
	addressBytes, err := encodeUTF16BE(params.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to encode address to UTF-16: %w", err)
	}

	// Ensure AlarmSMS is exactly 8 bytes
	alarmSMS := padOrTruncate(params.AlarmSMS, 8)

	// Ensure PhoneNumber is exactly 21 bytes
	phoneNumber := padOrTruncate(params.PhoneNumber, 21)

	// Build content
	var content bytes.Buffer

	// Content Length = ServerFlag(4) + AlarmSMS(8) + "&&"(2) + Address(N) + "&&"(2) + Phone(21) + "##"(2)
	contentLength := 4 + 8 + 2 + len(addressBytes) + 2 + 21 + 2
	content.WriteByte(byte(contentLength))

	// Server Flag (4 bytes)
	content.Write(params.ServerFlag[:])

	// ALARMSMS (8 bytes)
	content.WriteString(alarmSMS)

	// First separator
	content.WriteString("&&")

	// Address (UTF-16 BE)
	content.Write(addressBytes)

	// Second separator
	content.WriteString("&&")

	// Phone Number (21 bytes)
	content.WriteString(phoneNumber)

	// Third separator
	content.WriteString("##")

	// Build packet
	// Packet Length = Protocol(1) + Content + Serial(2) + CRC(2)
	packetLength := 1 + content.Len() + 2 + 2

	if packetLength > 255 {
		return nil, fmt.Errorf("packet too large for short format: %d bytes (max 255)", packetLength)
	}

	var packet bytes.Buffer

	// Start Bit (short packet)
	packet.WriteByte(0x78)
	packet.WriteByte(0x78)

	// Packet Length (1 byte)
	packet.WriteByte(byte(packetLength))

	// Protocol Number
	packet.WriteByte(protocol.ProtocolAddressResponseChinese)

	// Content
	packet.Write(content.Bytes())

	// Serial Number (2 bytes, big-endian)
	packet.WriteByte(byte(params.SerialNumber >> 8))
	packet.WriteByte(byte(params.SerialNumber & 0xFF))

	// Calculate CRC for: Length + Protocol + Content + Serial
	crcData := packet.Bytes()[2:] // Skip start bit
	crc := validator.CalculateCRC(crcData)

	// CRC (2 bytes, big-endian)
	packet.WriteByte(byte(crc >> 8))
	packet.WriteByte(byte(crc & 0xFF))

	// Stop Bit
	packet.WriteByte(0x0D)
	packet.WriteByte(0x0A)

	return packet.Bytes(), nil
}

// EnglishAddressResponse creates an English address response packet (0x97)
//
// Packet structure:
// - Start Bit: 0x7979 (long packet)
// - Packet Length: 2 bytes
// - Protocol Number: 0x97
// - Content Length: 1 byte
// - Server Flag: 4 bytes
// - ALARMSMS: 8 bytes (ASCII)
// - "&&": 2 bytes
// - Address: N bytes (ASCII/UTF-8)
// - "&&": 2 bytes
// - Phone Number: 21 bytes (ASCII)
// - "##": 2 bytes
// - Serial Number: 2 bytes
// - CRC: 2 bytes
// - Stop Bit: 0x0D0A
func EnglishAddressResponse(params AddressResponseParams) ([]byte, error) {
	// Validate parameters
	if err := validateAddressParams(params); err != nil {
		return nil, err
	}

	// Ensure AlarmSMS is exactly 8 bytes
	alarmSMS := padOrTruncate(params.AlarmSMS, 8)

	// Ensure PhoneNumber is exactly 21 bytes
	phoneNumber := padOrTruncate(params.PhoneNumber, 21)

	// Build content
	var content bytes.Buffer

	// Content Length = ServerFlag(4) + AlarmSMS(8) + "&&"(2) + Address(N) + "&&"(2) + Phone(21) + "##"(2)
	contentLength := 4 + 8 + 2 + len(params.Address) + 2 + 21 + 2
	content.WriteByte(byte(contentLength))

	// Server Flag (4 bytes)
	content.Write(params.ServerFlag[:])

	// ALARMSMS (8 bytes)
	content.WriteString(alarmSMS)

	// First separator
	content.WriteString("&&")

	// Address (ASCII/UTF-8)
	content.WriteString(params.Address)

	// Second separator
	content.WriteString("&&")

	// Phone Number (21 bytes)
	content.WriteString(phoneNumber)

	// Third separator
	content.WriteString("##")

	// Build packet
	// Packet Length = Protocol(1) + Content + Serial(2) + CRC(2)
	packetLength := 1 + content.Len() + 2 + 2

	if packetLength > 65535 {
		return nil, fmt.Errorf("packet too large: %d bytes (max 65535)", packetLength)
	}

	var packet bytes.Buffer

	// Start Bit (long packet)
	packet.WriteByte(0x79)
	packet.WriteByte(0x79)

	// Packet Length (2 bytes, big-endian)
	packet.WriteByte(byte(packetLength >> 8))
	packet.WriteByte(byte(packetLength & 0xFF))

	// Protocol Number
	packet.WriteByte(protocol.ProtocolAddressResponseEnglish)

	// Content
	packet.Write(content.Bytes())

	// Serial Number (2 bytes, big-endian)
	packet.WriteByte(byte(params.SerialNumber >> 8))
	packet.WriteByte(byte(params.SerialNumber & 0xFF))

	// Calculate CRC for: Length + Protocol + Content + Serial
	crcData := packet.Bytes()[2:] // Skip start bit
	crc := validator.CalculateCRC(crcData)

	// CRC (2 bytes, big-endian)
	packet.WriteByte(byte(crc >> 8))
	packet.WriteByte(byte(crc & 0xFF))

	// Stop Bit
	packet.WriteByte(0x0D)
	packet.WriteByte(0x0A)

	return packet.Bytes(), nil
}

// AddressResponse is a convenience function that automatically chooses
// between Chinese (0x17) and English (0x97) based on the Language parameter
func AddressResponse(params AddressResponseParams) ([]byte, error) {
	if params.Language == protocol.LanguageChinese {
		return ChineseAddressResponse(params)
	}
	return EnglishAddressResponse(params)
}

// validateAddressParams validates the address response parameters
func validateAddressParams(params AddressResponseParams) error {
	if params.AlarmSMS == "" {
		return fmt.Errorf("alarm SMS is required")
	}
	if params.Address == "" {
		return fmt.Errorf("address is required")
	}
	if params.PhoneNumber == "" {
		return fmt.Errorf("phone number is required")
	}
	if len(params.PhoneNumber) > 21 {
		return fmt.Errorf("phone number too long: %d bytes (max 21)", len(params.PhoneNumber))
	}
	return nil
}

// encodeUTF16BE encodes a UTF-8 string to UTF-16 Big Endian bytes
func encodeUTF16BE(s string) ([]byte, error) {
	// Convert string to runes
	runes := []rune(s)

	// Encode to UTF-16
	u16s := utf16.Encode(runes)

	// Convert to bytes (big-endian)
	bytes := make([]byte, len(u16s)*2)
	for i, u16 := range u16s {
		bytes[i*2] = byte(u16 >> 8)
		bytes[i*2+1] = byte(u16 & 0xFF)
	}

	return bytes, nil
}

// padOrTruncate pads or truncates a string to exactly n bytes
func padOrTruncate(s string, n int) string {
	if len(s) >= n {
		return s[:n]
	}
	// Pad with spaces
	padding := make([]byte, n-len(s))
	for i := range padding {
		padding[i] = ' '
	}
	return s + string(padding)
}
