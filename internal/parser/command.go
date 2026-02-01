package parser

import (
	"fmt"
	"strings"
	"time"

	"github.com/intelcon-group/jimi-vl103m/pkg/jimi/packet"
	"github.com/intelcon-group/jimi-vl103m/pkg/jimi/protocol"
	"github.com/intelcon-group/jimi-vl103m/pkg/jimi/types"
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
// Content structure (41 bytes according to JM-VL03 spec):
// - DateTime: 6 bytes (YY MM DD HH MM SS)
// - GPS Info: 1 byte (satellites in low nibble)
// - Latitude: 4 bytes (raw / 1,800,000)
// - Longitude: 4 bytes (raw / 1,800,000)
// - Speed: 1 byte (km/h)
// - Course/Status: 2 bytes (heading + status flags)
// - Phone Number: 21 bytes (ASCII)
// - Alert/Language: 2 bytes (AlarmType + Language)
// Total content: 41 bytes
func (p *GPSAddressRequestParser) Parse(data []byte) (packet.Packet, error) {
	content, err := ExtractContent(data)
	if err != nil {
		return nil, fmt.Errorf("gps_address_request: %w", err)
	}

	// STRICT: Require exactly 41 bytes according to specification
	if len(content) < 41 {
		return nil, fmt.Errorf("gps_address_request: content too short: got %d bytes, need exactly 41", len(content))
	}

	// If more than 41 bytes, use first 41 and ignore extras
	if len(content) > 41 {
		content = content[:41]
	}

	offset := 0

	// 1. Parse DateTime (6 bytes)
	dt, err := types.DateTimeFromBytes(content[offset : offset+6])
	if err != nil {
		return nil, fmt.Errorf("gps_address_request: failed to parse datetime: %w", err)
	}
	offset += 6

	// 2. Parse GPS Info byte (1 byte)
	// Low nibble (bits 3-0): Number of satellites
	gpsInfoByte := content[offset]
	satellites := gpsInfoByte & 0x0F
	offset++

	// 3. Parse Latitude (4 bytes)
	latBytes := content[offset : offset+4]
	offset += 4

	// 4. Parse Longitude (4 bytes)
	lonBytes := content[offset : offset+4]
	offset += 4

	// 5. Parse Speed (1 byte)
	speed := content[offset]
	offset++

	// 6. Parse Course/Status (2 bytes)
	courseStatus, err := types.NewCourseStatusFromBytes(content[offset : offset+2])
	if err != nil {
		return nil, fmt.Errorf("gps_address_request: failed to parse course/status: %w", err)
	}
	offset += 2

	// Create coordinates with hemisphere info from course status
	coords, err := types.NewCoordinatesFromBytes(
		latBytes, lonBytes,
		courseStatus.IsNorthLatitude,
		courseStatus.IsEastLongitude,
	)
	if err != nil {
		return nil, fmt.Errorf("gps_address_request: failed to parse coordinates: %w", err)
	}

	// 7. Parse Phone Number (21 bytes ASCII)
	phoneBytes := content[offset : offset+21]
	// Trim null bytes and whitespace from the end
	phoneNumber := strings.TrimRight(string(phoneBytes), "\x00 \t\n\r")
	offset += 21

	// 8. Parse Alert and Language (2 bytes)
	alarmType := protocol.AlarmType(content[offset])
	offset++
	language := protocol.Language(content[offset])

	// Extract serial number
	serialNum, _ := ExtractSerialNumber(data)

	return &packet.GPSAddressRequestPacket{
		BasePacket: packet.BasePacket{
			ProtocolNum: protocol.ProtocolGPSAddressRequest,
			SerialNum:   serialNum,
			RawData:     data,
			ParsedAt:    time.Now(),
		},
		DateTime:     dt,
		Satellites:   satellites,
		Coordinates:  coords,
		Speed:        speed,
		CourseStatus: courseStatus,
		PhoneNumber:  phoneNumber,
		AlarmType:    alarmType,
		Language:     language,
	}, nil
}

// init registers command-related parsers with the default registry
func init() {
	MustRegister(NewOnlineCommandParser())
	MustRegister(NewCommandResponseParser())
	MustRegister(NewCommandResponseOldParser())
	MustRegister(NewGPSAddressRequestParser())
}
