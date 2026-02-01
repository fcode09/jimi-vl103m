package jimi

import (
	"encoding/hex"
	"testing"

	"github.com/fcode09/jimi-vl103m/pkg/jimi/packet"
)

func TestDecoder_WithoutIMEIValidation(t *testing.T) {
	// Test that WithoutIMEIValidation() is respected by LoginParser

	// Create decoder WITH IMEI validation (default)
	decoderWithValidation := NewDecoder(
		WithSkipCRC(),
		WithStrictMode(false),
	)

	// Create decoder WITHOUT IMEI validation
	decoderWithoutValidation := NewDecoder(
		WithSkipCRC(),
		WithStrictMode(false),
		WithoutIMEIValidation(),
	)

	// Use a packet that should work - this is the same packet used in integration tests
	// Login packet: IMEI 0123456789012348, Model 0x044D, Timezone 0x0320 (UTC+8)
	data, _ := hex.DecodeString("787811010123456789012348044D03200001ABCD0D0A")

	// Test decoder WITH validation
	pkt1, err1 := decoderWithValidation.Decode(data)
	if err1 != nil {
		t.Logf("Decoder with validation error (may be due to IMEI checksum): %v", err1)
	} else {
		// In lenient mode, if parsing fails, it returns BasePacket
		if login1, ok := pkt1.(*packet.LoginPacket); ok {
			t.Logf("Decoder with validation parsed: IMEI=%s", login1.IMEI.String())
		} else {
			t.Logf("Decoder with validation returned: %T (parser may have failed)", pkt1)
		}
	}

	// Test decoder WITHOUT validation - should always succeed for valid format
	pkt2, err2 := decoderWithoutValidation.Decode(data)
	if err2 != nil {
		t.Errorf("Decoder without validation should parse valid format packets: %v", err2)
	} else {
		// Should be able to cast to LoginPacket
		login2, ok := pkt2.(*packet.LoginPacket)
		if !ok {
			t.Errorf("Expected *LoginPacket, got %T", pkt2)
		} else {
			t.Logf("Decoder without validation parsed: IMEI=%s", login2.IMEI.String())

			// Verify the IMEI was extracted
			if login2.IMEI.String() == "" {
				t.Error("IMEI should not be empty")
			}
		}
	}

	// The key test: decoder without validation should succeed and return a LoginPacket
	// even when the IMEI checksum might be invalid
	if err2 != nil {
		t.Error("WithoutIMEIValidation() option should allow parsing of packets with invalid IMEI checksum")
	}

	// Verify that we got a LoginPacket (not just a BasePacket)
	if login2, ok := pkt2.(*packet.LoginPacket); ok && err2 == nil {
		if login2.IMEI.String() != "" {
			t.Logf("SUCCESS: Decoder correctly parsed login packet without IMEI validation: %s", login2.IMEI.String())
		}
	}
}

func TestDecoder_IMEIValidationOption(t *testing.T) {
	// Verify that the ValidateIMEIChecksum option is correctly passed to the registry context

	// Default should have validation enabled
	decoderDefault := NewDecoder()
	opts := decoderDefault.GetOptions()

	if !opts.ValidateIMEIChecksum {
		t.Error("Default decoder should have ValidateIMEIChecksum=true")
	}

	// With WithoutIMEIValidation option
	decoderNoValidate := NewDecoder(WithoutIMEIValidation())
	opts2 := decoderNoValidate.GetOptions()

	if opts2.ValidateIMEIChecksum {
		t.Error("Decoder with WithoutIMEIValidation() should have ValidateIMEIChecksum=false")
	}
}
