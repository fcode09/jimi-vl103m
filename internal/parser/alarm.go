package parser

import (
	"fmt"
	"time"

	"github.com/fcode09/jimi-vl103m/pkg/jimi/packet"
	"github.com/fcode09/jimi-vl103m/pkg/jimi/protocol"
	"github.com/fcode09/jimi-vl103m/pkg/jimi/types"
)

// AlarmParser parses alarm packets (Protocol 0x26)
type AlarmParser struct {
	BaseParser
}

// NewAlarmParser creates a new alarm parser
func NewAlarmParser() *AlarmParser {
	return &AlarmParser{
		BaseParser: NewBaseParser(protocol.ProtocolAlarm, "Alarm"),
	}
}

// Parse implements Parser interface
// Alarm packet content structure (Protocol 0x26):
// - DateTime: 6 bytes (YY MM DD HH MM SS)
// - GPS Info Length: 1 byte
// - Latitude: 4 bytes
// - Longitude: 4 bytes
// - Speed: 1 byte
// - Course/Status: 2 bytes
// - LBS Length: 1 byte (total length of LBS info)
// - MCC: 2 bytes
// - MNC: 1 byte
// - LAC: 2 bytes
// - CellID: 3 bytes
// - Terminal Info: 1 byte
// - Voltage Level: 1 byte
// - GSM Signal: 1 byte
// - Alarm Type: 1 byte (Alert and Language byte 1)
// - Language: 1 byte (Alert and Language byte 2)
// - Mileage: 4 bytes
// Total content: 33 bytes minimum
func (p *AlarmParser) Parse(data []byte, ctx Context) (packet.Packet, error) {
	content, err := ExtractContent(data)
	if err != nil {
		return nil, fmt.Errorf("alarm: %w", err)
	}

	if len(content) < 32 {
		return nil, fmt.Errorf("alarm: content too short: %d bytes (need at least 32)", len(content))
	}

	offset := 0

	// Parse DateTime (6 bytes)
	dt, err := types.DateTimeFromBytes(content[offset : offset+6])
	if err != nil {
		return nil, fmt.Errorf("alarm: failed to parse datetime: %w", err)
	}
	offset += 6

	// Parse GPS Info byte (1 byte)
	// High nibble (bits 7-4): GPS Info Length indicator
	// Low nibble (bits 3-0): Number of satellites
	gpsInfoByte := content[offset]
	satellites := gpsInfoByte & 0x0F // Low nibble = satellites
	offset++

	// Parse Latitude (4 bytes)
	latBytes := content[offset : offset+4]
	offset += 4

	// Parse Longitude (4 bytes)
	lonBytes := content[offset : offset+4]
	offset += 4

	// Parse Speed (1 byte)
	speed := content[offset]
	offset++

	// Parse Course/Status (2 bytes)
	courseStatus, err := types.NewCourseStatusFromBytes(content[offset : offset+2])
	if err != nil {
		return nil, fmt.Errorf("alarm: failed to parse course status: %w", err)
	}
	offset += 2

	// Create coordinates with hemisphere info from course status
	coords, err := types.NewCoordinatesFromBytes(
		latBytes, lonBytes,
		courseStatus.IsNorthLatitude,
		courseStatus.IsEastLongitude,
	)
	if err != nil {
		return nil, fmt.Errorf("alarm: failed to parse coordinates: %w", err)
	}

	// Parse LBS Length (1 byte) - indicates total length of LBS info
	lbsLength := content[offset]
	offset++

	// Parse LBS Info based on length
	// For 2G: MCC(2) + MNC(1) + LAC(2) + CellID(3) = 8 bytes
	var lbsInfo types.LBSInfo
	if lbsLength > 0 && offset+8 <= len(content) {
		lbsInfo, _, err = types.NewLBSInfoFromBytes(content[offset:offset+8], false)
		if err != nil {
			lbsInfo = types.LBSInfo{}
		}
		offset += 8
	} else {
		// Skip LBS bytes if present but length is unexpected
		offset += 8
	}

	// Parse Terminal Info (1 byte)
	terminalInfo := types.NewTerminalInfo(content[offset])
	offset++

	// Parse Voltage Level (1 byte)
	voltageLevel := protocol.VoltageLevel(content[offset])
	offset++

	// Parse GSM Signal (1 byte)
	gsmSignal := protocol.GSMSignalStrength(content[offset])
	offset++

	// Parse Alert and Language (2 bytes)
	alarmType := protocol.AlarmType(content[offset])
	offset++
	language := protocol.Language(content[offset])
	offset++

	// Parse Mileage Statistics (4 bytes)
	var mileage uint32
	if offset+4 <= len(content) {
		mileage = uint32(content[offset])<<24 | uint32(content[offset+1])<<16 |
			uint32(content[offset+2])<<8 | uint32(content[offset+3])
		offset += 4
	}

	// Extract serial number
	serialNum, _ := ExtractSerialNumber(data)

	pkt := &packet.AlarmPacket{
		BasePacket: packet.BasePacket{
			ProtocolNum: protocol.ProtocolAlarm,
			SerialNum:   serialNum,
			RawData:     data,
			ParsedAt:    time.Now(),
		},
		DateTime:     dt,
		Satellites:   satellites,
		Coordinates:  coords,
		Speed:        speed,
		CourseStatus: courseStatus,
		LBSInfo:      lbsInfo,
		TerminalInfo: terminalInfo,
		VoltageLevel: voltageLevel,
		GSMSignal:    gsmSignal,
		AlarmType:    alarmType,
		Language:     language,
		Mileage:      mileage,
	}

	return pkt, nil
}

