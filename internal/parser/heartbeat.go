package parser

import (
	"fmt"
	"time"

	"github.com/fcode09/jimi-vl103m/pkg/jimi/packet"
	"github.com/fcode09/jimi-vl103m/pkg/jimi/protocol"
	"github.com/fcode09/jimi-vl103m/pkg/jimi/types"
)

// HeartbeatParser parses heartbeat packets (Protocol 0x13)
type HeartbeatParser struct {
	BaseParser
}

// NewHeartbeatParser creates a new heartbeat parser
func NewHeartbeatParser() *HeartbeatParser {
	return &HeartbeatParser{
		BaseParser: NewBaseParser(protocol.ProtocolHeartbeat, "Heartbeat"),
	}
}

// Parse implements Parser interface
// Heartbeat packet content structure:
// - Terminal Info: 1 byte
// - Voltage Level: 1 byte
// - GSM Signal: 1 byte
// - Extended Info: 2 bytes (optional, in newer devices)
// Total content: 3-5 bytes
func (p *HeartbeatParser) Parse(data []byte, ctx Context) (packet.Packet, error) {
	content, err := ExtractContent(data)
	if err != nil {
		return nil, fmt.Errorf("heartbeat: %w", err)
	}

	if len(content) < 3 {
		return nil, fmt.Errorf("heartbeat: content too short: %d bytes (need at least 3)", len(content))
	}

	// Parse Terminal Info (1 byte)
	terminalInfo := types.NewTerminalInfo(content[0])

	// Parse Voltage Level (1 byte)
	voltageLevel := protocol.VoltageLevel(content[1])

	// Parse GSM Signal (1 byte)
	gsmSignal := protocol.GSMSignalStrength(content[2])

	// Parse Extended Info (2 bytes, optional)
	var extendedInfo uint16
	hasExtended := false
	if len(content) >= 5 {
		extendedInfo = uint16(content[3])<<8 | uint16(content[4])
		hasExtended = true
	}

	// Extract serial number
	serialNum, _ := ExtractSerialNumber(data)

	pkt := &packet.HeartbeatPacket{
		BasePacket: packet.BasePacket{
			ProtocolNum: protocol.ProtocolHeartbeat,
			SerialNum:   serialNum,
			RawData:     data,
			ParsedAt:    time.Now(),
		},
		TerminalInfo: terminalInfo,
		VoltageLevel: voltageLevel,
		GSMSignal:    gsmSignal,
		ExtendedInfo: extendedInfo,
		HasExtended:  hasExtended,
	}

	return pkt, nil
}

// init registers the heartbeat parser with the default registry
func init() {
	MustRegister(NewHeartbeatParser())
}
