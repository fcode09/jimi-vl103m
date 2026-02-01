# API Usage Examples

This document provides detailed examples for using each packet type in the jimi-vl103m library.

## Table of Contents

1. [GPS Location Packets (0x22, 0xA0)](#gps-location-packets)
2. [Heartbeat Packet (0x13)](#heartbeat-packet)
3. [Alarm Packets (0x26, 0x27, 0xA4)](#alarm-packets)
4. [Login Packet (0x01)](#login-packet)
5. [GPS Address Request (0x2A)](#gps-address-request)
6. [Information Transfer (0x94)](#information-transfer)
7. [LBS Packets (0x28, 0xA1)](#lbs-packets)
8. [Command Response (0x21, 0x15)](#command-response)

---

## GPS Location Packets

### 0x22 - GPS Location (2G)

```go
package main

import (
    "fmt"
    "github.com/intelcon-group/jimi-vl103m/pkg/jimi/packet"
)

func handleLocation2G(p *packet.LocationPacket) {
    // GPS Coordinates
    fmt.Printf("Latitude: %.6f\n", p.Latitude())
    fmt.Printf("Longitude: %.6f\n", p.Longitude())
    
    // Movement data
    fmt.Printf("Speed: %d km/h\n", p.Speed)
    fmt.Printf("Heading: %d° (%s)\n", p.Heading(), p.HeadingName())
    
    // ACC Status - Uses dedicated ACC byte (0x00=OFF, 0x01=ON)
    if p.ACC {
        fmt.Println("ACC: ON (vehicle ignition on)")
    } else {
        fmt.Println("ACC: OFF (vehicle ignition off)")
    }
    
    // GPS Quality
    fmt.Printf("Satellites: %d\n", p.Satellites)
    fmt.Printf("GPS Positioned: %v\n", p.IsPositioned())
    fmt.Printf("GPS Real-time: %v\n", p.CourseStatus.IsGPSRealtime)
    
    // LBS Information (2G format)
    if p.LBSInfo.IsValid() {
        fmt.Printf("Cell Tower: MCC=%d MNC=%d LAC=%d CellID=%d\n",
            p.LBSInfo.MCC, p.LBSInfo.MNC, p.LBSInfo.LAC, p.LBSInfo.CellID)
    }
    
    // Upload mode (why this location was sent)
    fmt.Printf("Upload Mode: %s\n", p.UploadMode.String())
    
    // Mileage (if available)
    fmt.Printf("Mileage: %d meters\n", p.Mileage)
    
    // Timestamp
    fmt.Printf("Device Time: %s\n", p.DateTime)
    fmt.Printf("Parsed At: %s\n", p.ParsedAt)
}
```

### 0xA0 - GPS Location 4G

```go
func handleLocation4G(p *packet.Location4GPacket) {
    // All LocationPacket fields are available
    fmt.Printf("Position: %.6f, %.6f\n", p.Latitude(), p.Longitude())
    
    // 4G-specific fields
    fmt.Printf("MCC/MNC Combined: %d\n", p.MCCMNC)
    
    // LBS with extended fields (4G format)
    if p.LBSInfo.IsValid() {
        // MNC can be 1 or 2 bytes in 4G
        fmt.Printf("MNC: %d (length depends on MCC Bit15)\n", p.LBSInfo.MNC)
        // LAC is 4 bytes in 4G
        fmt.Printf("LAC: %d\n", p.LBSInfo.LAC)
        // CellID is 8 bytes in 4G
        fmt.Printf("CellID: %d\n", p.LBSInfo.CellID)
    }
    
    // Extended LBS (neighboring cells) if available
    for i, lbs := range p.ExtendedLBS {
        fmt.Printf("Neighbor Cell %d: MCC=%d MNC=%d LAC=%d CellID=%d\n",
            i, lbs.MCC, lbs.MNC, lbs.LAC, lbs.CellID)
    }
}
```

---

## Heartbeat Packet

### 0x13 - Heartbeat with Status

```go
func handleHeartbeat(p *packet.HeartbeatPacket) {
    fmt.Println("=== Heartbeat Received ===")
    
    // Terminal Information (status bits)
    info := p.TerminalInfo
    
    fmt.Printf("ACC Status: %v\n", info.ACCOn())
    fmt.Printf("Charging: %v\n", info.IsCharging())
    fmt.Printf("GPS Tracking: %v\n", info.GPSTrackingEnabled())
    fmt.Printf("Defense/Armed: %v\n", info.IsArmed())
    fmt.Printf("Power Cut: %v\n", info.OilElectricityDisconnected())
    
    // Battery Level
    fmt.Printf("Voltage Level: %s\n", p.VoltageLevel.String())
    fmt.Printf("Battery: %d%%\n", p.VoltageLevel.Percentage())
    
    // GSM Signal
    fmt.Printf("GSM Signal: %s\n", p.GSMSignal.String())
    fmt.Printf("Signal Bars: %d/4\n", p.GSMSignal.Bars())
    
    // Extended info (if present)
    if p.HasExtended {
        fmt.Printf("Extended Info: 0x%04X\n", p.ExtendedInfo)
    }
}
```

**Terminal Information Bit Layout:**
- Bit 0: Defense/Armed (1=armed, 0=not armed)
- Bit 1: ACC Status (1=on, 0=off)
- Bit 2: Charging (1=charging with power, 0=without)
- Bit 3-5: Alarm type bits (000=Normal, 001=Vibration, 010=PowerCut, etc.)
- Bit 6: GPS Positioned (1=positioned, 0=not)
- Bit 7: Oil/Electricity Cut (1=cut, 0=restore)

---

## Alarm Packets

### 0x26, 0x27, 0xA4 - Alarm Packets

```go
func handleAlarm(p *packet.AlarmPacket) {
    fmt.Println("ALARM RECEIVED")
    
    // Alarm type
    fmt.Printf("Alarm Type: %s\n", p.AlarmType.String())
    fmt.Printf("Critical: %v\n", p.IsCritical())
    
    // GPS Data
    fmt.Printf("Position: %.6f, %.6f\n", p.Latitude(), p.Longitude())
    fmt.Printf("Speed: %d km/h\n", p.Speed)
    
    // For multi-fence alarms (0x27, 0xA4)
    if multiFence, ok := p.(*packet.AlarmMultiFencePacket); ok {
        fmt.Printf("Fence ID: %d\n", multiFence.FenceID)
    }
    
    if alarm4G, ok := p.(*packet.Alarm4GPacket); ok {
        fmt.Printf("MCC/MNC: %d\n", alarm4G.MCCMNC)
    }
    
    // Terminal Info (bit fields)
    info := p.TerminalInfo
    fmt.Printf("ACC: %v\n", info.ACCOn())
    
    // Alarm type from terminal info bits (backup/alternative)
    alarmBits := info.AlarmTypeBits()
    switch alarmBits {
    case 0:
        fmt.Println("Terminal Status: Normal")
    case 1:
        fmt.Println("Terminal Status: Vibration")
    case 2:
        fmt.Println("Terminal Status: Power Cut")
    case 3:
        fmt.Println("Terminal Status: Low Battery")
    case 4:
        fmt.Println("Terminal Status: SOS")
    }
    
    // Device status at alarm time
    fmt.Printf("Voltage: %s\n", p.VoltageLevel.String())
    fmt.Printf("GSM Signal: %s\n", p.GSMSignal.String())
    
    // Response required
    fmt.Printf("Language: %s\n", p.Language.String())
}
```

### Alarm Types Reference

```go
// All alarm types defined in pkg/jimi/protocol/types.go
alarmTypes := []protocol.AlarmType{
    protocol.AlarmNormal,                    // 0x00
    protocol.AlarmSOS,                       // 0x01
    protocol.AlarmPowerCut,                  // 0x02
    protocol.AlarmVibration,                 // 0x03
    protocol.AlarmGeofenceEnter,             // 0x04
    protocol.AlarmGeofenceExit,              // 0x05
    protocol.AlarmSpeed,                     // 0x06
    protocol.AlarmTowTheft,                  // 0x09
    protocol.AlarmGPSBlindSpotEnter,         // 0x0A
    protocol.AlarmGPSBlindSpotExit,          // 0x0B
    protocol.AlarmPowerOn,                   // 0x0C
    protocol.AlarmPowerOff,                  // 0x11
    protocol.AlarmTamper,                    // 0x13
    protocol.AlarmDoor,                      // 0x14
    protocol.AlarmHarshAcceleration,         // 0x29
    protocol.AlarmSharpLeftCorner,           // 0x2A
    protocol.AlarmSharpRightCorner,          // 0x2B
    protocol.AlarmCollision,                 // 0x2C
    protocol.AlarmHarshBraking,              // 0x30
    protocol.AlarmACCOn,                     // 0xFE
    protocol.AlarmACCOff,                    // 0xFF
}
```

---

## Login Packet

### 0x01 - Device Authentication

```go
func handleLogin(p *packet.LoginPacket) {
    fmt.Println("Device Login")
    
    // Device identification
    fmt.Printf("IMEI: %s\n", p.GetIMEI())
    fmt.Printf("Model ID: 0x%04X\n", p.ModelID)
    
    // Timezone configuration
    fmt.Printf("Timezone: %s\n", p.Timezone.String())
    fmt.Printf("UTC Offset: %d minutes\n", p.Timezone.OffsetMinutes)
    fmt.Printf("Language: %s\n", p.Timezone.LanguageString())
    
    // Example: Send login response
    response := encoder.EncodeLoginResponse(p.SerialNumber())
    conn.Write(response)
}
```

---

## GPS Address Request

### 0x2A - Request Address from Coordinates

```go
func handleGPSAddressRequest(p *packet.GPSAddressRequestPacket) {
    fmt.Println("GPS Address Request")
    
    // GPS coordinates (device sends these to request address)
    lat := p.Latitude()
    lon := p.Longitude()
    fmt.Printf("Coordinates: %.6f, %.6f\n", lat, lon)
    
    // Google Maps link for quick verification
    fmt.Printf("Maps: https://maps.google.com/?q=%.6f,%.6f\n", lat, lon)
    
    // Request metadata
    fmt.Printf("Phone Number: %s\n", p.PhoneNumber)
    fmt.Printf("Satellites: %d\n", p.Satellites)
    fmt.Printf("Speed: %d km/h\n", p.Speed)
    fmt.Printf("Heading: %d°\n", p.Heading())
    
    // Context
    fmt.Printf("Alarm Type: %s\n", p.AlarmType.String())
    fmt.Printf("Language: %s\n", p.Language.String())
    
    // You should respond with address (0x17 Chinese or 0x97 English)
    address := lookupAddress(lat, lon) // Your geocoding function
    response := encoder.EncodeAddressResponse(
        p.SerialNumber(),
        p.AlarmType,
        address,
        p.Language,
    )
    conn.Write(response)
}
```

---

## Information Transfer

### 0x94 - Various Device Information

```go
func handleInfoTransfer(p *packet.InfoTransferPacket) {
    fmt.Printf("Info Transfer Type: %s (0x%02X)\n", 
        p.SubProtocol.String(), p.SubProtocol)
    
    switch p.SubProtocol {
    case protocol.InfoTypeExternalVoltage:
        // External battery voltage
        voltage := p.GetExternalVoltageVolts()
        fmt.Printf("External Battery: %.2fV\n", voltage)
        
    case protocol.InfoTypeTerminalStatus:
        // Terminal status synchronization
        fmt.Println("Terminal Status Sync:")
        fmt.Printf("  ALM1: 0x%02X\n", p.ALM1)
        fmt.Printf("  ALM2: 0x%02X\n", p.ALM2)
        fmt.Printf("  ALM3: 0x%02X\n", p.ALM3)
        fmt.Printf("  ALM4: 0x%02X\n", p.ALM4)
        fmt.Printf("  STA1: 0x%02X\n", p.STA1)
        fmt.Printf("  SOS Numbers: %v\n", p.SOSNumbers)
        fmt.Printf("  Center Number: %s\n", p.CenterNumber)
        fmt.Printf("  ICCID: %s\n", p.ICCID)
        fmt.Printf("  IMSI: %s\n", p.IMSI)
        
    case protocol.InfoTypeDoorStatus:
        fmt.Printf("Door Status: %v\n", p.DoorOpen)
        
    case protocol.InfoTypeGPSStatus:
        fmt.Println("GPS Status:")
        fmt.Printf("  GPS Module: %s\n", p.GPSStatus.GPSModuleStatus)
        fmt.Printf("  GPS Satellites: %d in fix\n", p.GPSStatus.SatellitesInFix)
        fmt.Printf("  BDS Module: %s\n", p.GPSStatus.BDSModuleStatus)
        fmt.Printf("  BDS Satellites: %d in fix\n", p.GPSStatus.BDSSatellitesInFix)
        
    case protocol.InfoTypeICCID:
        fmt.Printf("IMEI: %s\n", p.IMEI)
        fmt.Printf("IMSI: %s\n", p.IMSI)
        fmt.Printf("ICCID: %s\n", p.ICCID)
    }
}
```

---

## LBS Packets

### 0x28, 0xA1 - LBS Multi-Base Station

Used when GPS is not available, provides location via cell towers.

```go
func handleLBS(p *packet.LBSPacket) {
    fmt.Println("LBS Location (No GPS)")
    
    // Main cell tower
    main := p.LBSInfo
    fmt.Printf("Main Cell: MCC=%d MNC=%d LAC=%d CellID=%d\n",
        main.MCC, main.MNC, main.LAC, main.CellID)
    
    // Signal strength
    fmt.Printf("RSSI: %d (0x%02X)\n", main.RSSI, main.RSSI)
    
    // Neighboring cells (for triangulation)
    for i, neighbor := range p.NeighborCells {
        fmt.Printf("Neighbor %d: LAC=%d CellID=%d RSSI=%d\n",
            i, neighbor.LAC, neighbor.CellID, neighbor.RSSI)
    }
    
    // Timing advance
    fmt.Printf("Timing Advance: %d\n", p.TimingAdvance)
    
    // You can use these cell IDs with an LBS geolocation service
    // to approximate device location when GPS is unavailable
}
```

---

## Command Response

### 0x21, 0x15 - Response to Online Commands

```go
func handleCommandResponse(p *packet.CommandResponsePacket) {
    fmt.Printf("Command Response (Server Flag: 0x%08X)\n", p.ServerFlag)
    fmt.Printf("Response: %s\n", p.Response)
    fmt.Printf("Length: %d bytes\n", p.ResponseLength)
    
    // Protocol version
    if p.ProtocolNum == protocol.ProtocolCommandResponseOld {
        fmt.Println("Protocol: Old format (0x15)")
    } else {
        fmt.Println("Protocol: Universal format (0x21)")
    }
}
```

---

## Complete TCP Handler Example

Here is a complete example showing how to handle all packet types:

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
    
    log.Println("GPS Server started on port 5023")
    
    for {
        conn, _ := listener.Accept()
        go handleClient(conn, decoder, enc)
    }
}

func handleClient(conn net.Conn, decoder *jimi.Decoder, enc *encoder.Encoder) {
    defer conn.Close()
    
    buffer := make([]byte, 0, 4096)
    tmp := make([]byte, 1024)
    
    for {
        n, err := conn.Read(tmp)
        if err != nil {
            return
        }
        
        buffer = append(buffer, tmp[:n]...)
        
        packets, residue, _ := decoder.DecodeStream(buffer)
        buffer = residue
        
        for _, pkt := range packets {
            dispatchPacket(pkt, conn, enc)
        }
    }
}

func dispatchPacket(pkt packet.Packet, conn net.Conn, enc *encoder.Encoder) {
    switch p := pkt.(type) {
    case *packet.LoginPacket:
        log.Printf("Login: %s", p.GetIMEI())
        conn.Write(enc.EncodeLoginResponse(p.SerialNumber()))
        
    case *packet.HeartbeatPacket:
        log.Printf("Heartbeat: ACC=%v Battery=%d%%",
            p.TerminalInfo.ACCOn(), p.VoltageLevel.Percentage())
        conn.Write(enc.EncodeHeartbeatResponse(p.SerialNumber()))
        
    case *packet.LocationPacket:
        log.Printf("Location 2G: %.6f,%.6f ACC=%s Speed=%d",
            p.Latitude(), p.Longitude(),
            map[bool]string{true: "ON", false: "OFF"}[p.ACC],
            p.Speed)
        
    case *packet.Location4GPacket:
        log.Printf("Location 4G: %.6f,%.6f MCC/MNC=%d",
            p.Latitude(), p.Longitude(), p.MCCMNC)
        
    case *packet.AlarmPacket:
        log.Printf("ALARM: %s (Critical=%v)",
            p.AlarmType.String(), p.IsCritical())
        conn.Write(enc.EncodeAlarmResponse(p.SerialNumber()))
        
    case *packet.GPSAddressRequestPacket:
        log.Printf("Address Request: %.6f,%.6f Phone=%s",
            p.Latitude(), p.Longitude(), p.PhoneNumber)
        // TODO: Send address response (0x17 or 0x97)
        
    case *packet.InfoTransferPacket:
        log.Printf("Info Transfer: Type=%s", p.SubProtocol.String())
        
    case *packet.TimeCalibrationPacket:
        log.Printf("Time Calibration Request")
        // TODO: Send time response (0x8A)
        
    case *packet.LBSPacket:
        log.Printf("LBS: MCC=%d MNC=%d LAC=%d CellID=%d",
            p.LBSInfo.MCC, p.LBSInfo.MNC, p.LBSInfo.LAC, p.LBSInfo.CellID)
    }
}
```

---

## Error Handling

### Common Errors and Solutions

```go
// Invalid packet format
if err != nil {
    if errors.Is(err, jimi.ErrInvalidPacket) {
        log.Println("Invalid packet format")
        return
    }
    
    if errors.Is(err, jimi.ErrCRCMismatch) {
        log.Println("CRC validation failed - packet corrupted")
        return
    }
    
    if errors.Is(err, jimi.ErrUnknownProtocol) {
        log.Println("Unknown protocol number")
        return
    }
}

// Use lenient mode for testing
decoder := jimi.NewDecoder(
    jimi.WithSkipCRC(),
    jimi.WithStrictMode(false),
)
```

---

## Best Practices

1. **Always type assert safely:**
```go
if loc, ok := pkt.(*packet.LocationPacket); ok {
    // Handle location
}
```

2. **Use switch for clean dispatch:**
```go
switch p := pkt.(type) {
case *packet.LocationPacket:
    // Handle 2G
case *packet.Location4GPacket:
    // Handle 4G
}
```

3. **Check field validity:**
```go
if p.LBSInfo.IsValid() {
    // Use LBS data
}

if p.HasTimestamp() {
    // Use timestamp
}
```

4. **ACC Status - Important Distinction:**
```go
// GPS Location packets use dedicated ACC byte
if loc, ok := pkt.(*packet.LocationPacket); ok {
    accOn := loc.ACC  // Use field directly
}

// Heartbeat/Alarm packets use bit field
if hb, ok := pkt.(*packet.HeartbeatPacket); ok {
    accOn := hb.TerminalInfo.ACCOn()  // Use method
}
```

---

**For more examples, see the `/examples` directory in the repository.**
