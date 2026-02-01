package packet

import (
	"fmt"
	"time"

	"github.com/intelcon-group/jimi-vl103m/pkg/jimi/protocol"
	"github.com/intelcon-group/jimi-vl103m/pkg/jimi/types"
)

// OnlineCommandPacket represents an online command packet (Protocol 0x80)
// This packet is sent by the server to the device to execute commands
//
// Content structure:
// - Server Flag: 4 bytes (identifies the command)
// - Command Content: variable length (ASCII string)
type OnlineCommandPacket struct {
	BasePacket

	// ServerFlag is a 4-byte identifier for the command
	ServerFlag uint32

	// Command is the ASCII command string
	Command string

	// CommandLength is the length of the command
	CommandLength uint8
}

// NewOnlineCommandPacket creates a new OnlineCommandPacket
func NewOnlineCommandPacket(serverFlag uint32, command string) *OnlineCommandPacket {
	return &OnlineCommandPacket{
		BasePacket: BasePacket{
			ProtocolNum: protocol.ProtocolOnlineCommand,
			ParsedAt:    time.Now(),
		},
		ServerFlag:    serverFlag,
		Command:       command,
		CommandLength: uint8(len(command)),
	}
}

// Type implements Packet interface
func (p *OnlineCommandPacket) Type() string {
	return "Online Command"
}

// Timestamp implements Packet interface
func (p *OnlineCommandPacket) Timestamp() time.Time {
	return p.ParsedAt
}

// Validate implements Packet interface
func (p *OnlineCommandPacket) Validate() error {
	return nil
}

// String returns a human-readable representation
func (p *OnlineCommandPacket) String() string {
	return fmt.Sprintf("OnlineCommandPacket{ServerFlag: 0x%08X, Command: %q}", p.ServerFlag, p.Command)
}

// CommandResponsePacket represents a command response packet (Protocol 0x21 or 0x15)
// This packet is sent by the device in response to an online command
//
// Content structure:
// - Server Flag: 4 bytes (echo of command flag)
// - Response Content: variable length (ASCII string)
type CommandResponsePacket struct {
	BasePacket

	// ServerFlag echoes the flag from the original command
	ServerFlag uint32

	// Response is the ASCII response string
	Response string

	// ResponseLength is the length of the response
	ResponseLength uint8
}

// NewCommandResponsePacket creates a new CommandResponsePacket
func NewCommandResponsePacket(serverFlag uint32, response string) *CommandResponsePacket {
	return &CommandResponsePacket{
		BasePacket: BasePacket{
			ProtocolNum: protocol.ProtocolCommandResponse,
			ParsedAt:    time.Now(),
		},
		ServerFlag:     serverFlag,
		Response:       response,
		ResponseLength: uint8(len(response)),
	}
}

// Type implements Packet interface
func (p *CommandResponsePacket) Type() string {
	if p.ProtocolNum == protocol.ProtocolCommandResponseOld {
		return "Command Response (Old)"
	}
	return "Command Response"
}

// Timestamp implements Packet interface
func (p *CommandResponsePacket) Timestamp() time.Time {
	return p.ParsedAt
}

// Validate implements Packet interface
func (p *CommandResponsePacket) Validate() error {
	return nil
}

// String returns a human-readable representation
func (p *CommandResponsePacket) String() string {
	return fmt.Sprintf("CommandResponsePacket{ServerFlag: 0x%08X, Response: %q}", p.ServerFlag, p.Response)
}

// GPSAddressRequestPacket represents a GPS address request packet (Protocol 0x2A)
// Structure according to JM-VL03 documentation:
// - DateTime: 6 bytes (YY MM DD HH MM SS)
// - GPS Info: 1 byte (satellites in low nibble)
// - Latitude: 4 bytes (raw / 1,800,000)
// - Longitude: 4 bytes (raw / 1,800,000)
// - Speed: 1 byte (km/h)
// - Course/Status: 2 bytes (heading + status flags)
// - Phone Number: 21 bytes (ASCII)
// - Alert/Language: 2 bytes (AlarmType + Language)
// Total content: 41 bytes
type GPSAddressRequestPacket struct {
	BasePacket

	// GPS Data
	DateTime     types.DateTime
	Satellites   uint8
	Coordinates  types.Coordinates
	Speed        uint8
	CourseStatus types.CourseStatus

	// Request-specific data
	PhoneNumber string // 21 bytes ASCII, trimmed
	AlarmType   protocol.AlarmType
	Language    protocol.Language
}

// NewGPSAddressRequestPacket creates a new GPSAddressRequestPacket
func NewGPSAddressRequestPacket(coords types.Coordinates, phone string, lang protocol.Language) *GPSAddressRequestPacket {
	return &GPSAddressRequestPacket{
		BasePacket: BasePacket{
			ProtocolNum: protocol.ProtocolGPSAddressRequest,
			ParsedAt:    time.Now(),
		},
		Coordinates: coords,
		PhoneNumber: phone,
		Language:    lang,
	}
}

// Type implements Packet interface
func (p *GPSAddressRequestPacket) Type() string {
	return "GPS Address Request"
}

// Timestamp implements PacketWithTimestamp interface
func (p *GPSAddressRequestPacket) Timestamp() time.Time {
	return p.DateTime.Time
}

// HasTimestamp implements PacketWithTimestamp interface
func (p *GPSAddressRequestPacket) HasTimestamp() bool {
	return !p.DateTime.IsZero()
}

// HasLocation implements PacketWithLocation interface
func (p *GPSAddressRequestPacket) HasLocation() bool {
	return p.Coordinates.IsValid()
}

// IsPositioned implements PacketWithLocation interface
func (p *GPSAddressRequestPacket) IsPositioned() bool {
	return p.CourseStatus.GetIsPositioned()
}

// Latitude returns the signed latitude
func (p *GPSAddressRequestPacket) Latitude() float64 {
	return p.Coordinates.SignedLatitude()
}

// Longitude returns the signed longitude
func (p *GPSAddressRequestPacket) Longitude() float64 {
	return p.Coordinates.SignedLongitude()
}

// Heading returns the course/heading in degrees
func (p *GPSAddressRequestPacket) Heading() uint16 {
	return p.CourseStatus.GetCourse()
}

// Validate implements Packet interface
func (p *GPSAddressRequestPacket) Validate() error {
	return nil
}

// String returns a human-readable representation
func (p *GPSAddressRequestPacket) String() string {
	return fmt.Sprintf("GPSAddressRequestPacket{Time: %s, Pos: [%.6f, %.6f], Speed: %d km/h, Heading: %d°, Phone: %s, Alarm: %s, Lang: %s}",
		p.DateTime,
		p.Latitude(),
		p.Longitude(),
		p.Speed,
		p.Heading(),
		p.PhoneNumber,
		p.AlarmType,
		p.Language)
}

// AddressResponsePacket is defined in address.go
// It represents address response packets (Protocol 0x17 or 0x97)
