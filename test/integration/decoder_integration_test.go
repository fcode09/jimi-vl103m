package integration

import (
	"encoding/hex"
	"testing"

	"github.com/fcode09/jimi-vl103m/internal/testdata/packets"
	"github.com/fcode09/jimi-vl103m/pkg/jimi"
	"github.com/fcode09/jimi-vl103m/pkg/jimi/packet"
	"github.com/fcode09/jimi-vl103m/pkg/jimi/protocol"
)

// TestDecodeAllValidPackets tests decoding of all valid test packets
func TestDecodeAllValidPackets(t *testing.T) {
	// Use lenient decoder for test packets (some may have incorrect CRC)
	decoder := jimi.NewDecoder(
		jimi.WithSkipCRC(),
		jimi.WithStrictMode(false),
		jimi.WithAllowUnknownProtocols(),
	)

	allPackets := packets.GetAllValidPackets()

	for _, tp := range allPackets {
		t.Run(tp.Name, func(t *testing.T) {
			data, err := hex.DecodeString(tp.Hex)
			if err != nil {
				t.Fatalf("Failed to decode hex: %v", err)
			}

			pkt, err := decoder.Decode(data)
			if tp.Valid {
				if err != nil {
					t.Errorf("Expected valid packet, got error: %v", err)
					return
				}

				if pkt.ProtocolNumber() != tp.Protocol {
					t.Errorf("Protocol mismatch: expected 0x%02X, got 0x%02X",
						tp.Protocol, pkt.ProtocolNumber())
				}

				t.Logf("Decoded: %s (protocol 0x%02X)", pkt.Type(), pkt.ProtocolNumber())
			} else {
				if err == nil {
					t.Errorf("Expected error for invalid packet, got success")
				}
			}
		})
	}
}

// TestDecodeInvalidPackets tests that invalid packets are rejected
func TestDecodeInvalidPackets(t *testing.T) {
	// Use strict decoder to catch all errors
	decoder := jimi.NewDecoder(jimi.WithStrictMode(true))

	for _, tp := range packets.InvalidPackets {
		t.Run(tp.Name, func(t *testing.T) {
			data, err := hex.DecodeString(tp.Hex)
			if err != nil {
				t.Fatalf("Failed to decode hex: %v", err)
			}

			_, err = decoder.Decode(data)
			if err == nil {
				t.Errorf("Expected error for invalid packet %s", tp.Name)
			} else {
				t.Logf("Correctly rejected: %s (%v)", tp.Name, err)
			}
		})
	}
}

// TestDecodeLoginPackets tests login packet decoding
func TestDecodeLoginPackets(t *testing.T) {
	decoder := jimi.NewDecoder(
		jimi.WithSkipCRC(),
		jimi.WithStrictMode(false),
	)

	for _, tp := range packets.LoginPackets {
		t.Run(tp.Name, func(t *testing.T) {
			data, err := hex.DecodeString(tp.Hex)
			if err != nil {
				t.Fatalf("Failed to decode hex: %v", err)
			}

			pkt, err := decoder.Decode(data)
			if err != nil {
				t.Fatalf("Decode error: %v", err)
			}

			loginPkt, ok := pkt.(*packet.LoginPacket)
			if !ok {
				t.Fatalf("Expected *LoginPacket, got %T", pkt)
			}

			// Verify IMEI is valid
			if !loginPkt.IMEI.IsValid() {
				t.Error("IMEI is not valid")
			}

			t.Logf("Login: IMEI=%s, ModelID=0x%04X, TZ=%s",
				loginPkt.IMEI, loginPkt.ModelID, loginPkt.Timezone)
		})
	}
}

