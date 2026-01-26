// Package jimi provides a complete decoder for the Jimi VL103M GPS tracker protocol.
//
// The VL103M is a 4G GPS tracking device used in vehicle tracking systems.
// This library decodes the binary protocol used by the device.
//
// # Quick Start
//
// Create a decoder and decode packets:
//
//	decoder := jimi.NewDecoder()
//
//	// Decode a single packet
//	data, _ := hex.DecodeString("787811010359339073930523044D014E0F0A0D0A")
//	pkt, err := decoder.Decode(data)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Type assert to specific packet type
//	if login, ok := pkt.(*packet.LoginPacket); ok {
//	    fmt.Printf("Device IMEI: %s\n", login.IMEI)
//	}
//
// # Decoding TCP Streams
//
// For TCP connections where packets may be concatenated:
//
//	buffer := make([]byte, 0, 4096)
//	readBuf := make([]byte, 1024)
//
//	for {
//	    n, err := conn.Read(readBuf)
//	    if err != nil {
//	        return
//	    }
//	    buffer = append(buffer, readBuf[:n]...)
//
//	    packets, residue, err := decoder.DecodeStream(buffer)
//	    buffer = residue
//
//	    for _, pkt := range packets {
//	        handlePacket(pkt)
//	    }
//	}
//
// # Supported Protocols
//
// The decoder supports the following protocol numbers:
//   - 0x01: Login
//   - 0x13: Heartbeat
//   - 0x22: GPS Location (2G/3G)
//   - 0xA0: GPS Location (4G)
//   - 0x26: Alarm
//   - 0x27: Alarm Multi-Fence
//   - 0xA4: Alarm (4G)
//   - 0x28: LBS Multi-Base
//   - 0xA1: LBS Multi-Base (4G)
//   - And more...
//
// # Configuration Options
//
// The decoder supports various configuration options:
//
//	decoder := jimi.NewDecoder(
//	    jimi.WithStrictMode(false),       // Enable lenient parsing
//	    jimi.WithAllowUnknownProtocols(), // Don't error on unknown protocols
//	    jimi.WithSkipCRC(),               // Skip CRC validation (development only)
//	)
package jimi

// Re-export commonly used types for convenience
// Users can also import subpackages directly for more control

// Version information
const (
	// Version is the current library version
	Version = "0.2.0"

	// ProtocolVersion is the supported VL103M protocol version
	ProtocolVersion = "JM-VL03"
)
