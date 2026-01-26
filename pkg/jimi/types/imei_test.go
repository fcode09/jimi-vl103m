package types

import (
	"testing"
)

func TestNewIMEI(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid IMEI with correct checksum",
			input:   "353456789012348",
			wantErr: false,
		},
		{
			name:    "invalid length - too short",
			input:   "12345678901234",
			wantErr: true,
		},
		{
			name:    "invalid length - too long",
			input:   "1234567890123456",
			wantErr: true,
		},
		{
			name:    "invalid characters",
			input:   "35345678901234A",
			wantErr: true,
		},
		{
			name:    "invalid checksum",
			input:   "353456789012340",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewIMEI(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewIMEI() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewIMEIUnchecked(t *testing.T) {
	// This should succeed even with invalid checksum
	imei, err := NewIMEIUnchecked("353456789012340")
	if err != nil {
		t.Errorf("NewIMEIUnchecked() error = %v", err)
	}
	if imei.String() != "353456789012340" {
		t.Errorf("Expected IMEI string 353456789012340, got %s", imei.String())
	}
}

func TestIMEIFromBytes(t *testing.T) {
	// BCD encoded IMEI: 353456789012348 -> 0x35 0x34 0x56 0x78 0x90 0x12 0x34 0x80
	bcdBytes := []byte{0x35, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x80}

	imei, err := NewIMEIFromBytesUnchecked(bcdBytes)
	if err != nil {
		t.Fatalf("NewIMEIFromBytesUnchecked() error = %v", err)
	}

	expected := "353456789012348"
	if imei.String() != expected {
		t.Errorf("Expected IMEI %s, got %s", expected, imei.String())
	}
}

func TestIMEIToBytes(t *testing.T) {
	imei, _ := NewIMEIUnchecked("353456789012348")

	bytes := imei.Bytes()
	if len(bytes) != 8 {
		t.Fatalf("Expected 8 bytes, got %d", len(bytes))
	}

	// Verify BCD encoding
	expected := []byte{0x35, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x80}
	for i, b := range bytes {
		if b != expected[i] {
			t.Errorf("Byte %d: expected 0x%02X, got 0x%02X", i, expected[i], b)
		}
	}
}

func TestIMEIParts(t *testing.T) {
	imei, _ := NewIMEIUnchecked("353456789012348")

	// TAC is first 8 digits
	if tac := imei.TAC(); tac != "35345678" {
		t.Errorf("Expected TAC 35345678, got %s", tac)
	}

	// SNR is next 6 digits
	if snr := imei.SNR(); snr != "901234" {
		t.Errorf("Expected SNR 901234, got %s", snr)
	}

	// Check digit is last digit
	if cd := imei.CheckDigit(); cd != 8 {
		t.Errorf("Expected check digit 8, got %d", cd)
	}
}

func TestIMEIIsValid(t *testing.T) {
	imei, _ := NewIMEIUnchecked("353456789012348")
	if !imei.IsValid() {
		t.Error("Expected IsValid() to return true")
	}

	emptyIMEI := IMEI{}
	if emptyIMEI.IsValid() {
		t.Error("Expected empty IMEI IsValid() to return false")
	}
}

func BenchmarkNewIMEI(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = NewIMEI("353456789012348")
	}
}

func BenchmarkIMEIFromBytes(b *testing.B) {
	bcdBytes := []byte{0x35, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x80}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewIMEIFromBytesUnchecked(bcdBytes)
	}
}
