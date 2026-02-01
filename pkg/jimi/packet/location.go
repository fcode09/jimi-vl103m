package packet

import (
	"fmt"
	"time"

	"github.com/fcode09/jimi-vl103m/pkg/jimi/protocol"
	"github.com/fcode09/jimi-vl103m/pkg/jimi/types"
)

// LocationPacket represents a GPS location packet (Protocol 0x22)
//
// Content structure:
// - DateTime: 6 bytes (YY MM DD HH MM SS)
// - GPS Info Length: 1 byte (high nibble: satellites, low nibble: length/2)
// - Latitude: 4 bytes
// - Longitude: 4 bytes
// - Speed: 1 byte (km/h)
// - Course/Status: 2 bytes
// - LBS Info: 8 bytes (2G) or more (4G)
type LocationPacket struct {
	BasePacket

	// DateTime is when the location was recorded by the device
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

	// TerminalInfo contains device status (if present)
	// Note: For GPS Location packets, this contains other status bits but NOT ACC status.
	// ACC status is stored in the dedicated 'ACC' field below.
	TerminalInfo types.TerminalInfo

	// ACC indicates whether the vehicle ignition is ON (true) or OFF (false).
	// For GPS Location packets (0x22 and 0xA0), ACC is a dedicated byte where:
	// - 0x00 = ACC off
	// - 0x01 = ACC on
	// This is different from heartbeat/alarm packets where ACC is bit 1 of a status byte.
	ACC bool

	// VoltageLevel indicates battery level (if present)
	VoltageLevel protocol.VoltageLevel

	// GSMSignal indicates network signal strength (if present)
	GSMSignal protocol.GSMSignalStrength

	// UploadMode indicates why this location was uploaded
	UploadMode protocol.UploadMode

	// IsReupload indicates if this is a re-uploaded packet
	IsReupload bool

	// Mileage is the mileage statistics from the device
	Mileage uint32

	// HasStatus indicates if terminal status fields are present
	HasStatus bool
}

// NewLocationPacket creates a new LocationPacket
func NewLocationPacket(dt types.DateTime, coords types.Coordinates, speed uint8, course types.CourseStatus) *LocationPacket {
	return &LocationPacket{
		BasePacket: BasePacket{
			ProtocolNum: protocol.ProtocolGPSLocation,
			ParsedAt:    time.Now(),
		},
		DateTime:     dt,
		Coordinates:  coords,
		Speed:        speed,
		CourseStatus: course,
	}
}

// Type implements Packet interface
func (p *LocationPacket) Type() string {
	if p.ProtocolNum == protocol.ProtocolGPSLocation4G {
		return "GPS Location 4G"
	}
	return "GPS Location"
}

// Timestamp implements Packet interface
func (p *LocationPacket) Timestamp() time.Time {
	return p.DateTime.Time
}

// Validate implements Packet interface
func (p *LocationPacket) Validate() error {
	if !p.Coordinates.IsValid() {
		return &ValidationError{Field: "Coordinates", Reason: "invalid coordinates"}
	}
	if p.DateTime.IsZero() {
		return &ValidationError{Field: "DateTime", Reason: "missing timestamp"}
	}
	return nil
}

// HasTimestamp implements PacketWithTimestamp interface
func (p *LocationPacket) HasTimestamp() bool {
	return !p.DateTime.IsZero()
}

// HasLocation implements PacketWithLocation interface
func (p *LocationPacket) HasLocation() bool {
	return p.Coordinates.IsValid()
}

// IsPositioned implements PacketWithLocation interface
func (p *LocationPacket) IsPositioned() bool {
	return p.CourseStatus.GetIsPositioned()
}

// Latitude returns the signed latitude
func (p *LocationPacket) Latitude() float64 {
	return p.Coordinates.SignedLatitude()
}

// Longitude returns the signed longitude
func (p *LocationPacket) Longitude() float64 {
	return p.Coordinates.SignedLongitude()
}

// Heading returns the course/heading in degrees (0-360)
func (p *LocationPacket) Heading() uint16 {
	return p.CourseStatus.GetCourse()
}

// HeadingName returns the heading as a compass direction (N, NE, E, etc.)
func (p *LocationPacket) HeadingName() string {
	return p.CourseStatus.DirectionName()
}

// ACCOn returns true if ACC (ignition) is on
// For GPS Location packets (0x22 and 0xA0), this reads the dedicated ACC field.
// Note: This is different from heartbeat/alarm packets where ACC is stored in TerminalInfo.
func (p *LocationPacket) ACCOn() bool {
	return p.ACC
}

// IsCharging returns true if the device is charging
func (p *LocationPacket) IsCharging() bool {
	if p.HasStatus {
		return p.TerminalInfo.IsCharging()
	}
	return false
}

// IsGPSPositioned returns true if GPS is positioned
// Falls back to CourseStatus if TerminalInfo is not available
func (p *LocationPacket) IsGPSPositioned() bool {
	if p.HasStatus {
		return p.TerminalInfo.GPSTrackingEnabled()
	}
	return p.CourseStatus.GetIsPositioned()
}

// IsPowerCut returns true if fuel/power is cut
func (p *LocationPacket) IsPowerCut() bool {
	if p.HasStatus {
		return p.TerminalInfo.OilElectricityDisconnected()
	}
	return false
}

// IsArmed returns true if device is armed
func (p *LocationPacket) IsArmed() bool {
	if p.HasStatus {
		return p.TerminalInfo.IsArmed()
	}
	return false
}

// String returns a human-readable representation
func (p *LocationPacket) String() string {
	return fmt.Sprintf("LocationPacket{Time: %s, Pos: [%.6f, %.6f], Speed: %d km/h, Heading: %d° (%s), Satellites: %d}",
		p.DateTime,
		p.Latitude(),
		p.Longitude(),
		p.Speed,
		p.Heading(),
		p.HeadingName(),
		p.Satellites)
}

// Location4GPacket represents a 4G GPS location packet (Protocol 0xA0)
// It extends LocationPacket with additional 4G-specific fields
type Location4GPacket struct {
	LocationPacket

	// MCCMNC is the Mobile Country Code + Mobile Network Code
	MCCMNC uint32

	// ExtendedLBS contains additional LBS information for 4G
	ExtendedLBS []types.LBSInfo
}

// Type implements Packet interface
func (p *Location4GPacket) Type() string {
	return "GPS Location 4G"
}

// String returns a human-readable representation
func (p *Location4GPacket) String() string {
	return fmt.Sprintf("Location4GPacket{Time: %s, Pos: [%.6f, %.6f], Speed: %d km/h, Heading: %d° (%s), Satellites: %d, MCCMNC: %d}",
		p.DateTime,
		p.Latitude(),
		p.Longitude(),
		p.Speed,
		p.Heading(),
		p.HeadingName(),
		p.Satellites,
		p.MCCMNC)
}
