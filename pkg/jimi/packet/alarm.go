package packet

import (
	"fmt"
	"time"

	"github.com/intelcon-group/jimi-vl103m/pkg/jimi/protocol"
	"github.com/intelcon-group/jimi-vl103m/pkg/jimi/types"
)

// AlarmPacket represents an alarm packet (Protocol 0x26)
// Alarm packets have the same structure as location packets but include alarm info
//
// Content structure:
// - DateTime: 6 bytes
// - GPS Info Length: 1 byte
// - Latitude: 4 bytes
// - Longitude: 4 bytes
// - Speed: 1 byte
// - Course/Status: 2 bytes
// - LBS Info: 8 bytes
// - Terminal Info: 1 byte
// - Voltage Level: 1 byte
// - GSM Signal: 1 byte
// - Alarm Type: 1 byte
// - Language: 1 byte
type AlarmPacket struct {
	BasePacket

	// DateTime is when the alarm was triggered
	DateTime types.DateTime

	// Satellites is the number of GPS satellites used
	Satellites uint8

	// Coordinates contains the GPS position
	Coordinates types.Coordinates

	// Speed in km/h
	Speed uint8

	// CourseStatus contains heading and GPS status flags
	CourseStatus types.CourseStatus

	// LBSInfo contains cell tower information
	LBSInfo types.LBSInfo

	// TerminalInfo contains device status
	TerminalInfo types.TerminalInfo

	// VoltageLevel indicates battery level
	VoltageLevel protocol.VoltageLevel

	// GSMSignal indicates network signal strength
	GSMSignal protocol.GSMSignalStrength

	// AlarmType indicates the type of alarm triggered
	AlarmType protocol.AlarmType

	// Language is the device language setting
	Language protocol.Language

	// Mileage is the mileage statistics from the device
	Mileage uint32
}

// NewAlarmPacket creates a new AlarmPacket
func NewAlarmPacket(dt types.DateTime, coords types.Coordinates, alarmType protocol.AlarmType) *AlarmPacket {
	return &AlarmPacket{
		BasePacket: BasePacket{
			ProtocolNum: protocol.ProtocolAlarm,
			ParsedAt:    time.Now(),
		},
		DateTime:    dt,
		Coordinates: coords,
		AlarmType:   alarmType,
	}
}

// Type implements Packet interface
func (p *AlarmPacket) Type() string {
	return "Alarm"
}

// Timestamp implements Packet interface
func (p *AlarmPacket) Timestamp() time.Time {
	return p.DateTime.Time
}

// Validate implements Packet interface
func (p *AlarmPacket) Validate() error {
	if p.DateTime.IsZero() {
		return &ValidationError{Field: "DateTime", Reason: "missing timestamp"}
	}
	return nil
}

// HasTimestamp implements PacketWithTimestamp interface
func (p *AlarmPacket) HasTimestamp() bool {
	return !p.DateTime.IsZero()
}

// HasLocation implements PacketWithLocation interface
func (p *AlarmPacket) HasLocation() bool {
	return p.Coordinates.IsValid()
}

// IsPositioned implements PacketWithLocation interface
func (p *AlarmPacket) IsPositioned() bool {
	return p.CourseStatus.GetIsPositioned()
}

// GetAlarmType implements PacketWithAlarm interface
func (p *AlarmPacket) GetAlarmType() protocol.AlarmType {
	return p.AlarmType
}

// IsCritical implements PacketWithAlarm interface
func (p *AlarmPacket) IsCritical() bool {
	return p.AlarmType.IsCritical()
}

// Latitude returns the signed latitude
func (p *AlarmPacket) Latitude() float64 {
	return p.Coordinates.SignedLatitude()
}

// Longitude returns the signed longitude
func (p *AlarmPacket) Longitude() float64 {
	return p.Coordinates.SignedLongitude()
}

// String returns a human-readable representation
func (p *AlarmPacket) String() string {
	return fmt.Sprintf("AlarmPacket{Type: %s, Time: %s, Pos: [%.6f, %.6f], Critical: %v}",
		p.AlarmType,
		p.DateTime,
		p.Latitude(),
		p.Longitude(),
		p.IsCritical())
}

// AlarmMultiFencePacket represents an alarm packet with geo-fence data (Protocol 0x27)
type AlarmMultiFencePacket struct {
	AlarmPacket

	// FenceID is the geo-fence identifier
	FenceID uint8
}

// Type implements Packet interface
func (p *AlarmMultiFencePacket) Type() string {
	return "Alarm Multi-Fence"
}

// String returns a human-readable representation
func (p *AlarmMultiFencePacket) String() string {
	return fmt.Sprintf("AlarmMultiFencePacket{Type: %s, FenceID: %d, Time: %s, Pos: [%.6f, %.6f]}",
		p.AlarmType,
		p.FenceID,
		p.DateTime,
		p.Latitude(),
		p.Longitude())
}

// Alarm4GPacket represents a 4G alarm packet (Protocol 0xA4)
type Alarm4GPacket struct {
	AlarmPacket

	// MCCMNC is the Mobile Country Code + Mobile Network Code
	MCCMNC uint32

	// ExtendedLBS contains additional LBS information for 4G
	ExtendedLBS []types.LBSInfo

	// FenceID is the geo-fence identifier (for multi-fence alarms)
	FenceID uint8
}

// Type implements Packet interface
func (p *Alarm4GPacket) Type() string {
	return "Alarm 4G"
}

// String returns a human-readable representation
func (p *Alarm4GPacket) String() string {
	return fmt.Sprintf("Alarm4GPacket{Type: %s, Time: %s, Pos: [%.6f, %.6f], MCCMNC: %d, Critical: %v}",
		p.AlarmType,
		p.DateTime,
		p.Latitude(),
		p.Longitude(),
		p.MCCMNC,
		p.IsCritical())
}
