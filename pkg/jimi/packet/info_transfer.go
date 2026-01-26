package packet

import (
	"fmt"
	"time"

	"github.com/fcode09/jimi-vl103m/pkg/jimi/protocol"
)

// InfoTransferPacket represents an information transfer packet (Protocol 0x94)
// This is a general-purpose packet for transferring various types of information
//
// Content structure:
// - Sub-protocol: 1 byte (identifies the type of information)
// - Data: variable length (depends on sub-protocol)
type InfoTransferPacket struct {
	BasePacket

	// SubProtocol identifies the type of information being transferred
	SubProtocol protocol.InfoType

	// Data contains the raw information data
	Data []byte

	// Parsed fields depending on SubProtocol:

	// ExternalVoltage in millivolts (when SubProtocol = InfoTypeExternalVoltage)
	ExternalVoltage uint16

	// ICCID (when SubProtocol = InfoTypeICCID)
	ICCID string

	// GPSModuleStatus (when SubProtocol = InfoTypeGPSStatus)
	GPSStatus protocol.GPSModuleStatus
}

// NewInfoTransferPacket creates a new InfoTransferPacket
func NewInfoTransferPacket(subProtocol protocol.InfoType, data []byte) *InfoTransferPacket {
	return &InfoTransferPacket{
		BasePacket: BasePacket{
			ProtocolNum: protocol.ProtocolInfoTransfer,
			ParsedAt:    time.Now(),
		},
		SubProtocol: subProtocol,
		Data:        data,
	}
}

// Type implements Packet interface
func (p *InfoTransferPacket) Type() string {
	return "Information Transfer"
}

// Timestamp implements Packet interface
func (p *InfoTransferPacket) Timestamp() time.Time {
	return p.ParsedAt
}

// Validate implements Packet interface
func (p *InfoTransferPacket) Validate() error {
	return nil
}

// String returns a human-readable representation
func (p *InfoTransferPacket) String() string {
	switch p.SubProtocol {
	case protocol.InfoTypeExternalVoltage:
		return fmt.Sprintf("InfoTransferPacket{Type: External Voltage, Value: %d mV}", p.ExternalVoltage)
	case protocol.InfoTypeICCID:
		return fmt.Sprintf("InfoTransferPacket{Type: ICCID, Value: %s}", p.ICCID)
	case protocol.InfoTypeGPSStatus:
		return fmt.Sprintf("InfoTransferPacket{Type: GPS Status, Value: %s}", p.GPSStatus)
	default:
		return fmt.Sprintf("InfoTransferPacket{Type: %s, DataLen: %d}", p.SubProtocol, len(p.Data))
	}
}

// GetExternalVoltageVolts returns the external voltage in volts
func (p *InfoTransferPacket) GetExternalVoltageVolts() float64 {
	return float64(p.ExternalVoltage) / 1000.0
}
