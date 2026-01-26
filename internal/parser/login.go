package parser

import (
	"fmt"
	"time"

	"github.com/fcode09/jimi-vl103m/pkg/jimi/packet"
	"github.com/fcode09/jimi-vl103m/pkg/jimi/protocol"
	"github.com/fcode09/jimi-vl103m/pkg/jimi/types"
)

// LoginParser parses login packets (Protocol 0x01)
type LoginParser struct {
	BaseParser
	validateIMEI bool
}

// NewLoginParser creates a new login parser
func NewLoginParser() *LoginParser {
	return &LoginParser{
		BaseParser:   NewBaseParser(protocol.ProtocolLogin, "Login"),
		validateIMEI: true,
	}
}

// NewLoginParserWithOptions creates a login parser with options
func NewLoginParserWithOptions(validateIMEI bool) *LoginParser {
	return &LoginParser{
		BaseParser:   NewBaseParser(protocol.ProtocolLogin, "Login"),
		validateIMEI: validateIMEI,
	}
}

// Parse implements Parser interface
// Login packet content structure:
// - IMEI: 8 bytes (BCD encoded, 15 digits + padding)
// - Model Identification Code: 2 bytes
// - Timezone/Language: 2 bytes
// Total content: 12 bytes
func (p *LoginParser) Parse(data []byte) (packet.Packet, error) {
	content, err := ExtractContent(data)
	if err != nil {
		return nil, fmt.Errorf("login: %w", err)
	}

	if len(content) < 12 {
		return nil, fmt.Errorf("login: content too short: %d bytes (need 12)", len(content))
	}

	// Parse IMEI (8 bytes BCD)
	var imei types.IMEI
	if p.validateIMEI {
		imei, err = types.NewIMEIFromBytes(content[0:8])
	} else {
		imei, err = types.NewIMEIFromBytesUnchecked(content[0:8])
	}
	if err != nil {
		return nil, fmt.Errorf("login: failed to parse IMEI: %w", err)
	}

	// Parse Model ID (2 bytes big-endian)
	modelID := uint16(content[8])<<8 | uint16(content[9])

	// Parse Timezone/Language (2 bytes)
	timezone, err := types.TimezoneFromBytes(content[10:12])
	if err != nil {
		return nil, fmt.Errorf("login: failed to parse timezone: %w", err)
	}

	// Extract serial number
	serialNum, _ := ExtractSerialNumber(data)

	pkt := &packet.LoginPacket{
		BasePacket: packet.BasePacket{
			ProtocolNum: protocol.ProtocolLogin,
			SerialNum:   serialNum,
			RawData:     data,
			ParsedAt:    time.Now(),
		},
		IMEI:     imei,
		ModelID:  modelID,
		Timezone: timezone,
	}

	return pkt, nil
}

// init registers the login parser with the default registry
func init() {
	MustRegister(NewLoginParser())
}
