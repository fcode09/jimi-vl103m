// Example: TCP Server for GPS Trackers
//
// This example demonstrates how to create a TCP server
// that receives and processes GPS tracker packets.
// Enhanced version with raw data capture and comprehensive logging.
package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fcode09/jimi-vl103m/pkg/jimi"
	"github.com/fcode09/jimi-vl103m/pkg/jimi/encoder"
	"github.com/fcode09/jimi-vl103m/pkg/jimi/packet"
	"github.com/fcode09/jimi-vl103m/pkg/jimi/protocol"
)

var (
	port       = flag.Int("port", 8080, "TCP server port")
	logDir     = flag.String("logdir", "logs", "Directory to store raw packet logs")
	verbose    = flag.Bool("verbose", false, "Enable verbose logging")
	saveRaw    = flag.Bool("save-raw", true, "Save raw packets to files")
	strictMode = flag.Bool("strict", false, "Enable strict mode parsing")
)

// DeviceSession represents a connected device
type DeviceSession struct {
	conn        net.Conn
	decoder     *jimi.Decoder
	encoder     *encoder.Encoder
	imei        string
	lastSeen    time.Time
	mu          sync.Mutex
	rawLogFile  *os.File
	packetCount int
	connectedAt time.Time
	remoteAddr  string
}

// Global session manager
var (
	sessions   = make(map[string]*DeviceSession)
	sessionsMu sync.RWMutex
)

func main() {
	flag.Parse()

	// Create log directory if saving raw data
	if *saveRaw {
		if err := os.MkdirAll(*logDir, 0755); err != nil {
			log.Fatalf("Failed to create log directory: %v", err)
		}
	}

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.Printf("==============================================")
	log.Printf("GPS Tracker Server v1.0")
	log.Printf("==============================================")
	log.Printf("Port: %d", *port)
	log.Printf("Log Directory: %s", *logDir)
	log.Printf("Verbose: %v", *verbose)
	log.Printf("Save Raw Packets: %v", *saveRaw)
	log.Printf("Strict Mode: %v", *strictMode)
	log.Printf("==============================================")

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer listener.Close()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("\n==============================================")
		log.Println("Shutting down server...")
		printSessionSummary()
		listener.Close()
		os.Exit(0)
	}()

	log.Println("Server started. Waiting for connections...")
	log.Println("")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}

		log.Printf(">>> New connection from %s", conn.RemoteAddr())
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	remoteAddr := conn.RemoteAddr().String()
	connectedAt := time.Now()

	session := &DeviceSession{
		conn:        conn,
		decoder:     jimi.NewDecoder(jimi.WithStrictMode(*strictMode)),
		encoder:     encoder.New(),
		lastSeen:    time.Now(),
		connectedAt: connectedAt,
		remoteAddr:  remoteAddr,
	}

	// Create raw log file for this connection
	if *saveRaw {
		filename := fmt.Sprintf("raw_%s_%s.log",
			strings.ReplaceAll(remoteAddr, ":", "-"),
			connectedAt.Format("20060102_150405"))
		filepath := filepath.Join(*logDir, filename)
		f, err := os.Create(filepath)
		if err != nil {
			log.Printf("[%s] Warning: Failed to create raw log file: %v", remoteAddr, err)
		} else {
			session.rawLogFile = f
			defer f.Close()
			// Write header
			f.WriteString(fmt.Sprintf("# GPS Tracker Raw Packet Log\n"))
			f.WriteString(fmt.Sprintf("# Connection: %s\n", remoteAddr))
			f.WriteString(fmt.Sprintf("# Started: %s\n", connectedAt.Format(time.RFC3339)))
			f.WriteString(fmt.Sprintf("# Format: [timestamp] [direction] [hex_data]\n"))
			f.WriteString(fmt.Sprintf("#\n"))
		}
	}

	buffer := make([]byte, 0, 4096)
	readBuf := make([]byte, 1024)

	// Set read timeout
	conn.SetReadDeadline(time.Now().Add(5 * time.Minute))

	for {
		n, err := conn.Read(readBuf)
		if err != nil {
			if err != io.EOF {
				log.Printf("[%s] Read error: %v", session.getIdentifier(), err)
			}
			break
		}

		if n == 0 {
			continue
		}

		// Log raw received data
		rawData := readBuf[:n]
		session.logRawData("RX", rawData)

		if *verbose {
			log.Printf("[%s] RAW RX (%d bytes): %s",
				session.getIdentifier(), n, hex.EncodeToString(rawData))
		}

		buffer = append(buffer, rawData...)
		session.lastSeen = time.Now()

		// Reset read deadline
		conn.SetReadDeadline(time.Now().Add(5 * time.Minute))

		// Try to decode packets
		packets, residue, err := session.decoder.DecodeStream(buffer)
		if err != nil {
			log.Printf("[%s] Decode error: %v", session.getIdentifier(), err)
			// Log the problematic data for analysis
			if *verbose {
				log.Printf("[%s] Buffer at error (%d bytes): %s",
					session.getIdentifier(), len(buffer), hex.EncodeToString(buffer))
			}
		}

		// Check for unknown/unprocessed data in residue
		if len(residue) > 0 && *verbose {
			log.Printf("[%s] Residue (%d bytes): %s",
				session.getIdentifier(), len(residue), hex.EncodeToString(residue))
		}

		buffer = residue

		// Process each packet
		for _, pkt := range packets {
			session.packetCount++
			session.handlePacket(pkt)
		}
	}

	// Connection closed - print summary
	duration := time.Since(connectedAt)
	log.Printf("<<< [%s] Connection closed. Duration: %s, Packets: %d",
		session.getIdentifier(), duration.Round(time.Second), session.packetCount)

	// Remove from sessions
	sessionsMu.Lock()
	delete(sessions, session.imei)
	sessionsMu.Unlock()
}

