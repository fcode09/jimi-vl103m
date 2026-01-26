package jimi

// Options contains configuration for the decoder
type Options struct {
	// StrictMode enables strict validation (fail on any validation error)
	// When false, the decoder will attempt to decode packets with minor issues
	StrictMode bool

	// SkipCRCValidation skips CRC checksum validation
	// Useful for development/debugging, DO NOT use in production
	SkipCRCValidation bool

	// SkipStructureValidation skips basic structure validation (start/stop bits, length)
	// Useful for debugging corrupted packets, DO NOT use in production
	SkipStructureValidation bool

	// MaxPacketSize sets the maximum allowed packet size in bytes
	// Packets larger than this will be rejected
	// Default: 65535 + overhead
	MaxPacketSize int

	// AllowUnknownProtocols allows decoding of unknown protocol numbers
	// When true, unknown protocols return a generic packet
	// When false, unknown protocols return an error
	AllowUnknownProtocols bool

	// EnableLogging enables internal debug logging
	// Logs are written to the configured logger (if any)
	EnableLogging bool

	// TimeLocation sets the default timezone for datetime decoding
	// If nil, UTC is used
	TimeLocation *int // Timezone offset in minutes

	// ValidateIMEIChecksum enables IMEI Luhn checksum validation
	// When false, IMEI format is validated but checksum is not verified
	ValidateIMEIChecksum bool

	// EnableAutoCorrection enables automatic correction of minor packet issues
	// For example: auto-trimming trailing zeros, fixing minor length mismatches
	EnableAutoCorrection bool
}

// Option is a functional option for configuring the Decoder
type Option func(*Options)

// DefaultOptions returns the default decoder options
func DefaultOptions() Options {
	return Options{
		StrictMode:              true,
		SkipCRCValidation:       false,
		SkipStructureValidation: false,
		MaxPacketSize:           65535 + 2 + 2 + 2, // Max length + start + length field + stop
		AllowUnknownProtocols:   false,
		EnableLogging:           false,
		TimeLocation:            nil, // UTC
		ValidateIMEIChecksum:    true,
		EnableAutoCorrection:    false,
	}
}

// WithStrictMode enables or disables strict validation mode
func WithStrictMode(strict bool) Option {
	return func(o *Options) {
		o.StrictMode = strict
	}
}

// WithSkipCRC skips CRC validation (for development only)
func WithSkipCRC() Option {
	return func(o *Options) {
		o.SkipCRCValidation = true
	}
}

// WithSkipStructureValidation skips structure validation (for debugging only)
func WithSkipStructureValidation() Option {
	return func(o *Options) {
		o.SkipStructureValidation = true
	}
}

// WithMaxPacketSize sets the maximum allowed packet size
func WithMaxPacketSize(size int) Option {
	return func(o *Options) {
		if size > 0 {
			o.MaxPacketSize = size
		}
	}
}

// WithAllowUnknownProtocols allows decoding of unknown protocol numbers
func WithAllowUnknownProtocols() Option {
	return func(o *Options) {
		o.AllowUnknownProtocols = true
	}
}

// WithLogging enables internal debug logging
func WithLogging() Option {
	return func(o *Options) {
		o.EnableLogging = true
	}
}

// WithTimeLocation sets the timezone for datetime decoding
// offset: timezone offset in minutes (e.g., 480 for UTC+8, -300 for UTC-5)
func WithTimeLocation(offset int) Option {
	return func(o *Options) {
		o.TimeLocation = &offset
	}
}

// WithoutIMEIValidation disables IMEI Luhn checksum validation
func WithoutIMEIValidation() Option {
	return func(o *Options) {
		o.ValidateIMEIChecksum = false
	}
}

// WithAutoCorrection enables automatic correction of minor issues
func WithAutoCorrection() Option {
	return func(o *Options) {
		o.EnableAutoCorrection = true
	}
}

// WithLenientMode configures the decoder for lenient/permissive decoding
// This is a convenience option that sets multiple flags for maximum compatibility
func WithLenientMode() Option {
	return func(o *Options) {
		o.StrictMode = false
		o.AllowUnknownProtocols = true
		o.EnableAutoCorrection = true
		o.ValidateIMEIChecksum = false
	}
}

// WithDevelopmentMode configures the decoder for development/debugging
// WARNING: DO NOT use in production
func WithDevelopmentMode() Option {
	return func(o *Options) {
		o.StrictMode = false
		o.SkipCRCValidation = true
		o.AllowUnknownProtocols = true
		o.EnableLogging = true
		o.EnableAutoCorrection = true
		o.ValidateIMEIChecksum = false
	}
}

// Validate checks if the options are valid
func (o *Options) Validate() error {
	if o.MaxPacketSize < 10 {
		return NewValidationError("MaxPacketSize", "must be at least 10 bytes", o.MaxPacketSize)
	}

	if o.MaxPacketSize > 1<<20 { // 1 MB
		return NewValidationError("MaxPacketSize", "must not exceed 1 MB", o.MaxPacketSize)
	}

	return nil
}

// IsProduction returns true if options are safe for production use
func (o *Options) IsProduction() bool {
	return !o.SkipCRCValidation &&
		!o.SkipStructureValidation &&
		o.StrictMode &&
		o.ValidateIMEIChecksum
}

// Clone creates a deep copy of the options
func (o *Options) Clone() Options {
	clone := *o
	if o.TimeLocation != nil {
		offset := *o.TimeLocation
		clone.TimeLocation = &offset
	}
	return clone
}
