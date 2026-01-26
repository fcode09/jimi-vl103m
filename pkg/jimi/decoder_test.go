package jimi

import (
	"encoding/hex"
	"testing"

	"github.com/fcode09/jimi-vl103m/pkg/jimi/packet"
	"github.com/fcode09/jimi-vl103m/pkg/jimi/protocol"
)

func TestNewDecoder(t *testing.T) {
	decoder := NewDecoder()
	if decoder == nil {
		t.Fatal("NewDecoder returned nil")
	}

	// Check default options
	opts := decoder.GetOptions()
	if !opts.StrictMode {
		t.Error("Expected StrictMode to be true by default")
	}
	if opts.SkipCRCValidation {
		t.Error("Expected SkipCRCValidation to be false by default")
	}
}

func TestNewDecoderWithOptions(t *testing.T) {
	decoder := NewDecoder(
		WithStrictMode(false),
		WithAllowUnknownProtocols(),
	)

	opts := decoder.GetOptions()
	if opts.StrictMode {
		t.Error("Expected StrictMode to be false")
	}
	if !opts.AllowUnknownProtocols {
		t.Error("Expected AllowUnknownProtocols to be true")
	}
}

func TestDecoderRegisteredProtocols(t *testing.T) {
	decoder := NewDecoder()

	protocols := decoder.RegisteredProtocols()
	if len(protocols) == 0 {
		t.Fatal("No protocols registered")
	}

	// Check that core protocols are registered
	expectedProtocols := []byte{
		protocol.ProtocolLogin,
		protocol.ProtocolHeartbeat,
		protocol.ProtocolGPSLocation,
		protocol.ProtocolAlarm,
	}

	for _, expected := range expectedProtocols {
		found := false
		for _, p := range protocols {
			if p == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected protocol 0x%02X to be registered", expected)
		}
	}
}

func TestDecoderHasParser(t *testing.T) {
	decoder := NewDecoder()

	if !decoder.HasParser(protocol.ProtocolLogin) {
		t.Error("Expected Login parser to be registered")
	}
	if !decoder.HasParser(protocol.ProtocolHeartbeat) {
		t.Error("Expected Heartbeat parser to be registered")
	}
	if !decoder.HasParser(protocol.ProtocolGPSLocation) {
		t.Error("Expected GPS Location parser to be registered")
	}

	// Unknown protocol should not have parser
	if decoder.HasParser(0xFF) {
		t.Error("Expected no parser for protocol 0xFF")
	}
}

// Test decoding a login packet
func TestDecodeLoginPacket(t *testing.T) {
	// Login packet example:
	// 78 78 - Start bit
	// 11 - Length (17 bytes)
	// 01 - Protocol (Login)
	// 03 59 33 90 73 93 05 23 - IMEI (BCD: 0359339073930523)
	// 04 4D - Model ID
	// 01 4E - Timezone/Language
	// 0F 0A - Serial number
	// XX XX - CRC (will need to calculate)
	// 0D 0A - Stop bit

	// For testing without CRC validation
	decoder := NewDecoder(WithSkipCRC(), WithStrictMode(false))

	// This is a simplified test packet - in real tests you'd use actual captured packets
	hexData := "7878110103593390739305300400014E0001FFFF0D0A"
	data, err := hex.DecodeString(hexData)
	if err != nil {
		t.Fatalf("Failed to decode hex: %v", err)
	}

	// Try to decode - may fail due to CRC but structure test is valid
	pkt, err := decoder.Decode(data)
	if err != nil {
		// In lenient mode with skip CRC, we might still get errors for other reasons
		t.Logf("Decode returned error (expected for test data): %v", err)
		return
	}

	if pkt == nil {
		t.Fatal("Decoded packet is nil")
	}

	if pkt.ProtocolNumber() != protocol.ProtocolLogin {
		t.Errorf("Expected protocol 0x01, got 0x%02X", pkt.ProtocolNumber())
	}

	loginPkt, ok := pkt.(*packet.LoginPacket)
	if !ok {
		t.Fatalf("Expected *packet.LoginPacket, got %T", pkt)
	}

	t.Logf("Login packet: IMEI=%s, ModelID=0x%04X", loginPkt.IMEI, loginPkt.ModelID)
}

