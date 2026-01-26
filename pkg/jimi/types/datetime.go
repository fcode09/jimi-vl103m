package types

import (
	"fmt"
	"time"
)

// DateTime represents a timestamp from a VL103M device
// The protocol uses 6 bytes: YY MM DD HH MM SS
type DateTime struct {
	Time time.Time
}

// NewDateTime creates a DateTime from a time.Time value
func NewDateTime(t time.Time) DateTime {
	return DateTime{Time: t}
}

// Now returns the current time as a DateTime
func Now() DateTime {
	return DateTime{Time: time.Now().UTC()}
}

// DateTimeFromBytes decodes a DateTime from 6 protocol bytes
// Format: YY MM DD HH MM SS (decimal values, not BCD)
func DateTimeFromBytes(data []byte) (DateTime, error) {
	if len(data) < 6 {
		return DateTime{}, fmt.Errorf("datetime requires 6 bytes, got %d", len(data))
	}

	year := 2000 + int(data[0])
	month := int(data[1])
	day := int(data[2])
	hour := int(data[3])
	minute := int(data[4])
	second := int(data[5])

	// Validate ranges
	if month < 1 || month > 12 {
		return DateTime{}, fmt.Errorf("invalid month: %d", month)
	}
	if day < 1 || day > 31 {
		return DateTime{}, fmt.Errorf("invalid day: %d", day)
	}
	if hour > 23 {
		return DateTime{}, fmt.Errorf("invalid hour: %d", hour)
	}
	if minute > 59 {
		return DateTime{}, fmt.Errorf("invalid minute: %d", minute)
	}
	if second > 59 {
		return DateTime{}, fmt.Errorf("invalid second: %d", second)
	}

	t := time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC)
	return DateTime{Time: t}, nil
}

// ToBytes encodes the DateTime to 6 protocol bytes
func (dt DateTime) ToBytes() []byte {
	t := dt.Time.UTC()

	year := t.Year() - 2000
	if year < 0 {
		year = 0
	} else if year > 255 {
		year = 255
	}

	return []byte{
		byte(year),
		byte(t.Month()),
		byte(t.Day()),
		byte(t.Hour()),
		byte(t.Minute()),
		byte(t.Second()),
	}
}

// String returns the datetime in ISO format
func (dt DateTime) String() string {
	return dt.Time.Format("2006-01-02 15:04:05 UTC")
}

// Format returns the datetime using a custom format
func (dt DateTime) Format(layout string) string {
	return dt.Time.Format(layout)
}

// IsZero returns true if the datetime is the zero value
func (dt DateTime) IsZero() bool {
	return dt.Time.IsZero()
}

// Unix returns the Unix timestamp
func (dt DateTime) Unix() int64 {
	return dt.Time.Unix()
}

// Before returns true if dt is before other
func (dt DateTime) Before(other DateTime) bool {
	return dt.Time.Before(other.Time)
}

// After returns true if dt is after other
func (dt DateTime) After(other DateTime) bool {
	return dt.Time.After(other.Time)
}

// Add returns a new DateTime with the given duration added
func (dt DateTime) Add(d time.Duration) DateTime {
	return DateTime{Time: dt.Time.Add(d)}
}

// Sub returns the duration between two DateTimes
func (dt DateTime) Sub(other DateTime) time.Duration {
	return dt.Time.Sub(other.Time)
}

// InLocation returns the DateTime in the specified timezone
func (dt DateTime) InLocation(loc *time.Location) DateTime {
	return DateTime{Time: dt.Time.In(loc)}
}

// WithTimezoneOffset applies a timezone offset in minutes
func (dt DateTime) WithTimezoneOffset(offsetMinutes int) DateTime {
	loc := time.FixedZone("", offsetMinutes*60)
	return DateTime{Time: dt.Time.In(loc)}
}

// Timezone represents the timezone/language field from the protocol
type Timezone struct {
	OffsetMinutes int  // Timezone offset in minutes (e.g., 480 for UTC+8)
	Language      byte // Language code (0x01=Chinese, 0x02=English)
}

// TimezoneFromBytes decodes the 2-byte timezone/language field
//
// Format (2 bytes):
// Bit15-Bit4: Timezone value (expanded by 100, e.g., 800 for +08:00)
// Bit3: Direction (0=East, 1=West)
// Bit2: Undefined
// Bit1-Bit0: Language (0x01=Chinese, 0x02=English)
func TimezoneFromBytes(data []byte) (Timezone, error) {
	if len(data) < 2 {
		return Timezone{}, fmt.Errorf("timezone requires 2 bytes, got %d", len(data))
	}

	value := uint16(data[0])<<8 | uint16(data[1])

	// Extract timezone (bit15-bit4)
	timezoneValue := (value >> 4) & 0x0FFF

	// Extract direction bit (bit3): 0=East (positive), 1=West (negative)
	isWest := (value & 0x0008) != 0

	// Extract language (bit1-bit0)
	language := byte(value & 0x0003)

	// Convert to minutes: value = hours*100 + minutes
	hours := int(timezoneValue / 100)
	minutes := int(timezoneValue % 100)
	offsetMinutes := hours*60 + minutes

	if isWest {
		offsetMinutes = -offsetMinutes
	}

	return Timezone{
		OffsetMinutes: offsetMinutes,
		Language:      language,
	}, nil
}

// ToBytes encodes the Timezone to 2 bytes
func (tz Timezone) ToBytes() []byte {
	var value uint16

	offsetMinutes := tz.OffsetMinutes
	isWest := offsetMinutes < 0
	if isWest {
		offsetMinutes = -offsetMinutes
	}

	// Convert to protocol format
	hours := offsetMinutes / 60
	mins := offsetMinutes % 60
	timezoneValue := uint16(hours*100 + mins)

	value = (timezoneValue << 4)

	if isWest {
		value |= 0x0008
	}

	value |= uint16(tz.Language & 0x03)

	return []byte{
		byte(value >> 8),
		byte(value & 0xFF),
	}
}

// String returns the timezone as a string (e.g., "+08:00")
func (tz Timezone) String() string {
	sign := "+"
	offset := tz.OffsetMinutes
	if offset < 0 {
		sign = "-"
		offset = -offset
	}

	hours := offset / 60
	mins := offset % 60

	return fmt.Sprintf("UTC%s%02d:%02d", sign, hours, mins)
}

// Location returns a time.Location for this timezone
func (tz Timezone) Location() *time.Location {
	return time.FixedZone(tz.String(), tz.OffsetMinutes*60)
}

// LanguageString returns the language name
func (tz Timezone) LanguageString() string {
	switch tz.Language {
	case 0x01:
		return "Chinese"
	case 0x02:
		return "English"
	default:
		return "Unknown"
	}
}
