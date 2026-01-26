// Example: Basic Decoder Usage
//
// This example demonstrates how to use the jimi-vl103m library
// to decode GPS tracker packets.
package main

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/intelcon-group/jimi-vl103m/pkg/jimi"
	"github.com/intelcon-group/jimi-vl103m/pkg/jimi/packet"
)

func main() {
	fmt.Println("=== Jimi VL103M Decoder Example ===")

	// Create a decoder with default options
	decoder := jimi.NewDecoder()

	// Show registered protocols
	protocols := decoder.RegisteredProtocols()
	fmt.Printf("Registered protocols: %d\n", len(protocols))
	for _, p := range protocols {
		fmt.Printf("  - 0x%02X\n", p)
	}
	fmt.Println()

	// Example packets (these would normally come from a TCP connection)
	// Note: These are example structures - in production you'd use real captured packets
	examplePackets := []struct {
		name string
		hex  string
	}{
		// These hex strings are placeholders - replace with actual captured packets
		{"Heartbeat", "78780513000100011A2D0D0A"},
	}

	for _, example := range examplePackets {
		fmt.Printf("--- Decoding %s Packet ---\n", example.name)

		data, err := hex.DecodeString(example.hex)
		if err != nil {
			log.Printf("Failed to decode hex: %v\n", err)
			continue
		}

		fmt.Printf("Raw bytes: %s\n", hex.EncodeToString(data))
		fmt.Printf("Length: %d bytes\n", len(data))

		// Check if it's a complete packet
		if !decoder.HasCompletePacket(data) {
			fmt.Println("Incomplete packet")
			continue
		}

		// Get protocol number without full decode
		proto, err := decoder.GetProtocolNumber(data)
		if err != nil {
			log.Printf("Failed to get protocol: %v\n", err)
			continue
		}
		fmt.Printf("Protocol: 0x%02X\n", proto)

		// Full decode (with lenient mode for example data)
		lenientDecoder := jimi.NewDecoder(
			jimi.WithSkipCRC(),
			jimi.WithStrictMode(false),
			jimi.WithAllowUnknownProtocols(),
		)

		pkt, err := lenientDecoder.Decode(data)
		if err != nil {
			log.Printf("Decode error: %v\n", err)
			continue
		}

		fmt.Printf("Packet Type: %s\n", pkt.Type())
		fmt.Printf("Serial Number: %d\n", pkt.SerialNumber())

		// Type-specific handling
		handlePacket(pkt)

		fmt.Println()
	}

	// Demonstrate stream decoding
	demonstrateStreamDecoding()
}

func handlePacket(pkt packet.Packet) {
	switch p := pkt.(type) {
	case *packet.LoginPacket:
		fmt.Printf("Login - IMEI: %s, ModelID: 0x%04X\n", p.IMEI, p.ModelID)

	case *packet.HeartbeatPacket:
		fmt.Printf("Heartbeat - Voltage: %s, GSM: %s, ACC: %v\n",
			p.VoltageLevel, p.GSMSignal, p.ACCOn())

	case *packet.LocationPacket:
		fmt.Printf("Location - Lat: %.6f, Lon: %.6f, Speed: %d km/h\n",
			p.Latitude(), p.Longitude(), p.Speed)
		fmt.Printf("  Heading: %d° (%s), Satellites: %d\n",
			p.Heading(), p.HeadingName(), p.Satellites)

	case *packet.AlarmPacket:
		fmt.Printf("Alarm - Type: %s, Critical: %v\n",
			p.AlarmType, p.IsCritical())
		fmt.Printf("  Location: %.6f, %.6f\n",
			p.Latitude(), p.Longitude())

	default:
		fmt.Printf("Generic packet: %s\n", pkt.Type())
	}
}

func demonstrateStreamDecoding() {
	fmt.Println("=== Stream Decoding Example ===")

	decoder := jimi.NewDecoder(
		jimi.WithStrictMode(false),
		jimi.WithAllowUnknownProtocols(),
	)

	// Simulate receiving data in chunks (like from TCP)
	// In a real application, this would come from conn.Read()
	buffer := make([]byte, 0, 4096)

	// Simulate receiving first chunk (incomplete packet)
	chunk1, _ := hex.DecodeString("787805") // Partial packet
	buffer = append(buffer, chunk1...)
	fmt.Printf("Received chunk 1: %d bytes, buffer: %d bytes\n",
		len(chunk1), len(buffer))

	// Simulate receiving second chunk (completes first packet + starts another)
	chunk2, _ := hex.DecodeString("1300010001FFFF0D0A7878") // Rest of packet + start of another
	buffer = append(buffer, chunk2...)
	fmt.Printf("Received chunk 2: %d bytes, buffer: %d bytes\n",
		len(chunk2), len(buffer))

	// Try to decode
	packets, residue, err := decoder.DecodeStream(buffer)
	if err != nil {
		fmt.Printf("Stream decode error: %v\n", err)
	}

	fmt.Printf("Decoded %d complete packets\n", len(packets))
	fmt.Printf("Residue: %d bytes\n", len(residue))

	for i, pkt := range packets {
		fmt.Printf("  Packet %d: %s (protocol 0x%02X)\n",
			i+1, pkt.Type(), pkt.ProtocolNumber())
	}

	// Keep residue for next iteration
	buffer = residue

	fmt.Println()
}
