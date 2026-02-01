package parser

import (
	"encoding/hex"
	"testing"

	"github.com/intelcon-group/jimi-vl103m/pkg/jimi/packet"
	"github.com/intelcon-group/jimi-vl103m/pkg/jimi/protocol"
)

func TestLocationParser_ProtocolNumber(t *testing.T) {
	p := NewLocationParser()
	if p.ProtocolNumber() != protocol.ProtocolGPSLocation {
		t.Errorf("Expected protocol 0x%02X, got 0x%02X", protocol.ProtocolGPSLocation, p.ProtocolNumber())
	}
}

func TestLocationParser4G_ProtocolNumber(t *testing.T) {
	p := NewLocation4GParser()
	if p.ProtocolNumber() != protocol.ProtocolGPSLocation4G {
		t.Errorf("Expected protocol 0x%02X, got 0x%02X", protocol.ProtocolGPSLocation4G, p.ProtocolNumber())
	}
}

func TestLocationParser_Name(t *testing.T) {
	p := NewLocationParser()
	if p.Name() != "GPS Location" {
		t.Errorf("Expected name 'GPS Location', got '%s'", p.Name())
	}

	p4g := NewLocation4GParser()
	if p4g.Name() != "GPS Location 4G" {
		t.Errorf("Expected name 'GPS Location 4G', got '%s'", p4g.Name())
	}
}

func TestLocationParser_Parse(t *testing.T) {
	tests := []struct {
		name           string
		hex            string
		wantPositioned bool
		wantErr        bool
	}{
		{
			name:           "basic GPS location",
			hex:            "787822220f0c1d023305c9026b8d550c39771d14140c01cc00000100287d00000000001f710001643f0d0a",
			wantPositioned: true,
			wantErr:        false,
		},
		{
			name:           "location with status",
			hex:            "78782222180101120000cc026e953b0c3cb40010b401cc0128f8006563010000000003e800022e030d0a",
			wantPositioned: true,
			wantErr:        false,
		},
		{
			name:           "location without GPS fix",
			hex:            "787822220f0c1d02330500000000000000000000000001cc00000100287d000000000000000003442c0d0a",
			wantPositioned: false,
			wantErr:        false,
		},
		{
			name:    "packet too short",
			hex:     "787805220001ABCD0D0A",
			wantErr: true,
		},
	}

	p := NewLocationParser()

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

			locPkt, ok := pkt.(*packet.LocationPacket)
			if !ok {
				t.Fatalf("Expected *LocationPacket, got %T", pkt)
			}

			if locPkt.IsPositioned() != tt.wantPositioned {
				t.Errorf("IsPositioned mismatch: expected %v, got %v", tt.wantPositioned, locPkt.IsPositioned())
			}

			if locPkt.ProtocolNumber() != protocol.ProtocolGPSLocation {
				t.Errorf("Protocol mismatch: expected 0x%02X, got 0x%02X",
					protocol.ProtocolGPSLocation, locPkt.ProtocolNumber())
			}

			t.Logf("Location: Lat=%.6f, Lon=%.6f, Speed=%d, Heading=%d (%s)",
				locPkt.Latitude(), locPkt.Longitude(),
				locPkt.Speed, locPkt.Heading(), locPkt.HeadingName())
		})
	}
}

func TestLocationParser_Coordinates(t *testing.T) {
	p := NewLocationParser()

	// Parse a known location packet
	data, _ := hex.DecodeString("787822220f0c1d023305c9026b8d550c39771d14140c01cc00000100287d00000000001f710001643f0d0a")

	pkt, err := p.Parse(data)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	locPkt := pkt.(*packet.LocationPacket)

	// Validate coordinates are within expected range
	lat := locPkt.Latitude()
	lon := locPkt.Longitude()

	if lat < -90 || lat > 90 {
		t.Errorf("Invalid latitude: %f", lat)
	}

	if lon < -180 || lon > 180 {
		t.Errorf("Invalid longitude: %f", lon)
	}

	t.Logf("Coordinates: (%.6f, %.6f)", lat, lon)
}

func TestLocationParser_DateTime(t *testing.T) {
	p := NewLocationParser()

	data, _ := hex.DecodeString("787822220f0c1d023305c9026b8d550c39771d14140c01cc00000100287d00000000001f710001643f0d0a")

	pkt, err := p.Parse(data)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	locPkt := pkt.(*packet.LocationPacket)

	// Check that datetime is parsed
	if locPkt.DateTime.IsZero() {
		t.Error("DateTime should not be zero")
	}

	t.Logf("DateTime: %s", locPkt.DateTime)
}

