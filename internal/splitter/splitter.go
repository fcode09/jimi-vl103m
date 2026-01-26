package splitter

import (
	"fmt"

	"github.com/intelcon-group/jimi-vl103m/pkg/jimi/protocol"
)

// SplitPackets splits concatenated packets from a TCP stream
// GPS devices often send multiple packets in a single TCP write,
// or packets may be fragmented across multiple reads.
//
// Returns:
// - packets: Complete packets found in the data
// - residue: Incomplete packet data to be prepended to next read
// - error: Any error encountered during splitting
//
// Example:
//
//	Input:  [CompletePacket1][CompletePacket2][PartialPacket3]
//	Output: packets=[Packet1, Packet2], residue=[PartialPacket3]
func SplitPackets(data []byte) (packets [][]byte, residue []byte, err error) {
	if len(data) == 0 {
		return nil, nil, nil
	}

	packets = make([][]byte, 0)
	offset := 0

	for offset < len(data) {
		// Need at least 4 bytes to determine packet type and length
		if len(data)-offset < 4 {
			// Not enough data for a packet header, keep as residue
			residue = data[offset:]
			return packets, residue, nil
		}

		// Check for valid start bit
		startBit := uint16(data[offset])<<8 | uint16(data[offset+1])

		var lengthFieldSize int
		var packetLengthField int

		switch startBit {
		case protocol.StartBitShort: // 0x7878
			lengthFieldSize = protocol.LengthFieldSizeShort // 1 byte
			if len(data)-offset < 3 {
				residue = data[offset:]
				return packets, residue, nil
			}
			packetLengthField = int(data[offset+2])

		case protocol.StartBitLong: // 0x7979
			lengthFieldSize = protocol.LengthFieldSizeLong // 2 bytes
			if len(data)-offset < 4 {
				residue = data[offset:]
				return packets, residue, nil
			}
			packetLengthField = int(data[offset+2])<<8 | int(data[offset+3])

		default:
			// Invalid start bit - try to find next valid start bit
			nextOffset := findNextStartBit(data, offset+1)
			if nextOffset == -1 {
				// No valid start bit found, discard all remaining data
				return packets, nil, fmt.Errorf("no valid start bit found at offset %d: 0x%04X", offset, startBit)
			}
			// Skip to next valid start bit
			offset = nextOffset
			continue
		}

		// Calculate total packet size
		// Packet = StartBit + LengthField + PacketLengthField + StopBit
		// where PacketLengthField = ProtocolNum + Content + SerialNum + CRC
		totalSize := protocol.StartBitSize + lengthFieldSize + packetLengthField + protocol.StopBitSize

		// Check if we have enough data for the complete packet
		if len(data)-offset < totalSize {
			// Incomplete packet, keep as residue
			residue = data[offset:]
			return packets, residue, nil
		}

		// Extract the packet
		packet := data[offset : offset+totalSize]

		// Validate stop bit
		stopBitOffset := totalSize - 2
		stopBit := uint16(packet[stopBitOffset])<<8 | uint16(packet[stopBitOffset+1])
		if stopBit != protocol.StopBit {
			// Invalid stop bit - might be corrupted packet
			// Try to find next valid start bit
			nextOffset := findNextStartBit(data, offset+1)
			if nextOffset == -1 {
				return packets, nil, fmt.Errorf("invalid stop bit at offset %d: expected 0x%04X, got 0x%04X",
					offset+stopBitOffset, protocol.StopBit, stopBit)
			}
			offset = nextOffset
			continue
		}

		// Valid packet found
		packets = append(packets, packet)
		offset += totalSize
	}

	return packets, nil, nil
}

// findNextStartBit searches for the next valid start bit in the data
// Returns the offset of the start bit, or -1 if not found
func findNextStartBit(data []byte, startOffset int) int {
	for i := startOffset; i < len(data)-1; i++ {
		startBit := uint16(data[i])<<8 | uint16(data[i+1])
		if startBit == protocol.StartBitShort || startBit == protocol.StartBitLong {
			return i
		}
	}
	return -1
}

