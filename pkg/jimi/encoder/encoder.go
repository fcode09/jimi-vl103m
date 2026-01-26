// Package encoder provides functionality to encode server responses for the VL103M protocol.
//
// The encoder creates properly formatted response packets that can be sent back
// to GPS tracking devices. Each response type has its own builder for convenience.
//
// Example usage:
//
//	encoder := encoder.New()
//
//	// Create a login response
//	response, err := encoder.LoginResponse(serialNum)
//
//	// Send response to device
//	conn.Write(response)
package encoder

import (
	"time"

	"github.com/fcode09/jimi-vl103m/internal/validator"
	"github.com/fcode09/jimi-vl103m/pkg/jimi/protocol"
)

// Encoder creates response packets for the VL103M protocol
type Encoder struct {
	// UseShortFormat uses 0x7878 format (default)
	// Set to false to use 0x7979 format for long packets
	UseShortFormat bool
}

// New creates a new Encoder with default settings
func New() *Encoder {
	return &Encoder{
		UseShortFormat: true,
	}
}

// buildPacket creates a complete packet with start bit, length, content, CRC, and stop bit
func (e *Encoder) buildPacket(protocolNum byte, content []byte, serialNum uint16) []byte {
	// Calculate total content length: protocol(1) + content + serial(2) + crc(2)
	contentLen := 1 + len(content) + 2 + 2

	var packet []byte

	if e.UseShortFormat && contentLen <= 255 {
		// Short format: 0x7878 + length(1 byte)
		packet = make([]byte, 0, 2+1+contentLen+2)
		packet = append(packet, 0x78, 0x78)
		packet = append(packet, byte(contentLen))
	} else {
		// Long format: 0x7979 + length(2 bytes)
		packet = make([]byte, 0, 2+2+contentLen+2)
		packet = append(packet, 0x79, 0x79)
		packet = append(packet, byte(contentLen>>8), byte(contentLen&0xFF))
	}

	// Add protocol number
	packet = append(packet, protocolNum)

	// Add content
	packet = append(packet, content...)

	// Add serial number
	packet = append(packet, byte(serialNum>>8), byte(serialNum&0xFF))

	// Calculate CRC on: length field + protocol + content + serial
	var crcStart int
	if e.UseShortFormat && contentLen <= 255 {
		crcStart = 2 // After start bit
	} else {
		crcStart = 2 // After start bit
	}
	crcData := packet[crcStart:]
	crc := validator.CalculateCRC(crcData)
	packet = append(packet, byte(crc>>8), byte(crc&0xFF))

	// Add stop bit
	packet = append(packet, 0x0D, 0x0A)

	return packet
}

// LoginResponse creates a response to a login packet
// The device expects this response to confirm successful login
func (e *Encoder) LoginResponse(serialNum uint16) []byte {
	// Login response has no content, just echoes the serial number
	return e.buildPacket(protocol.ProtocolLogin, nil, serialNum)
}

// HeartbeatResponse creates a response to a heartbeat packet
func (e *Encoder) HeartbeatResponse(serialNum uint16) []byte {
	// Heartbeat response has no content
	return e.buildPacket(protocol.ProtocolHeartbeat, nil, serialNum)
}

// AlarmResponse creates a response to an alarm packet
func (e *Encoder) AlarmResponse(serialNum uint16) []byte {
	// Alarm response has no content
	return e.buildPacket(protocol.ProtocolAlarm, nil, serialNum)
}

// AlarmMultiFenceResponse creates a response to a multi-fence alarm packet
func (e *Encoder) AlarmMultiFenceResponse(serialNum uint16) []byte {
	return e.buildPacket(protocol.ProtocolAlarmMultiFence, nil, serialNum)
}

// Alarm4GResponse creates a response to a 4G alarm packet
func (e *Encoder) Alarm4GResponse(serialNum uint16) []byte {
	return e.buildPacket(protocol.ProtocolAlarmMultiFence4G, nil, serialNum)
}

