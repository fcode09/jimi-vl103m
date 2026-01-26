package types

import "fmt"

// LBSInfo represents Location-Based Service information from cell towers
// Used when GPS signal is not available
type LBSInfo struct {
	MCC    uint16 // Mobile Country Code
	MNC    uint16 // Mobile Network Code (1 or 2 bytes depending on 2G/4G)
	LAC    uint32 // Location Area Code (2 bytes for 2G, 4 bytes for 4G)
	CellID uint64 // Cell Tower ID (3 bytes for 2G, 8 bytes for 4G)
}

// NewLBSInfo creates a new LBSInfo with validation
func NewLBSInfo(mcc, mnc uint16, lac uint32, cellID uint64) LBSInfo {
	return LBSInfo{
		MCC:    mcc,
		MNC:    mnc,
		LAC:    lac,
		CellID: cellID,
	}
}

// NewLBSInfoFromBytes creates LBSInfo from raw bytes and returns bytes consumed.
// is4G: true for 4G packets (15-16 bytes), false for 2G/3G packets (8 bytes)
func NewLBSInfoFromBytes(data []byte, is4G bool) (LBSInfo, int, error) {
	if is4G {
		// For 4G, check if bit15 of MCC is set (indicates 2-byte MNC)
		if len(data) < 2 {
			return LBSInfo{}, 0, fmt.Errorf("insufficient data for LBS info")
		}
		mccHasBit15 := (data[0] & 0x80) != 0
		return NewLBSInfoFromBytes4G(data, mccHasBit15)
	}
	info, err := NewLBSInfoFromBytes2G(data)
	return info, 8, err
}

// NewLBSInfoFromBytes2G creates LBSInfo from 2G/3G packet bytes (8 bytes total)
// Format: MCC(2) + MNC(1) + LAC(2) + CellID(3)
func NewLBSInfoFromBytes2G(data []byte) (LBSInfo, error) {
	if len(data) < 8 {
		return LBSInfo{}, fmt.Errorf("2G LBS info requires at least 8 bytes, got %d", len(data))
	}

	mcc := uint16(data[0])<<8 | uint16(data[1])
	mnc := uint16(data[2])
	lac := uint32(data[3])<<8 | uint32(data[4])
	cellID := uint64(data[5])<<16 | uint64(data[6])<<8 | uint64(data[7])

	return LBSInfo{
		MCC:    mcc,
		MNC:    mnc,
		LAC:    lac,
		CellID: cellID,
	}, nil
}

// NewLBSInfoFromBytes4G creates LBSInfo from 4G packet bytes and returns bytes consumed.
// Format varies based on MCC bit15:
//   - If bit15=0: MCC(2) + MNC(1) + LAC(4) + CellID(8) = 15 bytes
//   - If bit15=1: MCC(2) + MNC(2) + LAC(4) + CellID(8) = 16 bytes
func NewLBSInfoFromBytes4G(data []byte, mccHasBit15 bool) (LBSInfo, int, error) {
	minSize := 15
	if mccHasBit15 {
		minSize = 16
	}

	if len(data) < minSize {
		return LBSInfo{}, 0, fmt.Errorf("4G LBS info requires at least %d bytes, got %d", minSize, len(data))
	}

	// MCC (2 bytes)
	mcc := uint16(data[0])<<8 | uint16(data[1])

	// Remove bit 15 from MCC if present (it indicates MNC length)
	if mccHasBit15 {
		mcc &= 0x7FFF // Clear bit 15
	}

	// MNC (1 or 2 bytes depending on bit15)
	var mnc uint16
	var offset int
	if mccHasBit15 {
		mnc = uint16(data[2])<<8 | uint16(data[3])
		offset = 4
	} else {
		mnc = uint16(data[2])
		offset = 3
	}

	// LAC (4 bytes for 4G)
	lac := uint32(data[offset])<<24 | uint32(data[offset+1])<<16 |
		uint32(data[offset+2])<<8 | uint32(data[offset+3])
	offset += 4

	// CellID (8 bytes for 4G)
	cellID := uint64(data[offset])<<56 | uint64(data[offset+1])<<48 |
		uint64(data[offset+2])<<40 | uint64(data[offset+3])<<32 |
		uint64(data[offset+4])<<24 | uint64(data[offset+5])<<16 |
		uint64(data[offset+6])<<8 | uint64(data[offset+7])
	offset += 8

	info := LBSInfo{
		MCC:    mcc,
		MNC:    mnc,
		LAC:    lac,
		CellID: cellID,
	}
	return info, offset, nil
}

// String returns a human-readable representation of LBS info
func (l LBSInfo) String() string {
	return fmt.Sprintf("MCC:%d MNC:%d LAC:%d CellID:%d", l.MCC, l.MNC, l.LAC, l.CellID)
}

// IsValid returns true if LBS info appears to be valid (MCC is non-zero)
func (l LBSInfo) IsValid() bool {
	return l.MCC != 0
}

// Bytes2G returns the LBSInfo encoded as 2G/3G format (8 bytes)
func (l LBSInfo) Bytes2G() []byte {
	return []byte{
		byte(l.MCC >> 8),
		byte(l.MCC),
		byte(l.MNC), // Only 1 byte for 2G
		byte(l.LAC >> 8),
		byte(l.LAC),
		byte(l.CellID >> 16),
		byte(l.CellID >> 8),
		byte(l.CellID),
	}
}

// Bytes4G returns the LBSInfo encoded as 4G format
// useTwoBytesMNC: if true, uses 2 bytes for MNC (sets bit15 of MCC)
func (l LBSInfo) Bytes4G(useTwoBytesMNC bool) []byte {
	mcc := l.MCC
	if useTwoBytesMNC {
		mcc |= 0x8000 // Set bit 15 to indicate 2-byte MNC
	}

	bytes := []byte{
		byte(mcc >> 8),
		byte(mcc),
	}

	// MNC (1 or 2 bytes)
	if useTwoBytesMNC {
		bytes = append(bytes,
			byte(l.MNC>>8),
			byte(l.MNC),
		)
	} else {
		bytes = append(bytes, byte(l.MNC))
	}

	// LAC (4 bytes)
	bytes = append(bytes,
		byte(l.LAC>>24),
		byte(l.LAC>>16),
		byte(l.LAC>>8),
		byte(l.LAC),
	)

	// CellID (8 bytes)
	bytes = append(bytes,
		byte(l.CellID>>56),
		byte(l.CellID>>48),
		byte(l.CellID>>40),
		byte(l.CellID>>32),
		byte(l.CellID>>24),
		byte(l.CellID>>16),
		byte(l.CellID>>8),
		byte(l.CellID),
	)

	return bytes
}

// CountryCode returns the country name based on MCC (common codes)
func (l LBSInfo) CountryCode() string {
	// This is a simplified mapping of common MCCs
	// For production, you'd want a complete lookup table
	switch l.MCC {
	case 310, 311, 312, 313, 316:
		return "US" // United States
	case 208:
		return "FR" // France
	case 234:
		return "GB" // United Kingdom
	case 262:
		return "DE" // Germany
	case 222:
		return "IT" // Italy
	case 214:
		return "ES" // Spain
	case 460:
		return "CN" // China
	case 440:
		return "JP" // Japan
	case 450:
		return "KR" // South Korea
	case 334:
		return "MX" // Mexico
	case 724:
		return "BR" // Brazil
	case 716:
		return "PE" // Peru
	default:
		return fmt.Sprintf("Unknown(%d)", l.MCC)
	}
}