func (s *DeviceSession) getIdentifier() string {
	if s.imei != "" {
		return s.imei
	}
	return s.remoteAddr
}

func (s *DeviceSession) logRawData(direction string, data []byte) {
	if s.rawLogFile == nil {
		return
	}
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	line := fmt.Sprintf("[%s] %s %s\n", timestamp, direction, hex.EncodeToString(data))
	s.rawLogFile.WriteString(line)
	s.rawLogFile.Sync()
}

func (s *DeviceSession) handlePacket(pkt packet.Packet) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Log raw packet bytes
	rawHex := hex.EncodeToString(pkt.Raw())
	log.Printf("[%s] PKT #%d: %s (0x%02X) serial=%d raw=%s",
		s.getIdentifier(),
		s.packetCount,
		pkt.Type(),
		pkt.ProtocolNumber(),
		pkt.SerialNumber(),
		rawHex)

	switch p := pkt.(type) {
	case *packet.LoginPacket:
		s.handleLogin(p)

	case *packet.HeartbeatPacket:
		s.handleHeartbeat(p)

	case *packet.LocationPacket:
		s.handleLocation(p)

	case *packet.AlarmPacket:
		s.handleAlarm(p)

	case *packet.TimeCalibrationPacket:
		s.handleTimeCalibration(p)

	case *packet.LBSPacket:
		s.handleLBS(p)

	case *packet.InfoTransferPacket:
		s.handleInfoTransfer(p)

	case *packet.CommandResponsePacket:
		s.handleCommandResponse(p)

	case *packet.GPSAddressRequestPacket:
		s.handleGPSAddressRequest(p)

	default:
		s.handleUnknown(pkt)
	}
}

