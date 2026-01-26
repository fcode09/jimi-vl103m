package codec

import "fmt"

// BCD (Binary-Coded Decimal) encoding/decoding
// Used for IMEI, ICCID, and other numeric fields in VL103M protocol

// DecodeBCD converts BCD-encoded bytes to a decimal string
// Each byte contains two decimal digits (high nibble and low nibble)
// Example: 0x12 0x34 -> "1234"
func DecodeBCD(data []byte) (string, error) {
	if len(data) == 0 {
		return "", nil
	}

	result := make([]byte, 0, len(data)*2)

	for i, b := range data {
		high := (b >> 4) & 0x0F
		low := b & 0x0F

		// Validate BCD digits (must be 0-9)
		if high > 9 {
			return "", fmt.Errorf("invalid BCD digit at byte %d (high nibble): 0x%X", i, high)
		}
		if low > 9 {
			return "", fmt.Errorf("invalid BCD digit at byte %d (low nibble): 0x%X", i, low)
		}

		result = append(result, '0'+high)
		result = append(result, '0'+low)
	}

	return string(result), nil
}

// DecodeBCDTrimmed decodes BCD and trims trailing zeros or padding
// Useful for variable-length BCD fields
func DecodeBCDTrimmed(data []byte) (string, error) {
	str, err := DecodeBCD(data)
	if err != nil {
		return "", err
	}

	// Trim trailing zeros
	for len(str) > 0 && str[len(str)-1] == '0' {
		str = str[:len(str)-1]
	}

	return str, nil
}

// DecodeBCDLength decodes BCD to a specific length
// Useful when you know the expected number of digits (e.g., IMEI is 15 digits)
func DecodeBCDLength(data []byte, length int) (string, error) {
	str, err := DecodeBCD(data)
	if err != nil {
		return "", err
	}

	if len(str) < length {
		return "", fmt.Errorf("BCD string too short: expected %d digits, got %d", length, len(str))
	}

	return str[:length], nil
}

// EncodeBCD converts a decimal string to BCD-encoded bytes
// The string must contain only digits 0-9
// Example: "1234" -> []byte{0x12, 0x34}
func EncodeBCD(str string) ([]byte, error) {
	// Validate input contains only digits
	for i, c := range str {
		if c < '0' || c > '9' {
			return nil, fmt.Errorf("invalid character at position %d: '%c' (must be 0-9)", i, c)
		}
	}

	// Pad with trailing zero if odd length
	if len(str)%2 != 0 {
		str = str + "0"
	}

	result := make([]byte, len(str)/2)

	for i := 0; i < len(str); i += 2 {
		high := str[i] - '0'
		low := str[i+1] - '0'
		result[i/2] = (high << 4) | low
	}

	return result, nil
}

// EncodeBCDFixed encodes a string to BCD with fixed output size
// Pads with zeros if the string is shorter than required
// Truncates if the string is longer than required
func EncodeBCDFixed(str string, byteLength int) ([]byte, error) {
	// Validate input
	for i, c := range str {
		if c < '0' || c > '9' {
			return nil, fmt.Errorf("invalid character at position %d: '%c'", i, c)
		}
	}

	// Pad or truncate to exact length
	digitLength := byteLength * 2
	if len(str) > digitLength {
		str = str[:digitLength]
	} else if len(str) < digitLength {
		// Pad with trailing zeros
		str = str + string(make([]byte, digitLength-len(str)))
		for i := len(str) - (digitLength - len(str)); i < len(str); i++ {
			str = str[:i] + "0" + str[i:]
		}
	}

	return EncodeBCD(str)
}

// DecodeIMEI decodes an IMEI from 8 BCD bytes (16 digits)
// Returns only the first 15 digits (last digit is padding)
func DecodeIMEI(data []byte) (string, error) {
	if len(data) != 8 {
		return "", fmt.Errorf("IMEI must be exactly 8 bytes, got %d", len(data))
	}

	str, err := DecodeBCD(data)
	if err != nil {
		return "", err
	}

	if len(str) < 15 {
		return "", fmt.Errorf("invalid IMEI: too short")
	}

	return str[:15], nil
}

// EncodeIMEI encodes a 15-digit IMEI to 8 BCD bytes
// Pads the 16th position with 0
func EncodeIMEI(imei string) ([]byte, error) {
	if len(imei) != 15 {
		return nil, fmt.Errorf("IMEI must be exactly 15 digits, got %d", len(imei))
	}

	// Validate all digits
	for i, c := range imei {
		if c < '0' || c > '9' {
			return nil, fmt.Errorf("invalid IMEI character at position %d: '%c'", i, c)
		}
	}

	// Add padding digit
	padded := imei + "0"
	return EncodeBCD(padded)
}

// DecodeICCID decodes an ICCID from 10 BCD bytes (20 digits)
// ICCID is typically 19-20 digits
func DecodeICCID(data []byte) (string, error) {
	if len(data) != 10 {
		return "", fmt.Errorf("ICCID must be exactly 10 bytes, got %d", len(data))
	}

	return DecodeBCD(data)
}

// EncodeICCID encodes a 20-digit ICCID to 10 BCD bytes
func EncodeICCID(iccid string) ([]byte, error) {
	if len(iccid) != 20 {
		return nil, fmt.Errorf("ICCID must be exactly 20 digits, got %d", len(iccid))
	}

	return EncodeBCD(iccid)
}

// DecodeIMSI decodes an IMSI from 8 BCD bytes
func DecodeIMSI(data []byte) (string, error) {
	if len(data) != 8 {
		return "", fmt.Errorf("IMSI must be exactly 8 bytes, got %d", len(data))
	}

	return DecodeBCD(data)
}

// EncodeIMSI encodes an IMSI to 8 BCD bytes
// IMSI is typically 15 digits, so we pad to 16
func EncodeIMSI(imsi string) ([]byte, error) {
	if len(imsi) < 14 || len(imsi) > 15 {
		return nil, fmt.Errorf("IMSI must be 14-15 digits, got %d", len(imsi))
	}

	// Pad to 16 digits if needed
	if len(imsi) == 14 {
		imsi = imsi + "00"
	} else if len(imsi) == 15 {
		imsi = imsi + "0"
	}

	return EncodeBCD(imsi)
}

// IsBCDValid checks if a byte is a valid BCD byte (both nibbles 0-9)
func IsBCDValid(b byte) bool {
	high := (b >> 4) & 0x0F
	low := b & 0x0F
	return high <= 9 && low <= 9
}

// ValidateBCD validates that all bytes in the slice are valid BCD
func ValidateBCD(data []byte) error {
	for i, b := range data {
		if !IsBCDValid(b) {
			return fmt.Errorf("invalid BCD byte at position %d: 0x%02X", i, b)
		}
	}
	return nil
}
