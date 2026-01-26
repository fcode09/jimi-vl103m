package packet

import (
	"fmt"
	"time"

	"github.com/intelcon-group/jimi-vl103m/pkg/jimi/protocol"
)

// TimeCalibrationPacket represents a time calibration request packet (Protocol 0x8A)
// The device sends this packet to request the current server time
//
// This packet typically has no content - it's just a request for the server
// to respond with the current time.
type TimeCalibrationPacket struct {
	BasePacket
}

// NewTimeCalibrationPacket creates a new TimeCalibrationPacket
func NewTimeCalibrationPacket() *TimeCalibrationPacket {
	return &TimeCalibrationPacket{
		BasePacket: BasePacket{
			ProtocolNum: protocol.ProtocolTimeCalibration,
			ParsedAt:    time.Now(),
		},
	}
}

// Type implements Packet interface
func (p *TimeCalibrationPacket) Type() string {
	return "Time Calibration"
}

// Timestamp implements Packet interface
// Time calibration packets don't have a timestamp
func (p *TimeCalibrationPacket) Timestamp() time.Time {
	return p.ParsedAt
}

// Validate implements Packet interface
func (p *TimeCalibrationPacket) Validate() error {
	return nil // Time calibration packets are always valid if parsed
}

// String returns a human-readable representation
func (p *TimeCalibrationPacket) String() string {
	return fmt.Sprintf("TimeCalibrationPacket{Serial: %d}", p.SerialNum)
}

// RequiresResponse returns true - time calibration always requires response
func (p *TimeCalibrationPacket) RequiresResponse() bool {
	return true
}
