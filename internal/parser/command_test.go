package parser

import (
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/fcode09/jimi-vl103m/pkg/jimi/packet"
	"github.com/fcode09/jimi-vl103m/pkg/jimi/protocol"
)

func TestGPSAddressRequestParser_ProtocolNumber(t *testing.T) {
	p := NewGPSAddressRequestParser()
	if p.ProtocolNumber() != protocol.ProtocolGPSAddressRequest {
		t.Errorf("Expected protocol 0x%02X, got 0x%02X",
			protocol.ProtocolGPSAddressRequest, p.ProtocolNumber())
	}
}

func TestGPSAddressRequestParser_Name(t *testing.T) {
	p := NewGPSAddressRequestParser()
	if p.Name() != "GPS Address Request" {
		t.Errorf("Expected name 'GPS Address Request', got '%s'", p.Name())
	}
}

func TestGPSAddressRequestParser_Parse(t *testing.T) {
	tests := []struct {
		name        string
		hex         string
		wantErr     bool
		errContains string
		checkFunc   func(*packet.GPSAddressRequestPacket) error
	}{
		{
			name:        "packet too short",
			hex:         "7878052A0001ABCD0D0A",
			wantErr:     true,
			errContains: "content too short",
		},
		{
			name: "packet with valid 41 bytes",
			// Build a valid packet with exact 41 bytes of content:
			// Length byte = 0x2B = 43 (protocol + 41 content + serial)
			// Content (41 bytes):
			// - DateTime: 1A 02 01 0E 0D 35 (6 bytes)
			// - GPS Info: 8A (1 byte)
			// - Latitude: 01 C3 BF 03 (4 bytes)
			// - Longitude: 07 AB CB 85 (4 bytes)
			// - Speed: 00 (1 byte)
			// - Course: 18 00 (2 bytes)
			// - Phone: 21 bytes
			// - Alarm/Language: 03 00 (2 bytes)
			hex: "78782B2A" + // header + length (43) + protocol
				"1A02010E0D35" + // DateTime (6 bytes)
				"8A" + // GPS Info (1 byte)
				"01C3BF03" + // Latitude (4 bytes)
				"07ABCB85" + // Longitude (4 bytes)
				"00" + // Speed (1 byte)
				"1800" + // Course (2 bytes)
				"3132333435363738393031323334353637383930" + // Phone 20 chars
				"31" + // Phone 21st char
				"0300" + // Alarm/Language (2 bytes)
				"1234" + // Serial (2 bytes)
				"00000D0A", // CRC (placeholder) + stop
			wantErr: false,
			checkFunc: func(p *packet.GPSAddressRequestPacket) error {
				// Check DateTime is parsed
				if p.DateTime.IsZero() {
					return fmt.Errorf("DateTime should not be zero")
				}

				// Check satellites
				if p.Satellites != 10 {
					return fmt.Errorf("expected 10 satellites, got %d", p.Satellites)
				}

				// Check coordinates are valid (non-zero)
				if p.Latitude() == 0 && p.Longitude() == 0 {
					return fmt.Errorf("coordinates should not be zero")
				}

				// Check phone number contains expected digits
				if !strings.Contains(p.PhoneNumber, "1234567890") {
					return fmt.Errorf("phone number not parsed correctly: %s", p.PhoneNumber)
				}

				// Check alarm type (Vibration = 0x03)
				if p.AlarmType != protocol.AlarmVibration {
					return fmt.Errorf("expected Vibration alarm (0x03), got %s", p.AlarmType)
				}

				return nil
			},
		},
		{
			name: "packet with SOS alarm",
			// Same structure but with SOS alarm (0x01)
			hex: "78782B2A" + // header + length (43) + protocol
				"1A02010E0D35" + // DateTime (6 bytes)
				"8A" + // GPS Info (1 byte)
				"01C3BF03" + // Latitude (4 bytes)
				"07ABCB85" + // Longitude (4 bytes)
				"00" + // Speed (1 byte)
				"1800" + // Course (2 bytes)
				"3132333435363738393031323334353637383930" + // Phone 20 chars
				"31" + // Phone 21st char
				"0100" + // Alarm/Language (2 bytes) - SOS
				"1234" + // Serial (2 bytes)
				"00000D0A", // CRC + stop
			wantErr: false,
			checkFunc: func(p *packet.GPSAddressRequestPacket) error {
				if p.AlarmType != protocol.AlarmSOS {
					return fmt.Errorf("expected SOS alarm (0x01), got %s", p.AlarmType)
				}
				return nil
			},
		},
	}

	parser := NewGPSAddressRequestParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := hex.DecodeString(tt.hex)
			if err != nil {
				t.Fatalf("Failed to decode hex: %v", err)
			}

			pkt, err := parser.Parse(data)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Expected error containing '%s', got: %v", tt.errContains, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			gpsPkt, ok := pkt.(*packet.GPSAddressRequestPacket)
			if !ok {
				t.Fatalf("Expected *GPSAddressRequestPacket, got %T", pkt)
			}

			if tt.checkFunc != nil {
				if err := tt.checkFunc(gpsPkt); err != nil {
					t.Errorf("Check failed: %v", err)
				}
			}
		})
	}
}

func TestGPSAddressRequestPacket_Methods(t *testing.T) {
	// Create a test packet with sample data
	pkt := &packet.GPSAddressRequestPacket{
		Satellites:  8,
		Speed:       60,
		PhoneNumber: "+51987654321",
		AlarmType:   protocol.AlarmSpeed,
		Language:    protocol.LanguageEnglish,
	}

	// Test interface implementations
	if pkt.HasTimestamp() {
		// Should be false because DateTime is zero
		t.Error("HasTimestamp should return false for zero DateTime")
	}

	// Test String() doesn't panic
	str := pkt.String()
	if str == "" {
		t.Error("String() should return non-empty string")
	}

	t.Logf("Packet string: %s", str)
}

func TestOnlineCommandParser_ProtocolNumber(t *testing.T) {
	p := NewOnlineCommandParser()
	if p.ProtocolNumber() != protocol.ProtocolOnlineCommand {
		t.Errorf("Expected protocol 0x%02X, got 0x%02X",
			protocol.ProtocolOnlineCommand, p.ProtocolNumber())
	}
}

func TestCommandResponseParser_ProtocolNumber(t *testing.T) {
	p := NewCommandResponseParser()
	if p.ProtocolNumber() != protocol.ProtocolCommandResponse {
		t.Errorf("Expected protocol 0x%02X, got 0x%02X",
			protocol.ProtocolCommandResponse, p.ProtocolNumber())
	}
}

func TestCommandResponseOldParser_ProtocolNumber(t *testing.T) {
	p := NewCommandResponseOldParser()
	if p.ProtocolNumber() != protocol.ProtocolCommandResponseOld {
		t.Errorf("Expected protocol 0x%02X, got 0x%02X",
			protocol.ProtocolCommandResponseOld, p.ProtocolNumber())
	}
}
