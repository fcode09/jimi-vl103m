package encoder

import (
	"fmt"
	"strings"
)

// Command constants for common device operations
// These are the AT-style commands supported by VL103M devices
const (
	// Device Information Commands
	CmdGetIMEI    = "IMEI#"    // Get device IMEI
	CmdGetVersion = "VERSION#" // Get firmware version
	CmdGetStatus  = "STATUS#"  // Get device status
	CmdGetParam   = "PARAM#"   // Get all parameters
	CmdGetICCID   = "ICCID#"   // Get SIM ICCID

	// Tracking Commands
	CmdSingleLocation = "WHERE#"   // Request single location
	CmdStartTracking  = "TRACK,1#" // Start continuous tracking
	CmdStopTracking   = "TRACK,0#" // Stop tracking

	// Configuration Commands
	CmdReboot       = "RESET#"   // Reboot device
	CmdFactoryReset = "FACTORY#" // Factory reset
	CmdRestore      = "RESTORE#" // Restore default settings

	// Power Management
	CmdPowerOff = "POWEROFF#" // Power off device
	CmdSleep    = "SLEEP#"    // Enter sleep mode
	CmdWake     = "WAKE#"     // Wake from sleep

	// Alarm Configuration
	CmdAlarmOn  = "ALARM,1#" // Enable alarms
	CmdAlarmOff = "ALARM,0#" // Disable alarms
)

// CommandBuilder helps construct device commands with parameters
type CommandBuilder struct {
	encoder    *Encoder
	serialNum  uint16
	serverFlag uint32
}

// NewCommandBuilder creates a new command builder
func (e *Encoder) NewCommandBuilder(serialNum uint16, serverFlag uint32) *CommandBuilder {
	return &CommandBuilder{
		encoder:    e,
		serialNum:  serialNum,
		serverFlag: serverFlag,
	}
}

// Send sends a raw command string
func (cb *CommandBuilder) Send(command string) []byte {
	return cb.encoder.OnlineCommand(cb.serialNum, cb.serverFlag, command)
}

// GetIMEI requests the device IMEI
func (cb *CommandBuilder) GetIMEI() []byte {
	return cb.Send(CmdGetIMEI)
}

// GetVersion requests firmware version
func (cb *CommandBuilder) GetVersion() []byte {
	return cb.Send(CmdGetVersion)
}

// GetStatus requests device status
func (cb *CommandBuilder) GetStatus() []byte {
	return cb.Send(CmdGetStatus)
}

// GetICCID requests SIM card ICCID
func (cb *CommandBuilder) GetICCID() []byte {
	return cb.Send(CmdGetICCID)
}

// RequestLocation requests current location
func (cb *CommandBuilder) RequestLocation() []byte {
	return cb.Send(CmdSingleLocation)
}

// StartTracking starts continuous location tracking
// interval: tracking interval in seconds (10-65535)
func (cb *CommandBuilder) StartTracking(interval int) []byte {
	if interval < 10 {
		interval = 10
	}
	if interval > 65535 {
		interval = 65535
	}
	return cb.Send(fmt.Sprintf("TRACK,%d#", interval))
}

// StopTracking stops continuous tracking
func (cb *CommandBuilder) StopTracking() []byte {
	return cb.Send(CmdStopTracking)
}

// Reboot reboots the device
func (cb *CommandBuilder) Reboot() []byte {
	return cb.Send(CmdReboot)
}

// FactoryReset performs a factory reset
func (cb *CommandBuilder) FactoryReset() []byte {
	return cb.Send(CmdFactoryReset)
}

// SetUploadInterval sets the GPS upload interval
// interval: upload interval in seconds
func (cb *CommandBuilder) SetUploadInterval(interval int) []byte {
	return cb.Send(fmt.Sprintf("TIMER,%d#", interval))
}

// SetAPN configures the APN settings
// apn: Access Point Name
// user: APN username (optional)
// pass: APN password (optional)
func (cb *CommandBuilder) SetAPN(apn, user, pass string) []byte {
	if user == "" && pass == "" {
		return cb.Send(fmt.Sprintf("APN,%s#", apn))
	}
	return cb.Send(fmt.Sprintf("APN,%s,%s,%s#", apn, user, pass))
}

