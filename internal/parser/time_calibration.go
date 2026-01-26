package parser

import (
	"time"

	"github.com/fcode09/jimi-vl103m/pkg/jimi/packet"
	"github.com/fcode09/jimi-vl103m/pkg/jimi/protocol"
)

// TimeCalibrationParser parses time calibration request packets (Protocol 0x8A)
type TimeCalibrationParser struct {
	BaseParser
}

// NewTimeCalibrationParser creates a new time calibration parser
func NewTimeCalibrationParser() *TimeCalibrationParser {
	return &TimeCalibrationParser{
		BaseParser: NewBaseParser(protocol.ProtocolTimeCalibration, "Time Calibration"),
	}
}

// Parse implements Parser interface
// Time calibration packet has no content - it's just a request
func (p *TimeCalibrationParser) Parse(data []byte) (packet.Packet, error) {
	// Extract serial number
	serialNum, _ := ExtractSerialNumber(data)

	pkt := &packet.TimeCalibrationPacket{
		BasePacket: packet.BasePacket{
			ProtocolNum: protocol.ProtocolTimeCalibration,
			SerialNum:   serialNum,
			RawData:     data,
			ParsedAt:    time.Now(),
		},
	}

	return pkt, nil
}

// init registers the time calibration parser with the default registry
func init() {
	MustRegister(NewTimeCalibrationParser())
}