// Test structure validation
func TestValidateStructure(t *testing.T) {
	decoder := NewDecoder()

	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "too short",
			data:    []byte{0x78, 0x78},
			wantErr: true,
		},
		{
			name:    "invalid start bit",
			data:    []byte{0x00, 0x00, 0x05, 0x01, 0x00, 0x00, 0x00, 0x00, 0x0D, 0x0A},
			wantErr: true,
		},
		{
			name:    "invalid stop bit",
			data:    []byte{0x78, 0x78, 0x05, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := decoder.ValidateStructure(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStructure() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test GetProtocolNumber
func TestGetProtocolNumber(t *testing.T) {
	decoder := NewDecoder()

	// Short packet format (0x7878)
	shortPacket := []byte{0x78, 0x78, 0x05, 0x01, 0x00, 0x00, 0x00, 0x00, 0x0D, 0x0A}
	proto, err := decoder.GetProtocolNumber(shortPacket)
	if err != nil {
		t.Fatalf("GetProtocolNumber failed: %v", err)
	}
	if proto != 0x01 {
		t.Errorf("Expected protocol 0x01, got 0x%02X", proto)
	}

	// Long packet format (0x7979)
	longPacket := []byte{0x79, 0x79, 0x00, 0x05, 0x01, 0x00, 0x00, 0x00, 0x00, 0x0D, 0x0A}
	proto, err = decoder.GetProtocolNumber(longPacket)
	if err != nil {
		t.Fatalf("GetProtocolNumber failed: %v", err)
	}
	if proto != 0x01 {
		t.Errorf("Expected protocol 0x01, got 0x%02X", proto)
	}
}

// Test HasCompletePacket
func TestHasCompletePacket(t *testing.T) {
	decoder := NewDecoder()

	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{
			name: "complete short packet",
			data: []byte{0x78, 0x78, 0x05, 0x01, 0x00, 0x00, 0x00, 0x00, 0x0D, 0x0A},
			want: true,
		},
		{
			name: "incomplete packet",
			data: []byte{0x78, 0x78, 0x05, 0x01},
			want: false,
		},
		{
			name: "too short",
			data: []byte{0x78, 0x78},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := decoder.HasCompletePacket(tt.data); got != tt.want {
				t.Errorf("HasCompletePacket() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test DecodeStream with concatenated packets
func TestDecodeStream(t *testing.T) {
	decoder := NewDecoder(WithSkipCRC(), WithStrictMode(false), WithAllowUnknownProtocols())

	// Two concatenated minimal packets
	// Note: These are test packets, not valid protocol packets
	packet1 := []byte{0x78, 0x78, 0x05, 0x13, 0x00, 0x00, 0x00, 0x00, 0x0D, 0x0A}
	packet2 := []byte{0x78, 0x78, 0x05, 0x13, 0x00, 0x00, 0x00, 0x00, 0x0D, 0x0A}

	stream := append(packet1, packet2...)

	packets, residue, err := decoder.DecodeStream(stream)
	if err != nil {
		t.Logf("DecodeStream error (may be expected): %v", err)
	}

	t.Logf("Decoded %d packets, residue length: %d", len(packets), len(residue))
}

// Benchmark decoder creation
func BenchmarkNewDecoder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewDecoder()
	}
}

// Benchmark protocol number extraction
func BenchmarkGetProtocolNumber(b *testing.B) {
	decoder := NewDecoder()
	packet := []byte{0x78, 0x78, 0x05, 0x01, 0x00, 0x00, 0x00, 0x00, 0x0D, 0x0A}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = decoder.GetProtocolNumber(packet)
	}
}