// TestDecodeHeartbeatPackets tests heartbeat packet decoding
func TestDecodeHeartbeatPackets(t *testing.T) {
	decoder := jimi.NewDecoder(
		jimi.WithSkipCRC(),
		jimi.WithStrictMode(false),
	)

	for _, tp := range packets.HeartbeatPackets {
		t.Run(tp.Name, func(t *testing.T) {
			data, err := hex.DecodeString(tp.Hex)
			if err != nil {
				t.Fatalf("Failed to decode hex: %v", err)
			}

			pkt, err := decoder.Decode(data)
			if err != nil {
				t.Fatalf("Decode error: %v", err)
			}

			hbPkt, ok := pkt.(*packet.HeartbeatPacket)
			if !ok {
				t.Fatalf("Expected *HeartbeatPacket, got %T", pkt)
			}

			t.Logf("Heartbeat: Voltage=%s (%d%%), GSM=%s (%d bars), ACC=%v",
				hbPkt.VoltageLevel, hbPkt.BatteryPercentage(),
				hbPkt.GSMSignal, hbPkt.SignalBars(),
				hbPkt.ACCOn())
		})
	}
}

// TestDecodeLocationPackets tests location packet decoding
func TestDecodeLocationPackets(t *testing.T) {
	decoder := jimi.NewDecoder(
		jimi.WithSkipCRC(),
		jimi.WithStrictMode(false),
	)

	for _, tp := range packets.LocationPackets {
		t.Run(tp.Name, func(t *testing.T) {
			data, err := hex.DecodeString(tp.Hex)
			if err != nil {
				t.Fatalf("Failed to decode hex: %v", err)
			}

			pkt, err := decoder.Decode(data)
			if err != nil {
				t.Fatalf("Decode error: %v", err)
			}

			switch p := pkt.(type) {
			case *packet.LocationPacket:
				t.Logf("Location: Time=%s, Lat=%.6f, Lon=%.6f, Speed=%d, Heading=%d (%s)",
					p.DateTime, p.Latitude(), p.Longitude(), p.Speed, p.CourseStatus.Course,
					[]string{"N", "NE", "E", "SE", "S", "SW", "W", "NW"}[p.CourseStatus.Course/45%8])
				t.Logf("  Satellites=%d, Positioned=%v, LBS=MCC:%d MNC:%d LAC:%d CellID:%d",
					p.Satellites, p.IsPositioned(),
					p.LBSInfo.MCC, p.LBSInfo.MNC, p.LBSInfo.LAC, p.LBSInfo.CellID)
			case *packet.Location4GPacket:
				t.Logf("Location 4G: Time=%s, Lat=%.6f, Lon=%.6f, Speed=%d",
					p.DateTime, p.Latitude(), p.Longitude(), p.Speed)
				t.Logf("  Satellites=%d, Positioned=%v, MCCMNC:%d",
					p.Satellites, p.IsPositioned(), p.MCCMNC)
			default:
				t.Errorf("Expected *LocationPacket or *Location4GPacket, got %T", pkt)
			}
		})
	}
}

// TestDecodeAlarmPackets tests alarm packet decoding
func TestDecodeAlarmPackets(t *testing.T) {
	decoder := jimi.NewDecoder(
		jimi.WithSkipCRC(),
		jimi.WithStrictMode(false),
	)

	for _, tp := range packets.AlarmPackets {
		t.Run(tp.Name, func(t *testing.T) {
			data, err := hex.DecodeString(tp.Hex)
			if err != nil {
				t.Fatalf("Failed to decode hex: %v", err)
			}

			pkt, err := decoder.Decode(data)
			if err != nil {
				t.Fatalf("Decode error: %v", err)
			}

			switch p := pkt.(type) {
			case *packet.AlarmPacket:
				t.Logf("Alarm: Type=%s, Critical=%v, Time=%s",
					p.AlarmType, p.IsCritical(), p.DateTime)
				t.Logf("  Location: %.6f, %.6f", p.Latitude(), p.Longitude())
			case *packet.Alarm4GPacket:
				t.Logf("Alarm 4G: Type=%s, Critical=%v, Time=%s",
					p.AlarmType, p.IsCritical(), p.DateTime)
				t.Logf("  Location: %.6f, %.6f, MCCMNC:%d, FenceID:%d",
					p.Latitude(), p.Longitude(), p.MCCMNC, p.FenceID)
			default:
				t.Errorf("Expected *AlarmPacket or *Alarm4GPacket, got %T", pkt)
			}
		})
	}
}

