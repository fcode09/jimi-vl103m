package codec

import (
	"encoding/binary"
	"fmt"
)

// Binary encoding/decoding helpers for VL103M protocol

// ReadUint16BE reads a big-endian uint16 from 2 bytes
func ReadUint16BE(data []byte) uint16 {
	if len(data) < 2 {
		return 0
	}
	return binary.BigEndian.Uint16(data)
}

// ReadUint32BE reads a big-endian uint32 from 4 bytes
func ReadUint32BE(data []byte) uint32 {
	if len(data) < 4 {
		return 0
	}
	return binary.BigEndian.Uint32(data)
}

// ReadUint64BE reads a big-endian uint64 from 8 bytes
func ReadUint64BE(data []byte) uint64 {
	if len(data) < 8 {
		return 0
	}
	return binary.BigEndian.Uint64(data)
}

// WriteUint16BE writes a uint16 as big-endian to 2 bytes
func WriteUint16BE(value uint16) []byte {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, value)
	return buf
}

// WriteUint32BE writes a uint32 as big-endian to 4 bytes
func WriteUint32BE(value uint32) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, value)
	return buf
}

// WriteUint64BE writes a uint64 as big-endian to 8 bytes
func WriteUint64BE(value uint64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, value)
	return buf
}

// ReadUint24BE reads a 24-bit big-endian value (3 bytes) as uint32
// Used for Cell ID in 2G/3G packets
func ReadUint24BE(data []byte) uint32 {
	if len(data) < 3 {
		return 0
	}
	return uint32(data[0])<<16 | uint32(data[1])<<8 | uint32(data[2])
}

// WriteUint24BE writes a uint32 as 24-bit big-endian (3 bytes)
// Only the lower 24 bits are written
func WriteUint24BE(value uint32) []byte {
	return []byte{
		byte(value >> 16),
		byte(value >> 8),
		byte(value),
	}
}

// ReadNibbles reads high and low nibbles from a byte
// Returns (highNibble, lowNibble)
func ReadNibbles(b byte) (high, low byte) {
	high = (b >> 4) & 0x0F
	low = b & 0x0F
	return
}

// WriteNibbles combines high and low nibbles into a byte
func WriteNibbles(high, low byte) byte {
	return (high << 4) | (low & 0x0F)
}

// IsBitSet returns true if the specified bit is set (1)
// bit: bit position (0-7, where 0 is LSB)
func IsBitSet(value byte, bit uint) bool {
	if bit > 7 {
		return false
	}
	return (value & (1 << bit)) != 0
}

// SetBit sets the specified bit to 1
func SetBit(value byte, bit uint) byte {
	if bit > 7 {
		return value
	}
	return value | (1 << bit)
}

// ClearBit sets the specified bit to 0
func ClearBit(value byte, bit uint) byte {
	if bit > 7 {
		return value
	}
	return value &^ (1 << bit)
}

// ToggleBit toggles the specified bit
func ToggleBit(value byte, bit uint) byte {
	if bit > 7 {
		return value
	}
	return value ^ (1 << bit)
}

// GetBits extracts bits from a value
// start: starting bit position (0-based, LSB = 0)
// length: number of bits to extract
func GetBits(value byte, start, length uint) byte {
	if start > 7 || length > 8 || start+length > 8 {
		return 0
	}
	mask := byte((1 << length) - 1)
	return (value >> start) & mask
}

// SetBits sets specific bits in a value
// start: starting bit position
// length: number of bits to set
// bits: the bits to set (only lower 'length' bits are used)
func SetBits(value byte, start, length uint, bits byte) byte {
	if start > 7 || length > 8 || start+length > 8 {
		return value
	}
	mask := byte((1 << length) - 1)
	bits = bits & mask
	value = value &^ (mask << start) // Clear the bits
	value = value | (bits << start)  // Set the bits
	return value
}

// ExtractBitfield extracts a multi-byte bitfield
// Useful for parsing complex protocol fields
type BitReader struct {
	data []byte
	pos  uint // Current bit position
}

// NewBitReader creates a new bit reader
func NewBitReader(data []byte) *BitReader {
	return &BitReader{
		data: data,
		pos:  0,
	}
}

// ReadBits reads the specified number of bits
func (r *BitReader) ReadBits(n uint) uint64 {
	if n > 64 {
		return 0
	}

	var result uint64
	for i := uint(0); i < n; i++ {
		bytePos := r.pos / 8
		bitPos := 7 - (r.pos % 8) // MSB first

		if int(bytePos) >= len(r.data) {
			return result
		}

		bit := (r.data[bytePos] >> bitPos) & 1
		result = (result << 1) | uint64(bit)
		r.pos++
	}

	return result
}

// Position returns the current bit position
func (r *BitReader) Position() uint {
	return r.pos
}

// Reset resets the reader to the beginning
func (r *BitReader) Reset() {
	r.pos = 0
}

// Remaining returns the number of bits remaining
func (r *BitReader) Remaining() uint {
	totalBits := uint(len(r.data) * 8)
	if r.pos >= totalBits {
		return 0
	}
	return totalBits - r.pos
}

// HexToBytes converts a hex string to bytes
// Example: "7878" -> []byte{0x78, 0x78}
func HexToBytes(hex string) ([]byte, error) {
	if len(hex)%2 != 0 {
		hex = "0" + hex // Pad with leading zero
	}

	bytes := make([]byte, len(hex)/2)
	for i := 0; i < len(hex); i += 2 {
		var b byte
		_, err := fmt.Sscanf(hex[i:i+2], "%02x", &b)
		if err != nil {
			return nil, err
		}
		bytes[i/2] = b
	}
	return bytes, nil
}

// BytesToHex converts bytes to hex string
// Example: []byte{0x78, 0x78} -> "7878"
func BytesToHex(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	hex := make([]byte, len(data)*2)
	const hexDigits = "0123456789ABCDEF"
	for i, b := range data {
		hex[i*2] = hexDigits[b>>4]
		hex[i*2+1] = hexDigits[b&0x0F]
	}
	return string(hex)
}
