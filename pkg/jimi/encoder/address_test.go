package encoder

import (
	"testing"

	"github.com/fcode09/jimi-vl103m/internal/parser"
	"github.com/fcode09/jimi-vl103m/internal/validator"
	"github.com/fcode09/jimi-vl103m/pkg/jimi/packet"
	"github.com/fcode09/jimi-vl103m/pkg/jimi/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChineseAddressResponse(t *testing.T) {
	tests := []struct {
		name    string
		params  AddressResponseParams
		wantErr bool
		check   func(t *testing.T, data []byte)
	}{
		{
			name: "valid chinese address response",
			params: AddressResponseParams{
				ServerFlag:   [4]byte{0x01, 0x02, 0x03, 0x04},
				AlarmSMS:     "ALARM001",
				Address:      "北京市朝阳区",
				PhoneNumber:  "861234567890",
				Language:     protocol.LanguageChinese,
				SerialNumber: 0x0001,
			},
			wantErr: false,
			check: func(t *testing.T, data []byte) {
				// Verify packet structure
				assert.True(t, len(data) > 10, "packet should have minimum length")
				assert.Equal(t, byte(0x78), data[0], "start bit 1")
				assert.Equal(t, byte(0x78), data[1], "start bit 2")
				assert.Equal(t, byte(0x17), data[3], "protocol number")
				assert.Equal(t, byte(0x0D), data[len(data)-2], "stop bit 1")
				assert.Equal(t, byte(0x0A), data[len(data)-1], "stop bit 2")

				// Verify CRC
				assert.True(t, validator.ValidateCRC(data), "CRC should be valid")
			},
		},
		{
			name: "chinese address with long text",
			params: AddressResponseParams{
				ServerFlag:   [4]byte{0xAA, 0xBB, 0xCC, 0xDD},
				AlarmSMS:     "TEST9999",
				Address:      "中华人民共和国北京市朝阳区建国路1号国贸大厦A座10层1001室",
				PhoneNumber:  "8613800138000",
				Language:     protocol.LanguageChinese,
				SerialNumber: 0xABCD,
			},
			wantErr: false,
			check: func(t *testing.T, data []byte) {
				assert.True(t, len(data) > 50, "long address should produce longer packet")
				assert.Equal(t, byte(0x78), data[0])
				assert.Equal(t, byte(0x17), data[3])
				assert.True(t, validator.ValidateCRC(data))
			},
		},
		{
			name: "alarm sms too long gets truncated",
			params: AddressResponseParams{
				ServerFlag:   [4]byte{0x01, 0x02, 0x03, 0x04},
				AlarmSMS:     "THISISAVERYLONGALARMSMS",
				Address:      "北京",
				PhoneNumber:  "123456",
				Language:     protocol.LanguageChinese,
				SerialNumber: 0x0001,
			},
			wantErr: false,
			check: func(t *testing.T, data []byte) {
				assert.Equal(t, byte(0x78), data[0])
				assert.Equal(t, byte(0x17), data[3])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := ChineseAddressResponse(tt.params)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tt.check != nil {
				tt.check(t, data)
			}
		})
	}
}

func TestEnglishAddressResponse(t *testing.T) {
	tests := []struct {
		name    string
		params  AddressResponseParams
		wantErr bool
		check   func(t *testing.T, data []byte)
	}{
		{
			name: "valid english address response",
			params: AddressResponseParams{
				ServerFlag:   [4]byte{0x01, 0x02, 0x03, 0x04},
				AlarmSMS:     "ALARM002",
				Address:      "123 Main St, New York, NY 10001",
				PhoneNumber:  "12125551234",
				Language:     protocol.LanguageEnglish,
				SerialNumber: 0x0002,
			},
			wantErr: false,
			check: func(t *testing.T, data []byte) {
				assert.True(t, len(data) > 10, "packet should have minimum length")
				assert.Equal(t, byte(0x79), data[0], "start bit 1")
				assert.Equal(t, byte(0x79), data[1], "start bit 2")
				assert.Equal(t, byte(0x97), data[4], "protocol number")
				assert.Equal(t, byte(0x0D), data[len(data)-2], "stop bit 1")
				assert.Equal(t, byte(0x0A), data[len(data)-1], "stop bit 2")
				assert.True(t, validator.ValidateCRC(data), "CRC should be valid")
			},
		},
		{
			name: "english address with special characters",
			params: AddressResponseParams{
				ServerFlag:   [4]byte{0xAA, 0xBB, 0xCC, 0xDD},
				AlarmSMS:     "TEST123",
				Address:      "Av. Libertador #1234, Apt. 5-B, Buenos Aires",
				PhoneNumber:  "541145678901",
				Language:     protocol.LanguageEnglish,
				SerialNumber: 0x1234,
			},
			wantErr: false,
			check: func(t *testing.T, data []byte) {
				assert.True(t, len(data) > 50)
				assert.Equal(t, byte(0x79), data[0])
				assert.Equal(t, byte(0x97), data[4])
				assert.True(t, validator.ValidateCRC(data))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := EnglishAddressResponse(tt.params)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tt.check != nil {
				tt.check(t, data)
			}
		})
	}
}

