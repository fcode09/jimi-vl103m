package packet

import (
	"fmt"
	"strings"
	"time"

	"github.com/intelcon-group/jimi-vl103m/pkg/jimi/protocol"
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

	// IMEI (when SubProtocol = InfoTypeICCID)
	IMEI string

	// IMSI (when SubProtocol = InfoTypeICCID)
	IMSI string

	// GPSModuleStatus (when SubProtocol = InfoTypeGPSStatus)
	GPSStatus protocol.GPSModuleStatus

	// TerminalSync contains parsed terminal sync data (when SubProtocol = InfoTypeTerminalSync)
	TerminalSync *TerminalSyncData

	// DoorStatus (when SubProtocol = InfoTypeDoorStatus)
	DoorStatus *DoorStatusData

	// GPSStatusInfo contains detailed GPS status (when SubProtocol = InfoTypeGPSStatus)
	GPSStatusInfo *GPSStatusData
}

// TerminalSyncData contains parsed terminal synchronization information
type TerminalSyncData struct {
	// Raw string data
	RawString string

	// Alarm configuration bytes (hex values)
	ALM1 string // Alarm byte 1
	ALM2 string // Alarm byte 2
	ALM3 string // Alarm byte 3
	ALM4 string // Alarm byte 4

	// Status byte
	STA1 string // Status byte 1

	// Fuel/power cutoff status
	DYD string

	// SOS numbers (comma separated)
	SOSNumbers []string

	// Center number
	CenterNumber string

	// Geofences
	Geofences []GeofenceConfig

	// Mode settings
	Mode string

	// IMSI from sync data
	IMSI string

	// ICCID from sync data
	ICCID string
}

// GeofenceConfig represents a geofence configuration
type GeofenceConfig struct {
	ID        int
	Enabled   bool
	Shape     int // 0=circle, 1=polygon
	Latitude  float64
	Longitude float64
	Radius    int    // meters
	Direction string // "IN", "OUT", "IN or OUT"
	AlarmType int
}

// DoorStatusData contains door status information
type DoorStatusData struct {
	DoorOpen    bool // bit0: 1=ON (open), 0=OFF (closed)
	TriggerHigh bool // bit1: 1=Level high, 0=Level low
	IOPortHigh  bool // bit2: 1=High, 0=Low
}

// GPSStatusData contains detailed GPS module status information
type GPSStatusData struct {
	ModuleStatus          protocol.GPSModuleStatus
	SatellitesInFix       int
	SatelliteStrengths    []int // Signal strength of satellites in fix
	VisibleSatellites     int
	VisibleStrengths      []int // Signal strength of visible satellites
	BDSModuleStatus       protocol.GPSModuleStatus
	BDSSatellitesInFix    int
	BDSSatelliteStrengths []int
	BDSVisibleSatellites  int
	BDSVisibleStrengths   []int
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
		return fmt.Sprintf("InfoTransferPacket{Type: External Voltage, Value: %d mV (%.2f V)}", p.ExternalVoltage, p.GetExternalVoltageVolts())
	case protocol.InfoTypeICCID:
		return fmt.Sprintf("InfoTransferPacket{Type: ICCID, ICCID: %s, IMEI: %s, IMSI: %s}", p.ICCID, p.IMEI, p.IMSI)
	case protocol.InfoTypeGPSStatus:
		return fmt.Sprintf("InfoTransferPacket{Type: GPS Status, Value: %s}", p.GPSStatus)
	case protocol.InfoTypeTerminalSync:
		if p.TerminalSync != nil {
			return fmt.Sprintf("InfoTransferPacket{Type: Terminal Sync, ICCID: %s, IMSI: %s, Center: %s, SOS: %v}",
				p.TerminalSync.ICCID, p.TerminalSync.IMSI, p.TerminalSync.CenterNumber, p.TerminalSync.SOSNumbers)
		}
		return fmt.Sprintf("InfoTransferPacket{Type: Terminal Sync, DataLen: %d}", len(p.Data))
	case protocol.InfoTypeDoorStatus:
		if p.DoorStatus != nil {
			return fmt.Sprintf("InfoTransferPacket{Type: Door Status, Open: %v, TriggerHigh: %v, IOHigh: %v}",
				p.DoorStatus.DoorOpen, p.DoorStatus.TriggerHigh, p.DoorStatus.IOPortHigh)
		}
		return fmt.Sprintf("InfoTransferPacket{Type: Door Status, DataLen: %d}", len(p.Data))
	default:
		return fmt.Sprintf("InfoTransferPacket{Type: %s, DataLen: %d}", p.SubProtocol, len(p.Data))
	}
}

