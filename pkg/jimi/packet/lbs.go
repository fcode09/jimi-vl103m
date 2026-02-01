package packet

import (
	"fmt"
	"time"

	"github.com/fcode09/jimi-vl103m/pkg/jimi/protocol"
	"github.com/fcode09/jimi-vl103m/pkg/jimi/types"
)

// LBSPacket represents an LBS (Location Based Service) packet (Protocol 0x28)
// LBS packets are sent when GPS signal is not available, using cell tower triangulation
//
// Content structure:
// - DateTime: 6 bytes
// - MCC: 2 bytes
// - MNC: 1 byte
// - LAC: 2 bytes
// - Cell ID: 3 bytes
// - Terminal Info: 1 byte (optional)
// - Voltage Level: 1 byte (optional)
// - GSM Signal: 1 byte (optional)
// - Upload Mode: 1 byte (optional)
type LBSPacket struct {
	BasePacket

	// DateTime is when the LBS data was recorded
	DateTime types.DateTime

	// LBSInfo contains primary cell tower information
	LBSInfo types.LBSInfo

	// NeighborCells contains information about neighboring cell towers
	NeighborCells []types.LBSInfo

	// TimingAdvance is the timing advance value
	TimingAdvance uint8

	// Language is the language setting for the device
	Language protocol.Language

	// TerminalInfo contains device status (if present)
	TerminalInfo types.TerminalInfo

	// VoltageLevel indicates battery level (if present)
	VoltageLevel protocol.VoltageLevel

	// GSMSignal indicates network signal strength (if present)
	GSMSignal protocol.GSMSignalStrength

	// UploadMode indicates why this data was uploaded
	UploadMode protocol.UploadMode

	// HasStatus indicates if terminal status fields are present
	HasStatus bool
}

// NewLBSPacket creates a new LBSPacket
func NewLBSPacket(dt types.DateTime, lbs types.LBSInfo) *LBSPacket {
	return &LBSPacket{
		BasePacket: BasePacket{
			ProtocolNum: protocol.ProtocolLBSMultiBase,
			ParsedAt:    time.Now(),
		},
		DateTime: dt,
		LBSInfo:  lbs,
	}
}

// Type implements Packet interface
func (p *LBSPacket) Type() string {
	return "LBS Multi-Base"
}

// Timestamp implements Packet interface
func (p *LBSPacket) Timestamp() time.Time {
	return p.DateTime.Time
}

// Validate implements Packet interface
func (p *LBSPacket) Validate() error {
	if p.DateTime.IsZero() {
		return &ValidationError{Field: "DateTime", Reason: "missing timestamp"}
	}
	if !p.LBSInfo.IsValid() {
		return &ValidationError{Field: "LBSInfo", Reason: "invalid LBS data"}
	}
	return nil
}

// HasTimestamp implements PacketWithTimestamp interface
func (p *LBSPacket) HasTimestamp() bool {
	return !p.DateTime.IsZero()
}

// String returns a human-readable representation
func (p *LBSPacket) String() string {
	return fmt.Sprintf("LBSPacket{Time: %s, %s}", p.DateTime, p.LBSInfo)
}

// LBS4GPacket represents a 4G LBS packet (Protocol 0xA1)
type LBS4GPacket struct {
	BasePacket

	// DateTime is when the LBS data was recorded
	DateTime types.DateTime

	// LBSInfo contains primary cell tower information
	LBSInfo types.LBSInfo

	// NeighborCells contains information about neighboring cell towers
	NeighborCells []types.LBSInfo

	// TerminalInfo contains device status
	TerminalInfo types.TerminalInfo

	// VoltageLevel indicates battery level
	VoltageLevel protocol.VoltageLevel

	// GSMSignal indicates network signal strength
	GSMSignal protocol.GSMSignalStrength

	// UploadMode indicates why this data was uploaded
	UploadMode protocol.UploadMode
}

// NewLBS4GPacket creates a new LBS4GPacket
func NewLBS4GPacket(dt types.DateTime, lbs types.LBSInfo) *LBS4GPacket {
	return &LBS4GPacket{
		BasePacket: BasePacket{
			ProtocolNum: protocol.ProtocolLBSMultiBase4G,
			ParsedAt:    time.Now(),
		},
		DateTime: dt,
		LBSInfo:  lbs,
	}
}

// Type implements Packet interface
func (p *LBS4GPacket) Type() string {
	return "LBS Multi-Base 4G"
}

// Timestamp implements Packet interface
func (p *LBS4GPacket) Timestamp() time.Time {
	return p.DateTime.Time
}

// Validate implements Packet interface
func (p *LBS4GPacket) Validate() error {
	if p.DateTime.IsZero() {
		return &ValidationError{Field: "DateTime", Reason: "missing timestamp"}
	}
	return nil
}

// HasTimestamp implements PacketWithTimestamp interface
func (p *LBS4GPacket) HasTimestamp() bool {
	return !p.DateTime.IsZero()
}

// String returns a human-readable representation
func (p *LBS4GPacket) String() string {
	return fmt.Sprintf("LBS4GPacket{Time: %s, %s, Neighbors: %d}",
		p.DateTime, p.LBSInfo, len(p.NeighborCells))
}
