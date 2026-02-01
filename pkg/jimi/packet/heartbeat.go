package packet

import (
	"fmt"
	"time"

	"github.com/fcode09/jimi-vl103m/pkg/jimi/protocol"
	"github.com/fcode09/jimi-vl103m/pkg/jimi/types"
)

// HeartbeatPacket represents a heartbeat packet (Protocol 0x13)
// Heartbeat packets are sent periodically to maintain the connection
//
// Content structure:
// - Terminal Info: 1 byte
// - Voltage Level: 1 byte
// - GSM Signal: 1 byte
// - Extended Info: 2 bytes (optional)
type HeartbeatPacket struct {
	BasePacket

	// TerminalInfo contains device status flags
	TerminalInfo types.TerminalInfo

	// VoltageLevel indicates battery level
	VoltageLevel protocol.VoltageLevel

	// GSMSignal indicates network signal strength
	GSMSignal protocol.GSMSignalStrength

	// ExtendedInfo contains additional status (if present)
	ExtendedInfo uint16

	// HasExtended indicates if extended info is present
	HasExtended bool
}

// NewHeartbeatPacket creates a new HeartbeatPacket
func NewHeartbeatPacket(termInfo types.TerminalInfo, voltage protocol.VoltageLevel, signal protocol.GSMSignalStrength) *HeartbeatPacket {
	return &HeartbeatPacket{
		BasePacket: BasePacket{
			ProtocolNum: protocol.ProtocolHeartbeat,
			ParsedAt:    time.Now(),
		},
		TerminalInfo: termInfo,
		VoltageLevel: voltage,
		GSMSignal:    signal,
	}
}

// Type implements Packet interface
func (p *HeartbeatPacket) Type() string {
	return "Heartbeat"
}

// Timestamp implements Packet interface
// Heartbeat packets don't have a timestamp field
func (p *HeartbeatPacket) Timestamp() time.Time {
	return p.ParsedAt
}

// Validate implements Packet interface
func (p *HeartbeatPacket) Validate() error {
	return nil // Heartbeat packets are always valid if parsed
}

// ACCOn returns true if ACC (ignition) is on
func (p *HeartbeatPacket) ACCOn() bool {
	return p.TerminalInfo.ACCOn()
}

// IsCharging returns true if the device is charging
func (p *HeartbeatPacket) IsCharging() bool {
	return p.TerminalInfo.IsCharging()
}

// BatteryPercentage returns approximate battery percentage
func (p *HeartbeatPacket) BatteryPercentage() int {
	return p.VoltageLevel.Percentage()
}

// SignalBars returns signal strength as 0-4 bars
func (p *HeartbeatPacket) SignalBars() int {
	return p.GSMSignal.Bars()
}

// String returns a human-readable representation
func (p *HeartbeatPacket) String() string {
	return fmt.Sprintf("HeartbeatPacket{Terminal: %s, Voltage: %s, GSM: %s}",
		p.TerminalInfo, p.VoltageLevel, p.GSMSignal)
}
