# Troubleshooting Guide

This guide helps you resolve common issues when using the jimi-vl103m library.

## Table of Contents

1. [ACC Status Issues](#acc-status-issues)
2. [Packet Decoding Errors](#packet-decoding-errors)
3. [TCP Stream Handling](#tcp-stream-handling)
4. [2G vs 4G Confusion](#2g-vs-4g-confusion)
5. [Performance Issues](#performance-issues)
6. [CRC Validation Failures](#crc-validation-failures)

---

## ACC Status Issues

### Problem: ACC Status Always Shows OFF for GPS Location Packets

**Status:** RESOLVED

**Description:** GPS Location packets (0x22, 0xA0) were showing ACC=OFF even when the raw data indicated ACC=ON.

**Root Cause:** The GPS Location protocol uses a dedicated byte for ACC (0x00=OFF, 0x01=ON), while Heartbeat and Alarm packets use a bit field (Bit 1). Earlier code was treating the GPS ACC byte as a bit field, so 0x01 was being interpreted as Bit 1 = 0 (OFF).

**Solution in Current Version:**

```go
// GPS Location packets (0x22, 0xA0)
// CORRECT: Read the dedicated ACC field
if loc, ok := pkt.(*packet.LocationPacket); ok {
    accOn := loc.ACC  // Returns true for 0x01, false for 0x00
    // OR
    accOn := loc.ACCOn()  // Same result
}

// Heartbeat/Alarm packets (0x13, 0x26, 0x27, 0xA4)
// CORRECT: Read Bit 1 from TerminalInfo
if hb, ok := pkt.(*packet.HeartbeatPacket); ok {
    accOn := hb.TerminalInfo.ACCOn()  // Checks Bit 1
}
```

**Verification:**

```go
// Test with known packet
raw := "787822221A02010E02118901C31ADC07ABA0CA00189301361A1234005678010000003C00B90D0A"
// This packet has ACC=0x01 (ON at position after LBS data)
```

---

## Packet Decoding Errors

### Error: "content too short"

**Cause:** The packet content does not match the expected length for the protocol.

**Solutions:**

1. **For development/testing** - Use lenient mode:
```go
decoder := jimi.NewDecoder(
    jimi.WithStrictMode(false),  // Do not enforce exact lengths
)
```

2. **Check packet structure** - Verify you are using the correct protocol:
```go
// Wrong: Trying to parse 4G packet as 2G
hexData := "78782DA01A02010E02118A..."  // This is 0xA0 (4G)
pkt, err := decoder.Decode(hexData)  // Works but may fail if content short

// Check protocol number
proto, _ := decoder.GetProtocolNumber(data)
fmt.Printf("Protocol: 0x%02X\n", proto)
```

3. **For GPS Address Request (0x2A)** - Must have exactly 41 bytes:
```go
// Content length breakdown:
// DateTime: 6 bytes
// GPS Info: 1 byte
// Latitude: 4 bytes
// Longitude: 4 bytes
// Speed: 1 byte
// Course/Status: 2 bytes
// Phone Number: 21 bytes
// Alarm/Language: 2 bytes
// Total: 41 bytes
```

---

## TCP Stream Handling

### Problem: Packets Are Concatenated or Fragmented

**Symptom:** Receiving multiple packets in one read, or incomplete packets.

**Solution:** Use `DecodeStream`:

```go
buffer := make([]byte, 0, 4096)
tmp := make([]byte, 1024)

for {
    n, err := conn.Read(tmp)
    if err != nil {
        return
    }
    
    buffer = append(buffer, tmp[:n]...)
    
    // Decode handles concatenation and fragmentation
    packets, residue, err := decoder.DecodeStream(buffer)
    if err != nil {
        log.Printf("Decode error: %v", err)
        // Keep residue for next iteration
        buffer = residue
        continue
    }
    
    buffer = residue
    
    for _, pkt := range packets {
        processPacket(pkt)
    }
}
```

**Important:** Always keep the `residue` for the next read. It contains incomplete packet data.

---

## 2G vs 4G Confusion

### Problem: LBS Data Looks Wrong

**Symptom:** MNC, LAC, or CellID values do not match expected format.

**Check Protocol Numbers:**

| Protocol | Code | Format | MNC | LAC | CellID |
|----------|------|--------|-----|-----|--------|
| GPS Location 2G | 0x22 | 2G | 1 byte | 2 bytes | 3 bytes |
| GPS Location 4G | 0xA0 | 4G | 1-2 bytes | 4 bytes | 8 bytes |
| LBS 2G | 0x28 | 2G | 1 byte | 2 bytes | 3 bytes |
| LBS 4G | 0xA1 | 4G | 1-2 bytes | 4 bytes | 8 bytes |
| Alarm 2G | 0x26 | 2G | 1 byte | 2 bytes | 3 bytes |
| Alarm 4G | 0xA4 | 4G | 1-2 bytes | 4 bytes | 8 bytes |

**How MNC Length is Determined:**

```go
// MCC Bit15 = 0 -> MNC is 1 byte
// MCC Bit15 = 1 -> MNC is 2 bytes

mcc := p.LBSInfo.MCC  // 2 bytes
if mcc & 0x8000 != 0 {
    fmt.Println("MNC is 2 bytes (new device)")
} else {
    fmt.Println("MNC is 1 byte (legacy device)")
}
```

**The library handles this automatically**, but you should be aware when debugging.

---

## Performance Issues

### Problem: High CPU or Memory Usage

**1. Optimize buffer sizes:**
```go
// Do not allocate large buffers per connection
buffer := make([]byte, 0, 4096)  // Good
buffer := make([]byte, 0, 65536) // Overkill for most cases
```

**2. Reuse decoders:**
```go
// Create once, use for all connections
decoder := jimi.NewDecoder()

for each connection {
    go handleConnection(conn, decoder)  // Reuse same decoder
}
```

**3. Skip CRC if needed:**
```go
// CRC validation uses CPU
decoder := jimi.NewDecoder(
    jimi.WithSkipCRC(),  // Skip if network is reliable
)
```

**4. Profile your code:**
```bash
go test -bench=. -cpuprofile=cpu.prof
go tool pprof cpu.prof
```

---

## CRC Validation Failures

### Problem: "CRC mismatch" Errors

**Causes:**
1. Corrupted data in transmission
2. Incomplete packets
3. Wrong CRC calculation (bug in device firmware)

**Solutions:**

**Option 1: Skip CRC (Development/Testing)**
```go
decoder := jimi.NewDecoder(
    jimi.WithSkipCRC(),
)
```

**Option 2: Log and Continue**
```go
pkt, err := decoder.Decode(data)
if err != nil {
    if strings.Contains(err.Error(), "CRC") {
        log.Printf("CRC error (ignoring): %v", err)
        // Try to decode anyway with lenient mode
        lenientDecoder := jimi.NewDecoder(jimi.WithSkipCRC())
        pkt, _ = lenientDecoder.Decode(data)
    }
}
```

**Option 3: Validate Raw Data**
```go
// Check packet structure before decoding
if len(data) < 5 {
    return  // Too short
}

// Verify start bits
if data[0] != 0x78 || data[1] != 0x78 {
    if data[0] != 0x79 || data[1] != 0x79 {
        log.Println("Invalid start bits")
        return
    }
}
```

---

## Debug Mode

### Enable Detailed Logging

```go
import "log"

// Set log level
decoder := jimi.NewDecoder(
    jimi.WithDebug(true),  // If available
)

// Or manual logging
func debugPacket(data []byte) {
    log.Printf("Raw packet: %X", data)
    log.Printf("Length: %d bytes", len(data))
    
    if len(data) >= 3 {
        log.Printf("Protocol: 0x%02X", data[2])
    }
}
```

### Common Debug Commands

```bash
# Enable Go runtime debugging
export GODEBUG=gctrace=1

# Run with race detector
go run -race main.go

# Profile memory
go run main.go -memprofile=mem.prof
go tool pprof mem.prof
```

---

## Getting Help

### Checklist Before Reporting Issues

1. Verify packet hex string is correct
2. Confirm protocol number (0x22 vs 0xA0, etc.)
3. Check content length matches expected
4. Try with CRC skipped
5. Test with lenient mode
6. Check if using latest library version

### Information to Include in Bug Reports

```go
// Include this in your bug report:
fmt.Printf("Library Version: %s\n", jimi.Version)
fmt.Printf("Go Version: %s\n", runtime.Version())
fmt.Printf("Packet Hex: %X\n", rawData)
fmt.Printf("Error: %v\n", err)
fmt.Printf("Protocol: 0x%02X\n", protocolNum)
```

### Useful Debugging Script

```go
package main

import (
    "encoding/hex"
    "fmt"
    "log"
    
    "github.com/intelcon-group/jimi-vl103m/pkg/jimi"
)

func main() {
    hexData := "YOUR_PACKET_HEX_HERE"
    data, _ := hex.DecodeString(hexData)
    
    // Debug info
    fmt.Printf("Raw: %X\n", data)
    fmt.Printf("Length: %d\n", len(data))
    
    // Try decoding
    decoder := jimi.NewDecoder()
    
    // Get protocol without full decode
    if proto, err := decoder.GetProtocolNumber(data); err == nil {
        fmt.Printf("Protocol: 0x%02X\n", proto)
    }
    
    // Try lenient decode
    lenientDecoder := jimi.NewDecoder(
        jimi.WithSkipCRC(),
        jimi.WithStrictMode(false),
    )
    
    pkt, err := lenientDecoder.Decode(data)
    if err != nil {
        log.Fatalf("Decode failed: %v", err)
    }
    
    fmt.Printf("Type: %s\n", pkt.Type())
    fmt.Printf("Serial: %d\n", pkt.SerialNumber())
}
```

---

## FAQ

**Q: Why does my packet fail with "content too short"?**

A: Check that you are using the correct protocol variant (2G vs 4G). 4G protocols have longer LBS sections.

**Q: How do I handle both 2G and 4G devices?**

A: The library automatically handles both. Just decode the packet and type assert:
```go
switch p := pkt.(type) {
case *packet.LocationPacket:
    // Handle 2G
case *packet.Location4GPacket:
    // Handle 4G
}
```

**Q: Can I use this library with other Jimi/Concox devices?**

A: This library is specifically for JM-VL03 protocol. Other devices may use different protocols.

**Q: How do I send commands to the device?**

A: Use the encoder:
```go
enc := encoder.NewEncoder()
cmd := enc.EncodeCutOilCommand(serverFlag, "CUT,OIL#")
conn.Write(cmd)
```

**Q: What if I receive an unknown protocol?**

A: Use `WithAllowUnknownProtocols()` option to skip unknown protocols without error.

---

**Still having issues?**

- **Issues:** https://github.com/intelcon-group/jimi-vl103m/issues
- **Documentation:** This README and `/docs` folder
