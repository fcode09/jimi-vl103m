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
		p.parseExternalVoltage(pkt, infoData)

	case protocol.InfoTypeICCID:
		p.parseICCIDInfo(pkt, infoData)

	case protocol.InfoTypeGPSStatus:
		p.parseGPSStatus(pkt, infoData)

	case protocol.InfoTypeTerminalSync:
		p.parseTerminalSync(pkt, infoData)

	case protocol.InfoTypeDoorStatus:
		p.parseDoorStatus(pkt, infoData)

	case protocol.InfoTypeSelfCheck:
		// Self check result - variable structure
		// Just store raw data for now
	}

	return pkt, nil
}

// parseExternalVoltage parses external battery voltage
// Format: 2 bytes, value = decimal / 100 = voltage in V
// Example: 0x04 0x9F = 1183 decimal = 11.83V
func (p *InfoTransferParser) parseExternalVoltage(pkt *packet.InfoTransferPacket, data []byte) {
	if len(data) >= 2 {
		pkt.ExternalVoltage = uint16(data[0])<<8 | uint16(data[1])
	}
}

// parseICCIDInfo parses ICCID/IMEI/IMSI information
// Format:
// - IMEI: 8 bytes BCD
// - IMSI: 8 bytes BCD
// - ICCID: 10 bytes BCD
func (p *InfoTransferParser) parseICCIDInfo(pkt *packet.InfoTransferPacket, data []byte) {
	offset := 0

	// IMEI: 8 bytes
	if len(data) >= offset+8 {
		imei, err := codec.DecodeBCD(data[offset : offset+8])
		if err == nil {
			pkt.IMEI = imei
		}
		offset += 8
	}

	// IMSI: 8 bytes
	if len(data) >= offset+8 {
		imsi, err := codec.DecodeBCD(data[offset : offset+8])
		if err == nil {
			pkt.IMSI = imsi
		}
		offset += 8
	}

	// ICCID: 10 bytes
	if len(data) >= offset+10 {
		iccid, err := codec.DecodeICCID(data[offset : offset+10])
		if err == nil {
			pkt.ICCID = iccid
		}
	}
}

// parseGPSStatus parses GPS module status
// Format:
// - GPS module status: 1 byte
// - Number of satellites in fix: 1 byte
// - GPS1-N strength: 1 byte each
// - Number of visible GPS satellites: 1 byte
// - Visible GPS1-N strength: 1 byte each
// - BDS module status: 1 byte (if present)
// - ... (similar to GPS)
func (p *InfoTransferParser) parseGPSStatus(pkt *packet.InfoTransferPacket, data []byte) {
	if len(data) < 1 {
		return
	}

	pkt.GPSStatus = protocol.GPSModuleStatus(data[0])

	// Parse detailed GPS status if more data available
	if len(data) >= 2 {
		gpsInfo := &packet.GPSStatusData{
			ModuleStatus: protocol.GPSModuleStatus(data[0]),
		}

		offset := 1

		// Number of satellites in fix
		if offset < len(data) {
			gpsInfo.SatellitesInFix = int(data[offset])
			offset++

			// Read satellite strengths
			for i := 0; i < gpsInfo.SatellitesInFix && offset < len(data); i++ {
				gpsInfo.SatelliteStrengths = append(gpsInfo.SatelliteStrengths, int(data[offset]))
				offset++
			}
		}

		// Number of visible satellites
		if offset < len(data) {
			gpsInfo.VisibleSatellites = int(data[offset])
			offset++

			// Read visible satellite strengths
			for i := 0; i < gpsInfo.VisibleSatellites && offset < len(data); i++ {
				gpsInfo.VisibleStrengths = append(gpsInfo.VisibleStrengths, int(data[offset]))
				offset++
			}
		}

		// BDS module status (if present)
		if offset < len(data) {
			gpsInfo.BDSModuleStatus = protocol.GPSModuleStatus(data[offset])
			offset++

			// BDS satellites in fix
			if offset < len(data) {
				gpsInfo.BDSSatellitesInFix = int(data[offset])
				offset++

				for i := 0; i < gpsInfo.BDSSatellitesInFix && offset < len(data); i++ {
					gpsInfo.BDSSatelliteStrengths = append(gpsInfo.BDSSatelliteStrengths, int(data[offset]))
					offset++
				}
			}

			// BDS visible satellites
			if offset < len(data) {
				gpsInfo.BDSVisibleSatellites = int(data[offset])
				offset++

				for i := 0; i < gpsInfo.BDSVisibleSatellites && offset < len(data); i++ {
					gpsInfo.BDSVisibleStrengths = append(gpsInfo.BDSVisibleStrengths, int(data[offset]))
					offset++
				}
			}
		}

		pkt.GPSStatusInfo = gpsInfo
	}
}

// parseTerminalSync parses terminal status synchronization data
// Format: ASCII string with key=value pairs separated by semicolons
// Example: ALM1=CC;ALM2=C4;ALM3=DC;STA1=C0;DYD=01;SOS=945538609,,;CENTER=+51974867548;GFENCE1,OFF,...;IMSI=...;ICCID=...;
func (p *InfoTransferParser) parseTerminalSync(pkt *packet.InfoTransferPacket, data []byte) {
	// Data is ASCII string
	dataStr := string(data)

	// Parse the terminal sync string
	pkt.TerminalSync = packet.ParseTerminalSyncString(dataStr)

	// Also populate top-level fields if available
	if pkt.TerminalSync != nil {
		if pkt.TerminalSync.ICCID != "" {
			pkt.ICCID = pkt.TerminalSync.ICCID
		}
		if pkt.TerminalSync.IMSI != "" {
			pkt.IMSI = pkt.TerminalSync.IMSI
		}
	}
}

// parseDoorStatus parses door status byte
// Format: 1 byte with bit flags
// - bit0: Door status (1=ON/Open, 0=OFF/Closed)
// - bit1: Trigger status (1=Level high, 0=Level low)
// - bit2: I/O port status (1=High, 0=Low)
func (p *InfoTransferParser) parseDoorStatus(pkt *packet.InfoTransferPacket, data []byte) {
	if len(data) >= 1 {
		pkt.DoorStatus = packet.ParseDoorStatusByte(data[0])
	}
}

// init registers the info transfer parser with the default registry
func init() {
	MustRegister(NewInfoTransferParser())
}
