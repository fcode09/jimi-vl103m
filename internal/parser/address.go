package parser

import (
	"bytes"
	"fmt"
	"strings"
	"time"
	"unicode/utf16"

	"github.com/fcode09/jimi-vl103m/pkg/jimi/packet"
	"github.com/fcode09/jimi-vl103m/pkg/jimi/protocol"
)

// ChineseAddressParser parses Chinese address response packets (Protocol 0x17)
type ChineseAddressParser struct {
	BaseParser
}

// NewChineseAddressParser creates a new Chinese address parser
func NewChineseAddressParser() *ChineseAddressParser {
	return &ChineseAddressParser{
		BaseParser: NewBaseParser(protocol.ProtocolAddressResponseChinese, "Chinese Address Response"),
	}
}

// Parse implements Parser interface
// Chinese Address packet content structure (Protocol 0x17):
// - Content Length: 1 byte (length of data between server flag and serial number)
// - Server Flag: 4 bytes (server marker)
// - ALARMSMS: 8 bytes (ASCII, typically "ALARMSMS")
// - Separator "&&": 2 bytes (ASCII)
// - Address Content: M bytes (UNICODE - UTF-16 BE)
// - Separator "&&": 2 bytes (ASCII)
// - Phone Number: 21 bytes (ASCII, "0" repeated for alarm packets)
// - Separator "##": 2 bytes (ASCII)
// Minimum content: 1 + 4 + 8 + 2 + 0 + 2 + 21 + 2 = 40 bytes
func (p *ChineseAddressParser) Parse(data []byte, ctx Context) (packet.Packet, error) {
	content, err := ExtractContent(data)
	if err != nil {
		return nil, fmt.Errorf("chinese_address: %w", err)
	}

	if len(content) < 40 {
		return nil, fmt.Errorf("chinese_address: content too short: %d bytes (need at least 40)", len(content))
	}

	offset := 0

	// Parse Content Length (1 byte)
	contentLength := content[offset]
	offset++

	// Parse Server Flag (4 bytes)
	var serverFlag [4]byte
	copy(serverFlag[:], content[offset:offset+4])
	offset += 4

	// Parse ALARMSMS (8 bytes ASCII)
	alarmSMS := strings.TrimSpace(string(content[offset : offset+8]))
	offset += 8

	// Parse first separator "&&" (2 bytes)
	separator1 := string(content[offset : offset+2])
	if separator1 != "&&" {
		return nil, fmt.Errorf("chinese_address: expected '&&' separator, got '%s'", separator1)
	}
	offset += 2

	// Find the next "&&" separator to determine address length
	// Address is in UNICODE (UTF-16 BE)
	remainingData := content[offset:]
	separatorIdx := bytes.Index(remainingData, []byte("&&"))
	if separatorIdx == -1 {
		return nil, fmt.Errorf("chinese_address: second '&&' separator not found")
	}

	// Parse Address Content (UNICODE - UTF-16 Big Endian)
	addressBytes := remainingData[:separatorIdx]
	address, err := decodeUTF16BE(addressBytes)
	if err != nil {
		return nil, fmt.Errorf("chinese_address: failed to decode address: %w", err)
	}
	offset += separatorIdx

	// Parse second separator "&&" (2 bytes)
	separator2 := string(content[offset : offset+2])
	if separator2 != "&&" {
		return nil, fmt.Errorf("chinese_address: expected second '&&' separator, got '%s'", separator2)
	}
	offset += 2

	// Parse Phone Number (21 bytes ASCII)
	if offset+21 > len(content) {
		return nil, fmt.Errorf("chinese_address: not enough data for phone number")
	}
	phoneNumber := strings.TrimSpace(string(content[offset : offset+21]))
	offset += 21

	// Parse third separator "##" (2 bytes)
	if offset+2 > len(content) {
		return nil, fmt.Errorf("chinese_address: not enough data for final separator")
	}
	separator3 := string(content[offset : offset+2])
	if separator3 != "##" {
		return nil, fmt.Errorf("chinese_address: expected '##' separator, got '%s'", separator3)
	}
	offset += 2

	// Extract serial number
	serialNum, _ := ExtractSerialNumber(data)

	pkt := &packet.AddressResponsePacket{
		BasePacket: packet.BasePacket{
			ProtocolNum: protocol.ProtocolAddressResponseChinese,
			SerialNum:   serialNum,
			RawData:     data,
			ParsedAt:    time.Now(),
		},
		ContentLength: contentLength,
		ServerFlag:    serverFlag,
		AlarmSMS:      alarmSMS,
		Address:       address,
		PhoneNumber:   phoneNumber,
		Language:      protocol.LanguageChinese,
	}

	return pkt, nil
}

