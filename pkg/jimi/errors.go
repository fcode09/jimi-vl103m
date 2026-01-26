package jimi

import (
	"errors"
	"fmt"
)

// Common errors returned by the decoder
var (
	// ErrInvalidPacketSize indicates the packet is too small to be valid
	ErrInvalidPacketSize = errors.New("invalid packet size: packet too small")

	// ErrInvalidStartBit indicates the packet doesn't start with valid start bits
	ErrInvalidStartBit = errors.New("invalid start bit: expected 0x7878 or 0x7979")

	// ErrInvalidStopBit indicates the packet doesn't end with valid stop bits
	ErrInvalidStopBit = errors.New("invalid stop bit: expected 0x0D0A")

	// ErrInvalidCRC indicates CRC validation failed
	ErrInvalidCRC = errors.New("invalid CRC: checksum mismatch")

	// ErrUnsupportedProtocol indicates the protocol number is not supported
	ErrUnsupportedProtocol = errors.New("unsupported protocol number")

	// ErrInvalidPacketLength indicates the packet length field is inconsistent
	ErrInvalidPacketLength = errors.New("invalid packet length: length mismatch")

	// ErrInsufficientData indicates not enough data to parse the packet
	ErrInsufficientData = errors.New("insufficient data: incomplete packet")

	// ErrInvalidData indicates the packet data is malformed or invalid
	ErrInvalidData = errors.New("invalid data: packet is malformed")

	// ErrBufferOverflow indicates the packet exceeds maximum size
	ErrBufferOverflow = errors.New("buffer overflow: packet too large")
)

// DecodeError represents a packet decoding error with additional context
type DecodeError struct {
	Protocol byte   // Protocol number that failed
	Offset   int    // Byte offset where error occurred
	Reason   string // Human-readable reason
	Err      error  // Underlying error
}

// Error implements the error interface
func (e *DecodeError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("decode error at offset %d (protocol 0x%02X): %s: %v",
			e.Offset, e.Protocol, e.Reason, e.Err)
	}
	return fmt.Sprintf("decode error at offset %d (protocol 0x%02X): %s",
		e.Offset, e.Protocol, e.Reason)
}

// Unwrap returns the underlying error
func (e *DecodeError) Unwrap() error {
	return e.Err
}

// NewDecodeError creates a new DecodeError
func NewDecodeError(protocol byte, offset int, reason string, err error) *DecodeError {
	return &DecodeError{
		Protocol: protocol,
		Offset:   offset,
		Reason:   reason,
		Err:      err,
	}
}

// ValidationError represents a packet validation error
type ValidationError struct {
	Field  string // Field that failed validation
	Value  any    // Actual value
	Reason string // Why validation failed
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for %s: %s (value: %v)",
		e.Field, e.Reason, e.Value)
}

// NewValidationError creates a new ValidationError
func NewValidationError(field, reason string, value any) *ValidationError {
	return &ValidationError{
		Field:  field,
		Value:  value,
		Reason: reason,
	}
}

// CRCError represents a CRC validation error with details
type CRCError struct {
	Expected   uint16 // Expected CRC value
	Received   uint16 // Received CRC value
	PacketSize int    // Size of the packet
}

// Error implements the error interface
func (e *CRCError) Error() string {
	return fmt.Sprintf("CRC mismatch: expected 0x%04X, got 0x%04X (packet size: %d bytes)",
		e.Expected, e.Received, e.PacketSize)
}

// NewCRCError creates a new CRCError
func NewCRCError(expected, received uint16, packetSize int) *CRCError {
	return &CRCError{
		Expected:   expected,
		Received:   received,
		PacketSize: packetSize,
	}
}

// ProtocolError represents an unsupported or invalid protocol error
type ProtocolError struct {
	Protocol byte   // Protocol number
	Message  string // Error message
}

// Error implements the error interface
func (e *ProtocolError) Error() string {
	return fmt.Sprintf("protocol 0x%02X: %s", e.Protocol, e.Message)
}

// NewProtocolError creates a new ProtocolError
func NewProtocolError(protocol byte, message string) *ProtocolError {
	return &ProtocolError{
		Protocol: protocol,
		Message:  message,
	}
}

// Helper functions for error checking

// IsInvalidCRC returns true if the error is a CRC error
func IsInvalidCRC(err error) bool {
	if err == nil {
		return false
	}
	var crcErr *CRCError
	if errors.As(err, &crcErr) {
		return true
	}
	return errors.Is(err, ErrInvalidCRC)
}

// IsUnsupportedProtocol returns true if the error is due to unsupported protocol
func IsUnsupportedProtocol(err error) bool {
	if err == nil {
		return false
	}
	var protocolErr *ProtocolError
	if errors.As(err, &protocolErr) {
		return true
	}
	return errors.Is(err, ErrUnsupportedProtocol)
}

// IsValidationError returns true if the error is a validation error
func IsValidationError(err error) bool {
	if err == nil {
		return false
	}
	var valErr *ValidationError
	return errors.As(err, &valErr)
}

// IsDecodeError returns true if the error is a decode error
func IsDecodeError(err error) bool {
	if err == nil {
		return false
	}
	var decErr *DecodeError
	return errors.As(err, &decErr)
}