// TestDecodeTimeCalibrationPackets tests time calibration packet decoding
func TestDecodeTimeCalibrationPackets(t *testing.T) {
	decoder := jimi.NewDecoder(
		jimi.WithSkipCRC(),
		jimi.WithStrictMode(false),
	)

	for _, tp := range packets.TimeCalibrationPackets {
		t.Run(tp.Name, func(t *testing.T) {
			data, err := hex.DecodeString(tp.Hex)
			if err != nil {
				t.Fatalf("Failed to decode hex: %v", err)
			}

			pkt, err := decoder.Decode(data)
			if err != nil {
				t.Fatalf("Decode error: %v", err)
			}

			tcPkt, ok := pkt.(*packet.TimeCalibrationPacket)
			if !ok {
				t.Fatalf("Expected *TimeCalibrationPacket, got %T", pkt)
			}

			t.Logf("Time Calibration Request: Serial=%d", tcPkt.SerialNumber())
		})
	}
}

// TestDecodeConcatenatedPackets tests decoding of concatenated packets
func TestDecodeConcatenatedPackets(t *testing.T) {
	decoder := jimi.NewDecoder(
		jimi.WithSkipCRC(),
		jimi.WithStrictMode(false),
		jimi.WithAllowUnknownProtocols(),
	)

	for _, cp := range packets.ConcatenatedPackets {
		t.Run(cp.Name, func(t *testing.T) {
			data, err := hex.DecodeString(cp.Hex)
			if err != nil {
				t.Fatalf("Failed to decode hex: %v", err)
			}

			pkts, residue, err := decoder.DecodeStream(data)
			if err != nil {
				t.Logf("DecodeStream warning: %v", err)
			}

			if len(pkts) != cp.PacketCount {
				t.Errorf("Expected %d packets, got %d", cp.PacketCount, len(pkts))
			}

			if len(residue) > 0 {
				t.Logf("Residue: %d bytes", len(residue))
			}

			for i, pkt := range pkts {
				t.Logf("Packet %d: %s (protocol 0x%02X)", i+1, pkt.Type(), pkt.ProtocolNumber())
			}
		})
	}
}

// TestAlarmCriticality tests that critical alarms are correctly identified
func TestAlarmCriticality(t *testing.T) {
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
		protocol.AlarmACCOn,
		protocol.AlarmACCOff,
	}

	for _, alarm := range criticalAlarms {
		if !alarm.IsCritical() {
			t.Errorf("Alarm %s should be critical", alarm)
		}
	}

	for _, alarm := range nonCriticalAlarms {
		if alarm.IsCritical() {
			t.Errorf("Alarm %s should not be critical", alarm)
		}
	}
}

// TestProtocolRegistration tests that all expected parsers are registered
func TestProtocolRegistration(t *testing.T) {
	decoder := jimi.NewDecoder()

	expectedProtocols := []byte{
		protocol.ProtocolLogin,
		protocol.ProtocolHeartbeat,
		protocol.ProtocolGPSLocation,
		protocol.ProtocolGPSLocation4G,
		protocol.ProtocolAlarm,
		protocol.ProtocolAlarmMultiFence,
		protocol.ProtocolAlarmMultiFence4G,
		protocol.ProtocolLBSMultiBase,
		protocol.ProtocolLBSMultiBase4G,
		protocol.ProtocolTimeCalibration,
		protocol.ProtocolInfoTransfer,
		protocol.ProtocolOnlineCommand,
		protocol.ProtocolCommandResponse,
		protocol.ProtocolCommandResponseOld,
		protocol.ProtocolGPSAddressRequest,
	}

	for _, proto := range expectedProtocols {
		if !decoder.HasParser(proto) {
			t.Errorf("Expected parser for protocol 0x%02X to be registered", proto)
		}
	}

	registeredCount := len(decoder.RegisteredProtocols())
	t.Logf("Total registered protocols: %d", registeredCount)

	if registeredCount < len(expectedProtocols) {
		t.Errorf("Expected at least %d protocols, got %d", len(expectedProtocols), registeredCount)
	}
}
