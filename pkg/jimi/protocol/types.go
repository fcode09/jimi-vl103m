package protocol

import "fmt"

// AlarmType represents the type of alarm triggered by the device
type AlarmType byte

// Alarm types as defined in JM-VL03 protocol specification
const (
	AlarmNormal                    AlarmType = 0x00 // Normal (no alarm)
	AlarmSOS                       AlarmType = 0x01 // SOS alert
	AlarmPowerCut                  AlarmType = 0x02 // Power cut alert
	AlarmVibration                 AlarmType = 0x03 // Vibrating alert
	AlarmGeofenceEnter             AlarmType = 0x04 // Entered fence alert
	AlarmGeofenceExit              AlarmType = 0x05 // Left fence alert
	AlarmSpeed                     AlarmType = 0x06 // Speed alert
	AlarmTowTheft                  AlarmType = 0x09 // Tow/theft alert
	AlarmGPSBlindSpotEnter         AlarmType = 0x0A // Entered GPS blind spot alert
	AlarmGPSBlindSpotExit          AlarmType = 0x0B // Left GPS blind spot alert
	AlarmPowerOn                   AlarmType = 0x0C // Powered on alert
	AlarmGPSFirstFix               AlarmType = 0x0D // GPS first fix alert
	AlarmExternalBatteryLow        AlarmType = 0x0E // Low external battery alert
	AlarmExternalBatteryProtection AlarmType = 0x0F // External battery low voltage protection alert
	AlarmSIMChanged                AlarmType = 0x10 // SIM changed alert
	AlarmPowerOff                  AlarmType = 0x11 // Powered off alert
	AlarmAirplaneMode              AlarmType = 0x12 // Airplane mode on following external battery low voltage protection
	AlarmTamper                    AlarmType = 0x13 // Tamper alert
	AlarmDoor                      AlarmType = 0x14 // Door alert
	AlarmPowerOffLowBattery        AlarmType = 0x15 // Powered off due to low battery
	AlarmSoundControl              AlarmType = 0x16 // Sound-control alert
	AlarmRogueBaseStation          AlarmType = 0x17 // Rogue base station detected alert
	AlarmCoverRemoved              AlarmType = 0x18 // Cover removed alert
	AlarmInternalBatteryLow        AlarmType = 0x19 // Low internal battery alert
	AlarmDeepSleep                 AlarmType = 0x20 // Entered deep sleep mode alert
	AlarmFall                      AlarmType = 0x23 // Fall alert
	AlarmHarshAcceleration         AlarmType = 0x29 // Harsh acceleration
	AlarmSharpLeftCorner           AlarmType = 0x2A // Sharp left cornering alert
	AlarmSharpRightCorner          AlarmType = 0x2B // Sharp right cornering alert
	AlarmCollision                 AlarmType = 0x2C // Collision alert
	AlarmHarshBraking              AlarmType = 0x30 // Harsh braking
	AlarmDeviceUnplugged           AlarmType = 0x32 // Device unplugged alert
	AlarmACCOff                    AlarmType = 0xFF // ACC OFF
	AlarmACCOn                     AlarmType = 0xFE // ACC ON
)

