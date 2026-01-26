package encoder

import (
	"testing"
	"time"

	"github.com/fcode09/jimi-vl103m/internal/validator"
	"github.com/fcode09/jimi-vl103m/pkg/jimi/protocol"
)

func TestNew(t *testing.T) {
	enc := New()
	if enc == nil {
		t.Fatal("New() returned nil")
	}
	if !enc.UseShortFormat {
		t.Error("Expected UseShortFormat to be true by default")
	}
}

func TestLoginResponse(t *testing.T) {
	enc := New()
	serialNum := uint16(0x0001)

	response := enc.LoginResponse(serialNum)

	// Verify structure
	if len(response) < 10 {
		t.Fatalf("Response too short: %d bytes", len(response))
	}

	// Check start bit
	if response[0] != 0x78 || response[1] != 0x78 {
		t.Errorf("Invalid start bit: %02X %02X", response[0], response[1])
	}

	// Check protocol number
	if response[3] != protocol.ProtocolLogin {
		t.Errorf("Expected protocol 0x%02X, got 0x%02X", protocol.ProtocolLogin, response[3])
	}

	// Check stop bit
	if response[len(response)-2] != 0x0D || response[len(response)-1] != 0x0A {
		t.Errorf("Invalid stop bit: %02X %02X", response[len(response)-2], response[len(response)-1])
	}

	// Verify CRC
	if !validator.ValidateCRC(response) {
		t.Error("CRC validation failed")
	}
}

func TestHeartbeatResponse(t *testing.T) {
	enc := New()
	serialNum := uint16(0x0123)

	response := enc.HeartbeatResponse(serialNum)

	// Check protocol number
	if response[3] != protocol.ProtocolHeartbeat {
		t.Errorf("Expected protocol 0x%02X, got 0x%02X", protocol.ProtocolHeartbeat, response[3])
	}

	// Verify CRC
	if !validator.ValidateCRC(response) {
		t.Error("CRC validation failed")
	}
}

func TestAlarmResponse(t *testing.T) {
	enc := New()
	serialNum := uint16(0x0456)

	response := enc.AlarmResponse(serialNum)

	if response[3] != protocol.ProtocolAlarm {
		t.Errorf("Expected protocol 0x%02X, got 0x%02X", protocol.ProtocolAlarm, response[3])
	}

	if !validator.ValidateCRC(response) {
		t.Error("CRC validation failed")
	}
}

func TestTimeCalibrationResponse(t *testing.T) {
	enc := New()
	serialNum := uint16(0x0789)

	// Use a specific time for testing
	testTime := time.Date(2024, 6, 15, 14, 30, 45, 0, time.UTC)

	response := enc.TimeCalibrationResponse(serialNum, testTime)

	// Check protocol number
	if response[3] != protocol.ProtocolTimeCalibration {
		t.Errorf("Expected protocol 0x%02X, got 0x%02X", protocol.ProtocolTimeCalibration, response[3])
	}

	// Check time bytes (after protocol number)
	// YY MM DD HH MM SS
	expectedYear := byte(24) // 2024 - 2000
	expectedMonth := byte(6) // June
	expectedDay := byte(15)
	expectedHour := byte(14)
	expectedMinute := byte(30)
	expectedSecond := byte(45)

	if response[4] != expectedYear {
		t.Errorf("Expected year %d, got %d", expectedYear, response[4])
	}
	if response[5] != expectedMonth {
		t.Errorf("Expected month %d, got %d", expectedMonth, response[5])
	}
	if response[6] != expectedDay {
		t.Errorf("Expected day %d, got %d", expectedDay, response[6])
	}
	if response[7] != expectedHour {
		t.Errorf("Expected hour %d, got %d", expectedHour, response[7])
	}
	if response[8] != expectedMinute {
		t.Errorf("Expected minute %d, got %d", expectedMinute, response[8])
	}
	if response[9] != expectedSecond {
		t.Errorf("Expected second %d, got %d", expectedSecond, response[9])
	}

	if !validator.ValidateCRC(response) {
		t.Error("CRC validation failed")
	}
}

