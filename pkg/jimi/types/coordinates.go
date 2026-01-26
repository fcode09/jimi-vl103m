package types

import (
	"fmt"
	"math"
)

// Coordinates represents GPS coordinates (latitude and longitude)
type Coordinates struct {
	Latitude  float64 // Decimal degrees (-90 to 90)
	Longitude float64 // Decimal degrees (-180 to 180)
	IsNorth   bool    // true if Northern hemisphere
	IsEast    bool    // true if Eastern hemisphere
}

// CoordinatesDivisor is the divisor used in VL103M protocol
// Lat/Lon values in protocol are stored as: actual_value * 1,800,000
const CoordinatesDivisor = 1800000.0

// NewCoordinates creates coordinates from decimal degrees
// Latitude range: -90 to 90 (negative = South, positive = North)
// Longitude range: -180 to 180 (negative = West, positive = East)
func NewCoordinates(lat, lon float64) (Coordinates, error) {
	if lat < -90 || lat > 90 {
		return Coordinates{}, fmt.Errorf("latitude out of range [-90, 90]: %f", lat)
	}
	if lon < -180 || lon > 180 {
		return Coordinates{}, fmt.Errorf("longitude out of range [-180, 180]: %f", lon)
	}

	return Coordinates{
		Latitude:  math.Abs(lat),
		Longitude: math.Abs(lon),
		IsNorth:   lat >= 0,
		IsEast:    lon >= 0,
	}, nil
}

// NewCoordinatesFromBytes creates coordinates from VL103M protocol bytes
// latBytes: 4 bytes representing latitude * 1,800,000
// lonBytes: 4 bytes representing longitude * 1,800,000
// isNorth: true if Northern hemisphere
// isEast: true if Eastern hemisphere
func NewCoordinatesFromBytes(latBytes, lonBytes []byte, isNorth, isEast bool) (Coordinates, error) {
	if len(latBytes) != 4 {
		return Coordinates{}, fmt.Errorf("latitude bytes must be 4 bytes, got %d", len(latBytes))
	}
	if len(lonBytes) != 4 {
		return Coordinates{}, fmt.Errorf("longitude bytes must be 4 bytes, got %d", len(lonBytes))
	}

	// Convert bytes to uint32 (big-endian)
	latRaw := uint32(latBytes[0])<<24 | uint32(latBytes[1])<<16 | uint32(latBytes[2])<<8 | uint32(latBytes[3])
	lonRaw := uint32(lonBytes[0])<<24 | uint32(lonBytes[1])<<16 | uint32(lonBytes[2])<<8 | uint32(lonBytes[3])

	// Convert to decimal degrees
	lat := float64(latRaw) / CoordinatesDivisor
	lon := float64(lonRaw) / CoordinatesDivisor

	return Coordinates{
		Latitude:  lat,
		Longitude: lon,
		IsNorth:   isNorth,
		IsEast:    isEast,
	}, nil
}

// MustNewCoordinates creates coordinates and panics if invalid
// Use this only when you're certain the coordinates are valid
func MustNewCoordinates(lat, lon float64) Coordinates {
	coords, err := NewCoordinates(lat, lon)
	if err != nil {
		panic(err)
	}
	return coords
}

// String returns coordinates in human-readable format
// Example: "23.125346°N, 113.251515°E"
func (c Coordinates) String() string {
	latDir := "N"
	if !c.IsNorth {
		latDir = "S"
	}

	lonDir := "E"
	if !c.IsEast {
		lonDir = "W"
	}

	return fmt.Sprintf("%.6f°%s, %.6f°%s", c.Latitude, latDir, c.Longitude, lonDir)
}

// SignedLatitude returns latitude with sign (negative for South)
func (c Coordinates) SignedLatitude() float64 {
	if c.IsNorth {
		return c.Latitude
	}
	return -c.Latitude
}

// SignedLongitude returns longitude with sign (negative for West)
func (c Coordinates) SignedLongitude() float64 {
	if c.IsEast {
		return c.Longitude
	}
	return -c.Longitude
}

// LatitudeBytes returns the latitude as 4 bytes (big-endian)
// for encoding back to VL103M protocol format
func (c Coordinates) LatitudeBytes() []byte {
	value := uint32(c.Latitude * CoordinatesDivisor)
	return []byte{
		byte(value >> 24),
		byte(value >> 16),
		byte(value >> 8),
		byte(value),
	}
}

// LongitudeBytes returns the longitude as 4 bytes (big-endian)
// for encoding back to VL103M protocol format
func (c Coordinates) LongitudeBytes() []byte {
	value := uint32(c.Longitude * CoordinatesDivisor)
	return []byte{
		byte(value >> 24),
		byte(value >> 16),
		byte(value >> 8),
		byte(value),
	}
}

// IsValid returns true if coordinates appear to be valid
// (non-zero and within reasonable range)
func (c Coordinates) IsValid() bool {
	return c.Latitude >= 0 && c.Latitude <= 90 &&
		c.Longitude >= 0 && c.Longitude <= 180
}

// IsZero returns true if coordinates are at 0,0 (likely invalid/no fix)
func (c Coordinates) IsZero() bool {
	return c.Latitude == 0 && c.Longitude == 0
}

// DistanceTo calculates the distance to another coordinate in meters
// using the Haversine formula
func (c Coordinates) DistanceTo(other Coordinates) float64 {
	const earthRadius = 6371000.0 // meters

	lat1 := c.SignedLatitude() * math.Pi / 180
	lat2 := other.SignedLatitude() * math.Pi / 180
	lon1 := c.SignedLongitude() * math.Pi / 180
	lon2 := other.SignedLongitude() * math.Pi / 180

	dlat := lat2 - lat1
	dlon := lon2 - lon1

	a := math.Sin(dlat/2)*math.Sin(dlat/2) +
		math.Cos(lat1)*math.Cos(lat2)*
			math.Sin(dlon/2)*math.Sin(dlon/2)

	c2 := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c2
}

// ToDecimalDegrees returns latitude and longitude as signed decimal degrees
// Returns (lat, lon) where negative values indicate South/West
func (c Coordinates) ToDecimalDegrees() (lat, lon float64) {
	return c.SignedLatitude(), c.SignedLongitude()
}