// String returns the human-readable name of the alarm type
func (a AlarmType) String() string {
	switch a {
	case AlarmNormal:
		return "Normal"
	case AlarmSOS:
		return "SOS"
	case AlarmPowerCut:
		return "Power Cut"
	case AlarmVibration:
		return "Vibration"
	case AlarmGeofenceEnter:
		return "Geofence Enter"
	case AlarmGeofenceExit:
		return "Geofence Exit"
	case AlarmSpeed:
		return "Speed"
	case AlarmTowTheft:
		return "Tow/Theft"
	case AlarmGPSBlindSpotEnter:
		return "GPS Blind Spot Enter"
	case AlarmGPSBlindSpotExit:
		return "GPS Blind Spot Exit"
	case AlarmPowerOn:
		return "Power On"
	case AlarmGPSFirstFix:
		return "GPS First Fix"
	case AlarmExternalBatteryLow:
		return "External Battery Low"
	case AlarmExternalBatteryProtection:
		return "External Battery Protection"
	case AlarmSIMChanged:
		return "SIM Changed"
	case AlarmPowerOff:
		return "Power Off"
	case AlarmAirplaneMode:
		return "Airplane Mode"
	case AlarmTamper:
		return "Tamper"
	case AlarmDoor:
		return "Door"
	case AlarmPowerOffLowBattery:
		return "Power Off Low Battery"
	case AlarmSoundControl:
		return "Sound Control"
	case AlarmRogueBaseStation:
		return "Rogue Base Station"
	case AlarmCoverRemoved:
		return "Cover Removed"
	case AlarmInternalBatteryLow:
		return "Internal Battery Low"
	case AlarmDeepSleep:
		return "Deep Sleep"
	case AlarmFall:
		return "Fall"
	case AlarmHarshAcceleration:
		return "Harsh Acceleration"
	case AlarmSharpLeftCorner:
		return "Sharp Left Corner"
	case AlarmSharpRightCorner:
		return "Sharp Right Corner"
	case AlarmCollision:
		return "Collision"
	case AlarmHarshBraking:
		return "Harsh Braking"
	case AlarmDeviceUnplugged:
		return "Device Unplugged"
	case AlarmACCOff:
		return "ACC Off"
	case AlarmACCOn:
		return "ACC On"
	default:
		return fmt.Sprintf("Unknown(0x%02X)", byte(a))
	}
}

// IsCritical returns true if the alarm is critical and requires immediate attention
func (a AlarmType) IsCritical() bool {
	switch a {
	case AlarmSOS, AlarmPowerCut, AlarmTowTheft, AlarmTamper, AlarmCollision:
		return true
	default:
		return false
	}
}

// VoltageLevel represents the battery voltage level
type VoltageLevel byte

const (
	VoltageNoPower       VoltageLevel = 0x00 // No power (shutdown)
	VoltageExtremelyLow  VoltageLevel = 0x01 // Extremely low (cannot make calls/SMS)
	VoltageVeryLow       VoltageLevel = 0x02 // Very low (will trigger low battery alert)
	VoltageLow           VoltageLevel = 0x03 // Low (normal use)
	VoltageMedium        VoltageLevel = 0x04 // Medium
	VoltageHigh          VoltageLevel = 0x05 // High
	VoltageExtremelyHigh VoltageLevel = 0x06 // Extremely high
)

// String returns the human-readable voltage level
func (v VoltageLevel) String() string {
	switch v {
	case VoltageNoPower:
		return "No Power"
	case VoltageExtremelyLow:
		return "Extremely Low"
	case VoltageVeryLow:
		return "Very Low"
	case VoltageLow:
		return "Low"
	case VoltageMedium:
		return "Medium"
	case VoltageHigh:
		return "High"
	case VoltageExtremelyHigh:
		return "Extremely High"
	default:
		return fmt.Sprintf("Unknown(0x%02X)", byte(v))
	}
}

// Percentage returns approximate battery percentage
func (v VoltageLevel) Percentage() int {
	switch v {
	case VoltageNoPower:
		return 0
	case VoltageExtremelyLow:
		return 5
	case VoltageVeryLow:
		return 15
	case VoltageLow:
		return 30
	case VoltageMedium:
		return 50
	case VoltageHigh:
		return 75
	case VoltageExtremelyHigh:
		return 100
	default:
		return 0
	}
}

// GSMSignalStrength represents the GSM signal strength
type GSMSignalStrength byte

const (
	SignalNone          GSMSignalStrength = 0x00 // No signal
	SignalExtremelyWeak GSMSignalStrength = 0x01 // Extremely weak
	SignalWeak          GSMSignalStrength = 0x02 // Weak
	SignalGood          GSMSignalStrength = 0x03 // Good
	SignalStrong        GSMSignalStrength = 0x04 // Strong
)

