package parser

import (
	"encoding/hex"
	"testing"

	"github.com/intelcon-group/jimi-vl103m/pkg/jimi/packet"
	"github.com/intelcon-group/jimi-vl103m/pkg/jimi/protocol"
)

func TestAlarmParser_ProtocolNumber(t *testing.T) {
	p := NewAlarmParser()
	if p.ProtocolNumber() != protocol.ProtocolAlarm {
		t.Errorf("Expected protocol 0x%02X, got 0x%02X", protocol.ProtocolAlarm, p.ProtocolNumber())
	}
}

func TestAlarmMultiFenceParser_ProtocolNumber(t *testing.T) {
	p := NewAlarmMultiFenceParser()
	if p.ProtocolNumber() != protocol.ProtocolAlarmMultiFence {
		t.Errorf("Expected protocol 0x%02X, got 0x%02X", protocol.ProtocolAlarmMultiFence, p.ProtocolNumber())
	}
}

func TestAlarm4GParser_ProtocolNumber(t *testing.T) {
	p := NewAlarm4GParser()
	if p.ProtocolNumber() != protocol.ProtocolAlarmMultiFence4G {
		t.Errorf("Expected protocol 0x%02X, got 0x%02X", protocol.ProtocolAlarmMultiFence4G, p.ProtocolNumber())
	}
}

func TestAlarmParser_Name(t *testing.T) {
	tests := []struct {
		parser Parser
		name   string
	}{
		{NewAlarmParser(), "Alarm"},
		{NewAlarmMultiFenceParser(), "Alarm Multi-Fence"},
		{NewAlarm4GParser(), "Alarm 4G"},
	}

	for _, tt := range tests {
		if tt.parser.Name() != tt.name {
			t.Errorf("Expected name '%s', got '%s'", tt.name, tt.parser.Name())
		}
	}
}

func TestAlarmParser_Parse(t *testing.T) {
	tests := []struct {
		name          string
		hex           string
		wantAlarmType protocol.AlarmType
		wantCritical  bool
		wantErr       bool
	}{
		{
			name:          "SOS alarm",
			hex:           "787825260F0C1D030B26C9027AC8180C4658600004000901CC00287D001F718004040102000C472A0D0A",
			wantAlarmType: protocol.AlarmSOS,
			wantCritical:  true,
			wantErr:       false,
		},
		{
			name:          "Power cut alarm",
			hex:           "787825260F0C1D030B26C9027AC8180C4658600004000901CC00287D001F718004040202000C472A0D0A",
			wantAlarmType: protocol.AlarmPowerCut,
			wantCritical:  true,
			wantErr:       false,
		},
		{
			name:          "Vibration alarm",
			hex:           "787825260F0C1D030B26C9027AC8180C4658600004000901CC00287D001F718004040302000C472A0D0A",
			wantAlarmType: protocol.AlarmVibration,
			wantCritical:  false,
			wantErr:       false,
		},
		{
			name:          "Geofence enter",
			hex:           "787825260F0C1D030B26C9027AC8180C4658600004000901CC00287D001F718004040402000C472A0D0A",
			wantAlarmType: protocol.AlarmGeofenceEnter,
			wantCritical:  false,
			wantErr:       false,
		},
		{
			name:          "Speed alarm",
			hex:           "787825260F0C1D030B26C9027AC8180C4658600004000901CC00287D001F718004040602000C472A0D0A",
			wantAlarmType: protocol.AlarmSpeed,
			wantCritical:  false,
			wantErr:       false,
		},
		{
			name:    "packet too short",
			hex:     "787805260001ABCD0D0A",
			wantErr: true,
		},
	}

	p := NewAlarmParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := hex.DecodeString(tt.hex)
			if err != nil {
				t.Fatalf("Failed to decode hex: %v", err)
			}

			pkt, err := p.Parse(data)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			alarmPkt, ok := pkt.(*packet.AlarmPacket)
			if !ok {
				t.Fatalf("Expected *AlarmPacket, got %T", pkt)
			}

			if alarmPkt.AlarmType != tt.wantAlarmType {
				t.Errorf("AlarmType mismatch: expected %s, got %s", tt.wantAlarmType, alarmPkt.AlarmType)
			}

			if alarmPkt.IsCritical() != tt.wantCritical {
				t.Errorf("IsCritical mismatch: expected %v, got %v", tt.wantCritical, alarmPkt.IsCritical())
			}

			if alarmPkt.ProtocolNumber() != protocol.ProtocolAlarm {
				t.Errorf("Protocol mismatch: expected 0x%02X, got 0x%02X",
					protocol.ProtocolAlarm, alarmPkt.ProtocolNumber())
			}
		})
	}
}

