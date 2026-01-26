package parser

import (
	"fmt"
	"time"

	"github.com/intelcon-group/jimi-vl103m/internal/codec"
	"github.com/intelcon-group/jimi-vl103m/pkg/jimi/packet"
	"github.com/intelcon-group/jimi-vl103m/pkg/jimi/protocol"
)

// InfoTransferParser parses information transfer packets (Protocol 0x94)
type InfoTransferParser struct {
	BaseParser
}

// NewInfoTransferParser creates a new info transfer parser
func NewInfoTransferParser() *InfoTransferParser {
	return &InfoTransferParser{
		BaseParser: NewBaseParser(protocol.ProtocolInfoTransfer, "Information Transfer"),
	}
}

// Parse implements Parser interface
// Info transfer packet content structure:
// - Sub-protocol: 1 byte
// - Data: variable length
func (p *InfoTransferParser) Parse(data []byte) (packet.Packet, error) {
	content, err := ExtractContent(data)
	if err != nil {
		return nil, fmt.Errorf("info_transfer: %w", err)
	}

	if len(content) < 1 {
		return nil, fmt.Errorf("info_transfer: content too short")
	}

	// Parse sub-protocol
	subProtocol := protocol.InfoType(content[0])
	infoData := content[1:]

	// Extract serial number
	serialNum, _ := ExtractSerialNumber(data)

	pkt := &packet.InfoTransferPacket{
		BasePacket: packet.BasePacket{
			ProtocolNum: protocol.ProtocolInfoTransfer,
			SerialNum:   serialNum,
			RawData:     data,
			ParsedAt:    time.Now(),
		},
		SubProtocol: subProtocol,
		Data:        infoData,
	}

	// Parse specific sub-protocol data
	switch subProtocol {
	case protocol.InfoTypeExternalVoltage:
		// External voltage: 2 bytes, value in millivolts
		if len(infoData) >= 2 {
			pkt.ExternalVoltage = uint16(infoData[0])<<8 | uint16(infoData[1])
		}

	case protocol.InfoTypeICCID:
		// ICCID: 10 bytes BCD encoded (20 digits)
		if len(infoData) >= 10 {
			iccid, err := codec.DecodeICCID(infoData[:10])
			if err == nil {
				pkt.ICCID = iccid
			}
		}

	case protocol.InfoTypeGPSStatus:
		// GPS status: 1 byte
		if len(infoData) >= 1 {
			pkt.GPSStatus = protocol.GPSModuleStatus(infoData[0])
		}

	case protocol.InfoTypeTerminalSync:
		// Terminal synchronization info - variable structure
		// Just store raw data for now

	case protocol.InfoTypeDoorStatus:
		// Door status - typically 1 byte
		// Just store raw data for now

	case protocol.InfoTypeSelfCheck:
		// Self check result - variable structure
		// Just store raw data for now
	}

	return pkt, nil
}

// init registers the info transfer parser with the default registry
func init() {
	MustRegister(NewInfoTransferParser())
}
