package types

import "fmt"

// CourseStatus represents the course (direction) and GPS status from 2 bytes
// BYTE_1: bit7-bit6(0), bit5(GPS Real-time/Differential), bit4(Positioned),
//
//	bit3(East/West), bit2(North/South), bit1-bit0(Course high bits)
//
// BYTE_2: bit7-bit0(Course low bits)
type CourseStatus struct {
	Course          uint16 // Direction in degrees (0-360°, 0=North, clockwise)
	IsGPSRealtime   bool   // true if GPS real-time positioning
	IsPositioned    bool   // true if GPS has valid fix
	IsEastLongitude bool   // true if East longitude
	IsNorthLatitude bool   // true if North latitude
}

// NewCourseStatusFromBytes creates CourseStatus from 2 bytes
func NewCourseStatusFromBytes(data []byte) (CourseStatus, error) {
	if len(data) < 2 {
		return CourseStatus{}, fmt.Errorf("course status requires 2 bytes, got %d", len(data))
	}

	byte1 := data[0]
	byte2 := data[1]

	// Extract course: bit1-bit0 of byte1 (high bits) + all bits of byte2 (low bits)
	// This gives us a 10-bit value (0-1023)
	courseHigh := uint16(byte1&0x03) << 8
	courseLow := uint16(byte2)
	course := courseHigh | courseLow

	// Ensure course is within 0-360 range
	if course > 360 {
		course = course % 360
	}

	return CourseStatus{
		Course:          course,
		IsGPSRealtime:   (byte1 & 0x20) != 0, // bit5
		IsPositioned:    (byte1 & 0x10) != 0, // bit4
		IsEastLongitude: (byte1 & 0x08) == 0, // bit3: 0=East, 1=West
		IsNorthLatitude: (byte1 & 0x04) != 0, // bit2: 1=North, 0=South
	}, nil
}

// MustNewCourseStatusFromBytes creates CourseStatus and panics if invalid
func MustNewCourseStatusFromBytes(data []byte) CourseStatus {
	cs, err := NewCourseStatusFromBytes(data)
	if err != nil {
		panic(err)
	}
	return cs
}

// NewCourseStatus creates a CourseStatus from individual fields
func NewCourseStatus(course uint16, isRealtime, isPositioned, isEast, isNorth bool) CourseStatus {
	// Ensure course is within 0-360 range
	if course > 360 {
		course = course % 360
	}

	return CourseStatus{
		Course:          course,
		IsGPSRealtime:   isRealtime,
		IsPositioned:    isPositioned,
		IsEastLongitude: isEast,
		IsNorthLatitude: isNorth,
	}
}

// Bytes returns the CourseStatus as 2 bytes for protocol encoding
func (c CourseStatus) Bytes() []byte {
	var byte1 byte

	// Set GPS realtime bit (bit5)
	if c.IsGPSRealtime {
		byte1 |= 0x20
	}

	// Set positioned bit (bit4)
	if c.IsPositioned {
		byte1 |= 0x10
	}

	// Set longitude hemisphere bit (bit3): 0=East, 1=West
	if !c.IsEastLongitude {
		byte1 |= 0x08
	}

	// Set latitude hemisphere bit (bit2): 1=North, 0=South
	if c.IsNorthLatitude {
		byte1 |= 0x04
	}

	// Set course high bits (bit1-bit0)
	courseHigh := byte((c.Course >> 8) & 0x03)
	byte1 |= courseHigh

	// Course low bits (all 8 bits of byte2)
	byte2 := byte(c.Course & 0xFF)

	return []byte{byte1, byte2}
}

// String returns a human-readable representation
func (c CourseStatus) String() string {
	direction := c.DirectionName()
	status := "No Fix"
	if c.IsPositioned {
		if c.IsGPSRealtime {
			status = "Real-time GPS"
		} else {
			status = "Differential GPS"
		}
	}

	hemisphere := ""
	if c.IsNorthLatitude {
		hemisphere += "N"
	} else {
		hemisphere += "S"
	}
	if c.IsEastLongitude {
		hemisphere += "E"
	} else {
		hemisphere += "W"
	}

	return fmt.Sprintf("%d° (%s) [%s, %s]", c.Course, direction, status, hemisphere)
}

// DirectionName returns the compass direction name
func (c CourseStatus) DirectionName() string {
	// Divide 360 degrees into 16 compass directions
	directions := []string{
		"N", "NNE", "NE", "ENE",
		"E", "ESE", "SE", "SSE",
		"S", "SSW", "SW", "WSW",
		"W", "WNW", "NW", "NNW",
	}

	// Calculate index (each direction covers 22.5 degrees)
	index := int((float64(c.Course) + 11.25) / 22.5)
	if index >= 16 {
		index = 0
	}

	return directions[index]
}

// GetIsNorthLatitude returns true if latitude is in Northern hemisphere
func (c CourseStatus) GetIsNorthLatitude() bool {
	return c.IsNorthLatitude
}

// GetIsEastLongitude returns true if longitude is in Eastern hemisphere
func (c CourseStatus) GetIsEastLongitude() bool {
	return c.IsEastLongitude
}

// GetCourse returns the course in degrees (0-360)
func (c CourseStatus) GetCourse() uint16 {
	return c.Course
}

// GetIsPositioned returns true if GPS has a valid fix
func (c CourseStatus) GetIsPositioned() bool {
	return c.IsPositioned
}

// GetIsGPSRealtime returns true if using real-time GPS positioning
func (c CourseStatus) GetIsGPSRealtime() bool {
	return c.IsGPSRealtime
}
