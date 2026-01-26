package parser

import (
	"fmt"
	"time"

	"github.com/intelcon-group/jimi-vl103m/pkg/jimi/packet"
	"github.com/intelcon-group/jimi-vl103m/pkg/jimi/protocol"
)

// OnlineCommandParser parses online command packets (Protocol 0x80)
type OnlineCommandParser struct {
	BaseParser
}

// NewOnlineCommandParser creates a new online command parser
func NewOnlineCommandParser() *OnlineCommandParser {
	return &OnlineCommandParser{
		BaseParser: NewBaseParser(protocol.ProtocolOnlineCommand, "Online Command"),
	}
}

// Parse implements Parser interface
// Online command packet content structure:
// - Command Length: 1 byte
// - Server Flag: 4 bytes
// - Command Content: variable length (ASCII)
func (p *OnlineCommandParser) Parse(data []byte) (packet.Packet, error) {
	content, err := ExtractContent(data)
	if err != nil {
		return nil, fmt.Errorf("online_command: %w", err)
	}

	if len(content) < 5 {
		return nil, fmt.Errorf("online_command: content too short: %d bytes (need at least 5)", len(content))
	}

	// Parse command length
	cmdLength := content[0]

	// Parse server flag (4 bytes)
	serverFlag := uint32(content[1])<<24 | uint32(content[2])<<16 |
		uint32(content[3])<<8 | uint32(content[4])

	// Parse command content
	var command string
	if int(cmdLength) > 4 && len(content) > 5 {
		commandBytes := content[5:]
		// Command length includes the server flag (4 bytes)
		actualCmdLen := int(cmdLength) - 4
		if actualCmdLen > 0 && actualCmdLen <= len(commandBytes) {
			command = string(commandBytes[:actualCmdLen])
		} else if len(commandBytes) > 0 {
			command = string(commandBytes)
		}
	}

	// Extract serial number
	serialNum, _ := ExtractSerialNumber(data)

	pkt := &packet.OnlineCommandPacket{
		BasePacket: packet.BasePacket{
			ProtocolNum: protocol.ProtocolOnlineCommand,
			SerialNum:   serialNum,
			RawData:     data,
			ParsedAt:    time.Now(),
		},
		ServerFlag:    serverFlag,
		Command:       command,
		CommandLength: cmdLength,
	}

	return pkt, nil
}

// CommandResponseParser parses command response packets (Protocol 0x21 and 0x15)
type CommandResponseParser struct {
	BaseParser
}

// NewCommandResponseParser creates a new command response parser
func NewCommandResponseParser() *CommandResponseParser {
	return &CommandResponseParser{
		BaseParser: NewBaseParser(protocol.ProtocolCommandResponse, "Command Response"),
	}
}

// Parse implements Parser interface
// Command response packet content structure:
// - Response Length: 1 byte
// - Server Flag: 4 bytes (echo of original command)
// - Response Content: variable length (ASCII)
func (p *CommandResponseParser) Parse(data []byte) (packet.Packet, error) {
	content, err := ExtractContent(data)
	if err != nil {
		return nil, fmt.Errorf("command_response: %w", err)
	}

	if len(content) < 5 {
		return nil, fmt.Errorf("command_response: content too short: %d bytes (need at least 5)", len(content))
	}

	// Parse response length
	respLength := content[0]

	// Parse server flag (4 bytes)
	serverFlag := uint32(content[1])<<24 | uint32(content[2])<<16 |
		uint32(content[3])<<8 | uint32(content[4])

	// Parse response content
	var response string
	if int(respLength) > 4 && len(content) > 5 {
		responseBytes := content[5:]
		actualRespLen := int(respLength) - 4
		if actualRespLen > 0 && actualRespLen <= len(responseBytes) {
			response = string(responseBytes[:actualRespLen])
		} else if len(responseBytes) > 0 {
			response = string(responseBytes)
		}
	}

	// Extract serial number
	serialNum, _ := ExtractSerialNumber(data)

	pkt := &packet.CommandResponsePacket{
		BasePacket: packet.BasePacket{
			ProtocolNum: protocol.ProtocolCommandResponse,
			SerialNum:   serialNum,
			RawData:     data,
			ParsedAt:    time.Now(),
		},
		ServerFlag:     serverFlag,
		Response:       response,
		ResponseLength: respLength,
	}

	return pkt, nil
}

// CommandResponseOldParser parses old-format command response packets (Protocol 0x15)
type CommandResponseOldParser struct {
	BaseParser
}

// NewCommandResponseOldParser creates a new old-format command response parser
func NewCommandResponseOldParser() *CommandResponseOldParser {
	return &CommandResponseOldParser{
		BaseParser: NewBaseParser(protocol.ProtocolCommandResponseOld, "Command Response (Old)"),
	}
}

// Parse implements Parser interface
func (p *CommandResponseOldParser) Parse(data []byte) (packet.Packet, error) {
	// Use the same parsing logic as CommandResponseParser
	parser := &CommandResponseParser{}
	pkt, err := parser.Parse(data)
	if err != nil {
		return nil, err
	}

	// Update protocol number
	if cmdResp, ok := pkt.(*packet.CommandResponsePacket); ok {
		cmdResp.ProtocolNum = protocol.ProtocolCommandResponseOld
	}

	return pkt, nil
}

// GPSAddressRequestParser parses GPS address request packets (Protocol 0x2A)
type GPSAddressRequestParser struct {
	BaseParser
}

// NewGPSAddressRequestParser creates a new GPS address request parser
func NewGPSAddressRequestParser() *GPSAddressRequestParser {
	return &GPSAddressRequestParser{
		BaseParser: NewBaseParser(protocol.ProtocolGPSAddressRequest, "GPS Address Request"),
	}
}

// Parse implements Parser interface
// GPS address request content contains coordinates for reverse geocoding
func (p *GPSAddressRequestParser) Parse(data []byte) (packet.Packet, error) {
	content, err := ExtractContent(data)
	if err != nil {
		return nil, fmt.Errorf("gps_address_request: %w", err)
	}

	// Extract serial number
	serialNum, _ := ExtractSerialNumber(data)

	// The content typically contains latitude and longitude in some format
	// For now, store as raw string
	coords := ""
	var language protocol.Language = protocol.LanguageEnglish

	if len(content) > 0 {
		// Try to extract as ASCII coordinates string
		coords = string(content)

		// Last byte might be language indicator
		if len(content) > 1 {
			lastByte := content[len(content)-1]
			if lastByte == 0x01 || lastByte == 0x02 {
				language = protocol.Language(lastByte)
				coords = string(content[:len(content)-1])
			}
		}
	}

	pkt := &packet.GPSAddressRequestPacket{
		BasePacket: packet.BasePacket{
			ProtocolNum: protocol.ProtocolGPSAddressRequest,
			SerialNum:   serialNum,
			RawData:     data,
			ParsedAt:    time.Now(),
		},
		Coordinates: coords,
		Language:    language,
	}

	return pkt, nil
}

// init registers command-related parsers with the default registry
func init() {
	MustRegister(NewOnlineCommandParser())
	MustRegister(NewCommandResponseParser())
	MustRegister(NewCommandResponseOldParser())
	MustRegister(NewGPSAddressRequestParser())
}
