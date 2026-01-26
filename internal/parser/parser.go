package parser

import (
	"fmt"
	"sync"

	"github.com/fcode09/jimi-vl103m/pkg/jimi/packet"
)

// Parser is the interface that all protocol parsers must implement
type Parser interface {
	// ProtocolNumber returns the protocol number this parser handles
	ProtocolNumber() byte

	// Parse decodes the raw packet data into a typed Packet
	// The data parameter contains the full packet (including start bit, length, etc.)
	// Returns the parsed packet or an error if parsing fails
	Parse(data []byte) (packet.Packet, error)

	// Name returns the human-readable name of this parser
	Name() string
}

// Context provides additional context for parsing
type Context struct {
	// StrictMode enables strict validation
	StrictMode bool

	// ValidateIMEI enables IMEI checksum validation
	ValidateIMEI bool

	// TimezoneOffset is the default timezone offset in minutes
	TimezoneOffset int
}

// DefaultContext returns the default parser context
func DefaultContext() Context {
	return Context{
		StrictMode:     true,
		ValidateIMEI:   true,
		TimezoneOffset: 0,
	}
}

// Registry maintains a mapping of protocol numbers to parsers
type Registry struct {
	mu      sync.RWMutex
	parsers map[byte]Parser
	context Context
}

// NewRegistry creates a new parser registry
func NewRegistry() *Registry {
	return &Registry{
		parsers: make(map[byte]Parser),
		context: DefaultContext(),
	}
}

// Register adds a parser to the registry
// Returns an error if a parser for the protocol is already registered
func (r *Registry) Register(p Parser) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	proto := p.ProtocolNumber()
	if _, exists := r.parsers[proto]; exists {
		return fmt.Errorf("parser for protocol 0x%02X already registered", proto)
	}

	r.parsers[proto] = p
	return nil
}

// MustRegister adds a parser and panics if registration fails
func (r *Registry) MustRegister(p Parser) {
	if err := r.Register(p); err != nil {
		panic(err)
	}
}

// Unregister removes a parser from the registry
func (r *Registry) Unregister(protocolNum byte) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.parsers, protocolNum)
}

// Get returns the parser for the given protocol number
func (r *Registry) Get(protocolNum byte) (Parser, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.parsers[protocolNum]
	return p, ok
}

// Parse uses the appropriate parser to decode the packet
func (r *Registry) Parse(protocolNum byte, data []byte) (packet.Packet, error) {
	p, ok := r.Get(protocolNum)
	if !ok {
		return nil, fmt.Errorf("no parser registered for protocol 0x%02X", protocolNum)
	}

	return p.Parse(data)
}

// Has returns true if a parser for the protocol number is registered
func (r *Registry) Has(protocolNum byte) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.parsers[protocolNum]
	return ok
}

// List returns all registered protocol numbers
func (r *Registry) List() []byte {
	r.mu.RLock()
	defer r.mu.RUnlock()

	protocols := make([]byte, 0, len(r.parsers))
	for proto := range r.parsers {
		protocols = append(protocols, proto)
	}
	return protocols
}

// Count returns the number of registered parsers
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.parsers)
}

// SetContext sets the parser context
func (r *Registry) SetContext(ctx Context) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.context = ctx
}

// Context returns the current parser context
func (r *Registry) Context() Context {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.context
}

// Global default registry
var defaultRegistry = NewRegistry()

// DefaultRegistry returns the default global registry
func DefaultRegistry() *Registry {
	return defaultRegistry
}

// Register adds a parser to the default registry
func Register(p Parser) error {
	return defaultRegistry.Register(p)
}

// MustRegister adds a parser to the default registry and panics on error
func MustRegister(p Parser) {
	defaultRegistry.MustRegister(p)
}

// Parse uses the default registry to parse a packet
func Parse(protocolNum byte, data []byte) (packet.Packet, error) {
	return defaultRegistry.Parse(protocolNum, data)
}

// Has returns true if the default registry has a parser for the protocol
func Has(protocolNum byte) bool {
	return defaultRegistry.Has(protocolNum)
}

// BaseParser provides common functionality for parsers
type BaseParser struct {
	protocolNum byte
	name        string
}

// NewBaseParser creates a new base parser
func NewBaseParser(protocolNum byte, name string) BaseParser {
	return BaseParser{
		protocolNum: protocolNum,
		name:        name,
	}
}

// ProtocolNumber implements Parser
func (p *BaseParser) ProtocolNumber() byte {
	return p.protocolNum
}

// Name implements Parser
func (p *BaseParser) Name() string {
	return p.name
}

// ExtractContent extracts the content portion of a packet (excluding header/footer)
// For 0x7878 packets: StartBit(2) + Length(1) + Protocol(1) + Content + Serial(2) + CRC(2) + StopBit(2)
// For 0x7979 packets: StartBit(2) + Length(2) + Protocol(1) + Content + Serial(2) + CRC(2) + StopBit(2)
func ExtractContent(data []byte) ([]byte, error) {
	if len(data) < 10 {
		return nil, fmt.Errorf("packet too small: %d bytes", len(data))
	}

	startBit := uint16(data[0])<<8 | uint16(data[1])

	var contentStart int
	var contentEnd int

	switch startBit {
	case 0x7878:
		// Short packet: 2 (start) + 1 (len) + 1 (proto) = 4
		contentStart = 4
		// End is 6 bytes from end: 2 (serial) + 2 (crc) + 2 (stop)
		contentEnd = len(data) - 6
	case 0x7979:
		// Long packet: 2 (start) + 2 (len) + 1 (proto) = 5
		contentStart = 5
		// End is same: 6 bytes from end
		contentEnd = len(data) - 6
	default:
		return nil, fmt.Errorf("invalid start bit: 0x%04X", startBit)
	}

	if contentEnd < contentStart {
		return nil, fmt.Errorf("invalid packet structure")
	}

	return data[contentStart:contentEnd], nil
}

// ExtractSerialNumber extracts the serial number from a packet
func ExtractSerialNumber(data []byte) (uint16, error) {
	if len(data) < 10 {
		return 0, fmt.Errorf("packet too small")
	}

	// Serial number is at position [len-6:len-4]
	serialOffset := len(data) - 6
	return uint16(data[serialOffset])<<8 | uint16(data[serialOffset+1]), nil
}

// IsShortPacket returns true if the packet uses 0x7878 format
func IsShortPacket(data []byte) bool {
	if len(data) < 2 {
		return false
	}
	return data[0] == 0x78 && data[1] == 0x78
}

// IsLongPacket returns true if the packet uses 0x7979 format
func IsLongPacket(data []byte) bool {
	if len(data) < 2 {
		return false
	}
	return data[0] == 0x79 && data[1] == 0x79
}
