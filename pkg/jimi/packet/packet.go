package packet

import (
	"time"

	"github.com/intelcon-group/jimi-vl103m/pkg/jimi/protocol"
)

// Packet is the base interface for all decoded packets
// All specific packet types (Login, Location, Alarm, etc.) implement this interface
type Packet interface {
	// ProtocolNumber returns the protocol number for this packet type
	ProtocolNumber() byte

	// SerialNumber returns the information serial number
	// This number auto-increments with each packet sent by the device
	SerialNumber() uint16

	// Timestamp returns the packet timestamp (if available)
	// Returns zero time if packet doesn't contain timestamp
	Timestamp() time.Time

	// Raw returns the raw packet bytes
	Raw() []byte

	// Type returns the human-readable packet type name
	Type() string

	// Validate performs validation on the packet data
	// Returns error if packet data is invalid
	Validate() error
}

// BasePacket contains fields common to all packets
// Specific packet types should embed this struct
type BasePacket struct {
	ProtocolNum byte      // Protocol number (e.g., 0x01, 0x22, etc.)
	SerialNum   uint16    // Information serial number
	RawData     []byte    // Original raw packet bytes
	ParsedAt    time.Time // When this packet was parsed
}

// ProtocolNumber implements Packet interface
func (p *BasePacket) ProtocolNumber() byte {
	return p.ProtocolNum
}

// SerialNumber implements Packet interface
func (p *BasePacket) SerialNumber() uint16 {
	return p.SerialNum
}

// Raw implements Packet interface
func (p *BasePacket) Raw() []byte {
	return p.RawData
}

// Timestamp implements Packet interface (base implementation)
// Specific packet types should override if they have a timestamp
func (p *BasePacket) Timestamp() time.Time {
	return time.Time{} // Zero time by default
}

// Type implements Packet interface (base implementation)
// Specific packet types should override with meaningful type name
func (p *BasePacket) Type() string {
	switch p.ProtocolNum {
	case protocol.ProtocolLogin:
		return "Login"
	case protocol.ProtocolHeartbeat:
		return "Heartbeat"
	case protocol.ProtocolGPSLocation:
		return "GPS Location"
	case protocol.ProtocolGPSLocation4G:
		return "GPS Location 4G"
	case protocol.ProtocolLBSMultiBase:
		return "LBS Multi-Base"
	case protocol.ProtocolLBSMultiBase4G:
		return "LBS Multi-Base 4G"
	case protocol.ProtocolAlarm:
		return "Alarm"
	case protocol.ProtocolAlarmMultiFence:
		return "Alarm Multi-Fence"
	case protocol.ProtocolAlarmMultiFence4G:
		return "Alarm 4G"
	case protocol.ProtocolGPSAddressRequest:
		return "GPS Address Request"
	case protocol.ProtocolOnlineCommand:
		return "Online Command"
	case protocol.ProtocolCommandResponse:
		return "Command Response"
	case protocol.ProtocolCommandResponseOld:
		return "Command Response (Old)"
	case protocol.ProtocolTimeCalibration:
		return "Time Calibration"
	case protocol.ProtocolInfoTransfer:
		return "Information Transfer"
	case protocol.ProtocolAddressResponseChinese:
		return "Address Response (Chinese)"
	case protocol.ProtocolAddressResponseEnglish:
		return "Address Response (English)"
	default:
		return "Unknown"
	}
}

// Validate implements Packet interface (base implementation)
// Specific packet types should override with actual validation logic
func (p *BasePacket) Validate() error {
	if len(p.RawData) < protocol.MinPacketSize {
		return ErrInvalidPacketSize
	}
	return nil
}

// Common errors for packet validation
var (
	ErrInvalidPacketSize = &ValidationError{
		Field:  "PacketSize",
		Reason: "packet too small",
	}
	ErrMissingTimestamp = &ValidationError{
		Field:  "Timestamp",
		Reason: "packet does not contain timestamp",
	}
)

// ValidationError represents a packet validation error
type ValidationError struct {
	Field  string
	Reason string
	Value  any
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	if e.Value != nil {
		return "validation error for " + e.Field + ": " + e.Reason + " (value: " + e.Reason + ")"
	}
	return "validation error for " + e.Field + ": " + e.Reason
}

// PacketWithTimestamp is an interface for packets that contain timestamp
type PacketWithTimestamp interface {
	Packet
	// HasTimestamp returns true if the packet contains a valid timestamp
	HasTimestamp() bool
}

// PacketWithLocation is an interface for packets that contain GPS location
type PacketWithLocation interface {
	Packet
	// HasLocation returns true if the packet contains valid GPS coordinates
	HasLocation() bool
	// IsPositioned returns true if GPS has a valid fix
	IsPositioned() bool
}

// PacketWithAlarm is an interface for alarm packets
type PacketWithAlarm interface {
	Packet
	// GetAlarmType returns the alarm type
	GetAlarmType() protocol.AlarmType
	// IsCritical returns true if the alarm is critical
	IsCritical() bool
}

// PacketWithIMEI is an interface for packets that contain device IMEI
type PacketWithIMEI interface {
	Packet
	// GetIMEI returns the device IMEI
	GetIMEI() string
}

// Helper functions

// IsLoginPacket returns true if the packet is a login packet
func IsLoginPacket(p Packet) bool {
	return p.ProtocolNumber() == protocol.ProtocolLogin
}

// IsHeartbeatPacket returns true if the packet is a heartbeat packet
func IsHeartbeatPacket(p Packet) bool {
	return p.ProtocolNumber() == protocol.ProtocolHeartbeat
}

// IsLocationPacket returns true if the packet is a location packet (2G or 4G)
func IsLocationPacket(p Packet) bool {
	proto := p.ProtocolNumber()
	return proto == protocol.ProtocolGPSLocation || proto == protocol.ProtocolGPSLocation4G
}

// IsAlarmPacket returns true if the packet is an alarm packet
func IsAlarmPacket(p Packet) bool {
	proto := p.ProtocolNumber()
	return proto == protocol.ProtocolAlarm ||
		proto == protocol.ProtocolAlarmMultiFence ||
		proto == protocol.ProtocolAlarmMultiFence4G
}

// IsLBSPacket returns true if the packet is an LBS packet
func IsLBSPacket(p Packet) bool {
	proto := p.ProtocolNumber()
	return proto == protocol.ProtocolLBSMultiBase || proto == protocol.ProtocolLBSMultiBase4G
}

// RequiresResponse returns true if the protocol requires a server response
func RequiresResponse(p Packet) bool {
	proto := p.ProtocolNumber()
	switch proto {
	case protocol.ProtocolLogin,
		protocol.ProtocolHeartbeat,
		protocol.ProtocolAlarm,
		protocol.ProtocolAlarmMultiFence,
		protocol.ProtocolAlarmMultiFence4G,
		protocol.ProtocolTimeCalibration:
		return true
	default:
		return false
	}
}

// GetProtocolName returns the human-readable protocol name
func GetProtocolName(protocolNum byte) string {
	p := &BasePacket{ProtocolNum: protocolNum}
	return p.Type()
}
