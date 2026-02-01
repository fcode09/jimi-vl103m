package parser

import (
	"fmt"
	"time"

	"github.com/fcode09/jimi-vl103m/pkg/jimi/packet"
	"github.com/fcode09/jimi-vl103m/pkg/jimi/protocol"
	"github.com/fcode09/jimi-vl103m/pkg/jimi/types"
)

// LBSParser parses LBS packets (Protocol 0x28)
type LBSParser struct {
	BaseParser
}

// NewLBSParser creates a new LBS parser
func NewLBSParser() *LBSParser {
	return &LBSParser{
		BaseParser: NewBaseParser(protocol.ProtocolLBSMultiBase, "LBS Multi-Base"),
	}
}

// Parse implements Parser interface
// LBS packet content structure:
// - DateTime: 6 bytes
// - MCC: 2 bytes
// - MNC: 1 byte
// - LAC: 2 bytes
// - Cell ID: 3 bytes
// Optional:
// - Terminal Info: 1 byte
// - Voltage Level: 1 byte
// - GSM Signal: 1 byte
// - Upload Mode: 1 byte
// Minimum content: 14 bytes, with status: 18 bytes
func (p *LBSParser) Parse(data []byte, ctx Context) (packet.Packet, error) {
	content, err := ExtractContent(data)
	if err != nil {
		return nil, fmt.Errorf("lbs: %w", err)
	}

	if len(content) < 14 {
		return nil, fmt.Errorf("lbs: content too short: %d bytes (need at least 14)", len(content))
	}

	offset := 0

	// Parse DateTime (6 bytes)
	dt, err := types.DateTimeFromBytes(content[offset : offset+6])
	if err != nil {
		return nil, fmt.Errorf("lbs: failed to parse datetime: %w", err)
	}
	offset += 6

	// Main Cell Tower
	mcc := uint16(content[offset])<<8 | uint16(content[offset+1])
	offset += 2
	mnc := uint16(content[offset])
	offset++
	lac := uint32(content[offset])<<8 | uint32(content[offset+1])
	offset += 2
	ci := uint32(content[offset])<<16 | uint32(content[offset+1])<<8 | uint32(content[offset+2])
	offset += 3
	_ = content[offset] // RSSI is not stored in LBSInfo
	offset++

	mainCell := types.NewLBSInfo(mcc, mnc, lac, uint64(ci))

	// Neighbor Cells
	var neighborCells []types.LBSInfo
	for i := 0; i < 6 && offset+6 <= len(content); i++ {
		nlac := uint32(content[offset])<<8 | uint32(content[offset+1])
		offset += 2
		nci := uint32(content[offset])<<16 | uint32(content[offset+1])<<8 | uint32(content[offset+2])
		offset += 3
		//nrssi := content[offset]
		offset++
		// The neighbor cells in the doc share MCC and MNC with the main cell.
		neighborCells = append(neighborCells, types.NewLBSInfo(mcc, mnc, nlac, uint64(nci)))
	}

	// Timing Advance
	var timingAdvance uint8
	if offset < len(content) {
		timingAdvance = content[offset]
		offset++
	}

	// Language
	var language protocol.Language
	if offset+2 <= len(content) {
		langCode := uint16(content[offset])<<8 | uint16(content[offset+1])
		language = protocol.Language(langCode) // This might be wrong, doc says 2 bytes but language is 1 byte
		offset += 2
	}

	// Extract serial number
	serialNum, _ := ExtractSerialNumber(data)

	pkt := &packet.LBSPacket{
		BasePacket: packet.BasePacket{
			ProtocolNum: protocol.ProtocolLBSMultiBase,
			SerialNum:   serialNum,
			RawData:     data,
			ParsedAt:    time.Now(),
		},
		DateTime:      dt,
		LBSInfo:       mainCell,
		NeighborCells: neighborCells,
		TimingAdvance: timingAdvance,
		Language:      language,
	}

	return pkt, nil
}

// LBS4GParser parses 4G LBS packets (Protocol 0xA1)
type LBS4GParser struct {
	BaseParser
}

// NewLBS4GParser creates a new 4G LBS parser
func NewLBS4GParser() *LBS4GParser {
	return &LBS4GParser{
		BaseParser: NewBaseParser(protocol.ProtocolLBSMultiBase4G, "LBS Multi-Base 4G"),
	}
}

// Parse implements Parser interface
// 4G LBS packet has extended LBS info with multiple cell towers
func (p *LBS4GParser) Parse(data []byte, ctx Context) (packet.Packet, error) {
	content, err := ExtractContent(data)
	if err != nil {
		return nil, fmt.Errorf("lbs_4g: %w", err)
	}

	if len(content) < 6 {
		return nil, fmt.Errorf("lbs_4g: content too short: %d bytes", len(content))
	}

	offset := 0

	// Parse DateTime (6 bytes)
	dt, err := types.DateTimeFromBytes(content[offset : offset+6])
	if err != nil {
		return nil, fmt.Errorf("lbs_4g: failed to parse datetime: %w", err)
	}
	offset += 6

	// Extract serial number
	serialNum, _ := ExtractSerialNumber(data)

	pkt := &packet.LBS4GPacket{
		BasePacket: packet.BasePacket{
			ProtocolNum: protocol.ProtocolLBSMultiBase4G,
			SerialNum:   serialNum,
			RawData:     data,
			ParsedAt:    time.Now(),
		},
		DateTime: dt,
	}

	// Parse 4G LBS info (variable length)
	// Structure: MCC(2) + MNC(1-2) + LAC(4) + CellID(8) = 15-16 bytes per cell
	remainingBytes := len(content) - offset

	// Try to parse at least one cell tower
	if remainingBytes >= 15 {
		lbsInfo, _, err := types.NewLBSInfoFromBytes(content[offset:], true)
		if err == nil {
			pkt.LBSInfo = lbsInfo
		}
	}

	// Parse status info from the end if present
	// Status is typically the last 4 bytes
	if remainingBytes >= 4 {
		statusOffset := len(content) - 4
		pkt.TerminalInfo = types.NewTerminalInfo(content[statusOffset])
		pkt.VoltageLevel = protocol.VoltageLevel(content[statusOffset+1])
		pkt.GSMSignal = protocol.GSMSignalStrength(content[statusOffset+2])
		pkt.UploadMode = protocol.UploadMode(content[statusOffset+3])
	}

	return pkt, nil
}

// init registers LBS parsers with the default registry
func init() {
	MustRegister(NewLBSParser())
	MustRegister(NewLBS4GParser())
}