// EnglishAddressParser parses English address response packets (Protocol 0x97)
type EnglishAddressParser struct {
	BaseParser
}

// NewEnglishAddressParser creates a new English address parser
func NewEnglishAddressParser() *EnglishAddressParser {
	return &EnglishAddressParser{
		BaseParser: NewBaseParser(protocol.ProtocolAddressResponseEnglish, "English Address Response"),
	}
}

// Parse implements Parser interface
// English Address packet content structure (Protocol 0x97):
// Same as Chinese but:
// - Uses long packet format (0x7979 start bit, 2-byte length)
// - Address content is ASCII/UTF-8 instead of UNICODE
// - Content Length: 1 byte
// - Server Flag: 4 bytes
// - ALARMSMS: 8 bytes (ASCII)
// - Separator "&&": 2 bytes (ASCII)
// - Address Content: M bytes (ASCII/UTF-8)
// - Separator "&&": 2 bytes (ASCII)
// - Phone Number: 21 bytes (ASCII)
// - Separator "##": 2 bytes (ASCII)
// Minimum content: 40 bytes
func (p *EnglishAddressParser) Parse(data []byte, ctx Context) (packet.Packet, error) {
	content, err := ExtractContent(data)
	if err != nil {
		return nil, fmt.Errorf("english_address: %w", err)
	}

	if len(content) < 40 {
		return nil, fmt.Errorf("english_address: content too short: %d bytes (need at least 40)", len(content))
	}

	offset := 0

	// Parse Content Length (1 byte)
	contentLength := content[offset]
	offset++

	// Parse Server Flag (4 bytes)
	var serverFlag [4]byte
	copy(serverFlag[:], content[offset:offset+4])
	offset += 4

	// Parse ALARMSMS (8 bytes ASCII)
	alarmSMS := strings.TrimSpace(string(content[offset : offset+8]))
	offset += 8

	// Parse first separator "&&" (2 bytes)
	separator1 := string(content[offset : offset+2])
	if separator1 != "&&" {
		return nil, fmt.Errorf("english_address: expected '&&' separator, got '%s'", separator1)
	}
	offset += 2

	// Find the next "&&" separator to determine address length
	remainingData := content[offset:]
	separatorIdx := bytes.Index(remainingData, []byte("&&"))
	if separatorIdx == -1 {
		return nil, fmt.Errorf("english_address: second '&&' separator not found")
	}

	// Parse Address Content (ASCII/UTF-8)
	address := string(remainingData[:separatorIdx])
	offset += separatorIdx

	// Parse second separator "&&" (2 bytes)
	separator2 := string(content[offset : offset+2])
	if separator2 != "&&" {
		return nil, fmt.Errorf("english_address: expected second '&&' separator, got '%s'", separator2)
	}
	offset += 2

	// Parse Phone Number (21 bytes ASCII)
	if offset+21 > len(content) {
		return nil, fmt.Errorf("english_address: not enough data for phone number")
	}
	phoneNumber := strings.TrimSpace(string(content[offset : offset+21]))
	offset += 21

	// Parse third separator "##" (2 bytes)
	if offset+2 > len(content) {
		return nil, fmt.Errorf("english_address: not enough data for final separator")
	}
	separator3 := string(content[offset : offset+2])
	if separator3 != "##" {
		return nil, fmt.Errorf("english_address: expected '##' separator, got '%s'", separator3)
	}
	offset += 2

	// Extract serial number
	serialNum, _ := ExtractSerialNumber(data)

	pkt := &packet.AddressResponsePacket{
		BasePacket: packet.BasePacket{
			ProtocolNum: protocol.ProtocolAddressResponseEnglish,
			SerialNum:   serialNum,
			RawData:     data,
			ParsedAt:    time.Now(),
		},
		ContentLength: contentLength,
		ServerFlag:    serverFlag,
		AlarmSMS:      alarmSMS,
		Address:       address,
		PhoneNumber:   phoneNumber,
		Language:      protocol.LanguageEnglish,
	}

	return pkt, nil
}

// decodeUTF16BE decodes UTF-16 Big Endian bytes to UTF-8 string
func decodeUTF16BE(b []byte) (string, error) {
	if len(b)%2 != 0 {
		return "", fmt.Errorf("invalid UTF-16 data: odd number of bytes")
	}

	// Convert bytes to uint16 slice
	u16s := make([]uint16, len(b)/2)
	for i := 0; i < len(b); i += 2 {
		u16s[i/2] = uint16(b[i])<<8 | uint16(b[i+1])
	}

	// Decode UTF-16 to runes
	runes := utf16.Decode(u16s)

	return string(runes), nil
}

// init registers address parsers with the default registry
func init() {
	MustRegister(NewChineseAddressParser())
	MustRegister(NewEnglishAddressParser())
}