// String returns the human-readable signal strength
func (s GSMSignalStrength) String() string {
	switch s {
	case SignalNone:
		return "No Signal"
	case SignalExtremelyWeak:
		return "Extremely Weak"
	case SignalWeak:
		return "Weak"
	case SignalGood:
		return "Good"
	case SignalStrong:
		return "Strong"
	default:
		return fmt.Sprintf("Unknown(0x%02X)", byte(s))
	}
}

// Bars returns the signal strength as number of bars (0-4)
func (s GSMSignalStrength) Bars() int {
	if s > 4 {
		return 0
	}
	return int(s)
}

// UploadMode represents the reason for GPS data upload
type UploadMode byte

// String returns the human-readable upload mode
func (u UploadMode) String() string {
	switch u {
	case UploadModeInterval:
		return "Fixed Interval"
	case UploadModeDistance:
		return "Fixed Distance"
	case UploadModeTurningPoint:
		return "Turning Point"
	case UploadModeACCChange:
		return "ACC Status Change"
	case UploadModeStopToStill:
		return "Stop to Still"
	case UploadModeNetworkReconnect:
		return "Network Reconnect"
	case UploadModeRefreshEphem:
		return "Refresh Ephemeris"
	case UploadModeKeyPress:
		return "Key Press"
	case UploadModePowerOn:
		return "Power On"
	case UploadModeAfterStill:
		return "After Still"
	case UploadModeWiFi:
		return "WiFi"
	case UploadModeImmediate:
		return "Immediate (LJDW)"
	case UploadModeLastStill:
		return "Last Still"
	case UploadModeGPSDUP:
		return "GPSDUP (Still Interval)"
	case UploadModeExitTracking:
		return "Exit Tracking"
	default:
		return fmt.Sprintf("Unknown(0x%02X)", byte(u))
	}
}

// Language represents the language code
type Language byte

// String returns the human-readable language name
func (l Language) String() string {
	switch l {
	case LanguageChinese:
		return "Chinese"
	case LanguageEnglish:
		return "English"
	default:
		return fmt.Sprintf("Unknown(0x%02X)", byte(l))
	}
}

// InfoType represents the information transfer sub-protocol type
type InfoType byte

// String returns the human-readable information type
func (i InfoType) String() string {
	switch i {
	case InfoTypeExternalVoltage:
		return "External Voltage"
	case InfoTypeTerminalSync:
		return "Terminal Sync"
	case InfoTypeDoorStatus:
		return "Door Status"
	case InfoTypeSelfCheck:
		return "Self Check"
	case InfoTypeGPSStatus:
		return "GPS Status"
	case InfoTypeICCID:
		return "ICCID"
	case InfoTypeRFID:
		return "RFID"
	default:
		return fmt.Sprintf("Custom(0x%02X)", byte(i))
	}
}

// GPSModuleStatus represents the GPS module operational status
type GPSModuleStatus byte

const (
	GPSStatusNoFeature GPSModuleStatus = 0x00 // No feature
	GPSStatusSearching GPSModuleStatus = 0x01 // Searching
	GPSStatus2D        GPSModuleStatus = 0x02 // 2D fix
	GPSStatus3D        GPSModuleStatus = 0x03 // 3D fix
	GPSStatusSleeping  GPSModuleStatus = 0x04 // Sleeping
)

// String returns the human-readable GPS module status
func (g GPSModuleStatus) String() string {
	switch g {
	case GPSStatusNoFeature:
		return "No Feature"
	case GPSStatusSearching:
		return "Searching"
	case GPSStatus2D:
		return "2D Fix"
	case GPSStatus3D:
		return "3D Fix"
	case GPSStatusSleeping:
		return "Sleeping"
	default:
		return fmt.Sprintf("Unknown(0x%02X)", byte(g))
	}
}

// HasFix returns true if the GPS has a valid fix
func (g GPSModuleStatus) HasFix() bool {
	return g == GPSStatus2D || g == GPSStatus3D
}
