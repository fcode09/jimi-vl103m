package packet

import (
	"fmt"
	"time"

	"github.com/fcode09/jimi-vl103m/pkg/jimi/protocol"
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
// The device sends GPS coordinates and requests the server to return the address
type GPSAddressRequestPacket struct {
	BasePacket

	// Coordinates are the GPS coordinates for address lookup
	Coordinates string

	// Language is the preferred response language
	Language protocol.Language
}

// NewGPSAddressRequestPacket creates a new GPSAddressRequestPacket
func NewGPSAddressRequestPacket(coords string, lang protocol.Language) *GPSAddressRequestPacket {
	return &GPSAddressRequestPacket{
		BasePacket: BasePacket{
			ProtocolNum: protocol.ProtocolGPSAddressRequest,
			ParsedAt:    time.Now(),
		},
		Coordinates: coords,
		Language:    lang,
	}
}

// Type implements Packet interface
func (p *GPSAddressRequestPacket) Type() string {
	return "GPS Address Request"
}

// Timestamp implements Packet interface
func (p *GPSAddressRequestPacket) Timestamp() time.Time {
	return p.ParsedAt
}

// Validate implements Packet interface
func (p *GPSAddressRequestPacket) Validate() error {
	return nil
}

// String returns a human-readable representation
func (p *GPSAddressRequestPacket) String() string {
	return fmt.Sprintf("GPSAddressRequestPacket{Coords: %s, Language: %s}", p.Coordinates, p.Language)
}

// AddressResponsePacket is defined in address.go
// It represents address response packets (Protocol 0x17 or 0x97)
