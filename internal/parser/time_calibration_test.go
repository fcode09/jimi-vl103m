package parser

import (
	"encoding/hex"
	"testing"

	"github.com/fcode09/jimi-vl103m/pkg/jimi/packet"
	"github.com/fcode09/jimi-vl103m/pkg/jimi/protocol"
)

func TestTimeCalibrationParser_ProtocolNumber(t *testing.T) {
	p := NewTimeCalibrationParser()
	if p.ProtocolNumber() != protocol.ProtocolTimeCalibration {
		t.Errorf("Expected protocol 0x%02X, got 0x%02X", protocol.ProtocolTimeCalibration, p.ProtocolNumber())
	}
}

func TestTimeCalibrationParser_Name(t *testing.T) {
	p := NewTimeCalibrationParser()
	if p.Name() != "Time Calibration" {
		t.Errorf("Expected name 'Time Calibration', got '%s'", p.Name())
	}
}

func TestTimeCalibrationParser_Parse(t *testing.T) {
	tests := []struct {
		name    string
		hex     string
		wantErr bool
	}{
		{
			name:    "time calibration request",
			hex:     "7878058A00010003870D0A",
			wantErr: false,
		},
		{
			name:    "another time calibration",
			hex:     "7878058A00020004980D0A",
			wantErr: false,
		},
	}

	p := NewTimeCalibrationParser()
	ctx := DefaultContext()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := hex.DecodeString(tt.hex)
			if err != nil {
				t.Fatalf("Failed to decode hex: %v", err)
			}

			pkt, err := p.Parse(data, ctx)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			tcPkt, ok := pkt.(*packet.TimeCalibrationPacket)
			if !ok {
				t.Fatalf("Expected *TimeCalibrationPacket, got %T", pkt)
			}

			if tcPkt.ProtocolNumber() != protocol.ProtocolTimeCalibration {
				t.Errorf("Protocol mismatch: expected 0x%02X, got 0x%02X",
					protocol.ProtocolTimeCalibration, tcPkt.ProtocolNumber())
			}

			t.Logf("Time Calibration: Serial=%d", tcPkt.SerialNumber())
		})
	}
}

func TestTimeCalibrationPacket_Type(t *testing.T) {
	pkt := &packet.TimeCalibrationPacket{}
	if pkt.Type() != "Time Calibration" {
		t.Errorf("Expected type 'Time Calibration', got '%s'", pkt.Type())
	}
}
