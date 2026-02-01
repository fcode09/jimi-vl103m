package types

import (
	"fmt"
	"strings"

	"github.com/intelcon-group/jimi-vl103m/pkg/jimi/protocol"
)

// TerminalInfo represents the terminal information status byte
// This byte contains various device status flags packed as bits
type TerminalInfo struct {
	raw byte
}

// Terminal information bit positions (per JM-VL03 protocol documentation)
// Bits 3-5 contain alarm type, extracted separately
const (
	terminalBitDefense        = 0 // Bit 0: Defense/armed status (1=armed, 0=not armed)
	terminalBitACCStatus      = 1 // Bit 1: ACC status (1=on, 0=off)
	terminalBitCharging       = 2 // Bit 2: Charging status (1=charging with power, 0=without)
	terminalBitAlarm0         = 3 // Bit 3: Alarm type bit 0
	terminalBitAlarm1         = 4 // Bit 4: Alarm type bit 1
	terminalBitAlarm2         = 5 // Bit 5: Alarm type bit 2
	terminalBitGPSTracking    = 6 // Bit 6: GPS positioned/tracking (1=positioned, 0=not)
	terminalBitOilElectricity = 7 // Bit 7: Oil/Electricity cut (1=cut, 0=restore)
)

// NewTerminalInfo creates a new TerminalInfo from a raw byte
func NewTerminalInfo(b byte) TerminalInfo {
	return TerminalInfo{raw: b}
}

// TerminalInfoFromByte creates a TerminalInfo from a byte
func TerminalInfoFromByte(b byte) TerminalInfo {
	return TerminalInfo{raw: b}
}

// Raw returns the raw byte value
func (t TerminalInfo) Raw() byte {
	return t.raw
}

// OilElectricityDisconnected returns true if oil/electricity is disconnected
// This indicates power has been cut to the vehicle
func (t TerminalInfo) OilElectricityDisconnected() bool {
	return t.raw&(1<<terminalBitOilElectricity) != 0
}

// GPSTrackingEnabled returns true if GPS tracking is enabled
func (t TerminalInfo) GPSTrackingEnabled() bool {
	return t.raw&(1<<terminalBitGPSTracking) != 0
}

// IsCharging returns true if the device is charging
func (t TerminalInfo) IsCharging() bool {
	return t.raw&(1<<terminalBitCharging) != 0
}

// ACCOn returns true if ACC (accessory power) is on
// ACC is typically on when the vehicle ignition is in accessory or on position
func (t TerminalInfo) ACCOn() bool {
	return t.raw&(1<<terminalBitACCStatus) != 0
}

// IsArmed returns true if the device is in defense/armed mode
func (t TerminalInfo) IsArmed() bool {
	return t.raw&(1<<terminalBitDefense) != 0
}

// AlarmTypeBits returns the alarm type encoded in bits 3-5
// Values: 000=Normal, 001=Vibration, 010=PowerCut, 011=LowBattery, 100=SOS
func (t TerminalInfo) AlarmTypeBits() byte {
	return (t.raw >> 3) & 0x07 // Extract bits 3-5
}

// String returns a human-readable representation
func (t TerminalInfo) String() string {
	var parts []string

	if t.OilElectricityDisconnected() {
		parts = append(parts, "PowerCut")
	}
	if t.GPSTrackingEnabled() {
		parts = append(parts, "GPSTracking")
	}
	if t.IsCharging() {
		parts = append(parts, "Charging")
	}
	if t.ACCOn() {
		parts = append(parts, "ACC:ON")
	} else {
		parts = append(parts, "ACC:OFF")
	}
	if t.IsArmed() {
		parts = append(parts, "Armed")
	}

	if len(parts) == 0 {
		return "Normal"
	}
	return strings.Join(parts, ", ")
}

// TerminalInfoBuilder helps construct TerminalInfo values
type TerminalInfoBuilder struct {
	value byte
}

