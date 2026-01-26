package types

import (
	"math"
	"testing"
)

func TestNewCoordinates(t *testing.T) {
	tests := []struct {
		name    string
		lat     float64
		lon     float64
		wantErr bool
	}{
		{
			name:    "valid coordinates",
			lat:     23.125346,
			lon:     113.251515,
			wantErr: false,
		},
		{
			name:    "valid negative coordinates",
			lat:     -33.8688,
			lon:     -151.2093,
			wantErr: false,
		},
		{
			name:    "latitude too high",
			lat:     91.0,
			lon:     0.0,
			wantErr: true,
		},
		{
			name:    "latitude too low",
			lat:     -91.0,
			lon:     0.0,
			wantErr: true,
		},
		{
			name:    "longitude too high",
			lat:     0.0,
			lon:     181.0,
			wantErr: true,
		},
		{
			name:    "longitude too low",
			lat:     0.0,
			lon:     -181.0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewCoordinates(tt.lat, tt.lon)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCoordinates() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCoordinatesFromBytes(t *testing.T) {
	// Latitude: 23.134439° = 23.134439 * 1800000 = 41641990.2 = 0x027B7806
	// Using 0x027B7807 from the original test data for consistency.
	latBytes := []byte{0x02, 0x7B, 0x78, 0x07}
	lonBytes := []byte{0x0C, 0x26, 0x57, 0x47}

	coords, err := NewCoordinatesFromBytes(latBytes, lonBytes, true, true)
	if err != nil {
		t.Fatalf("NewCoordinatesFromBytes() error = %v", err)
	}

	// Check latitude (allow small floating point error)
	expectedLat := 23.136715
	if math.Abs(coords.Latitude-expectedLat) > 0.000001 {
		t.Errorf("Expected latitude %f, got %f", expectedLat, coords.Latitude)
	}

	// Check longitude
	expectedLon := 113.244057
	if math.Abs(coords.Longitude-expectedLon) > 0.000001 {
		t.Errorf("Expected longitude %f, got %f", expectedLon, coords.Longitude)
	}

	if !coords.IsNorth {
		t.Error("Expected IsNorth to be true")
	}
	if !coords.IsEast {
		t.Error("Expected IsEast to be true")
	}
}

func TestCoordinatesSignedValues(t *testing.T) {
	// Test Southern/Western hemisphere
	coords := Coordinates{
		Latitude:  33.8688,
		Longitude: 151.2093,
		IsNorth:   false, // South
		IsEast:    false, // West
	}

	if signedLat := coords.SignedLatitude(); signedLat >= 0 {
		t.Errorf("Expected negative latitude for South, got %f", signedLat)
	}

	if signedLon := coords.SignedLongitude(); signedLon >= 0 {
		t.Errorf("Expected negative longitude for West, got %f", signedLon)
	}
}

func TestCoordinatesToBytes(t *testing.T) {
	coords, _ := NewCoordinates(23.125346, 113.251515)

	latBytes := coords.LatitudeBytes()
	lonBytes := coords.LongitudeBytes()

	if len(latBytes) != 4 {
		t.Errorf("Expected 4 bytes for latitude, got %d", len(latBytes))
	}
	if len(lonBytes) != 4 {
		t.Errorf("Expected 4 bytes for longitude, got %d", len(lonBytes))
	}
}

func TestCoordinatesDistance(t *testing.T) {
	// Distance between Sydney and Melbourne
	sydney, _ := NewCoordinates(-33.8688, 151.2093)
	melbourne, _ := NewCoordinates(-37.8136, 144.9631)

	distance := sydney.DistanceTo(melbourne)

	// Expected distance is approximately 714 km
	expectedKm := 714000.0
	tolerance := 50000.0 // 50 km tolerance

	if math.Abs(distance-expectedKm) > tolerance {
		t.Errorf("Expected distance ~%f meters, got %f meters", expectedKm, distance)
	}
}

func TestCoordinatesIsValid(t *testing.T) {
	valid, _ := NewCoordinates(23.125346, 113.251515)
	if !valid.IsValid() {
		t.Error("Expected valid coordinates")
	}

	invalid := Coordinates{Latitude: 100, Longitude: 200}
	if invalid.IsValid() {
		t.Error("Expected invalid coordinates")
	}
}

func TestCoordinatesIsZero(t *testing.T) {
	zero := Coordinates{Latitude: 0, Longitude: 0}
	if !zero.IsZero() {
		t.Error("Expected zero coordinates")
	}

	nonZero, _ := NewCoordinates(23.125346, 113.251515)
	if nonZero.IsZero() {
		t.Error("Expected non-zero coordinates")
	}
}

func TestCoordinatesString(t *testing.T) {
	coords := Coordinates{
		Latitude:  23.125346,
		Longitude: 113.251515,
		IsNorth:   true,
		IsEast:    true,
	}

	str := coords.String()
	if str == "" {
		t.Error("Expected non-empty string")
	}

	// Should contain N and E
	if !contains(str, "N") || !contains(str, "E") {
		t.Errorf("Expected string to contain N and E: %s", str)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr, 0))
}

func containsAt(s, substr string, start int) bool {
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func BenchmarkNewCoordinatesFromBytes(b *testing.B) {
	latBytes := []byte{0x02, 0x7B, 0x78, 0x07}
	lonBytes := []byte{0x0C, 0x26, 0x57, 0x47}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewCoordinatesFromBytes(latBytes, lonBytes, true, true)
	}
}

func BenchmarkDistanceTo(b *testing.B) {
	coord1, _ := NewCoordinates(-33.8688, 151.2093)
	coord2, _ := NewCoordinates(-37.8136, 144.9631)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = coord1.DistanceTo(coord2)
	}
}
