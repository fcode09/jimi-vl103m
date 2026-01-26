// Package packets contains test packet data for the VL103M protocol.
// These are example packets based on the protocol specification.
package packets

// TestPacket represents a test packet with metadata
type TestPacket struct {
	Name        string // Descriptive name
	Hex         string // Hex-encoded packet data
	Protocol    byte   // Expected protocol number
	Description string // What this packet represents
	Valid       bool   // Whether this packet should parse successfully
}

// LoginPackets contains example login packets
var LoginPackets = []TestPacket{
	{
		Name:        "login_basic",
		Hex:         "787811010359339073930530044D014E00015ED00D0A",
		Protocol:    0x01,
		Description: "Basic login packet with IMEI 035933907393052",
		Valid:       true,
	},
	{
		Name:        "login_timezone_east8",
		Hex:         "787811010123456789012378044D03200001ABCD0D0A",
		Protocol:    0x01,
		Description: "Login with UTC+8 timezone",
		Valid:       true,
	},
}

// HeartbeatPackets contains example heartbeat packets
var HeartbeatPackets = []TestPacket{
	{
		Name:        "heartbeat_normal",
		Hex:         "78780813040300010006950D0A",
		Protocol:    0x13,
		Description: "Normal heartbeat: voltage=medium, GSM=good, ACC off",
		Valid:       true,
	},
	{
		Name:        "heartbeat_acc_on",
		Hex:         "78780813240400010007A50D0A",
		Protocol:    0x13,
		Description: "Heartbeat with ACC on, voltage=high",
		Valid:       true,
	},
	{
		Name:        "heartbeat_charging",
		Hex:         "78780813140500010008B50D0A",
		Protocol:    0x13,
		Description: "Heartbeat while charging",
		Valid:       true,
	},
	{
		Name:        "heartbeat_extended",
		Hex:         "78780B130403001234000100C5D50D0A",
		Protocol:    0x13,
		Description: "Heartbeat with extended info",
		Valid:       true,
	},
}

// LocationPackets contains example GPS location packets
var LocationPackets = []TestPacket{
	{
		Name:        "location_basic",
		Hex:         "787822220F0C1D023305C9027AC8180C46586000140001CC00287D001F71000001000820860D0A",
		Protocol:    0x22,
		Description: "Basic GPS location packet",
		Valid:       true,
	},
	{
		Name:        "location_with_status",
		Hex:         "787822220F0C1D023305C9027AC8180C46586000140001CC00287D001F71010001000820860D0A",
		Protocol:    0x22,
		Description: "Location with terminal status",
		Valid:       true,
	},
	{
		Name:        "location_no_gps",
		Hex:         "787822220F0C1D02330509000000000000000000000001CC00287D001F71000001000820860D0A",
		Protocol:    0x22,
		Description: "Location packet without GPS fix",
		Valid:       true,
	},
	{
		Name:        "location_4g_basic",
		Hex:         "78782DA01A011A033305CA027AC8180C46586000001902CC100000C60D0000000002CE68EA00000000005B8C0001309E0D0A",
		Protocol:    0xA0,
		Description: "GPS Location 4G packet (Sanitized)",
		Valid:       true,
	},
}

// AlarmPackets contains example alarm packets
var AlarmPackets = []TestPacket{
	{
		Name:        "alarm_sos",
		Hex:         "787825260F0C1D030B26C9027AC8180C4658600004000901CC00287D001F718004040102000C472A0D0A",
		Protocol:    0x26,
		Description: "SOS alarm",
		Valid:       true,
	},
	{
		Name:        "alarm_power_cut",
		Hex:         "787825260F0C1D030B26C9027AC8180C4658600004000901CC00287D001F718004040202000C472A0D0A",
		Protocol:    0x26,
		Description: "Power cut alarm",
		Valid:       true,
	},
	{
		Name:        "alarm_vibration",
		Hex:         "787825260F0C1D030B26C9027AC8180C4658600004000901CC00287D001F718004040302000C472A0D0A",
		Protocol:    0x26,
		Description: "Vibration alarm",
		Valid:       true,
	},
	{
		Name:        "alarm_geofence_enter",
		Hex:         "787825260F0C1D030B26C9027AC8180C4658600004000901CC00287D001F718004040402000C472A0D0A",
		Protocol:    0x26,
		Description: "Geofence enter alarm",
		Valid:       true,
	},
	{
		Name:        "alarm_speed",
		Hex:         "787825260F0C1D030B26C9027AC8180C4658600004000901CC00287D001F718004040602000C472A0D0A",
		Protocol:    0x26,
		Description: "Speed alarm",
		Valid:       true,
	},
	{
		Name:        "alarm_acc_on",
		Hex:         "787825260F0C1D030B26C9027AC8180C4658600004000901CC00287D001F71800404FE02000C472A0D0A",
		Protocol:    0x26,
		Description: "ACC ON alarm",
		Valid:       true,
	},
	{
		Name:        "alarm_acc_off",
		Hex:         "787825260F0C1D030B26C9027AC8180C4658600004000901CC00287D001F71800404FF02000C472A0D0A",
		Protocol:    0x26,
		Description: "ACC OFF alarm",
		Valid:       true,
	},
	{
		Name:        "alarm_4g_basic",
		Hex:         "78782DA41A011A03121FCA01C3AF5407AC8D200519191002CC100000C60D0000000002CE68EA4106040600FF00726F810D0A",
		Protocol:    0xA4,
		Description: "Alarm 4G packet (Real Sanitized Structure)",
		Valid:       true,
	},
}