func (s *DeviceSession) handleLogin(p *packet.LoginPacket) {
	s.imei = p.IMEI.String()

	// Register session
	sessionsMu.Lock()
	sessions[s.imei] = s
	sessionsMu.Unlock()

	// Rename raw log file with IMEI if possible
	if s.rawLogFile != nil {
		oldPath := s.rawLogFile.Name()
		newFilename := fmt.Sprintf("raw_%s_%s.log",
			s.imei,
			s.connectedAt.Format("20060102_150405"))
		newPath := filepath.Join(*logDir, newFilename)
		s.rawLogFile.Close()
		os.Rename(oldPath, newPath)
		f, err := os.OpenFile(newPath, os.O_APPEND|os.O_WRONLY, 0644)
		if err == nil {
			s.rawLogFile = f
			f.WriteString(fmt.Sprintf("# IMEI identified: %s\n", s.imei))
		}
	}

	log.Printf("[%s] LOGIN", s.imei)
	log.Printf("[%s]   IMEI: %s", s.imei, p.IMEI)
	log.Printf("[%s]   Model ID: 0x%04X (%d)", s.imei, p.ModelID, p.ModelID)
	log.Printf("[%s]   Timezone: %s", s.imei, p.Timezone)

	// Send login response
	response := s.encoder.LoginResponse(p.SerialNumber())
	s.sendResponse(response)
}

func (s *DeviceSession) handleHeartbeat(p *packet.HeartbeatPacket) {
	log.Printf("[%s] HEARTBEAT", s.imei)
	log.Printf("[%s]   Voltage: %s (%d%%)", s.imei, p.VoltageLevel, p.BatteryPercentage())
	log.Printf("[%s]   GSM Signal: %s (%d bars)", s.imei, p.GSMSignal, p.SignalBars())
	log.Printf("[%s]   ACC: %v, Charging: %v",
		s.imei, p.ACCOn(), p.IsCharging())
	log.Printf("[%s]   Terminal Info: %s", s.imei, p.TerminalInfo)

	// Send heartbeat response
	response := s.encoder.HeartbeatResponse(p.SerialNumber())
	s.sendResponse(response)
}

func (s *DeviceSession) handleLocation(p *packet.LocationPacket) {
	log.Printf("[%s] LOCATION", s.imei)
	log.Printf("[%s]   Time: %s", s.imei, p.DateTime)
	log.Printf("[%s]   Position: %.6f, %.6f", s.imei, p.Latitude(), p.Longitude())
	log.Printf("[%s]   Speed: %d km/h", s.imei, p.Speed)
	log.Printf("[%s]   Heading: %d° (%s)", s.imei, p.Heading(), p.HeadingName())
	log.Printf("[%s]   Satellites: %d, Positioned: %v", s.imei, p.Satellites, p.IsPositioned())

	if p.LBSInfo.IsValid() {
		log.Printf("[%s]   LBS: MCC=%d MNC=%d LAC=%d CellID=%d",
			s.imei, p.LBSInfo.MCC, p.LBSInfo.MNC, p.LBSInfo.LAC, p.LBSInfo.CellID)
	}

	if p.HasStatus {
		log.Printf("[%s]   Terminal: %s", s.imei, p.TerminalInfo)
	}

	log.Printf("[%s]   Upload Mode: %s", s.imei, p.UploadMode)

	if p.Mileage > 0 {
		log.Printf("[%s]   Mileage: %d meters", s.imei, p.Mileage)
	}

	// Location packets typically don't require response
}

func (s *DeviceSession) handleAlarm(p *packet.AlarmPacket) {
	log.Printf("[%s] ALARM *** %s ***", s.imei, p.AlarmType)
	log.Printf("[%s]   Critical: %v", s.imei, p.IsCritical())
	log.Printf("[%s]   Time: %s", s.imei, p.DateTime)
	log.Printf("[%s]   Position: %.6f, %.6f", s.imei, p.Latitude(), p.Longitude())
	log.Printf("[%s]   Speed: %d km/h", s.imei, p.Speed)
	log.Printf("[%s]   Positioned: %v", s.imei, p.IsPositioned())

	if p.LBSInfo.IsValid() {
		log.Printf("[%s]   LBS: MCC=%d MNC=%d LAC=%d CellID=%d",
			s.imei, p.LBSInfo.MCC, p.LBSInfo.MNC, p.LBSInfo.LAC, p.LBSInfo.CellID)
	}

	log.Printf("[%s]   Terminal: %s", s.imei, p.TerminalInfo)

	// Handle critical alarms
	if p.IsCritical() {
		log.Printf("[%s] !!! CRITICAL ALARM: %s !!!", s.imei, p.AlarmType)
	}

	// Send alarm response (required by protocol)
	response := s.encoder.AlarmResponse(p.SerialNumber())
	s.sendResponse(response)
}

