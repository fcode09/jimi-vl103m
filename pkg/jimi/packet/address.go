package packet

import (
	"fmt"
	"time"

	"github.com/intelcon-group/jimi-vl103m/pkg/jimi/protocol"
)

// AddressResponsePacket represents an address response packet from server
// This is sent by the server to the terminal after receiving an alarm packet
// with parsed address information for the GPS coordinates
//
// Protocols:
// - 0x17: Chinese Address Packet (short packet format)
// - 0x97: English Address Packet (long packet format)
type AddressResponsePacket struct {
	BasePacket

	// ContentLength is the length of data between server flag and serial number
	ContentLength uint8

	// ServerFlag is a 4-byte marker used by server to identify the alarm
	ServerFlag [4]byte

	// AlarmSMS is the alarm code flag (typically "ALARMSMS" in ASCII)
	AlarmSMS string

	// Address is the parsed address string
	// - For 0x17 (Chinese): UNICODE encoded
	// - For 0x97 (English): ASCII/UTF-8 encoded
	Address string

	// PhoneNumber is the destination phone number for SMS
	// Typically "0" repeated 21 times for alarm packets uploaded to server
	PhoneNumber string

	// Language indicates the language of the address
	Language protocol.Language
}

// NewAddressResponsePacket creates a new address response packet
func NewAddressResponsePacket(protocolNum byte, address string, lang protocol.Language) *AddressResponsePacket {
	return &AddressResponsePacket{
		BasePacket: BasePacket{
			ProtocolNum: protocolNum,
			ParsedAt:    time.Now(),
		},
		Address:  address,
		Language: lang,
	}
}

// Type implements Packet interface
func (p *AddressResponsePacket) Type() string {
	if p.ProtocolNum == protocol.ProtocolAddressResponseChinese {
		return "Address Response (Chinese)"
	}
	return "Address Response (English)"
}

// Timestamp implements Packet interface
// Address packets don't have timestamp
func (p *AddressResponsePacket) Timestamp() time.Time {
	return time.Time{} // Zero time
}

// Validate implements Packet interface
func (p *AddressResponsePacket) Validate() error {
	if len(p.Address) == 0 {
		return &ValidationError{
			Field:  "Address",
			Reason: "address is empty",
		}
	}
	if len(p.PhoneNumber) != 21 {
		return &ValidationError{
			Field:  "PhoneNumber",
			Reason: fmt.Sprintf("phone number must be 21 bytes, got %d", len(p.PhoneNumber)),
			Value:  len(p.PhoneNumber),
		}
	}
	return nil
}

// IsChinese returns true if this is a Chinese address packet
func (p *AddressResponsePacket) IsChinese() bool {
	return p.ProtocolNum == protocol.ProtocolAddressResponseChinese
}

// IsEnglish returns true if this is an English address packet
func (p *AddressResponsePacket) IsEnglish() bool {
	return p.ProtocolNum == protocol.ProtocolAddressResponseEnglish
}

// String returns a human-readable representation
func (p *AddressResponsePacket) String() string {
	lang := "Chinese"
	if p.IsEnglish() {
		lang = "English"
	}
	return fmt.Sprintf("AddressResponsePacket{Type: %s, Address: %s, AlarmSMS: %s}",
		lang,
		p.Address,
		p.AlarmSMS)
}