// NewTerminalInfoBuilder creates a new builder
func NewTerminalInfoBuilder() *TerminalInfoBuilder {
	return &TerminalInfoBuilder{}
}

// SetOilElectricityDisconnected sets the oil/electricity disconnected flag
func (b *TerminalInfoBuilder) SetOilElectricityDisconnected(v bool) *TerminalInfoBuilder {
	if v {
		b.value |= 1 << terminalBitOilElectricity
	} else {
		b.value &^= 1 << terminalBitOilElectricity
	}
	return b
}

// SetGPSTracking sets the GPS tracking flag
func (b *TerminalInfoBuilder) SetGPSTracking(v bool) *TerminalInfoBuilder {
	if v {
		b.value |= 1 << terminalBitGPSTracking
	} else {
		b.value &^= 1 << terminalBitGPSTracking
	}
	return b
}

// SetCharging sets the charging flag
func (b *TerminalInfoBuilder) SetCharging(v bool) *TerminalInfoBuilder {
	if v {
		b.value |= 1 << terminalBitCharging
	} else {
		b.value &^= 1 << terminalBitCharging
	}
	return b
}

// SetACCOn sets the ACC status
func (b *TerminalInfoBuilder) SetACCOn(v bool) *TerminalInfoBuilder {
	if v {
		b.value |= 1 << terminalBitACCStatus
	} else {
		b.value &^= 1 << terminalBitACCStatus
	}
	return b
}

// SetArmed sets the defense/armed status
func (b *TerminalInfoBuilder) SetArmed(v bool) *TerminalInfoBuilder {
	if v {
		b.value |= 1 << terminalBitDefense
	} else {
		b.value &^= 1 << terminalBitDefense
	}
	return b
}

// Build creates the TerminalInfo
func (b *TerminalInfoBuilder) Build() TerminalInfo {
	return TerminalInfo{raw: b.value}
}

// DeviceStatus represents extended device status information
// Some protocols include additional status bytes
type DeviceStatus struct {
	TerminalInfo    TerminalInfo
	VoltageLevel    protocol.VoltageLevel
	GSMSignal       protocol.GSMSignalStrength
	ExtendedStatus  byte // Additional status byte if present
	HasExtendedInfo bool
}

// DeviceStatusFromBytes parses device status from protocol bytes
// The number of bytes varies by protocol:
// - Basic: 2 bytes (terminal info + voltage)
// - Extended: 3+ bytes (+ GSM signal, etc.)
func DeviceStatusFromBytes(data []byte) (DeviceStatus, error) {
	if len(data) < 2 {
		return DeviceStatus{}, fmt.Errorf("device status requires at least 2 bytes, got %d", len(data))
	}

	status := DeviceStatus{
		TerminalInfo: NewTerminalInfo(data[0]),
		VoltageLevel: protocol.VoltageLevel(data[1]),
	}

	if len(data) >= 3 {
		status.GSMSignal = protocol.GSMSignalStrength(data[2])
		status.HasExtendedInfo = true
	}

	if len(data) >= 4 {
		status.ExtendedStatus = data[3]
	}

	return status, nil
}

// ToBytes serializes the device status to bytes
func (s DeviceStatus) ToBytes() []byte {
	result := []byte{
		s.TerminalInfo.Raw(),
		byte(s.VoltageLevel),
	}

	if s.HasExtendedInfo {
		result = append(result, byte(s.GSMSignal))
		if s.ExtendedStatus != 0 {
			result = append(result, s.ExtendedStatus)
		}
	}

	return result
}

// String returns a human-readable representation
func (s DeviceStatus) String() string {
	parts := []string{
		fmt.Sprintf("Terminal: %s", s.TerminalInfo),
		fmt.Sprintf("Voltage: %s", s.VoltageLevel),
	}

	if s.HasExtendedInfo {
		parts = append(parts, fmt.Sprintf("GSM: %s", s.GSMSignal))
	}

	return strings.Join(parts, ", ")
}
