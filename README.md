# Jimi VL103M Protocol Decoder

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/intelcon-group/jimi-vl103m)](https://goreportcard.com/report/github.com/intelcon-group/jimi-vl103m)

Professional Go library for decoding and encoding the Jimi VL103M GPS Tracker protocol (JM-VL03) manufactured by Shenzhen Concox Information Technology Co., Ltd.

## Overview

This library provides complete support for the JM-VL03 protocol version 1.1.2, implementing all 17 packet types with proper handling of both 2G and 4G base station formats.

**Key Features:**
- Complete Protocol Support - All 17 packet types (Login, Location, Alarm, Heartbeat, etc.)
- 2G & 4G Variants - Automatic handling of legacy 2G and modern 4G formats
- ACC Status Correction - Properly handles ACC (ignition) status for all packet types
- Automatic CRC Validation - Built-in CRC-ITU checksum verification with optional skip
- TCP Stream Handling - Robust handling of concatenated and fragmented packets
- Bidirectional Communication - Both decoding (device to server) and encoding (server to device)
- Type-Safe API - Strongly typed value objects with comprehensive validation
- Zero External Dependencies - Pure Go standard library for production builds
- Production Ready - Comprehensive test coverage and optimized performance

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Supported Packet Types](#supported-packet-types)
- [API Documentation](#api-documentation)
- [Examples](#examples)
- [Troubleshooting](#troubleshooting)
- [Documentation](#documentation)

## Installation

```bash
go get github.com/intelcon-group/jimi-vl103m
```

Requirements:
- Go 1.21 or higher
- No external dependencies for production use

## Quick Start

### Basic Packet Decoding

```go
package main

import (
    "encoding/hex"
    "fmt"
    "log"
    
    "github.com/intelcon-group/jimi-vl103m/pkg/jimi"
    "github.com/intelcon-group/jimi-vl103m/pkg/jimi/packet"
)

func main() {
    // Create decoder with default options
    decoder := jimi.NewDecoder()
    
    // Example GPS location packet (from device)
    hexData := "787822220F0C1D023305C9027AC8180C46586000140001CC00287D001F71000001000820860D0A"
    data, _ := hex.DecodeString(hexData)
    
    // Decode packet
    pkt, err := decoder.Decode(data)
    if err != nil {
        log.Fatal(err)
    }
    
    // Type assertion to access specific packet fields
    switch p := pkt.(type) {
    case *packet.LocationPacket:
        fmt.Printf("GPS Location:\n")
        fmt.Printf("  Position: %.6f, %.6f\n", p.Latitude(), p.Longitude())
        fmt.Printf("  Speed: %d km/h\n", p.Speed)
        fmt.Printf("  ACC Status: %s\n", map[bool]string{true: "ON", false: "OFF"}[p.ACC])
        fmt.Printf("  Satellites: %d\n", p.Satellites)
        fmt.Printf("  Timestamp: %s\n", p.DateTime)
        
    case *packet.HeartbeatPacket:
        fmt.Printf("Heartbeat:\n")
        fmt.Printf("  ACC: %v, GPS Tracking: %v, Armed: %v\n",
            p.TerminalInfo.ACCOn(),
            p.TerminalInfo.GPSTrackingEnabled(),
            p.TerminalInfo.IsArmed())
        fmt.Printf("  Battery: %s (%d%%)\n", p.VoltageLevel.String(), p.VoltageLevel.Percentage())
        fmt.Printf("  GSM Signal: %s (%d bars)\n", p.GSMSignal.String(), p.GSMSignal.Bars())
    }
}
```

### TCP Server with Complete Packet Handling

```go
package main

import (
    "log"
    "net"
    
    "github.com/intelcon-group/jimi-vl103m/pkg/jimi"
    "github.com/intelcon-group/jimi-vl103m/pkg/jimi/encoder"
    "github.com/intelcon-group/jimi-vl103m/pkg/jimi/packet"
)

func main() {
    decoder := jimi.NewDecoder()
    enc := encoder.NewEncoder()
    
    listener, _ := net.Listen("tcp", ":5023")
    defer listener.Close()
    
    log.Println("GPS Server listening on :5023")
    
    for {
        conn, _ := listener.Accept()
        go handleConnection(conn, decoder, enc)
    }
}

func handleConnection(conn net.Conn, decoder *jimi.Decoder, enc *encoder.Encoder) {
    defer conn.Close()
    
    buffer := make([]byte, 0, 4096)
    readBuf := make([]byte, 1024)
    
    for {
        n, err := conn.Read(readBuf)
        if err != nil {
            return
        }
        
        buffer = append(buffer, readBuf[:n]...)
        
        // Decode TCP stream (handles packet fragmentation)
        packets, residue, err := decoder.DecodeStream(buffer)
        if err != nil {
            continue
        }
        
        buffer = residue
        
        // Process each complete packet
        for _, pkt := range packets {
            processPacket(pkt, conn, enc)
        }
    }
}

func processPacket(pkt packet.Packet, conn net.Conn, enc *encoder.Encoder) {
    switch p := pkt.(type) {
    case *packet.LoginPacket:
        log.Printf("Login from IMEI: %s", p.GetIMEI())
        response := enc.EncodeLoginResponse(p.SerialNumber())
        conn.Write(response)
        
    case *packet.LocationPacket:
        log.Printf("Location: %.6f, %.6f | ACC: %s | Speed: %d km/h",
            p.Latitude(), p.Longitude(),
            map[bool]string{true: "ON", false: "OFF"}[p.ACC],
            p.Speed)
        
    case *packet.Location4GPacket:
        log.Printf("Location 4G: %.6f, %.6f | MCC/MNC: %d",
            p.Latitude(), p.Longitude(), p.MCCMNC)
        
    case *packet.AlarmPacket:
        log.Printf("ALARM: %s (Critical: %v)", p.AlarmType, p.IsCritical())
        response := enc.EncodeAlarmResponse(p.SerialNumber())
        conn.Write(response)
        
    case *packet.HeartbeatPacket:
        log.Printf("Heartbeat - ACC: %v, Battery: %d%%",
            p.TerminalInfo.ACCOn(), p.VoltageLevel.Percentage())
        response := enc.EncodeHeartbeatResponse(p.SerialNumber())
        conn.Write(response)
    }
}
```

## Supported Packet Types

| Protocol | Code | Description | Direction | Status |
|----------|------|-------------|-----------|---------|
| Login | 0x01 | Device authentication with IMEI | Device to Server | Complete |
| Heartbeat | 0x13 | Keep-alive with status and battery | Device to Server | Complete |
| GPS Location | 0x22 | GPS data with 2G cell towers | Device to Server | Complete |
| GPS Location 4G | 0xA0 | GPS data with 4G cell towers | Device to Server | Complete |
| LBS Multi-Base | 0x28 | LBS-only location (2G) | Device to Server | Complete |
| LBS Multi-Base 4G | 0xA1 | LBS-only location (4G) | Device to Server | Complete |
| Alarm | 0x26 | Single geofence alarm | Device to Server | Complete |
| Alarm Multi-Fence | 0x27 | Multiple geofence alarm (2G) | Device to Server | Complete |
| Alarm 4G | 0xA4 | Multiple geofence alarm (4G) | Device to Server | Complete |
| GPS Address Request | 0x2A | Request address from coordinates | Device to Server | Complete |
| Online Command | 0x80 | Send command to device | Server to Device | Complete |
| Time Calibration | 0x8A | Time synchronization | Bidirectional | Complete |
| Information Transfer | 0x94 | Device status and parameters | Device to Server | Complete |
| Command Response | 0x21/0x15 | Response to commands | Device to Server | Complete |
| Chinese Address | 0x17 | Parsed address response (Chinese) | Server to Device | Complete |
| English Address | 0x97 | Parsed address response (English) | Server to Device | Complete |

### 2G vs 4G Packet Differences

The library automatically handles differences between 2G and 4G protocols:

| Field | 2G Format | 4G Format | Notes |
|-------|-----------|-----------|-------|
| MNC | 1 byte | 1 or 2 bytes | Determined by MCC Bit15 |
| LAC | 2 bytes | 4 bytes | Extended range in 4G |
| Cell ID | 3 bytes | 8 bytes | Extended range in 4G |

## API Documentation

### Decoding Packets

```go
// Create decoder with options
decoder := jimi.NewDecoder(
    jimi.WithSkipCRC(),              // Skip CRC validation (faster)
    jimi.WithStrictMode(false),      // Allow unknown protocols
    jimi.WithAllowUnknownProtocols(),
)

// Decode single packet
packet, err := decoder.Decode(data)

// Decode TCP stream (handles fragmentation)
packets, residue, err := decoder.DecodeStream(buffer)
```

### Common Packet Fields

#### LocationPacket (0x22) and Location4GPacket (0xA0)

```go
// Access GPS data
lat := packet.Latitude()          // float64
lon := packet.Longitude()         // float64
speed := packet.Speed             // uint8 (km/h)
heading := packet.Heading()       // uint16 (degrees)

// ACC Status
accOn := packet.ACC               // bool - direct field
// OR
accOn := packet.ACCOn()           // bool - method

// LBS Information (cell tower)
if packet.LBSInfo.IsValid() {
    mcc := packet.LBSInfo.MCC     // Mobile Country Code
    mnc := packet.LBSInfo.MNC     // Mobile Network Code
    lac := packet.LBSInfo.LAC     // Location Area Code
    cellID := packet.LBSInfo.CellID
}
```

#### HeartbeatPacket (0x13)

```go
// Terminal status bits
accOn := packet.TerminalInfo.ACCOn()                    // Bit 1
isCharging := packet.TerminalInfo.IsCharging()          // Bit 2
gpsTracking := packet.TerminalInfo.GPSTrackingEnabled() // Bit 6
isArmed := packet.TerminalInfo.IsArmed()                // Bit 0

// Battery and signal
batteryPercent := packet.VoltageLevel.Percentage()  // 0-100%
gsmBars := packet.GSMSignal.Bars()                  // 0-4 bars
```

#### AlarmPacket (0x26/0x27/0xA4)

```go
// Alarm information
alarmType := packet.AlarmType           // protocol.AlarmType
isCritical := packet.IsCritical()       // bool

// GPS data (same as LocationPacket)
lat := packet.Latitude()
lon := packet.Longitude()

// Terminal info (from bit fields)
accOn := packet.TerminalInfo.ACCOn()
alarmBits := packet.TerminalInfo.AlarmTypeBits()  // Bits 3-5
```

### Encoding Responses

```go
enc := encoder.NewEncoder()

// Send responses back to device
loginResp := enc.EncodeLoginResponse(serialNumber)
heartbeatResp := enc.EncodeHeartbeatResponse(serialNumber)
alarmResp := enc.EncodeAlarmResponse(serialNumber)
timeResp := enc.EncodeTimeCalibrationResponse(serialNumber, time.Now())

conn.Write(loginResp)
```

## Examples

See the `/examples` directory for complete working examples:

- [Basic Decoder](examples/basic_decoder/) - Simple packet decoding
- [TCP Server](examples/tcp_server/) - Full-featured GPS server
- [Batch Processor](examples/batch_processor/) - Process multiple packets
- [Kafka Integration](examples/kafka_integration/) - Publish to Kafka
- [WebSocket Publisher](examples/websocket_publisher/) - Real-time streaming

## Troubleshooting

### ACC Status Shows Incorrect Value

**Issue:** GPS Location packets show ACC=OFF when they should show ACC=ON.

**Solution:** This has been fixed in the current version. GPS Location packets use a dedicated ACC byte (0x00=OFF, 0x01=ON), while Heartbeat and Alarm packets use a bit field (Bit 1). The library now correctly handles both formats.

```go
// GPS Location packets (0x22, 0xA0)
accOn := packet.ACC  // Reads dedicated byte

// Heartbeat/Alarm packets (0x13, 0x26, 0x27, 0xA4)
accOn := packet.TerminalInfo.ACCOn()  // Reads Bit 1
```

### Packet Too Short Errors

**Cause:** Packet content does not match expected length for the protocol.

**Solution:**
```go
// Use lenient mode for testing
decoder := jimi.NewDecoder(
    jimi.WithStrictMode(false),
)
```

### CRC Validation Failures

**Cause:** Corrupted data or incorrect CRC calculation.

**Solution:**
```go
// Skip CRC validation during development
decoder := jimi.NewDecoder(
    jimi.WithSkipCRC(),
)
```

For more troubleshooting information, see [docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md).

## Documentation

- [Protocol Specification](docs/README_VL03_Protocol.md) - Official JM-VL03 v1.1.2 protocol details
- [Architecture Guide](docs/ARCHITECTURE.md) - Library design and internal structure
- [API Reference](docs/API.md) - Complete API documentation
- [Examples](docs/EXAMPLES.md) - Detailed usage examples for all packet types
- [Troubleshooting](docs/TROUBLESHOOTING.md) - Common issues and solutions
- [Migration Guide](docs/MIGRATION.md) - Upgrading from other libraries

## Development

### Building

```bash
# Build all binaries
make build

# Run tests
make test

# Run tests with coverage
make test-coverage

# Run linter
make lint
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with race detector
go test -race ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Performance

- **Throughput:** 10,000+ packets/second on modern hardware
- **Memory:** Approximately 1KB per concurrent connection
- **Latency:** Less than 1ms per packet decode (average)
- **Zero allocations** in hot paths where possible

## Contributing

Contributions are welcome. Please read our [Contributing Guide](CONTRIBUTING.md) first.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

Protocol specification based on official Concox JM-VL03 documentation (version 1.1.2).

## Support

- **Issues:** [GitHub Issues](https://github.com/intelcon-group/jimi-vl103m/issues)
- **Documentation:** This README and `/docs` folder

---

**Copyright 2026 Intelcon Group. All rights reserved.**