func TestAddressResponse(t *testing.T) {
	tests := []struct {
		name     string
		params   AddressResponseParams
		wantErr  bool
		wantByte byte
	}{
		{
			name: "chinese language selects chinese protocol",
			params: AddressResponseParams{
				ServerFlag:   [4]byte{0x01, 0x02, 0x03, 0x04},
				AlarmSMS:     "TEST",
				Address:      "北京市",
				PhoneNumber:  "123456",
				Language:     protocol.LanguageChinese,
				SerialNumber: 0x0001,
			},
			wantErr:  false,
			wantByte: 0x17,
		},
		{
			name: "english language selects english protocol",
			params: AddressResponseParams{
				ServerFlag:   [4]byte{0x01, 0x02, 0x03, 0x04},
				AlarmSMS:     "TEST",
				Address:      "New York",
				PhoneNumber:  "123456",
				Language:     protocol.LanguageEnglish,
				SerialNumber: 0x0001,
			},
			wantErr:  false,
			wantByte: 0x97,
		},
		{
			name: "default to english for unknown language",
			params: AddressResponseParams{
				ServerFlag:   [4]byte{0x01, 0x02, 0x03, 0x04},
				AlarmSMS:     "TEST",
				Address:      "Address",
				PhoneNumber:  "123456",
				Language:     protocol.Language(99),
				SerialNumber: 0x0001,
			},
			wantErr:  false,
			wantByte: 0x97,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := AddressResponse(tt.params)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			var protocolIdx int
			if data[0] == 0x78 {
				protocolIdx = 3
			} else {
				protocolIdx = 4
			}
			assert.Equal(t, tt.wantByte, data[protocolIdx])
		})
	}
}

func TestRoundTripChineseAddress(t *testing.T) {
	params := AddressResponseParams{
		ServerFlag:   [4]byte{0x11, 0x22, 0x33, 0x44},
		AlarmSMS:     "ALARM999",
		Address:      "中华人民共和国北京市",
		PhoneNumber:  "8613912345678",
		Language:     protocol.LanguageChinese,
		SerialNumber: 0x5678,
	}

	encoded, err := ChineseAddressResponse(params)
	require.NoError(t, err)

	chineseParser := parser.NewChineseAddressParser()
	ctx := parser.DefaultContext()
	pkt, err := chineseParser.Parse(encoded, ctx)
	require.NoError(t, err)

	addrPkt, ok := pkt.(*packet.AddressResponsePacket)
	require.True(t, ok)
	assert.Equal(t, params.ServerFlag, addrPkt.ServerFlag)
	assert.Equal(t, params.AlarmSMS, addrPkt.AlarmSMS)
	assert.Equal(t, params.Address, addrPkt.Address)
	assert.Equal(t, params.PhoneNumber, addrPkt.PhoneNumber)
	assert.Equal(t, params.Language, addrPkt.Language)
}

func TestRoundTripEnglishAddress(t *testing.T) {
	params := AddressResponseParams{
		ServerFlag:   [4]byte{0xAA, 0xBB, 0xCC, 0xDD},
		AlarmSMS:     "TEST1234",
		Address:      "1600 Pennsylvania Avenue NW, Washington, DC 20500",
		PhoneNumber:  "12025551234",
		Language:     protocol.LanguageEnglish,
		SerialNumber: 0x9ABC,
	}

	encoded, err := EnglishAddressResponse(params)
	require.NoError(t, err)

	englishParser := parser.NewEnglishAddressParser()
	ctx := parser.DefaultContext()
	pkt, err := englishParser.Parse(encoded, ctx)
	require.NoError(t, err)

	addrPkt, ok := pkt.(*packet.AddressResponsePacket)
	require.True(t, ok)
	assert.Equal(t, params.ServerFlag, addrPkt.ServerFlag)
	assert.Equal(t, params.AlarmSMS, addrPkt.AlarmSMS)
	assert.Equal(t, params.Address, addrPkt.Address)
	assert.Equal(t, params.PhoneNumber, addrPkt.PhoneNumber)
	assert.Equal(t, params.Language, addrPkt.Language)
}