// TimeCalibrationResponse creates a response with current server time
// The device uses this to synchronize its internal clock
func (e *Encoder) TimeCalibrationResponse(serialNum uint16, t time.Time) []byte {
	// Time response content: YY MM DD HH MM SS (6 bytes)
	t = t.UTC()
	content := []byte{
		byte(t.Year() - 2000),
		byte(t.Month()),
		byte(t.Day()),
		byte(t.Hour()),
		byte(t.Minute()),
		byte(t.Second()),
	}
	return e.buildPacket(protocol.ProtocolTimeCalibration, content, serialNum)
}

// TimeCalibrationResponseNow creates a time response with current time
func (e *Encoder) TimeCalibrationResponseNow(serialNum uint16) []byte {
	return e.TimeCalibrationResponse(serialNum, time.Now())
}

// OnlineCommand creates an online command packet to send to the device
// serverFlag: 4-byte identifier for tracking command responses
// command: ASCII command string
func (e *Encoder) OnlineCommand(serialNum uint16, serverFlag uint32, command string) []byte {
	// Content: length(1) + serverFlag(4) + command(variable)
	cmdBytes := []byte(command)
	contentLen := 4 + len(cmdBytes) // serverFlag + command

	content := make([]byte, 0, 1+4+len(cmdBytes))
	content = append(content, byte(contentLen))
	content = append(content,
		byte(serverFlag>>24),
		byte(serverFlag>>16),
		byte(serverFlag>>8),
		byte(serverFlag),
	)
	content = append(content, cmdBytes...)

	return e.buildPacket(protocol.ProtocolOnlineCommand, content, serialNum)
}

// AddressResponse creates an address response packet
// This is sent in response to a GPS address request (reverse geocoding)
func (e *Encoder) AddressResponse(serialNum uint16, address string, language protocol.Language) []byte {
	// Content: address string in specified language encoding
	content := []byte(address)

	var proto byte
	if language == protocol.LanguageChinese {
		proto = protocol.ProtocolAddressResponseChinese
	} else {
		proto = protocol.ProtocolAddressResponseEnglish
	}

	return e.buildPacket(proto, content, serialNum)
}

// AddressResponseChinese creates a Chinese address response
func (e *Encoder) AddressResponseChinese(serialNum uint16, address string) []byte {
	return e.AddressResponse(serialNum, address, protocol.LanguageChinese)
}

// AddressResponseEnglish creates an English address response
func (e *Encoder) AddressResponseEnglish(serialNum uint16, address string) []byte {
	return e.AddressResponse(serialNum, address, protocol.LanguageEnglish)
}

// LocationResponse creates a location packet response (if needed)
// Note: Location packets typically don't require responses
func (e *Encoder) LocationResponse(serialNum uint16) []byte {
	return e.buildPacket(protocol.ProtocolGPSLocation, nil, serialNum)
}

// LBSResponse creates an LBS packet response
func (e *Encoder) LBSResponse(serialNum uint16) []byte {
	return e.buildPacket(protocol.ProtocolLBSMultiBase, nil, serialNum)
}

// CustomResponse creates a response with custom protocol and content
// Use this for protocols not covered by specific methods
func (e *Encoder) CustomResponse(protocolNum byte, content []byte, serialNum uint16) []byte {
	return e.buildPacket(protocolNum, content, serialNum)
}

// ResponseBuilder provides a fluent interface for building responses
type ResponseBuilder struct {
	encoder     *Encoder
	protocolNum byte
	content     []byte
	serialNum   uint16
}

// NewResponseBuilder creates a new response builder
func (e *Encoder) NewResponseBuilder(protocolNum byte) *ResponseBuilder {
	return &ResponseBuilder{
		encoder:     e,
		protocolNum: protocolNum,
		content:     nil,
		serialNum:   0,
	}
}

// WithSerialNumber sets the serial number
func (b *ResponseBuilder) WithSerialNumber(serialNum uint16) *ResponseBuilder {
	b.serialNum = serialNum
	return b
}

// WithContent sets the content bytes
func (b *ResponseBuilder) WithContent(content []byte) *ResponseBuilder {
	b.content = content
	return b
}

// Build creates the final packet
func (b *ResponseBuilder) Build() []byte {
	return b.encoder.buildPacket(b.protocolNum, b.content, b.serialNum)
}