func TestAlarmType_Criticality(t *testing.T) {
	criticalAlarms := []protocol.AlarmType{
		protocol.AlarmSOS,
		protocol.AlarmPowerCut,
		protocol.AlarmTowTheft,
		protocol.AlarmTamper,
		protocol.AlarmCollision,
	}

	nonCriticalAlarms := []protocol.AlarmType{
		protocol.AlarmNormal,
		protocol.AlarmVibration,
		protocol.AlarmGeofenceEnter,
		protocol.AlarmGeofenceExit,
		protocol.AlarmSpeed,
		protocol.AlarmGPSBlindSpotEnter,
		protocol.AlarmGPSBlindSpotExit,
		protocol.AlarmPowerOn,
		protocol.AlarmPowerOff,
		protocol.AlarmACCOn,
		protocol.AlarmACCOff,
	}

	for _, alarm := range criticalAlarms {
		t.Run(alarm.String()+"_critical", func(t *testing.T) {
			if !alarm.IsCritical() {
				t.Errorf("Alarm %s should be critical", alarm)
			}
		})
	}

	for _, alarm := range nonCriticalAlarms {
		t.Run(alarm.String()+"_non_critical", func(t *testing.T) {
			if alarm.IsCritical() {
				t.Errorf("Alarm %s should NOT be critical", alarm)
			}
		})
	}
}

func TestAlarm4GParser_RealDevicePacket(t *testing.T) {
	// Real device packet from user showing "PowerCut" but byte 49 should not have Bit 7 set
	raw := "78782DA41A02010D2337CB01C3AF8807AC8B080018E41002CC100000C60D0000000002CE68CC4906040300FF001A3C160D0A"
	data, _ := hex.DecodeString(raw)

	parser := NewAlarm4GParser()
	pkt, err := parser.Parse(data)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	alarmPkt := pkt.(*packet.Alarm4GPacket)

	t.Logf("Raw TerminalInfo byte: 0x%02X (%08b)", alarmPkt.TerminalInfo.Raw(), alarmPkt.TerminalInfo.Raw())
	t.Logf("Bit 7 (OilElectricityDisconnected): %v", alarmPkt.TerminalInfo.OilElectricityDisconnected())
	t.Logf("Bit 1 (ACCOn): %v", alarmPkt.TerminalInfo.ACCOn())
	t.Logf("Bit 0 (IsArmed): %v", alarmPkt.TerminalInfo.IsArmed())
	t.Logf("Alarm bits (3-5): %03b", alarmPkt.TerminalInfo.AlarmTypeBits())
	t.Logf("AlarmType from packet: %s (0x%02X)", alarmPkt.AlarmType.String(), alarmPkt.AlarmType)
	t.Logf("TerminalInfo.String(): %s", alarmPkt.TerminalInfo.String())

	// Verify the byte is actually 49
	if alarmPkt.TerminalInfo.Raw() != 0x49 {
		t.Errorf("Expected TerminalInfo byte 0x49, got 0x%02X", alarmPkt.TerminalInfo.Raw())
	}

	// If byte is 49 (01001001), Bit 7 should be 0, so OilElectricityDisconnected should be false
	if alarmPkt.TerminalInfo.OilElectricityDisconnected() {
		t.Error("OilElectricityDisconnected should be FALSE for byte 0x49 (Bit 7 = 0)")
	}

	// Alarm type should be Vibration (0x03) from the alarm byte
	if alarmPkt.AlarmType != protocol.AlarmVibration {
		t.Errorf("Expected AlarmType Vibration (0x03), got %s (0x%02X)", alarmPkt.AlarmType.String(), alarmPkt.AlarmType)
	}
}

func TestAlarmType_String(t *testing.T) {
	tests := []struct {
		alarm protocol.AlarmType
		name  string
	}{
		{protocol.AlarmSOS, "SOS"},
		{protocol.AlarmPowerCut, "Power Cut"},
		{protocol.AlarmVibration, "Vibration"},
		{protocol.AlarmGeofenceEnter, "Geofence Enter"},
		{protocol.AlarmGeofenceExit, "Geofence Exit"},
		{protocol.AlarmSpeed, "Speed"},
		{protocol.AlarmTowTheft, "Tow/Theft"},
		{protocol.AlarmACCOn, "ACC On"},
		{protocol.AlarmACCOff, "ACC Off"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.alarm.String() != tt.name {
				t.Errorf("Expected '%s', got '%s'", tt.name, tt.alarm.String())
			}
		})
	}
}

func TestAlarmParser_Location(t *testing.T) {
	p := NewAlarmParser()

	// Parse a valid alarm packet
	data, _ := hex.DecodeString("787825260F0C1D030B26C9027AC8180C4658600004000901CC00287D001F718004040102000C472A0D0A")

	pkt, err := p.Parse(data)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	alarmPkt := pkt.(*packet.AlarmPacket)

	// Check coordinates
	lat := alarmPkt.Latitude()
	lon := alarmPkt.Longitude()

	if lat < -90 || lat > 90 {
		t.Errorf("Invalid latitude: %f", lat)
	}

	if lon < -180 || lon > 180 {
		t.Errorf("Invalid longitude: %f", lon)
	}

	t.Logf("Alarm location: (%.6f, %.6f)", lat, lon)
}
