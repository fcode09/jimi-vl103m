package types

import (
	"fmt"
	"regexp"
	"strconv"
)

// IMEI represents a validated International Mobile Equipment Identity number
// IMEI is a 15-digit unique identifier for mobile devices
type IMEI struct {
	value string // 15 digits
}

var imeiRegex = regexp.MustCompile(`^\d{15}$`)

// NewIMEI creates a new IMEI from a string with validation
// The string must be exactly 15 digits
func NewIMEI(s string) (IMEI, error) {
	if !imeiRegex.MatchString(s) {
		return IMEI{}, fmt.Errorf("invalid IMEI format: must be 15 digits, got: %s", s)
	}

	// Optionally validate checksum (Luhn algorithm)
	if !validateIMEIChecksum(s) {
		return IMEI{}, fmt.Errorf("invalid IMEI checksum: %s", s)
	}

	return IMEI{value: s}, nil
}

// NewIMEIFromBytes creates an IMEI from BCD-encoded bytes (8 bytes)
// The VL103M protocol encodes IMEI as 8 bytes in BCD format
// Example: IMEI "123456789012345" → 0x01 0x23 0x45 0x67 0x89 0x01 0x23 0x45
func NewIMEIFromBytes(data []byte) (IMEI, error) {
	if len(data) != 8 {
		return IMEI{}, fmt.Errorf("IMEI bytes must be exactly 8 bytes, got %d", len(data))
	}

	// Convert BCD bytes to string
	var digits []byte
	for _, b := range data {
		highNibble := (b >> 4) & 0x0F
		lowNibble := b & 0x0F

		if highNibble > 9 || lowNibble > 9 {
			return IMEI{}, fmt.Errorf("invalid BCD encoding in IMEI bytes: 0x%02X", b)
		}

		digits = append(digits, '0'+highNibble)
		digits = append(digits, '0'+lowNibble)
	}

	// IMEI is 15 digits, but we got 16 from 8 bytes
	// Typically the last nibble is padding (0 or F)
	imeiStr := string(digits[:15])

	return NewIMEI(imeiStr)
}

// NewIMEIUnchecked creates an IMEI without checksum validation
// Use this when dealing with devices that don't follow Luhn algorithm
func NewIMEIUnchecked(s string) (IMEI, error) {
	if !imeiRegex.MatchString(s) {
		return IMEI{}, fmt.Errorf("invalid IMEI format: must be 15 digits, got: %s", s)
	}
	return IMEI{value: s}, nil
}

// NewIMEIFromBytesUnchecked creates an IMEI from BCD bytes without checksum validation
func NewIMEIFromBytesUnchecked(data []byte) (IMEI, error) {
	if len(data) != 8 {
		return IMEI{}, fmt.Errorf("IMEI bytes must be exactly 8 bytes, got %d", len(data))
	}

	var digits []byte
	for _, b := range data {
		highNibble := (b >> 4) & 0x0F
		lowNibble := b & 0x0F

		if highNibble > 9 || lowNibble > 9 {
			return IMEI{}, fmt.Errorf("invalid BCD encoding in IMEI bytes: 0x%02X", b)
		}

		digits = append(digits, '0'+highNibble)
		digits = append(digits, '0'+lowNibble)
	}

	imeiStr := string(digits[:15])
	return NewIMEIUnchecked(imeiStr)
}

// MustNewIMEI creates a new IMEI and panics if invalid
// Use this only when you're certain the IMEI is valid (e.g., in tests)
func MustNewIMEI(s string) IMEI {
	imei, err := NewIMEI(s)
	if err != nil {
		panic(err)
	}
	return imei
}

// String returns the IMEI as a 15-digit string
func (i IMEI) String() string {
	return i.value
}

// Bytes returns the IMEI in BCD-encoded format (8 bytes)
// Each byte contains two digits (high nibble and low nibble)
// The 16th nibble is set to 0 for padding
func (i IMEI) Bytes() []byte {
	bytes := make([]byte, 8)

	// Convert each pair of digits to a byte
	for j := 0; j < 7; j++ {
		high := i.value[j*2] - '0'
		low := i.value[j*2+1] - '0'
		bytes[j] = (high << 4) | low
	}

	// Last byte: 15th digit + padding 0
	lastDigit := i.value[14] - '0'
	bytes[7] = lastDigit << 4

	return bytes
}

// IsValid returns true if the IMEI is valid (non-empty and validated)
func (i IMEI) IsValid() bool {
	return i.value != ""
}

// TAC returns the Type Allocation Code (first 8 digits)
// TAC identifies the device manufacturer and model
func (i IMEI) TAC() string {
	if !i.IsValid() {
		return ""
	}
	return i.value[:8]
}

// SNR returns the Serial Number (next 6 digits after TAC)
func (i IMEI) SNR() string {
	if !i.IsValid() {
		return ""
	}
	return i.value[8:14]
}

// CheckDigit returns the Luhn check digit (last digit)
func (i IMEI) CheckDigit() int {
	if !i.IsValid() {
		return 0
	}
	digit, _ := strconv.Atoi(string(i.value[14]))
	return digit
}

// validateIMEIChecksum validates the IMEI using the Luhn algorithm
func validateIMEIChecksum(imei string) bool {
	if len(imei) != 15 {
		return false
	}

	sum := 0
	for i := 0; i < 14; i++ {
		digit := int(imei[i] - '0')

		// Double every second digit (from right to left, so odd positions from left)
		if i%2 == 1 {
			digit *= 2
			if digit > 9 {
				digit = digit/10 + digit%10
			}
		}

		sum += digit
	}

	// Calculate check digit
	checkDigit := (10 - (sum % 10)) % 10
	expectedCheckDigit := int(imei[14] - '0')

	return checkDigit == expectedCheckDigit
}