// AlarmMultiFenceParser parses alarm packets with geo-fence data (Protocol 0x27)
type AlarmMultiFenceParser struct {
	BaseParser
}

// NewAlarmMultiFenceParser creates a new multi-fence alarm parser
func NewAlarmMultiFenceParser() *AlarmMultiFenceParser {
	return &AlarmMultiFenceParser{
		BaseParser: NewBaseParser(protocol.ProtocolAlarmMultiFence, "Alarm Multi-Fence"),
	}
}

// Parse implements Parser interface
// Multi-fence alarm packet has the same structure as alarm but with fence ID
func (p *AlarmMultiFenceParser) Parse(data []byte, ctx Context) (packet.Packet, error) {
	content, err := ExtractContent(data)
	if err != nil {
		return nil, fmt.Errorf("alarm_multi_fence: %w", err)
	}

	if len(content) < 29 {
		return nil, fmt.Errorf("alarm_multi_fence: content too short: %d bytes", len(content))
	}

	offset := 0

	// DateTime
	dt, err := types.DateTimeFromBytes(content[offset : offset+6])
	if err != nil {
		return nil, fmt.Errorf("alarm_multi_fence: %w", err)
	}
	offset += 6

	// GPS Info
	satellites := (content[offset] >> 4) & 0x0F
	offset++

	// Coordinates
	latBytes := content[offset : offset+4]
	offset += 4
	lonBytes := content[offset : offset+4]
	offset += 4

	// Speed
	speed := content[offset]
	offset++

	// Course/Status
	courseStatus, err := types.NewCourseStatusFromBytes(content[offset : offset+2])
	if err != nil {
		return nil, fmt.Errorf("alarm_multi_fence: %w", err)
	}
	offset += 2

	coords, err := types.NewCoordinatesFromBytes(latBytes, lonBytes, courseStatus.IsNorthLatitude, courseStatus.IsEastLongitude)
	if err != nil {
		return nil, fmt.Errorf("alarm_multi_fence: %w", err)
	}

	// LBS Length and Info
	var lbsInfo types.LBSInfo
	lbsLength := content[offset]
	offset++
	if lbsLength > 0 && offset+8 <= len(content) {
		lbsInfo, _, _ = types.NewLBSInfoFromBytes(content[offset:offset+8], false)
		offset += 8
	}

	// Status fields
	terminalInfo := types.NewTerminalInfo(content[offset])
	offset++
	voltageLevel := protocol.VoltageLevel(content[offset])
	offset++
	gsmSignal := protocol.GSMSignalStrength(content[offset])
	offset++

	// Alert and Language
	alarmType := protocol.AlarmType(content[offset])
	offset++
	language := protocol.Language(content[offset])
	offset++

	// Fence ID
	fenceID := content[offset]
	offset++

	// Create base alarm packet for embedding
	alarmPkt := packet.AlarmPacket{
		BasePacket: packet.BasePacket{
			ProtocolNum: protocol.ProtocolAlarmMultiFence,
			ParsedAt:    time.Now(),
		},
		DateTime:     dt,
		Satellites:   satellites,
		Coordinates:  coords,
		Speed:        speed,
		CourseStatus: courseStatus,
		LBSInfo:      lbsInfo,
		TerminalInfo: terminalInfo,
		VoltageLevel: voltageLevel,
		GSMSignal:    gsmSignal,
		AlarmType:    alarmType,
		Language:     language,
	}

	// Optional Mileage
	if offset+4 <= len(content) {
		alarmPkt.Mileage = uint32(content[offset])<<24 | uint32(content[offset+1])<<16 | uint32(content[offset+2])<<8 | uint32(content[offset+3])
	}

	pkt := &packet.AlarmMultiFencePacket{
		AlarmPacket: alarmPkt,
		FenceID:     fenceID,
	}

	// Set common BasePacket fields
	serialNum, _ := ExtractSerialNumber(data)
	pkt.SerialNum = serialNum
	pkt.RawData = data

	return pkt, nil
}

// Alarm4GParser parses 4G alarm packets (Protocol 0xA4)
type Alarm4GParser struct {
	BaseParser
}

// NewAlarm4GParser creates a new 4G alarm parser
func NewAlarm4GParser() *Alarm4GParser {
	return &Alarm4GParser{
		BaseParser: NewBaseParser(protocol.ProtocolAlarmMultiFence4G, "Alarm 4G"),
	}
}

