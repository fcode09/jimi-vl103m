package parser

import (
	"encoding/hex"
	"testing"

	"github.com/intelcon-group/jimi-vl103m/pkg/jimi/packet"
	"github.com/intelcon-group/jimi-vl103m/pkg/jimi/protocol"
)

func TestHeartbeatParser_ProtocolNumber(t *testing.T) {
	p := NewHeartbeatParser()
	if p.ProtocolNumber() != protocol.ProtocolHeartbeat {
		t.Errorf("Expected protocol 0x%02X, got 0x%02X", protocol.ProtocolHeartbeat, p.ProtocolNumber())
	}
}

func TestHeartbeatParser_Name(t *testing.T) {
	p := NewHeartbeatParser()
	if p.Name() != "Heartbeat" {
		t.Errorf("Expected name 'Heartbeat', got '%s'", p.Name())
	}
}

func TestHeartbeatParser_Parse(t *testing.T) {
	tests := []struct {
		name           string
		hex            string
		wantVoltage    protocol.VoltageLevel
		wantSignal     protocol.GSMSignalStrength
		wantACCOn      bool
		wantHasExtInfo bool
		wantErr        bool
	}{
		{
			name:        "normal heartbeat - ACC off",
			hex:         "78780513040300010006950D0A",
			wantVoltage: protocol.VoltageLow,
			wantSignal:  protocol.SignalNone,
			wantACCOn:   false,
			wantErr:     false,
		},
		{
			name:        "heartbeat with ACC on",
			hex:         "78780513240402010007A50D0A",
			wantVoltage: protocol.VoltageMedium,
			wantSignal:  protocol.SignalWeak,
			wantACCOn:   true,
			wantErr:     false,
		},
		{
			name:        "heartbeat while charging",
			hex:         "78780513140504010008B50D0A",
			wantVoltage: protocol.VoltageHigh,
			wantSignal:  protocol.SignalStrong,
			wantACCOn:   false,
			wantErr:     false,
		},
		{
			name:           "heartbeat with extended info",
			hex:            "787807130403011234000100C5D50D0A",
			wantVoltage:    protocol.VoltageLow,
			wantSignal:     protocol.SignalExtremelyWeak,
			wantACCOn:      false,
			wantHasExtInfo: true,
			wantErr:        false,
		},
		{
			name:    "packet too short",
			hex:     "78780313000001ABCD0D0A",
			wantErr: true,
		},
	}

	p := NewHeartbeatParser()

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

			hbPkt, ok := pkt.(*packet.HeartbeatPacket)
			if !ok {
				t.Fatalf("Expected *HeartbeatPacket, got %T", pkt)
			}

			if hbPkt.VoltageLevel != tt.wantVoltage {
				t.Errorf("Voltage mismatch: expected %s, got %s", tt.wantVoltage, hbPkt.VoltageLevel)
			}

			if hbPkt.GSMSignal != tt.wantSignal {
				t.Errorf("Signal mismatch: expected %s, got %s", tt.wantSignal, hbPkt.GSMSignal)
			}

			if tt.wantHasExtInfo && !hbPkt.HasExtended {
				t.Error("Expected extended info to be present")
			}

			if hbPkt.ProtocolNumber() != protocol.ProtocolHeartbeat {
				t.Errorf("Protocol mismatch: expected 0x%02X, got 0x%02X",
					protocol.ProtocolHeartbeat, hbPkt.ProtocolNumber())
			}
		})
	}
}

func TestHeartbeatParser_BatteryPercentage(t *testing.T) {
	tests := []struct {
		voltage protocol.VoltageLevel
		percent int
	}{
		{protocol.VoltageNoPower, 0},
		{protocol.VoltageExtremelyLow, 5},
		{protocol.VoltageVeryLow, 15},
		{protocol.VoltageLow, 30},
		{protocol.VoltageMedium, 50},
		{protocol.VoltageHigh, 75},
		{protocol.VoltageExtremelyHigh, 100},
	}

	for _, tt := range tests {
		t.Run(tt.voltage.String(), func(t *testing.T) {
			pkt := &packet.HeartbeatPacket{
				VoltageLevel: tt.voltage,
			}

			if pkt.BatteryPercentage() != tt.percent {
				t.Errorf("Expected %d%%, got %d%%", tt.percent, pkt.BatteryPercentage())
			}
		})
	}
}

func TestHeartbeatParser_SignalBars(t *testing.T) {
	tests := []struct {
		signal protocol.GSMSignalStrength
		bars   int
	}{
		{protocol.SignalNone, 0},
		{protocol.SignalExtremelyWeak, 1},
		{protocol.SignalWeak, 2},
		{protocol.SignalGood, 3},
		{protocol.SignalStrong, 4},
	}

	for _, tt := range tests {
		t.Run(tt.signal.String(), func(t *testing.T) {
			pkt := &packet.HeartbeatPacket{
				GSMSignal: tt.signal,
			}

			if pkt.SignalBars() != tt.bars {
				t.Errorf("Expected %d bars, got %d bars", tt.bars, pkt.SignalBars())
			}
		})
	}
}