// SetServer configures the server address
// ip: Server IP address or domain
// port: Server port
func (cb *CommandBuilder) SetServer(ip string, port int) []byte {
	return cb.Send(fmt.Sprintf("SERVER,1,%s,%d,0#", ip, port))
}

// SetTimezone sets the device timezone
// offset: timezone offset in hours (e.g., 8 for UTC+8, -5 for UTC-5)
func (cb *CommandBuilder) SetTimezone(offset int) []byte {
	sign := "E" // East
	if offset < 0 {
		sign = "W" // West
		offset = -offset
	}
	return cb.Send(fmt.Sprintf("GMT,%s,%d#", sign, offset))
}

// SetSpeedAlarm configures speed alarm
// speed: speed threshold in km/h (0 to disable)
func (cb *CommandBuilder) SetSpeedAlarm(speed int) []byte {
	if speed <= 0 {
		return cb.Send("SPEED,0#")
	}
	return cb.Send(fmt.Sprintf("SPEED,%d#", speed))
}

// SetGeofence configures a circular geofence
// index: geofence index (0-9)
// lat, lon: center coordinates
// radius: radius in meters
// enable: true to enable, false to disable
func (cb *CommandBuilder) SetGeofence(index int, lat, lon float64, radius int, enable bool) []byte {
	enableFlag := 0
	if enable {
		enableFlag = 1
	}
	return cb.Send(fmt.Sprintf("FENCE,%d,%d,%.6f,%.6f,%d#",
		index, enableFlag, lat, lon, radius))
}

// DeleteGeofence removes a geofence
func (cb *CommandBuilder) DeleteGeofence(index int) []byte {
	return cb.Send(fmt.Sprintf("FENCE,%d,0#", index))
}

// SetSOSNumbers configures SOS phone numbers
// numbers: up to 3 phone numbers
func (cb *CommandBuilder) SetSOSNumbers(numbers ...string) []byte {
	if len(numbers) == 0 {
		return cb.Send("SOS,#")
	}
	if len(numbers) > 3 {
		numbers = numbers[:3]
	}
	return cb.Send(fmt.Sprintf("SOS,%s#", strings.Join(numbers, ",")))
}

// SetCenterNumber sets the authorized center number
func (cb *CommandBuilder) SetCenterNumber(number string) []byte {
	return cb.Send(fmt.Sprintf("CENTER,%s#", number))
}

// EnableVibrationAlarm enables/disables vibration alarm
func (cb *CommandBuilder) EnableVibrationAlarm(enable bool) []byte {
	flag := 0
	if enable {
		flag = 1
	}
	return cb.Send(fmt.Sprintf("VIBRATION,%d#", flag))
}

// EnablePowerAlarm enables/disables power cut alarm
func (cb *CommandBuilder) EnablePowerAlarm(enable bool) []byte {
	flag := 0
	if enable {
		flag = 1
	}
	return cb.Send(fmt.Sprintf("POWERALARM,%d#", flag))
}

// SetLanguage sets the device language
// lang: "CN" for Chinese, "EN" for English
func (cb *CommandBuilder) SetLanguage(lang string) []byte {
	return cb.Send(fmt.Sprintf("LANG,%s#", strings.ToUpper(lang)))
}

// CutOil sends command to cut oil/electricity (immobilizer)
// WARNING: This should only be used when vehicle is stationary
func (cb *CommandBuilder) CutOil() []byte {
	return cb.Send("RELAY,1#")
}

// RestoreOil restores oil/electricity
func (cb *CommandBuilder) RestoreOil() []byte {
	return cb.Send("RELAY,0#")
}

// SetDefenseMode enables/disables defense (armed) mode
func (cb *CommandBuilder) SetDefenseMode(enable bool) []byte {
	flag := 0
	if enable {
		flag = 1
	}
	return cb.Send(fmt.Sprintf("DEFENSE,%d#", flag))
}
