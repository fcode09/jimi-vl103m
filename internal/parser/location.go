package parser

import (
	"fmt"
	"time"

	"github.com/fcode09/jimi-vl103m/pkg/jimi/packet"
	"github.com/fcode09/jimi-vl103m/pkg/jimi/protocol"
	"github.com/fcode09/jimi-vl103m/pkg/jimi/types"
)

// LocationParser parses GPS location packets (Protocol 0x22)
type LocationParser struct {
	BaseParser
}

// NewLocationParser creates a new location parser
func NewLocationParser() *LocationParser {
	return &LocationParser{
		BaseParser: NewBaseParser(protocol.ProtocolGPSLocation, "GPS Location"),
	}
}

// Parse implements Parser interface
// Location packet content structure (Protocol 0x22):
// - DateTime: 6 bytes (YY MM DD HH MM SS)
// - GPS Info Length: 1 byte (high nibble: satellites, low nibble: GPS data length / 2)
// - Latitude: 4 bytes (raw value / 1800000 = decimal degrees)
// - Longitude: 4 bytes (raw value / 1800000 = decimal degrees)
// - Speed: 1 byte (km/h)
// - Course/Status: 2 bytes
// - MCC: 2 bytes (Mobile Country Code)
// - MNC: 1 byte (Mobile Network Code)
// - LAC: 2 bytes (Location Area Code)
// - CellID: 3 bytes (Cell Tower ID)
// - ACC: 1 byte (0x00=OFF, 0x01=ON)
// - Data Upload Mode: 1 byte
// - GPS Data Re-upload: 1 byte (0x00=Real-time, 0x01=Re-upload)
// - Mileage Statistics: 4 bytes
// Total content: 33 bytes minimum (all fields are MANDATORY per protocol spec)
func (p *LocationParser) Parse(data []byte) (packet.Packet, error) {
	content, err := ExtractContent(data)
	if err != nil {
		return nil, fmt.Errorf("location: %w", err)
	}

	if len(content) < 29 {
		return nil, fmt.Errorf("location: content too short: %d bytes (need at least 29)", len(content))
	}

	offset := 0

	// Parse DateTime (6 bytes)
	dt, err := types.DateTimeFromBytes(content[offset : offset+6])
	if err != nil {
		return nil, fmt.Errorf("location: failed to parse datetime: %w", err)
	}
	offset += 6

	// Parse GPS Info Length (1 byte)
	gpsInfoByte := content[offset]
	satellites := (gpsInfoByte >> 4) & 0x0F
	// gpsDataLength := (gpsInfoByte & 0x0F) * 2 // Not used directly
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
		return nil, fmt.Errorf("location: failed to parse course status: %w", err)
	}
	offset += 2

	// Create coordinates with hemisphere info from course status
	coords, err := types.NewCoordinatesFromBytes(
		latBytes, lonBytes,
		courseStatus.IsNorthLatitude,
		courseStatus.IsEastLongitude,
	)
	if err != nil {
		return nil, fmt.Errorf("location: failed to parse coordinates: %w", err)
	}

	// Parse LBS Info (8 bytes: MCC(2) + MNC(1) + LAC(2) + CellID(3))
	lbsInfo, lbsConsumed, err := types.NewLBSInfoFromBytes(content[offset:offset+8], false) // false = 2G/3G
	if err != nil {
		// LBS parse error is non-fatal, continue with empty LBS
		lbsInfo = types.LBSInfo{}
		lbsConsumed = 8 // Still consume 8 bytes to maintain offset alignment
	}
	offset += lbsConsumed

	// Parse ACC (1 byte) - Protocol 0x22 uses simple boolean ACC field
	// According to protocol spec: 0x00=ACC off, 0x01=ACC on
	// Convert to TerminalInfo format by setting bit 1 (ACC bit)
	accByte := content[offset]
	var terminalInfo types.TerminalInfo
	if accByte == 0x01 {
		terminalInfo = types.NewTerminalInfoBuilder().SetACCOn(true).Build()
	} else {
		terminalInfo = types.NewTerminalInfoBuilder().SetACCOn(false).Build()
	}
	offset++

	// Parse Data Upload Mode (1 byte)
	uploadMode := protocol.UploadMode(content[offset])
	offset++

	// Parse GPS Data Re-upload (1 byte)
	isReupload := content[offset] == 0x01
	offset++

	// Parse Mileage Statistics (4 bytes) - Optional
	var mileage uint32
	if offset+4 <= len(content) {
		mileage = uint32(content[offset])<<24 | uint32(content[offset+1])<<16 |
			uint32(content[offset+2])<<8 | uint32(content[offset+3])
		offset += 4
	}

	// Extract serial number
	serialNum, _ := ExtractSerialNumber(data)

	pkt := &packet.LocationPacket{
		BasePacket: packet.BasePacket{
			ProtocolNum: protocol.ProtocolGPSLocation,
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
		HasStatus:    true,
		TerminalInfo: terminalInfo,
		UploadMode:   uploadMode,
		IsReupload:   isReupload,
		Mileage:      mileage,
	}

	return pkt, nil
}