// LBSPackets contains example LBS packets
var LBSPackets = []TestPacket{
	{
		Name:        "lbs_2g",
		Hex:         "78781F281807151223100282F800289C8B4015A517001E0001CC00260101000143630D0A",
		Protocol:    0x28,
		Description: "2G LBS packet",
		Valid:       true,
	},
}

// TimeCalibrationPackets contains example time calibration packets
var TimeCalibrationPackets = []TestPacket{
	{
		Name:        "time_request",
		Hex:         "7878068A00010003870D0A",
		Protocol:    0x8A,
		Description: "Time calibration request",
		Valid:       true,
	},
}

// InfoTransferPackets contains example info transfer packets
var InfoTransferPackets = []TestPacket{
	{
		Name:        "info_external_voltage",
		Hex:         "7878099400002EE0000138AD0D0A",
		Protocol:    0x94,
		Description: "External voltage info (12V)",
		Valid:       true,
	},
	{
		Name:        "info_iccid",
		Hex:         "787815940289860112345678901234567890000100012A2D0D0A",
		Protocol:    0x94,
		Description: "ICCID info transfer",
		Valid:       true,
	},
}

// CommandPackets contains example command packets
var CommandPackets = []TestPacket{
	{
		Name:        "command_imei",
		Hex:         "79790010800B00000001494D454923000100B0950D0A",
		Protocol:    0x80,
		Description: "IMEI# command",
		Valid:       true,
	},
	{
		Name:        "command_response",
		Hex:         "797900192115000000013335393333393037333933303532000100B1A50D0A",
		Protocol:    0x21,
		Description: "Command response with IMEI",
		Valid:       true,
	},
}

// InvalidPackets contains packets that should fail to parse
var InvalidPackets = []TestPacket{
	{
		Name:        "invalid_start_bit",
		Hex:         "00001101035933907393052380044D014E00015ED00D0A",
		Protocol:    0x00,
		Description: "Invalid start bit",
		Valid:       false,
	},
	{
		Name:        "invalid_stop_bit",
		Hex:         "78781101035933907393052380044D014E00015ED00000",
		Protocol:    0x01,
		Description: "Invalid stop bit",
		Valid:       false,
	},
	{
		Name:        "too_short",
		Hex:         "7878050D0A",
		Protocol:    0x00,
		Description: "Packet too short",
		Valid:       false,
	},
	{
		Name:        "invalid_crc",
		Hex:         "78781101035933907393052380044D014E0001FFFF0D0A",
		Protocol:    0x01,
		Description: "Invalid CRC checksum",
		Valid:       false,
	},
}

// ConcatenatedPackets contains multiple packets concatenated together
var ConcatenatedPackets = []struct {
	Name        string
	Hex         string
	PacketCount int
	Description string
}{
	{
		Name:        "two_heartbeats",
		Hex:         "78780813040300010006950D0A78780813240400010007A50D0A",
		PacketCount: 2,
		Description: "Two heartbeat packets concatenated",
	},
	{
		Name:        "login_then_heartbeat",
		Hex:         "787811010359339073930523044D014E00015ED00D0A78780813040300010006950D0A",
		PacketCount: 2,
		Description: "Login followed by heartbeat",
	},
}

// GetAllValidPackets returns all valid test packets
func GetAllValidPackets() []TestPacket {
	var all []TestPacket
	all = append(all, LoginPackets...)
	all = append(all, HeartbeatPackets...)
	all = append(all, LocationPackets...)
	all = append(all, AlarmPackets...)
	all = append(all, LBSPackets...)
	all = append(all, TimeCalibrationPackets...)
	all = append(all, InfoTransferPackets...)
	all = append(all, CommandPackets...)
	return all
}

// GetPacketsByProtocol returns test packets for a specific protocol
func GetPacketsByProtocol(protocol byte) []TestPacket {
	var result []TestPacket
	for _, p := range GetAllValidPackets() {
		if p.Protocol == protocol {
			result = append(result, p)
		}
	}
	return result
}
