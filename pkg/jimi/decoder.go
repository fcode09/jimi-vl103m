package jimi

import (
	"fmt"
	"time"

	"github.com/intelcon-group/jimi-vl103m/internal/parser"
	"github.com/intelcon-group/jimi-vl103m/internal/splitter"
	"github.com/intelcon-group/jimi-vl103m/internal/validator"
	"github.com/intelcon-group/jimi-vl103m/pkg/jimi/packet"
	"github.com/intelcon-group/jimi-vl103m/pkg/jimi/protocol"
)

// Decoder is the main entry point for decoding VL103M protocol packets
type Decoder struct {
	opts     Options
	registry *parser.Registry
}

// NewDecoder creates a new decoder with optional configuration
//
// Example:
//
//	decoder := jimi.NewDecoder()  // Default options
//	decoder := jimi.NewDecoder(jimi.WithStrictMode(false))
//	decoder := jimi.NewDecoder(jimi.WithSkipCRC(), jimi.WithLogging())
func NewDecoder(opts ...Option) *Decoder {
	options := DefaultOptions()

	// Apply functional options
	for _, opt := range opts {
		opt(&options)
	}

	return &Decoder{
		opts:     options,
		registry: parser.DefaultRegistry(),
	}
}

// NewDecoderWithRegistry creates a decoder with a custom parser registry
func NewDecoderWithRegistry(registry *parser.Registry, opts ...Option) *Decoder {
	options := DefaultOptions()

	for _, opt := range opts {
		opt(&options)
	}

	return &Decoder{
		opts:     options,
		registry: registry,
	}
}

// Decode decodes a single complete packet
//
// The packet must be a complete packet with start bit, length, protocol,
// content, serial number, CRC, and stop bit.
//
// Returns:
//   - packet.Packet: The decoded packet (specific type depends on protocol)
//   - error: Any error encountered during decoding
//
// Example:
//
//	data, _ := hex.DecodeString("787822220F0C1D023305C9027AC8...")
//	pkt, err := decoder.Decode(data)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	if loc, ok := pkt.(*packet.LocationPacket); ok {
//	    fmt.Printf("Location: %s\n", loc.Coordinates)
//	}
func (d *Decoder) Decode(data []byte) (packet.Packet, error) {
	if len(data) < protocol.MinPacketSize {
		return nil, ErrInvalidPacketSize
	}

	// Validate structure (start bit, stop bit, length)
	if !d.opts.SkipStructureValidation {
		if err := d.validateStructure(data); err != nil {
			return nil, err
		}
	}

	// Validate CRC
	if !d.opts.SkipCRCValidation {
		if !validator.ValidateCRC(data) {
			received, calculated, _ := validator.VerifyPacketCRC(data)
			return nil, NewCRCError(calculated, received, len(data))
		}
	}

	// Extract protocol number
	protocolNum, err := splitter.GetPacketType(data)
	if err != nil {
		return nil, err
	}

	// Try to use registered parser
	if d.registry != nil && d.registry.Has(protocolNum) {
		pkt, parseErr := d.registry.Parse(protocolNum, data)
		if parseErr != nil {
			if d.opts.StrictMode {
				return nil, fmt.Errorf("failed to parse protocol 0x%02X: %w", protocolNum, parseErr)
			}
			// Fall through to return base packet in lenient mode
		} else {
			return pkt, nil
		}
	}

	// No parser registered or parse failed in lenient mode
	// Check if we should reject unknown protocols
	if !d.opts.AllowUnknownProtocols && (d.registry == nil || !d.registry.Has(protocolNum)) {
		return nil, NewProtocolError(protocolNum, "no parser registered for this protocol")
	}

	// Return a base packet for unknown protocols
	serialNum, _ := splitter.GetSerialNumber(data)

	basePacket := &packet.BasePacket{
		ProtocolNum: protocolNum,
		SerialNum:   serialNum,
		RawData:     data,
		ParsedAt:    time.Now(),
	}

	return basePacket, nil
}