// ValidatePacketStructure performs basic structural validation on a packet
// without decoding the full content
func ValidatePacketStructure(packet []byte) error {
	if len(packet) < protocol.MinPacketSize {
		return fmt.Errorf("packet too small: %d bytes (minimum %d)", len(packet), protocol.MinPacketSize)
	}

	// Check start bit
	startBit := uint16(packet[0])<<8 | uint16(packet[1])
	if startBit != protocol.StartBitShort && startBit != protocol.StartBitLong {
		return fmt.Errorf("invalid start bit: 0x%04X", startBit)
	}

	// Check stop bit
	stopBit := uint16(packet[len(packet)-2])<<8 | uint16(packet[len(packet)-1])
	if stopBit != protocol.StopBit {
		return fmt.Errorf("invalid stop bit: 0x%04X", stopBit)
	}

	// Validate length field
	var lengthFieldSize int
	var declaredLength int

	if startBit == protocol.StartBitShort {
		lengthFieldSize = 1
		declaredLength = int(packet[2])
	} else {
		lengthFieldSize = 2
		declaredLength = int(packet[2])<<8 | int(packet[3])
	}

	expectedSize := protocol.StartBitSize + lengthFieldSize + declaredLength + protocol.StopBitSize
	if len(packet) != expectedSize {
		return fmt.Errorf("packet length mismatch: declared %d, actual %d", expectedSize, len(packet))
	}

	return nil
}

// GetPacketType returns the protocol number from a packet
// Returns 0 and error if packet is invalid
func GetPacketType(packet []byte) (byte, error) {
	if len(packet) < 4 {
		return 0, fmt.Errorf("packet too small to determine type")
	}

	startBit := uint16(packet[0])<<8 | uint16(packet[1])

	var protocolOffset int
	if startBit == protocol.StartBitShort {
		protocolOffset = 3 // StartBit(2) + Length(1)
	} else if startBit == protocol.StartBitLong {
		protocolOffset = 4 // StartBit(2) + Length(2)
	} else {
		return 0, fmt.Errorf("invalid start bit: 0x%04X", startBit)
	}

	if len(packet) <= protocolOffset {
		return 0, fmt.Errorf("packet too small")
	}

	return packet[protocolOffset], nil
}

// GetSerialNumber extracts the serial number from a packet
// Serial number is always 2 bytes before CRC
func GetSerialNumber(packet []byte) (uint16, error) {
	if len(packet) < protocol.MinPacketSize {
		return 0, fmt.Errorf("packet too small")
	}

	// Serial number is at position [len-6:len-4]
	// Packet ends with: SerialNum(2) + CRC(2) + StopBit(2)
	serialOffset := len(packet) - 6
	serialNum := uint16(packet[serialOffset])<<8 | uint16(packet[serialOffset+1])

	return serialNum, nil
}

// EstimatePacketCount estimates the number of complete packets in the data
// This is a quick check without full parsing
func EstimatePacketCount(data []byte) int {
	count := 0
	offset := 0

	for offset < len(data)-3 {
		startBit := uint16(data[offset])<<8 | uint16(data[offset+1])

		if startBit == protocol.StartBitShort || startBit == protocol.StartBitLong {
			count++
			// Skip past this potential packet
			// This is just an estimate, so we use a simple heuristic
			if startBit == protocol.StartBitShort && offset+2 < len(data) {
				length := int(data[offset+2])
				offset += protocol.StartBitSize + 1 + length + protocol.StopBitSize
			} else if startBit == protocol.StartBitLong && offset+3 < len(data) {
				length := int(data[offset+2])<<8 | int(data[offset+3])
				offset += protocol.StartBitSize + 2 + length + protocol.StopBitSize
			} else {
				offset++
			}
		} else {
			offset++
		}
	}

	return count
}

// HasCompletePacket quickly checks if the data contains at least one complete packet
func HasCompletePacket(data []byte) bool {
	if len(data) < protocol.MinPacketSize {
		return false
	}

	startBit := uint16(data[0])<<8 | uint16(data[1])

	var totalSize int
	if startBit == protocol.StartBitShort && len(data) >= 3 {
		length := int(data[2])
		totalSize = protocol.StartBitSize + 1 + length + protocol.StopBitSize
	} else if startBit == protocol.StartBitLong && len(data) >= 4 {
		length := int(data[2])<<8 | int(data[3])
		totalSize = protocol.StartBitSize + 2 + length + protocol.StopBitSize
	} else {
		return false
	}

	return len(data) >= totalSize
}