func TestOnlineCommand(t *testing.T) {
	enc := New()
	serialNum := uint16(0x0001)
	serverFlag := uint32(0x12345678)
	command := "IMEI#"

	response := enc.OnlineCommand(serialNum, serverFlag, command)

	// Check protocol number
	if response[3] != protocol.ProtocolOnlineCommand {
		t.Errorf("Expected protocol 0x%02X, got 0x%02X", protocol.ProtocolOnlineCommand, response[3])
	}

	// Verify structure contains command
	// Content starts at offset 4: length(1) + serverFlag(4) + command
	if len(response) < 15 { // Minimum size with command
		t.Fatalf("Response too short for command: %d bytes", len(response))
	}

	// Check server flag is present (bytes 5-8)
	gotFlag := uint32(response[5])<<24 | uint32(response[6])<<16 |
		uint32(response[7])<<8 | uint32(response[8])
	if gotFlag != serverFlag {
		t.Errorf("Expected server flag 0x%08X, got 0x%08X", serverFlag, gotFlag)
	}

	if !validator.ValidateCRC(response) {
		t.Error("CRC validation failed")
	}
}

// TestAddressResponse is now in address_test.go with more comprehensive tests

func TestCustomResponse(t *testing.T) {
	enc := New()
	serialNum := uint16(0x0001)
	customProto := byte(0xAA)
	content := []byte{0x01, 0x02, 0x03}

	response := enc.CustomResponse(customProto, content, serialNum)

	if response[3] != customProto {
		t.Errorf("Expected protocol 0x%02X, got 0x%02X", customProto, response[3])
	}

	// Check content is present
	if response[4] != 0x01 || response[5] != 0x02 || response[6] != 0x03 {
		t.Error("Content not correctly placed in response")
	}

	if !validator.ValidateCRC(response) {
		t.Error("CRC validation failed")
	}
}

func TestResponseBuilder(t *testing.T) {
	enc := New()

	response := enc.NewResponseBuilder(protocol.ProtocolLogin).
		WithSerialNumber(0x0001).
		WithContent(nil).
		Build()

	if response[3] != protocol.ProtocolLogin {
		t.Errorf("Expected protocol 0x%02X, got 0x%02X", protocol.ProtocolLogin, response[3])
	}

	if !validator.ValidateCRC(response) {
		t.Error("CRC validation failed")
	}
}

func TestLongFormatPacket(t *testing.T) {
	enc := New()
	enc.UseShortFormat = false

	response := enc.LoginResponse(0x0001)

	// Check start bit for long format
	if response[0] != 0x79 || response[1] != 0x79 {
		t.Errorf("Expected long format start bit 0x7979, got %02X%02X", response[0], response[1])
	}

	// Length should be 2 bytes for long format
	// Protocol number is at offset 4 for long format
	if response[4] != protocol.ProtocolLogin {
		t.Errorf("Expected protocol 0x%02X at offset 4, got 0x%02X", protocol.ProtocolLogin, response[4])
	}
}

func TestCommandBuilder(t *testing.T) {
	enc := New()
	cb := enc.NewCommandBuilder(0x0001, 0x12345678)

	// Test GetIMEI
	response := cb.GetIMEI()
	if response[3] != protocol.ProtocolOnlineCommand {
		t.Errorf("Expected protocol 0x%02X, got 0x%02X", protocol.ProtocolOnlineCommand, response[3])
	}
	if !validator.ValidateCRC(response) {
		t.Error("CRC validation failed for GetIMEI")
	}

	// Test StartTracking
	response = cb.StartTracking(30)
	if !validator.ValidateCRC(response) {
		t.Error("CRC validation failed for StartTracking")
	}

	// Test SetServer
	response = cb.SetServer("192.168.1.1", 8080)
	if !validator.ValidateCRC(response) {
		t.Error("CRC validation failed for SetServer")
	}
}

// Benchmarks

func BenchmarkLoginResponse(b *testing.B) {
	enc := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = enc.LoginResponse(uint16(i))
	}
}

func BenchmarkTimeCalibrationResponse(b *testing.B) {
	enc := New()
	now := time.Now()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = enc.TimeCalibrationResponse(uint16(i), now)
	}
}

func BenchmarkOnlineCommand(b *testing.B) {
	enc := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = enc.OnlineCommand(uint16(i), 0x12345678, "IMEI#")
	}
}