func (s *DeviceSession) handleTimeCalibration(p *packet.TimeCalibrationPacket) {
	log.Printf("[%s] TIME CALIBRATION REQUEST", s.imei)

	// Send current server time
	now := time.Now().UTC()
	log.Printf("[%s]   Responding with: %s", s.imei, now.Format(time.RFC3339))

	response := s.encoder.TimeCalibrationResponseNow(p.SerialNumber())
	s.sendResponse(response)
}

func (s *DeviceSession) handleLBS(p *packet.LBSPacket) {
	log.Printf("[%s] LBS (Cell Tower Location)", s.imei)
	log.Printf("[%s]   Time: %s", s.imei, p.DateTime)

	if p.LBSInfo.IsValid() {
		log.Printf("[%s]   MCC: %d (Country: %s)", s.imei, p.LBSInfo.MCC, p.LBSInfo.CountryCode())
		log.Printf("[%s]   MNC: %d", s.imei, p.LBSInfo.MNC)
		log.Printf("[%s]   LAC: %d", s.imei, p.LBSInfo.LAC)
		log.Printf("[%s]   Cell ID: %d", s.imei, p.LBSInfo.CellID)
	}

	log.Printf("[%s]   Voltage: %s, GSM: %s", s.imei, p.VoltageLevel, p.GSMSignal)

	if len(p.NeighborCells) > 0 {
		log.Printf("[%s]   Neighbor Cells: %d", s.imei, len(p.NeighborCells))
		for i, cell := range p.NeighborCells {
			log.Printf("[%s]     [%d] LAC=%d CellID=%d", s.imei, i, cell.LAC, cell.CellID)
		}
	}
}

func (s *DeviceSession) handleInfoTransfer(p *packet.InfoTransferPacket) {
	log.Printf("[%s] INFO TRANSFER (SubProtocol: 0x%02X)", s.imei, p.SubProtocol)

	switch p.SubProtocol {
	case protocol.InfoTypeExternalVoltage:
		log.Printf("[%s]   External Voltage: %d mV (%.2f V)",
			s.imei, p.ExternalVoltage, p.GetExternalVoltageVolts())
	case protocol.InfoTypeICCID:
		log.Printf("[%s]   ICCID: %s", s.imei, p.ICCID)
	case protocol.InfoTypeGPSStatus:
		log.Printf("[%s]   GPS Status: %s", s.imei, p.GPSStatus)
	default:
		log.Printf("[%s]   Data (%d bytes): %s", s.imei, len(p.Data), hex.EncodeToString(p.Data))
	}
}

func (s *DeviceSession) handleCommandResponse(p *packet.CommandResponsePacket) {
	log.Printf("[%s] COMMAND RESPONSE", s.imei)
	log.Printf("[%s]   Server Flag: 0x%08X", s.imei, p.ServerFlag)
	log.Printf("[%s]   Response: %s", s.imei, p.Response)
}

func (s *DeviceSession) handleGPSAddressRequest(p *packet.GPSAddressRequestPacket) {
	log.Printf("[%s] GPS ADDRESS REQUEST", s.imei)
	log.Printf("[%s]   Coordinates: %s", s.imei, p.Coordinates)
	log.Printf("[%s]   Language: %s", s.imei, p.Language)

	// Here you would implement reverse geocoding
	// For now, we just log the request
	log.Printf("[%s]   (Reverse geocoding not implemented)", s.imei)
}

