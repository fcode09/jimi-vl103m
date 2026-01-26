package codec

import (
	"fmt"
	"time"
)

// DateTime encoding/decoding for VL103M protocol
// Protocol uses 6 bytes: YY MM DD HH MM SS (each as decimal byte)

// DecodeDateTime decodes 6 bytes to time.Time (UTC)
// Format: YY MM DD HH MM SS
// YY: year offset from 2000 (e.g., 0x0F = 2015)
// All values are in decimal (not BCD)
func DecodeDateTime(data []byte) (time.Time, error) {
	if len(data) < 6 {
		return time.Time{}, fmt.Errorf("datetime requires 6 bytes, got %d", len(data))
	}

	year := 2000 + int(data[0])
	month := int(data[1])
	day := int(data[2])
	hour := int(data[3])
	minute := int(data[4])
	second := int(data[5])

	// Validate ranges
	if month < 1 || month > 12 {
		return time.Time{}, fmt.Errorf("invalid month: %d", month)
	}
	if day < 1 || day > 31 {
		return time.Time{}, fmt.Errorf("invalid day: %d", day)
	}
	if hour > 23 {
		return time.Time{}, fmt.Errorf("invalid hour: %d", hour)
	}
	if minute > 59 {
		return time.Time{}, fmt.Errorf("invalid minute: %d", minute)
	}
	if second > 59 {
		return time.Time{}, fmt.Errorf("invalid second: %d", second)
	}

	// Create time in UTC
	t := time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC)

	return t, nil
}

// EncodeDateTime encodes time.Time to 6 bytes (UTC)
// Returns: YY MM DD HH MM SS
func EncodeDateTime(t time.Time) []byte {
	// Convert to UTC if not already
	t = t.UTC()

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

// DecodeDateTimeWithLocation decodes datetime and applies a timezone offset
// timezoneOffset: offset in minutes (e.g., 480 for UTC+8)
func DecodeDateTimeWithLocation(data []byte, timezoneOffset int) (time.Time, error) {
	t, err := DecodeDateTime(data)
	if err != nil {
		return time.Time{}, err
	}

	// Apply timezone offset
	t = t.Add(time.Duration(timezoneOffset) * time.Minute)

	return t, nil
}

// EncodeDateTimeWithLocation encodes datetime from a specific timezone
// The time is converted to UTC before encoding
func EncodeDateTimeWithLocation(t time.Time) []byte {
	return EncodeDateTime(t.UTC())
}

// IsValidDateTime checks if the 6 bytes represent a valid datetime
func IsValidDateTime(data []byte) bool {
	_, err := DecodeDateTime(data)
	return err == nil
}

// ParseDateTimeString parses a string in format "2006-01-02 15:04:05" to protocol bytes
func ParseDateTimeString(s string) ([]byte, error) {
	t, err := time.Parse("2006-01-02 15:04:05", s)
	if err != nil {
		return nil, fmt.Errorf("invalid datetime format: %v", err)
	}

	return EncodeDateTime(t), nil
}

// FormatDateTime formats the 6-byte datetime to a string
func FormatDateTime(data []byte) (string, error) {
	t, err := DecodeDateTime(data)
	if err != nil {
		return "", err
	}

	return t.Format("2006-01-02 15:04:05"), nil
}

// DecodeTimezone decodes the 2-byte timezone/language field
// Returns: timezone offset in minutes, language code, error
//
// Format (2 bytes):
// Bit15-Bit4: Timezone value (expanded by 100)
// Bit3: 0=East, 1=West
// Bit2: Undefined
// Bit1-Bit0: Language (0x00, 0x01=Chinese, 0x02=English)
func DecodeTimezone(data []byte) (timezoneMinutes int, language byte, err error) {
	if len(data) < 2 {
		return 0, 0, fmt.Errorf("timezone requires 2 bytes, got %d", len(data))
	}

	// Read as uint16 big-endian
	value := uint16(data[0])<<8 | uint16(data[1])

	// Extract timezone (bit15-bit4)
	timezoneValue := (value >> 4) & 0x0FFF

	// Extract direction bit (bit3): 0=East (positive), 1=West (negative)
	isWest := (value & 0x0008) != 0

	// Extract language (bit1-bit0)
	language = byte(value & 0x0003)

	// Calculate timezone in minutes
	// Protocol uses format: (hours * 100 + minutes) for the value
	// Example: GMT+8:00 = 800 = 0x0320
	// Example: GMT-12:45 = 1245 = 0x04DD
	hours := int(timezoneValue / 100)
	minutes := int(timezoneValue % 100)
	timezoneMinutes = hours*60 + minutes

	// Apply direction
	if isWest {
		timezoneMinutes = -timezoneMinutes
	}

	return timezoneMinutes, language, nil
}

// EncodeTimezone encodes timezone offset and language to 2 bytes
// timezoneMinutes: timezone offset in minutes (e.g., 480 for UTC+8, -720 for UTC-12)
// language: language code (0x01=Chinese, 0x02=English)
func EncodeTimezone(timezoneMinutes int, language byte) []byte {
	var value uint16

	// Determine if West (negative offset)
	isWest := timezoneMinutes < 0
	if isWest {
		timezoneMinutes = -timezoneMinutes
	}

	// Convert minutes to protocol format (hours * 100 + minutes)
	hours := timezoneMinutes / 60
	mins := timezoneMinutes % 60
	timezoneValue := uint16(hours*100 + mins)

	// Build the 16-bit value
	value = (timezoneValue << 4) // Timezone in bits 15-4

	if isWest {
		value |= 0x0008 // Set bit 3 for West
	}

	value |= uint16(language & 0x03) // Language in bits 1-0

	return []byte{
		byte(value >> 8),
		byte(value & 0xFF),
	}
}

// GetTimezoneString converts timezone minutes to string format
// Examples: 480 -> "+08:00", -720 -> "-12:00"
func GetTimezoneString(timezoneMinutes int) string {
	sign := "+"
	if timezoneMinutes < 0 {
		sign = "-"
		timezoneMinutes = -timezoneMinutes
	}

	hours := timezoneMinutes / 60
	mins := timezoneMinutes % 60

	return fmt.Sprintf("%s%02d:%02d", sign, hours, mins)
}

// ParseTimezoneString parses timezone string to minutes
// Examples: "+08:00" -> 480, "-12:00" -> -720
func ParseTimezoneString(tz string) (int, error) {
	if len(tz) < 6 {
		return 0, fmt.Errorf("invalid timezone format: %s", tz)
	}

	sign := tz[0]
	if sign != '+' && sign != '-' {
		return 0, fmt.Errorf("timezone must start with + or -: %s", tz)
	}

	var hours, mins int
	_, err := fmt.Sscanf(tz[1:], "%02d:%02d", &hours, &mins)
	if err != nil {
		return 0, fmt.Errorf("invalid timezone format: %s: %v", tz, err)
	}

	totalMinutes := hours*60 + mins

	if sign == '-' {
		totalMinutes = -totalMinutes
	}

	return totalMinutes, nil
}
