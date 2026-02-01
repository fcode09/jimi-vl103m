package packet

import (
	"time"

	"github.com/fcode09/jimi-vl103m/pkg/jimi/protocol"
	"github.com/fcode09/jimi-vl103m/pkg/jimi/types"
)

// LoginPacket represents a login packet (Protocol 0x01)
// This is the first packet sent by the device when connecting
//
// Content structure:
// - IMEI: 8 bytes (BCD encoded)
// - Model Identification Code: 2 bytes
// - Timezone/Language: 2 bytes
type LoginPacket struct {
	BasePacket

	// IMEI is the device identifier (15 digits)
	IMEI types.IMEI

	// ModelID is the device model identification code
	ModelID uint16

	// Timezone contains timezone offset and language setting
	Timezone types.Timezone
}

// NewLoginPacket creates a new LoginPacket
func NewLoginPacket(imei types.IMEI, modelID uint16, tz types.Timezone) *LoginPacket {
	return &LoginPacket{
		BasePacket: BasePacket{
			ProtocolNum: protocol.ProtocolLogin,
			ParsedAt:    time.Now(),
		},
		IMEI:     imei,
		ModelID:  modelID,
		Timezone: tz,
	}
}

// Type implements Packet interface
func (p *LoginPacket) Type() string {
	return "Login"
}

// Timestamp implements Packet interface
// Login packets don't have a timestamp field
func (p *LoginPacket) Timestamp() time.Time {
	return p.ParsedAt
}

// Validate implements Packet interface
func (p *LoginPacket) Validate() error {
	if !p.IMEI.IsValid() {
		return &ValidationError{Field: "IMEI", Reason: "invalid IMEI"}
	}
	return nil
}

// GetIMEI implements PacketWithIMEI interface
func (p *LoginPacket) GetIMEI() string {
	return p.IMEI.String()
}

// String returns a human-readable representation
func (p *LoginPacket) String() string {
	return "LoginPacket{IMEI: " + p.IMEI.String() + ", ModelID: " + formatModelID(p.ModelID) + ", Timezone: " + p.Timezone.String() + "}"
}

// formatModelID formats the model ID as hex
func formatModelID(id uint16) string {
	const hexDigits = "0123456789ABCDEF"
	return "0x" + string([]byte{
		hexDigits[(id>>12)&0xF],
		hexDigits[(id>>8)&0xF],
		hexDigits[(id>>4)&0xF],
		hexDigits[id&0xF],
	})
}