// DecodeStream decodes packets from a TCP stream
//
// This method handles the common scenario where multiple packets are
// concatenated in a single TCP read, or packets are fragmented across
// multiple reads.
//
// Returns:
//   - packets: All complete packets found in the stream
//   - residue: Incomplete packet data to prepend to next read
//   - error: Any error encountered
//
// Example:
//
//	buffer := make([]byte, 0, 4096)
//	readBuf := make([]byte, 1024)
//
//	for {
//	    n, err := conn.Read(readBuf)
//	    if err != nil {
//	        return
//	    }
//
//	    buffer = append(buffer, readBuf[:n]...)
//
//	    packets, residue, err := decoder.DecodeStream(buffer)
//	    if err != nil {
//	        log.Printf("Decode error: %v", err)
//	        continue
//	    }
//
//	    buffer = residue
//
//	    for _, pkt := range packets {
//	        processPacket(pkt)
//	    }
//	}
func (d *Decoder) DecodeStream(stream []byte) (packets []packet.Packet, residue []byte, err error) {
	// Split the stream into individual packets
	rawPackets, residue, err := splitter.SplitPackets(stream)
	if err != nil {
		// If split fails, try to continue with what we have
		if !d.opts.StrictMode {
			// In lenient mode, ignore split errors and return what we got
			if len(rawPackets) > 0 {
				err = nil
			}
		} else {
			return nil, residue, err
		}
	}

	// Decode each packet
	packets = make([]packet.Packet, 0, len(rawPackets))
	for i, raw := range rawPackets {
		pkt, decodeErr := d.Decode(raw)
		if decodeErr != nil {
			if d.opts.StrictMode {
				// In strict mode, fail on first error
				return packets, residue, fmt.Errorf("failed to decode packet %d: %w", i, decodeErr)
			}
			// In lenient mode, skip invalid packets
			if d.opts.EnableLogging {
				// TODO: Add logging
				_ = decodeErr
			}
			continue
		}
		packets = append(packets, pkt)
	}

	return packets, residue, nil
}

// SplitPackets splits concatenated packets without decoding them
//
// This is useful if you want to split packets but decode them later,
// or if you want to forward raw packets to another system.
//
// Returns the same values as splitter.SplitPackets
func (d *Decoder) SplitPackets(data []byte) (packets [][]byte, residue []byte, err error) {
	return splitter.SplitPackets(data)
}

// ValidateCRC validates the CRC checksum of a packet
//
// Returns nil if CRC is valid, error otherwise
func (d *Decoder) ValidateCRC(data []byte) error {
	if !validator.ValidateCRC(data) {
		received, calculated, _ := validator.VerifyPacketCRC(data)
		return NewCRCError(calculated, received, len(data))
	}
	return nil
}

// ValidateStructure validates the packet structure (start/stop bits, length)
//
// Returns nil if structure is valid, error otherwise
func (d *Decoder) ValidateStructure(data []byte) error {
	return d.validateStructure(data)
}

// GetOptions returns a copy of the decoder options
func (d *Decoder) GetOptions() Options {
	return d.opts.Clone()
}

// SetOptions updates the decoder options
func (d *Decoder) SetOptions(opts Options) error {
	if err := opts.Validate(); err != nil {
		return err
	}
	d.opts = opts
	return nil
}

// validateStructure performs basic packet structure validation
func (d *Decoder) validateStructure(data []byte) error {
	if len(data) < protocol.MinPacketSize {
		return ErrInvalidPacketSize
	}

	// Check start bit
	startBit := uint16(data[0])<<8 | uint16(data[1])
	if startBit != protocol.StartBitShort && startBit != protocol.StartBitLong {
		return ErrInvalidStartBit
	}

	// Check stop bit
	stopBit := uint16(data[len(data)-2])<<8 | uint16(data[len(data)-1])
	if stopBit != protocol.StopBit {
		return ErrInvalidStopBit
	}

	// Validate length field matches actual length
	var lengthFieldSize int
	var declaredLength int

	if startBit == protocol.StartBitShort {
		lengthFieldSize = 1
		declaredLength = int(data[2])
	} else {
		lengthFieldSize = 2
		declaredLength = int(data[2])<<8 | int(data[3])
	}

	expectedSize := protocol.StartBitSize + lengthFieldSize + declaredLength + protocol.StopBitSize
	if len(data) != expectedSize {
		return ErrInvalidPacketLength
	}

	// Check max packet size
	if len(data) > d.opts.MaxPacketSize {
		return ErrBufferOverflow
	}

	return nil
}

// GetProtocolNumber returns the protocol number from a packet without full decoding
func (d *Decoder) GetProtocolNumber(data []byte) (byte, error) {
	return splitter.GetPacketType(data)
}

// GetSerialNumber returns the serial number from a packet without full decoding
func (d *Decoder) GetSerialNumber(data []byte) (uint16, error) {
	return splitter.GetSerialNumber(data)
}

// HasCompletePacket checks if the data contains at least one complete packet
func (d *Decoder) HasCompletePacket(data []byte) bool {
	return splitter.HasCompletePacket(data)
}

// EstimatePacketCount estimates the number of packets in the data
func (d *Decoder) EstimatePacketCount(data []byte) int {
	return splitter.EstimatePacketCount(data)
}

// RegisteredProtocols returns a list of protocol numbers that have registered parsers
func (d *Decoder) RegisteredProtocols() []byte {
	if d.registry == nil {
		return nil
	}
	return d.registry.List()
}

// HasParser returns true if a parser is registered for the protocol number
func (d *Decoder) HasParser(protocolNum byte) bool {
	if d.registry == nil {
		return false
	}
	return d.registry.Has(protocolNum)
}

// RegisterParser registers a custom parser with the decoder
func (d *Decoder) RegisterParser(p parser.Parser) error {
	if d.registry == nil {
		d.registry = parser.NewRegistry()
	}
	return d.registry.Register(p)
}