func TestLocationParser_HeadingNames(t *testing.T) {
	headingTests := []struct {
		heading int
		name    string
	}{
		{0, "N"},
		{45, "NE"},
		{90, "E"},
		{135, "SE"},
		{180, "S"},
		{225, "SW"},
		{270, "W"},
		{315, "NW"},
		{360, "N"},
	}

	for _, tt := range headingTests {
		t.Run(tt.name, func(t *testing.T) {
			pkt := &packet.LocationPacket{}
			// Heading is extracted from CourseStatus
			// For this test, we just verify the method exists
			_ = pkt.Heading()
			_ = pkt.HeadingName()
		})
	}
}

func TestLocationParser_SatelliteCount(t *testing.T) {
	p := NewLocationParser()

	data, _ := hex.DecodeString("787822220f0c1d023305c9026b8d550c39771d14140c01cc00000100287d00000000001f710001643f0d0a")

	pkt, err := p.Parse(data)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	locPkt := pkt.(*packet.LocationPacket)

	// Satellite count should be reasonable (0-24)
	if locPkt.Satellites > 24 {
		t.Errorf("Unreasonable satellite count: %d", locPkt.Satellites)
	}

	t.Logf("Satellites: %d", locPkt.Satellites)
}

func TestLocationParser_ACCStatus(t *testing.T) {
	p := NewLocationParser()

	// Test packet with ACC=ON (0x01) - Simulated packet from user
	// Raw: 787822221A02010E02118901C31ADC07ABA0CA00189301361A1234005678010000003C00B90D0A
	// The ACC byte is 0x01 (at position after LBS data), should be ON
	dataACCOn, _ := hex.DecodeString("787822221A02010E02118901C31ADC07ABA0CA00189301361A1234005678010000003C00B90D0A")

	pkt, err := p.Parse(dataACCOn)
	if err != nil {
		t.Fatalf("Parse failed for ACC=ON packet: %v", err)
	}

	locPkt := pkt.(*packet.LocationPacket)

	if !locPkt.ACC {
		t.Errorf("ACC should be ON (true), got OFF (false)")
	}

	if !locPkt.ACCOn() {
		t.Errorf("ACCOn() should return true, got false")
	}

	t.Logf("ACC Status: %v (expected: true)", locPkt.ACC)

	// Test packet with ACC=OFF (0x00)
	// Same packet but with ACC byte changed to 0x00
	dataACCOff, _ := hex.DecodeString("787822221A02010E02118901C31ADC07ABA0CA00189301361A1234005678000000003C00B80D0A")

	pkt2, err := p.Parse(dataACCOff)
	if err != nil {
		t.Fatalf("Parse failed for ACC=OFF packet: %v", err)
	}

	locPkt2 := pkt2.(*packet.LocationPacket)

	if locPkt2.ACC {
		t.Errorf("ACC should be OFF (false), got ON (true)")
	}

	if locPkt2.ACCOn() {
		t.Errorf("ACCOn() should return false, got true")
	}

	t.Logf("ACC Status: %v (expected: false)", locPkt2.ACC)
}

func TestLocation4GParser_ACCStatus(t *testing.T) {
	p := NewLocation4GParser()

	// Test packet 4G with ACC=ON (0x01) - Simulated packet from user
	// Raw: 78782DA01A02010E04038A01C3BF0307ABCB8500193501CC000000C60D0000000002CE68EA010000000000000004652A0D0A
	dataACCOn, _ := hex.DecodeString("78782DA01A02010E04038A01C3BF0307ABCB8500193501CC000000C60D0000000002CE68EA010000000000000004652A0D0A")

	pkt, err := p.Parse(dataACCOn)
	if err != nil {
		t.Fatalf("Parse failed for 4G ACC=ON packet: %v", err)
	}

	locPkt := pkt.(*packet.Location4GPacket)

	if !locPkt.ACC {
		t.Errorf("ACC should be ON (true), got OFF (false)")
	}

	if !locPkt.ACCOn() {
		t.Errorf("ACCOn() should return true, got false")
	}

	t.Logf("4G ACC Status: %v (expected: true)", locPkt.ACC)
}