// GetExternalVoltageVolts returns the external voltage in volts
func (p *InfoTransferPacket) GetExternalVoltageVolts() float64 {
	return float64(p.ExternalVoltage) / 100.0 // Note: protocol says divide by 100, not 1000
}

// GetDataAsString returns the data as ASCII string (useful for Terminal Sync)
func (p *InfoTransferPacket) GetDataAsString() string {
	return string(p.Data)
}

// HasTerminalSync returns true if terminal sync data is available
func (p *InfoTransferPacket) HasTerminalSync() bool {
	return p.SubProtocol == protocol.InfoTypeTerminalSync && p.TerminalSync != nil
}

// HasDoorStatus returns true if door status data is available
func (p *InfoTransferPacket) HasDoorStatus() bool {
	return p.SubProtocol == protocol.InfoTypeDoorStatus && p.DoorStatus != nil
}

// HasGPSStatusInfo returns true if detailed GPS status is available
func (p *InfoTransferPacket) HasGPSStatusInfo() bool {
	return p.SubProtocol == protocol.InfoTypeGPSStatus && p.GPSStatusInfo != nil
}

// ParseTerminalSyncString parses the raw terminal sync string into structured data
func ParseTerminalSyncString(data string) *TerminalSyncData {
	result := &TerminalSyncData{
		RawString:  data,
		SOSNumbers: []string{},
		Geofences:  []GeofenceConfig{},
	}

	// Parse key=value pairs separated by semicolons
	parts := strings.Split(data, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Handle GFENCE separately (uses comma as separator)
		if strings.HasPrefix(part, "GFENCE") || strings.HasPrefix(part, "FENCE") {
			fence := parseGeofenceConfig(part)
			if fence != nil {
				result.Geofences = append(result.Geofences, *fence)
			}
			continue
		}

		// Regular key=value parsing
		if idx := strings.Index(part, "="); idx > 0 {
			key := strings.TrimSpace(part[:idx])
			value := strings.TrimSpace(part[idx+1:])

			switch key {
			case "ALM1":
				result.ALM1 = value
			case "ALM2":
				result.ALM2 = value
			case "ALM3":
				result.ALM3 = value
			case "ALM4":
				result.ALM4 = value
			case "STA1":
				result.STA1 = value
			case "DYD":
				result.DYD = value
			case "SOS":
				// SOS numbers are comma-separated
				sosNums := strings.Split(value, ",")
				for _, num := range sosNums {
					num = strings.TrimSpace(num)
					if num != "" {
						result.SOSNumbers = append(result.SOSNumbers, num)
					}
				}
			case "CENTER":
				result.CenterNumber = value
			case "MODE":
				result.Mode = value
			case "IMSI":
				result.IMSI = value
			case "ICCID":
				result.ICCID = value
			}
		}
	}

	return result
}

// parseGeofenceConfig parses a geofence configuration string
// Format: GFENCE1,OFF,0,0.000000,0.000000,300,IN or OUT,1
func parseGeofenceConfig(s string) *GeofenceConfig {
	// Remove GFENCE or FENCE prefix and parse
	s = strings.TrimPrefix(s, "GFENCE")
	s = strings.TrimPrefix(s, "FENCE")

	parts := strings.Split(s, ",")
	if len(parts) < 8 {
		return nil
	}

	fence := &GeofenceConfig{}

	// Parse ID from first part (e.g., "1" from "GFENCE1,...")
	if len(parts[0]) > 0 {
		fmt.Sscanf(parts[0], "%d", &fence.ID)
	}

	// Parse enabled status
	fence.Enabled = strings.ToUpper(strings.TrimSpace(parts[1])) == "ON"

	// Parse shape
	fmt.Sscanf(parts[2], "%d", &fence.Shape)

	// Parse coordinates
	fmt.Sscanf(parts[3], "%f", &fence.Latitude)
	fmt.Sscanf(parts[4], "%f", &fence.Longitude)

	// Parse radius
	fmt.Sscanf(parts[5], "%d", &fence.Radius)

	// Parse direction
	fence.Direction = strings.TrimSpace(parts[6])

	// Parse alarm type
	if len(parts) > 7 {
		fmt.Sscanf(parts[7], "%d", &fence.AlarmType)
	}

	return fence
}

// ParseDoorStatusByte parses a door status byte
func ParseDoorStatusByte(b byte) *DoorStatusData {
	return &DoorStatusData{
		DoorOpen:    (b & 0x01) != 0,
		TriggerHigh: (b & 0x02) != 0,
		IOPortHigh:  (b & 0x04) != 0,
	}
}