// Location4GParser parses 4G GPS location packets (Protocol 0xA0)
type Location4GParser struct {
	BaseParser
}

// NewLocation4GParser creates a new 4G location parser
func NewLocation4GParser() *Location4GParser {
	return &Location4GParser{
		BaseParser: NewBaseParser(protocol.ProtocolGPSLocation4G, "GPS Location 4G"),
	}
}

func (p *Location4GParser) Parse(data []byte) (packet.Packet, error) {
	content, err := ExtractContent(data)
	if err != nil {
		return nil, fmt.Errorf("location_4g: %w", err)
	}

	// Minimum: DateTime(6) + GPSInfo(1) + Lat(4) + Lon(4) + Speed(1) + Course(2) +
	// 4G LBS (variable, min 15) + ACC(1) + UploadMode(1) + Reupload(1) + Mileage(4)
	// = approx 40 bytes minimum
	if len(content) < 30 {
		return nil, fmt.Errorf("location_4g: content too short: %d bytes (need at least 30)", len(content))
	}

	offset := 0

	// Parse DateTime (6 bytes)
	dt, err := types.DateTimeFromBytes(content[offset : offset+6])
	if err != nil {
		return nil, fmt.Errorf("location_4g: failed to parse datetime: %w", err)
	}
	offset += 6

	// Parse GPS Info Length (1 byte)
	gpsInfoByte := content[offset]
	satellites := (gpsInfoByte >> 4) & 0x0F
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
		return nil, fmt.Errorf("location_4g: failed to parse course status: %w", err)
	}
	offset += 2

	// Create coordinates with hemisphere info from course status
	coords, err := types.NewCoordinatesFromBytes(
		latBytes, lonBytes,
		courseStatus.IsNorthLatitude,
		courseStatus.IsEastLongitude,
	)
	if err != nil {
		return nil, fmt.Errorf("location_4g: failed to parse coordinates: %w", err)
	}

	// Parse 4G LBS Info (variable length: 15-16 bytes typically)
	// MCC(2) + MNC(1-2 depending on bit15 of MCC) + LAC(4) + CellID(8)
	lbsInfo, lbsConsumed, err := types.NewLBSInfoFromBytes(content[offset:], true) // true = 4G
	if err != nil {
		// If LBS parsing fails, try to estimate consumed bytes
		lbsInfo = types.LBSInfo{}
		lbsConsumed = 15 // Minimum 4G LBS size
	}
	mccmnc := uint32(lbsInfo.MCC)*1000 + uint32(lbsInfo.MNC)
	offset += lbsConsumed

	// Parse ACC (1 byte) - Protocol 0xA0 uses simple boolean ACC field
	// According to protocol spec: 0x00=ACC off, 0x01=ACC on
	// Convert to TerminalInfo format by setting bit 1 (ACC bit)
	accByte := content[offset]
	var terminalInfo types.TerminalInfo
	if accByte == 0x01 {
		terminalInfo = types.NewTerminalInfoBuilder().SetACCOn(true).Build()
	} else {
		terminalInfo = types.NewTerminalInfoBuilder().SetACCOn(false).Build()
	}
	offset++

	// Parse Data Upload Mode (1 byte)
	uploadMode := protocol.UploadMode(content[offset])
	offset++

	// Parse GPS Data Re-upload (1 byte)
	isReupload := content[offset] == 0x01
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

	pkt := &packet.Location4GPacket{
		LocationPacket: packet.LocationPacket{
			BasePacket: packet.BasePacket{
				ProtocolNum: protocol.ProtocolGPSLocation4G,
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
			HasStatus:    true,
			TerminalInfo: terminalInfo,
			UploadMode:   uploadMode,
			IsReupload:   isReupload,
			Mileage:      mileage,
		},
		MCCMNC: mccmnc,
	}

	return pkt, nil
}

// init registers location parsers with the default registry
func init() {
	MustRegister(NewLocationParser())
	MustRegister(NewLocation4GParser())
}