func (p *Alarm4GParser) Parse(data []byte, ctx Context) (packet.Packet, error) {
	content, err := ExtractContent(data)
	if err != nil {
		return nil, fmt.Errorf("alarm_4g: %w", err)
	}

	if len(content) < 20 { // Absolute minimum
		return nil, fmt.Errorf("alarm_4g: content too short: %d bytes", len(content))
	}

	offset := 0

	// === Parse fixed-size fields from the start ===
	dt, err := types.DateTimeFromBytes(content[offset : offset+6])
	if err != nil {
		return nil, fmt.Errorf("alarm_4g: failed to parse datetime: %w", err)
	}
	offset += 6

	// Parse GPS Info byte (1 byte)
	// High nibble (bits 7-4): GPS Info Length indicator
	// Low nibble (bits 3-0): Number of satellites
	gpsInfoByte := content[offset]
	satellites := gpsInfoByte & 0x0F // Low nibble = satellites
	offset++

	latBytes := content[offset : offset+4]
	offset += 4
	lonBytes := content[offset : offset+4]
	offset += 4

	speed := content[offset]
	offset++

	courseStatus, err := types.NewCourseStatusFromBytes(content[offset : offset+2])
	if err != nil {
		return nil, fmt.Errorf("alarm_4g: failed to parse course status: %w", err)
	}
	offset += 2

	coords, err := types.NewCoordinatesFromBytes(
		latBytes, lonBytes,
		courseStatus.IsNorthLatitude,
		courseStatus.IsEastLongitude,
	)
	if err != nil {
		return nil, fmt.Errorf("alarm_4g: failed to parse coordinates: %w", err)
	}

	// === Sequentially parse the rest of the packet ===
	var lbsInfo types.LBSInfo
	var mccmnc uint32
	var extendedLBS []types.LBSInfo
	var terminalInfo types.TerminalInfo
	var voltageLevel protocol.VoltageLevel
	var gsmSignal protocol.GSMSignalStrength
	var alarmType protocol.AlarmType
	var language protocol.Language
	var mileage uint32

	// LBS Length and Data (variable)
	if offset < len(content) {
		lbsLength := content[offset]
		offset++
		if lbsLength > 1 {
			// lbsLength includes the length byte itself.
			// We need to parse lbsLength-1 bytes of LBS data.
			lbsData := content[offset : offset+int(lbsLength)-1]
			var errLbs error
			lbsInfo, _, errLbs = types.NewLBSInfoFromBytes(lbsData, true) // 4G
			if errLbs == nil {
				mccmnc = uint32(lbsInfo.MCC)*1000 + uint32(lbsInfo.MNC)
			}
			offset += int(lbsLength) - 1
		}
	}

	// The remaining fields should be the status block
	if offset < len(content) {
		terminalInfo = types.NewTerminalInfo(content[offset])
		offset++
	}
	if offset < len(content) {
		voltageLevel = protocol.VoltageLevel(content[offset])
		offset++
	}
	if offset < len(content) {
		gsmSignal = protocol.GSMSignalStrength(content[offset])
		offset++
	}
	if offset < len(content) {
		// Doc says Alert and Language is 2 bytes, but parser for 2G used 1 byte each
		// Let's assume 1 byte for alarm type
		alarmType = protocol.AlarmType(content[offset])
		offset++
	}
	if offset < len(content) {
		language = protocol.Language(content[offset])
		offset++
	}

	// Fence ID (for 0xA4, similar to 0x27)
	var fenceID uint8
	if offset < len(content) {
		fenceID = content[offset]
		offset++
	}

	if offset+4 <= len(content) {
		mileage = uint32(content[offset])<<24 | uint32(content[offset+1])<<16 | uint32(content[offset+2])<<8 | uint32(content[offset+3])
	}

	// Build the packet
	serialNum, _ := ExtractSerialNumber(data)
	pkt := &packet.Alarm4GPacket{
		AlarmPacket: packet.AlarmPacket{
			BasePacket: packet.BasePacket{
				ProtocolNum: protocol.ProtocolAlarmMultiFence4G,
				SerialNum:   serialNum,
				RawData:     data,
				ParsedAt:    time.Now(),
			},
			DateTime:     dt,
			Satellites:   satellites,
			Coordinates:  coords,
			Speed:        speed,
			CourseStatus: courseStatus,
			LBSInfo:      lbsInfo,
			TerminalInfo: terminalInfo,
			VoltageLevel: voltageLevel,
			GSMSignal:    gsmSignal,
			AlarmType:    alarmType,
			Language:     language,
			Mileage:      mileage,
		},
		MCCMNC:      mccmnc,
		ExtendedLBS: extendedLBS,
		FenceID:     fenceID,
	}

	return pkt, nil
}

// init registers alarm parsers with the default registry
func init() {
	MustRegister(NewAlarmParser())
	MustRegister(NewAlarmMultiFenceParser())
	MustRegister(NewAlarm4GParser())
}
