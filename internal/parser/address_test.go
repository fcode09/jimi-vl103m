package parser

import (
	"testing"

	"github.com/fcode09/jimi-vl103m/pkg/jimi/packet"
	"github.com/fcode09/jimi-vl103m/pkg/jimi/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChineseAddressParser_Parse(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
		check   func(t *testing.T, pkt packet.Packet)
	}{
		{
			name: "valid chinese address response",
			data: func() []byte {
				// Build content first
				// Content: ContentLength(1) + ServerFlag(4) + ALARMSMS(8) + && + Address(UTF-16BE) + && + PhoneNumber(21) + ##
				addressBytes := []byte{
					0x53, 0x17, 0x4E, 0xAC, 0x67, 0x2D, 0x9E, 0x33, 0x53, 0x0A, 0x67, 0x2D, // 北京市朝阳区 (UTF-16 BE)
				}

				content := make([]byte, 0)
				content = append(content, 0x30)                   // Content length placeholder
				content = append(content, 0x01, 0x02, 0x03, 0x04) // ServerFlag
				content = append(content, []byte("ALARM001")...)  // ALARMSMS (8 bytes)
				content = append(content, '&', '&')               // Separator
				content = append(content, addressBytes...)        // Address in UTF-16 BE
				content = append(content, '&', '&')               // Separator
				phoneBytes := []byte("861234567890         ")     // PhoneNumber (21 bytes)
				content = append(content, phoneBytes...)
				content = append(content, '#', '#') // End marker
				content[0] = byte(len(content) - 1) // Update content length

				// Build full packet: Start(2) + Length(1) + Protocol(1) + Content + Serial(2) + CRC(2) + Stop(2)
				pkt := []byte{0x78, 0x78}               // Start bits
				pkt = append(pkt, byte(len(content)+1)) // Length (protocol + content)
				pkt = append(pkt, 0x17)                 // Protocol
				pkt = append(pkt, content...)           // Content
				pkt = append(pkt, 0x00, 0x01)           // Serial
				pkt = append(pkt, 0x00, 0x00)           // CRC placeholder
				pkt = append(pkt, 0x0D, 0x0A)           // Stop bits
				return pkt
			}(),
			wantErr: false,
			check: func(t *testing.T, pkt packet.Packet) {
				addrPkt, ok := pkt.(*packet.AddressResponsePacket)
				require.True(t, ok, "packet should be AddressResponsePacket")
				assert.Equal(t, byte(protocol.ProtocolAddressResponseChinese), addrPkt.ProtocolNum)
				assert.Equal(t, [4]byte{0x01, 0x02, 0x03, 0x04}, addrPkt.ServerFlag)
				assert.Equal(t, "ALARM001", addrPkt.AlarmSMS)
				assert.NotEmpty(t, addrPkt.Address)
				assert.Equal(t, "861234567890", addrPkt.PhoneNumber)
				assert.Equal(t, byte(protocol.LanguageChinese), byte(addrPkt.Language))
			},
		},
		{
			name:    "too short packet",
			data:    []byte{0x78, 0x78, 0x05, 0x17, 0x01, 0x02, 0x03},
			wantErr: true,
		},
	}

	parser := NewChineseAddressParser()
	ctx := DefaultContext()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkt, err := parser.Parse(tt.data, ctx)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tt.check != nil {
				tt.check(t, pkt)
			}
		})
	}
}

func TestEnglishAddressParser_Parse(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
		check   func(t *testing.T, pkt packet.Packet)
	}{
		{
			name: "valid english address response",
			data: func() []byte {
				// Build content first
				// Note: Content Length is still 1 byte even for long packets (0x97)
				address := "123 Main St, New York, NY 10001"

				content := make([]byte, 0)
				content = append(content, 0x50)                   // Content length placeholder (1 byte)
				content = append(content, 0x01, 0x02, 0x03, 0x04) // ServerFlag
				content = append(content, []byte("ALARM002")...)  // ALARMSMS (8 bytes)
				content = append(content, '&', '&')               // Separator
				content = append(content, []byte(address)...)     // Address in ASCII
				content = append(content, '&', '&')               // Separator
				phoneBytes := []byte("12125551234          ")     // PhoneNumber (21 bytes)
				content = append(content, phoneBytes...)
				content = append(content, '#', '#') // End marker
				content[0] = byte(len(content) - 1) // Update content length

				// Build full packet: Start(2) + Length(2) + Protocol(1) + Content + Serial(2) + CRC(2) + Stop(2)
				pkt := []byte{0x79, 0x79}                    // Start bits (long packet)
				pkt = append(pkt, byte((len(content)+1)>>8)) // Length high byte
				pkt = append(pkt, byte(len(content)+1))      // Length low byte
				pkt = append(pkt, 0x97)                      // Protocol
				pkt = append(pkt, content...)                // Content
				pkt = append(pkt, 0x00, 0x02)                // Serial
				pkt = append(pkt, 0x00, 0x00)                // CRC placeholder
				pkt = append(pkt, 0x0D, 0x0A)                // Stop bits
				return pkt
			}(),
			wantErr: false,
			check: func(t *testing.T, pkt packet.Packet) {
				addrPkt, ok := pkt.(*packet.AddressResponsePacket)
				require.True(t, ok, "packet should be AddressResponsePacket")
				assert.Equal(t, byte(protocol.ProtocolAddressResponseEnglish), addrPkt.ProtocolNum)
				assert.Equal(t, [4]byte{0x01, 0x02, 0x03, 0x04}, addrPkt.ServerFlag)
				assert.Equal(t, "ALARM002", addrPkt.AlarmSMS)
				assert.Equal(t, "123 Main St, New York, NY 10001", addrPkt.Address)
				assert.Equal(t, "12125551234", addrPkt.PhoneNumber)
				assert.Equal(t, byte(protocol.LanguageEnglish), byte(addrPkt.Language))
			},
		},
		{
			name:    "too short packet",
			data:    []byte{0x79, 0x79, 0x00, 0x05, 0x97, 0x01, 0x02},
			wantErr: true,
		},
	}

	parser := NewEnglishAddressParser()
	ctx := DefaultContext()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkt, err := parser.Parse(tt.data, ctx)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tt.check != nil {
				tt.check(t, pkt)
			}
		})
	}
}
