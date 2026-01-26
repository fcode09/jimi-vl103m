// Example: TCP Server for GPS Trackers
//
// This example demonstrates how to create a TCP server
// that receives and processes GPS tracker packets.
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
	"sync"
	"syscall"
	"time"

	"github.com/intelcon-group/jimi-vl103m/pkg/jimi"
	"github.com/intelcon-group/jimi-vl103m/pkg/jimi/encoder"
	"github.com/intelcon-group/jimi-vl103m/pkg/jimi/packet"
	"github.com/intelcon-group/jimi-vl103m/pkg/jimi/protocol"
)

var (
	port = flag.Int("port", 8080, "TCP server port")
)

// DeviceSession represents a connected device
type DeviceSession struct {
	conn     net.Conn
	decoder  *jimi.Decoder
	encoder  *encoder.Encoder
	imei     string
	lastSeen time.Time
	mu       sync.Mutex
}

func main() {
	flag.Parse()

	log.Printf("Starting GPS Tracker Server on port %d...\n", *port)

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
		log.Println("Shutting down server...")
		listener.Close()
		os.Exit(0)
	}()

	log.Println("Server started. Waiting for connections...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}

		log.Printf("New connection from %s", conn.RemoteAddr())
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	session := &DeviceSession{
		conn:     conn,
		decoder:  jimi.NewDecoder(jimi.WithStrictMode(false)),
		encoder:  encoder.New(),
		lastSeen: time.Now(),
	}

	buffer := make([]byte, 0, 4096)
	readBuf := make([]byte, 1024)

	// Set read timeout
	conn.SetReadDeadline(time.Now().Add(5 * time.Minute))

	for {
		n, err := conn.Read(readBuf)
		if err != nil {
			if err != io.EOF {
				log.Printf("[%s] Read error: %v", session.imei, err)
			}
			break
		}

		if n == 0 {
			continue
		}

		buffer = append(buffer, readBuf[:n]...)
		session.lastSeen = time.Now()

		// Reset read deadline
		conn.SetReadDeadline(time.Now().Add(5 * time.Minute))

		// Try to decode packets
		packets, residue, err := session.decoder.DecodeStream(buffer)
		if err != nil {
			log.Printf("[%s] Decode error: %v", session.imei, err)
		}

		buffer = residue

		// Process each packet
		for _, pkt := range packets {
			session.handlePacket(pkt)
		}
	}

	log.Printf("[%s] Connection closed", session.imei)
}

func (s *DeviceSession) handlePacket(pkt packet.Packet) {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("[%s] Received: %s (protocol 0x%02X, serial %d)",
		s.imei, pkt.Type(), pkt.ProtocolNumber(), pkt.SerialNumber())

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

	default:
		log.Printf("[%s] Unhandled packet type: %s", s.imei, pkt.Type())
	}
}

func (s *DeviceSession) handleLogin(p *packet.LoginPacket) {
	s.imei = p.IMEI.String()

	log.Printf("[%s] Login - IMEI: %s, ModelID: 0x%04X, Timezone: %s",
		s.imei, p.IMEI, p.ModelID, p.Timezone)

	// Send login response
	response := s.encoder.LoginResponse(p.SerialNumber())
	s.sendResponse(response)
}

func (s *DeviceSession) handleHeartbeat(p *packet.HeartbeatPacket) {
	log.Printf("[%s] Heartbeat - Voltage: %s (%d%%), GSM: %s (%d bars), ACC: %v",
		s.imei,
		p.VoltageLevel, p.BatteryPercentage(),
		p.GSMSignal, p.SignalBars(),
		p.ACCOn())

	// Send heartbeat response
	response := s.encoder.HeartbeatResponse(p.SerialNumber())
	s.sendResponse(response)
}

func (s *DeviceSession) handleLocation(p *packet.LocationPacket) {
	log.Printf("[%s] Location - Time: %s, Lat: %.6f, Lon: %.6f, Speed: %d km/h, Heading: %d° (%s)",
		s.imei,
		p.DateTime,
		p.Latitude(), p.Longitude(),
		p.Speed,
		p.Heading(), p.HeadingName())

	log.Printf("[%s]   Satellites: %d, Positioned: %v, LBS: %s",
		s.imei, p.Satellites, p.IsPositioned(), p.LBSInfo)

	// Location packets typically don't require response
	// But you can send one if needed
}

func (s *DeviceSession) handleAlarm(p *packet.AlarmPacket) {
	log.Printf("[%s] ALARM - Type: %s, Critical: %v",
		s.imei, p.AlarmType, p.IsCritical())
	log.Printf("[%s]   Location: %.6f, %.6f at %s",
		s.imei, p.Latitude(), p.Longitude(), p.DateTime)

	// Handle critical alarms
	if p.IsCritical() {
		log.Printf("[%s] *** CRITICAL ALARM: %s ***", s.imei, p.AlarmType)
		// Here you would trigger notifications, alerts, etc.
	}

	// Send alarm response (required by protocol)
	response := s.encoder.AlarmResponse(p.SerialNumber())
	s.sendResponse(response)
}

func (s *DeviceSession) handleTimeCalibration(p *packet.TimeCalibrationPacket) {
	log.Printf("[%s] Time Calibration Request", s.imei)

	// Send current server time
	response := s.encoder.TimeCalibrationResponseNow(p.SerialNumber())
	s.sendResponse(response)
}

func (s *DeviceSession) handleLBS(p *packet.LBSPacket) {
	log.Printf("[%s] LBS - Time: %s, MCC: %d, MNC: %d, LAC: %d, CellID: %d",
		s.imei,
		p.DateTime,
		p.LBSInfo.MCC, p.LBSInfo.MNC,
		p.LBSInfo.LAC, p.LBSInfo.CellID)

	// LBS packets typically don't require response
}

func (s *DeviceSession) sendResponse(data []byte) {
	_, err := s.conn.Write(data)
	if err != nil {
		log.Printf("[%s] Failed to send response: %v", s.imei, err)
		return
	}

	log.Printf("[%s] Sent response: %s", s.imei, hex.EncodeToString(data))
}

// sendCommand sends a command to the device
func (s *DeviceSession) sendCommand(serverFlag uint32, command string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Use serial number 1 for commands (or track proper serial number)
	response := s.encoder.OnlineCommand(1, serverFlag, command)
	s.sendResponse(response)

	log.Printf("[%s] Sent command: %s", s.imei, command)
}

// Example: Request device location
func (s *DeviceSession) requestLocation() {
	s.sendCommand(0x00000001, "WHERE#")
}

// Example: Set tracking interval
func (s *DeviceSession) setTrackingInterval(seconds int) {
	cmd := fmt.Sprintf("TIMER,%d#", seconds)
	s.sendCommand(0x00000002, cmd)
}

// Example: Cut fuel (immobilizer)
func (s *DeviceSession) cutFuel() {
	// WARNING: Only use when vehicle is stationary!
	log.Printf("[%s] WARNING: Sending fuel cut command", s.imei)
	s.sendCommand(0x00000003, "RELAY,1#")
}

// Example: Restore fuel
func (s *DeviceSession) restoreFuel() {
	s.sendCommand(0x00000004, "RELAY,0#")
}

// Utility to check if alarm is critical
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
