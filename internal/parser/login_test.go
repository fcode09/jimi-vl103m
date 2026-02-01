package parser

import (
	"encoding/hex"
	"testing"

	"github.com/fcode09/jimi-vl103m/pkg/jimi/packet"
	"github.com/fcode09/jimi-vl103m/pkg/jimi/protocol"
)

func TestLoginParser_ProtocolNumber(t *testing.T) {
	p := NewLoginParser()
	if p.ProtocolNumber() != protocol.ProtocolLogin {
		t.Errorf("Expected protocol 0x%02X, got 0x%02X", protocol.ProtocolLogin, p.ProtocolNumber())
	}
}

func TestLoginParser_Name(t *testing.T) {
	p := NewLoginParser()
	if p.Name() != "Login" {
		t.Errorf("Expected name 'Login', got '%s'", p.Name())
	}
}

func TestLoginParser_Parse(t *testing.T) {
	tests := []struct {
		name        string
		hex         string
		wantIMEI    string
		wantModelID uint16
		wantErr     bool
	}{
		{
			name:        "valid login packet",
			hex:         "78781101035933907393052380044D014E00015ED00D0A",
			wantIMEI:    "359339073930523", // 16-digit BCD: 0359339073930523 → last 15 digits
			wantModelID: 0x8004,
			wantErr:     false,
		},
		{
			name:        "login with timezone",
			hex:         "787811010123456789012348044D03200001ABCD0D0A",
			wantIMEI:    "123456789012348", // 16-digit BCD: 0123456789012348 → last 15 digits
			wantModelID: 0x044D,
			wantErr:     false,
		},
		{
			name:    "packet too short",
			hex:     "787805010001ABCD0D0A",
			wantErr: true,
		},
	}

	p := NewLoginParser()
	ctx := Context{ValidateIMEI: false}

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

			loginPkt, ok := pkt.(*packet.LoginPacket)
			if !ok {
				t.Fatalf("Expected *LoginPacket, got %T", pkt)
			}

			if loginPkt.IMEI.String() != tt.wantIMEI {
				t.Errorf("IMEI mismatch: expected %s, got %s", tt.wantIMEI, loginPkt.IMEI.String())
			}

			if loginPkt.ModelID != tt.wantModelID {
				t.Errorf("ModelID mismatch: expected 0x%04X, got 0x%04X", tt.wantModelID, loginPkt.ModelID)
			}

			if loginPkt.ProtocolNumber() != protocol.ProtocolLogin {
				t.Errorf("Protocol mismatch: expected 0x%02X, got 0x%02X",
					protocol.ProtocolLogin, loginPkt.ProtocolNumber())
			}
		})
	}
}

func TestLoginParser_ParseIMEIValidation(t *testing.T) {
	p := NewLoginParser()
	ctx := Context{ValidateIMEI: false}

	// Valid login packet
	data, _ := hex.DecodeString("78781101035933907393052380044D014E00015ED00D0A")

	pkt, err := p.Parse(data, ctx)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	loginPkt := pkt.(*packet.LoginPacket)

	if !loginPkt.IMEI.IsValid() {
		t.Error("Expected IMEI to be valid")
	}

	// Check IMEI string format
	imeiStr := loginPkt.IMEI.String()
	if len(imeiStr) != 15 {
		t.Errorf("Expected IMEI length 15, got %d", len(imeiStr))
	}
}

func TestLoginParser_TimezoneExtraction(t *testing.T) {
	p := NewLoginParser()
	ctx := Context{ValidateIMEI: false}

	// Login packet with UTC+8 timezone (0x0320 = 800 = +8:00)
	data, _ := hex.DecodeString("787811010123456789012348044D03200001ABCD0D0A")

	pkt, err := p.Parse(data, ctx)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	loginPkt := pkt.(*packet.LoginPacket)

	// Verify timezone is extracted
	tz := loginPkt.Timezone
	t.Logf("Timezone: %s (offset: %d minutes)", tz, tz.OffsetMinutes)
}