func (s *DeviceSession) handleUnknown(pkt packet.Packet) {
	log.Printf("[%s] UNKNOWN PACKET TYPE", s.imei)
	log.Printf("[%s]   Protocol: 0x%02X (%s)", s.imei, pkt.ProtocolNumber(), pkt.Type())
	log.Printf("[%s]   Serial: %d", s.imei, pkt.SerialNumber())
	log.Printf("[%s]   Raw (%d bytes): %s", s.imei, len(pkt.Raw()), hex.EncodeToString(pkt.Raw()))

	// Try to identify patterns in the raw data
	raw := pkt.Raw()
	if len(raw) > 0 {
		log.Printf("[%s]   Analysis:", s.imei)
		if len(raw) >= 2 {
			log.Printf("[%s]     Start bits: 0x%02X%02X", s.imei, raw[0], raw[1])
		}
		if len(raw) >= 3 {
			log.Printf("[%s]     Length byte: %d", s.imei, raw[2])
		}
		if len(raw) >= 4 {
			log.Printf("[%s]     Protocol byte: 0x%02X", s.imei, raw[3])
		}
	}
}

func (s *DeviceSession) sendResponse(data []byte) {
	s.logRawData("TX", data)

	_, err := s.conn.Write(data)
	if err != nil {
		log.Printf("[%s] Failed to send response: %v", s.imei, err)
		return
	}

	log.Printf("[%s] TX: %s", s.imei, hex.EncodeToString(data))
}

// sendCommand sends a command to the device
func (s *DeviceSession) sendCommand(serverFlag uint32, command string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("[%s] Sending command: %s (flag: 0x%08X)", s.imei, command, serverFlag)

	// Use serial number 1 for commands (or track proper serial number)
	response := s.encoder.OnlineCommand(1, serverFlag, command)
	s.sendResponse(response)
}

// Example command methods

func (s *DeviceSession) requestLocation() {
	s.sendCommand(0x00000001, "WHERE#")
}

func (s *DeviceSession) setTrackingInterval(seconds int) {
	cmd := fmt.Sprintf("TIMER,%d#", seconds)
	s.sendCommand(0x00000002, cmd)
}

func (s *DeviceSession) cutFuel() {
	log.Printf("[%s] WARNING: Sending fuel cut command", s.imei)
	s.sendCommand(0x00000003, "RELAY,1#")
}

func (s *DeviceSession) restoreFuel() {
	s.sendCommand(0x00000004, "RELAY,0#")
}

func (s *DeviceSession) requestVersion() {
	s.sendCommand(0x00000005, "VERSION#")
}

func (s *DeviceSession) requestStatus() {
	s.sendCommand(0x00000006, "STATUS#")
}

// Utility functions

func isCriticalAlarm(alarmType protocol.AlarmType) bool {
	switch alarmType {
	case protocol.AlarmSOS,
		protocol.AlarmPowerCut,
		protocol.AlarmTowTheft,
		protocol.AlarmTamper,
		protocol.AlarmCollision:
		return true
	default:
		return false
	}
}

func printSessionSummary() {
	sessionsMu.RLock()
	defer sessionsMu.RUnlock()

	log.Printf("Active sessions: %d", len(sessions))
	for imei, session := range sessions {
		duration := time.Since(session.connectedAt)
		log.Printf("  - %s: connected %s ago, %d packets",
			imei, duration.Round(time.Second), session.packetCount)
	}
}

// GetSession returns a session by IMEI (for external use)
func GetSession(imei string) *DeviceSession {
	sessionsMu.RLock()
	defer sessionsMu.RUnlock()
	return sessions[imei]
}

// GetAllSessions returns all active sessions
func GetAllSessions() map[string]*DeviceSession {
	sessionsMu.RLock()
	defer sessionsMu.RUnlock()
	result := make(map[string]*DeviceSession)
	for k, v := range sessions {
		result[k] = v
	}
	return result
}
