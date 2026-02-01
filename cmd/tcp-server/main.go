// TCP Server for Jimi VL103M GPS Trackers
//
// This is a production-ready TCP server for receiving and processing
// GPS tracker packets with comprehensive logging and raw data capture.
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

// Configuration flags
var (
	port       = flag.Int("port", 5023, "TCP server port")
	logDir     = flag.String("logdir", "logs", "Directory to store raw packet logs")
	verbose    = flag.Bool("verbose", false, "Enable verbose raw data logging")
	saveRaw    = flag.Bool("save-raw", true, "Save raw packets to files")
	strictMode = flag.Bool("strict", false, "Enable strict mode parsing")
	timeout    = flag.Duration("timeout", 5*time.Minute, "Connection read timeout")
)

// DeviceSession represents a connected GPS tracker device
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
	printBanner()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Error starting TCP server: %v", err)
	}
	defer listener.Close()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("\n" + strings.Repeat("=", 60))
		log.Println("Shutting down server...")
		printSessionSummary()
		listener.Close()
		os.Exit(0)
	}()

	log.Printf("Server started. Waiting for connections...")
	log.Println("")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go handleConnection(conn)
	}
}

func printBanner() {
	log.Println(strings.Repeat("=", 60))
	log.Println("Jimi VL103M GPS Tracker Server")
	log.Println(strings.Repeat("=", 60))
	log.Printf("Port:            %d", *port)
	log.Printf("Log Directory:   %s", *logDir)
	log.Printf("Verbose:         %v", *verbose)
	log.Printf("Save Raw:        %v", *saveRaw)
	log.Printf("Strict Mode:     %v", *strictMode)
	log.Printf("Read Timeout:    %v", *timeout)
	log.Println(strings.Repeat("=", 60))
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	remoteAddr := conn.RemoteAddr().String()
	connectedAt := time.Now()

	log.Printf(">>> New connection from %s", remoteAddr)

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
			strings.ReplaceAll(strings.ReplaceAll(remoteAddr, ":", "-"), ".", "_"),
			connectedAt.Format("20060102_150405"))
		fpath := filepath.Join(*logDir, filename)
		f, err := os.Create(fpath)
		if err != nil {
			log.Printf("[%s] Warning: Failed to create raw log file: %v", remoteAddr, err)
		} else {
			session.rawLogFile = f
			defer f.Close()
			writeLogHeader(f, remoteAddr, connectedAt)
		}
	}

	buffer := make([]byte, 0, 4096)
	readBuf := make([]byte, 1024)

	conn.SetReadDeadline(time.Now().Add(*timeout))

	for {
		n, err := conn.Read(readBuf)
		if err != nil {
			if err != io.EOF {
				log.Printf("[%s] Read error: %v", session.getIdentifier(), err)
			} else {
				log.Printf("[%s] Client disconnected", session.getIdentifier())
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
		conn.SetReadDeadline(time.Now().Add(*timeout))

		// Try to decode packets
		packets, residue, err := session.decoder.DecodeStream(buffer)
		if err != nil {
			log.Printf("[%s] Decode error: %v", session.getIdentifier(), err)
			if *verbose {
				log.Printf("[%s] Buffer at error (%d bytes): %s",
					session.getIdentifier(), len(buffer), hex.EncodeToString(buffer))
			}
		}

		buffer = residue

		// Process each packet
		for _, p := range packets {
			session.packetCount++
			session.processPacket(p)
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

func writeLogHeader(f *os.File, remoteAddr string, connectedAt time.Time) {
	f.WriteString("# Jimi VL103M GPS Tracker Raw Packet Log\n")
	f.WriteString(fmt.Sprintf("# Connection: %s\n", remoteAddr))
	f.WriteString(fmt.Sprintf("# Started: %s\n", connectedAt.Format(time.RFC3339)))
	f.WriteString("# Format: [timestamp] [direction] [hex_data]\n")
	f.WriteString("#\n")
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

func (s *DeviceSession) processPacket(p packet.Packet) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Log the packet details
	logPacket(p, s.getIdentifier(), s.packetCount)

	// Handle IMEI registration on login
	if login, ok := p.(*packet.LoginPacket); ok {
		s.imei = login.GetIMEI()

		sessionsMu.Lock()
		sessions[s.imei] = s
		sessionsMu.Unlock()

		// Rename raw log file with IMEI
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
	}

	// Send response if required
	response := s.buildResponse(p)
	if response != nil {
		s.sendResponse(response)
	}
}

func (s *DeviceSession) buildResponse(p packet.Packet) []byte {
	switch p.ProtocolNumber() {
	case protocol.ProtocolLogin:
		return s.encoder.LoginResponse(p.SerialNumber())
	case protocol.ProtocolHeartbeat:
		return s.encoder.HeartbeatResponse(p.SerialNumber())
	case protocol.ProtocolAlarm, protocol.ProtocolAlarmMultiFence, protocol.ProtocolAlarmMultiFence4G:
		return s.encoder.AlarmResponse(p.SerialNumber())
	case protocol.ProtocolTimeCalibration:
		return s.encoder.TimeCalibrationResponseNow(p.SerialNumber())
	default:
		return nil
	}
}

func (s *DeviceSession) sendResponse(data []byte) {
	s.logRawData("TX", data)

	_, err := s.conn.Write(data)
	if err != nil {
		log.Printf("[%s] Failed to send response: %v", s.getIdentifier(), err)
		return
	}

	if *verbose {
		log.Printf("[%s] TX: %s", s.getIdentifier(), hex.EncodeToString(data))
	}
}

func logPacket(p packet.Packet, identifier string, packetNum int) {
	log.Println(strings.Repeat("-", 60))
	log.Printf("[%s] PKT #%d: %s (0x%02X) Serial: %d",
		identifier, packetNum, p.Type(), p.ProtocolNumber(), p.SerialNumber())
	log.Printf("[%s] Raw: %X", identifier, p.Raw())

	switch v := p.(type) {
	case *packet.LoginPacket:
		log.Printf("[%s] LOGIN", identifier)
		log.Printf("[%s]   IMEI: %s", identifier, v.GetIMEI())
		log.Printf("[%s]   Model ID: 0x%04X", identifier, v.ModelID)
		log.Printf("[%s]   Timezone: %s (Offset: %d mins, Lang: %s)",
			identifier, v.Timezone.String(), v.Timezone.OffsetMinutes, v.Timezone.LanguageString())

	case *packet.HeartbeatPacket:
		log.Printf("[%s] HEARTBEAT", identifier)
		log.Printf("[%s]   Terminal Info:", identifier)
		log.Printf("[%s]     ACC: %v, Charging: %v, GPS Tracking: %v, Armed: %v, Power Cut: %v",
			identifier,
			v.TerminalInfo.ACCOn(),
			v.TerminalInfo.IsCharging(),
			v.TerminalInfo.GPSTrackingEnabled(),
			v.TerminalInfo.IsArmed(),
			v.TerminalInfo.OilElectricityDisconnected())
		log.Printf("[%s]   Voltage: %s (%d%%)", identifier, v.VoltageLevel.String(), v.VoltageLevel.Percentage())
		log.Printf("[%s]   GSM Signal: %s (%d bars)", identifier, v.GSMSignal.String(), v.GSMSignal.Bars())
		if v.HasExtended {
			log.Printf("[%s]   Extended Info: 0x%04X", identifier, v.ExtendedInfo)
		}

	case *packet.LocationPacket:
		log.Printf("[%s] LOCATION", identifier)
		log.Printf("[%s]   Timestamp: %s", identifier, v.DateTime.Time)
		log.Printf("[%s]   Position: %.6f, %.6f", identifier, v.Coordinates.SignedLatitude(), v.Coordinates.SignedLongitude())
		log.Printf("[%s]   Google Maps: https://www.google.com/maps?q=%.6f,%.6f",
			identifier, v.Coordinates.SignedLatitude(), v.Coordinates.SignedLongitude())
		log.Printf("[%s]   Satellites: %d", identifier, v.Satellites)
		log.Printf("[%s]   Speed: %d km/h", identifier, v.Speed)
		log.Printf("[%s]   Course: %d° | Realtime: %v | Positioned: %v | East: %v | North: %v",
			identifier,
			v.CourseStatus.Course,
			v.CourseStatus.IsGPSRealtime,
			v.CourseStatus.IsPositioned,
			v.CourseStatus.IsEastLongitude,
			v.CourseStatus.IsNorthLatitude)
		if v.LBSInfo.IsValid() {
			log.Printf("[%s]   LBS: MCC=%d MNC=%d LAC=%d CellID=%d",
				identifier, v.LBSInfo.MCC, v.LBSInfo.MNC, v.LBSInfo.LAC, v.LBSInfo.CellID)
		}
		if v.HasStatus {
			// For GPS Location packets, ACC is in a dedicated field, not in TerminalInfo
			// TerminalInfo is only used for other status bits in this packet type
			terminalInfo := v.TerminalInfo.String()
			// Remove ACC status from TerminalInfo string since we show it separately
			terminalInfo = strings.ReplaceAll(terminalInfo, "ACC:OFF, ", "")
			terminalInfo = strings.ReplaceAll(terminalInfo, "ACC:ON, ", "")
			terminalInfo = strings.ReplaceAll(terminalInfo, ", ACC:OFF", "")
			terminalInfo = strings.ReplaceAll(terminalInfo, ", ACC:ON", "")
			terminalInfo = strings.ReplaceAll(terminalInfo, "ACC:OFF", "")
			terminalInfo = strings.ReplaceAll(terminalInfo, "ACC:ON", "")
			if terminalInfo == "" || terminalInfo == "Normal" {
				log.Printf("[%s]   Terminal: ACC:%s", identifier,
					map[bool]string{true: "ON", false: "OFF"}[v.ACC])
			} else {
				log.Printf("[%s]   Terminal: ACC:%s, %s", identifier,
					map[bool]string{true: "ON", false: "OFF"}[v.ACC],
					terminalInfo)
			}
		}
		log.Printf("[%s]   Upload Mode: %s", identifier, v.UploadMode.String())
		log.Printf("[%s]   Re-upload: %v | Mileage: %d m", identifier, v.IsReupload, v.Mileage)

	case *packet.Location4GPacket:
		log.Printf("[%s] LOCATION 4G", identifier)
		log.Printf("[%s]   Timestamp: %s", identifier, v.DateTime.Time)
		log.Printf("[%s]   Position: %.6f, %.6f", identifier, v.Coordinates.SignedLatitude(), v.Coordinates.SignedLongitude())
		log.Printf("[%s]   Google Maps: https://www.google.com/maps?q=%.6f,%.6f",
			identifier, v.Coordinates.SignedLatitude(), v.Coordinates.SignedLongitude())
		log.Printf("[%s]   Satellites: %d", identifier, v.Satellites)
		log.Printf("[%s]   Speed: %d km/h", identifier, v.Speed)
		log.Printf("[%s]   Course: %d° | Realtime: %v | Positioned: %v",
			identifier, v.CourseStatus.Course, v.CourseStatus.IsGPSRealtime, v.CourseStatus.IsPositioned)
		if v.LBSInfo.IsValid() {
			log.Printf("[%s]   LBS: MCC=%d MNC=%d LAC=%d CellID=%d",
				identifier, v.LBSInfo.MCC, v.LBSInfo.MNC, v.LBSInfo.LAC, v.LBSInfo.CellID)
		}
		if v.HasStatus {
			// For GPS Location 4G packets, ACC is in a dedicated field, not in TerminalInfo
			// TerminalInfo is only used for other status bits in this packet type
			terminalInfo := v.TerminalInfo.String()
			// Remove ACC status from TerminalInfo string since we show it separately
			terminalInfo = strings.ReplaceAll(terminalInfo, "ACC:OFF, ", "")
			terminalInfo = strings.ReplaceAll(terminalInfo, "ACC:ON, ", "")
			terminalInfo = strings.ReplaceAll(terminalInfo, ", ACC:OFF", "")
			terminalInfo = strings.ReplaceAll(terminalInfo, ", ACC:ON", "")
			terminalInfo = strings.ReplaceAll(terminalInfo, "ACC:OFF", "")
			terminalInfo = strings.ReplaceAll(terminalInfo, "ACC:ON", "")
			if terminalInfo == "" || terminalInfo == "Normal" {
				log.Printf("[%s]   Terminal: ACC:%s", identifier,
					map[bool]string{true: "ON", false: "OFF"}[v.ACC])
			} else {
				log.Printf("[%s]   Terminal: ACC:%s, %s", identifier,
					map[bool]string{true: "ON", false: "OFF"}[v.ACC],
					terminalInfo)
			}
		}
		log.Printf("[%s]   Upload Mode: %s | Mileage: %d m", identifier, v.UploadMode.String(), v.Mileage)
		log.Printf("[%s]   4G MCC/MNC: %d", identifier, v.MCCMNC)
		for i, lbs := range v.ExtendedLBS {
			log.Printf("[%s]   Extended LBS[%d]: MCC=%d MNC=%d LAC=%d CellID=%d",
				identifier, i, lbs.MCC, lbs.MNC, lbs.LAC, lbs.CellID)
		}

	case *packet.AlarmPacket:
		log.Printf("[%s] ALARM: %s (Critical: %v)", identifier, v.AlarmType.String(), v.AlarmType.IsCritical())
		log.Printf("[%s]   Timestamp: %s", identifier, v.DateTime.Time)
		log.Printf("[%s]   Position: %.6f, %.6f", identifier, v.Coordinates.SignedLatitude(), v.Coordinates.SignedLongitude())
		log.Printf("[%s]   Google Maps: https://www.google.com/maps?q=%.6f,%.6f",
			identifier, v.Coordinates.SignedLatitude(), v.Coordinates.SignedLongitude())
		log.Printf("[%s]   Satellites: %d | Speed: %d km/h", identifier, v.Satellites, v.Speed)
		log.Printf("[%s]   Course: %d° | Positioned: %v", identifier, v.CourseStatus.Course, v.CourseStatus.IsPositioned)
		if v.LBSInfo.IsValid() {
			log.Printf("[%s]   LBS: MCC=%d MNC=%d LAC=%d CellID=%d",
				identifier, v.LBSInfo.MCC, v.LBSInfo.MNC, v.LBSInfo.LAC, v.LBSInfo.CellID)
		}
		log.Printf("[%s]   Terminal: %s", identifier, v.TerminalInfo)
		log.Printf("[%s]   Voltage: %s | GSM: %s", identifier, v.VoltageLevel.String(), v.GSMSignal.String())
		log.Printf("[%s]   Mileage: %d m", identifier, v.Mileage)

	case *packet.AlarmMultiFencePacket:
		log.Printf("[%s] ALARM MULTI-FENCE: %s", identifier, v.AlarmType.String())
		log.Printf("[%s]   Fence ID: %d", identifier, v.FenceID)
		log.Printf("[%s]   Timestamp: %s", identifier, v.DateTime.Time)
		log.Printf("[%s]   Google Maps: https://www.google.com/maps?q=%.6f,%.6f",
			identifier, v.Coordinates.SignedLatitude(), v.Coordinates.SignedLongitude())

	case *packet.Alarm4GPacket:
		log.Printf("[%s] ALARM 4G: %s (Critical: %v)", identifier, v.AlarmType.String(), v.AlarmType.IsCritical())
		log.Printf("[%s]   Timestamp: %s", identifier, v.DateTime.Time)
		log.Printf("[%s]   Position: %.6f, %.6f", identifier, v.Coordinates.SignedLatitude(), v.Coordinates.SignedLongitude())
		log.Printf("[%s]   Google Maps: https://www.google.com/maps?q=%.6f,%.6f",
			identifier, v.Coordinates.SignedLatitude(), v.Coordinates.SignedLongitude())
		log.Printf("[%s]   Satellites: %d | Speed: %d km/h", identifier, v.Satellites, v.Speed)
		log.Printf("[%s]   Terminal: %s", identifier, v.TerminalInfo)
		log.Printf("[%s]   Voltage: %s | GSM: %s", identifier, v.VoltageLevel.String(), v.GSMSignal.String())
		log.Printf("[%s]   4G MCC/MNC: %d | Mileage: %d m", identifier, v.MCCMNC, v.Mileage)
		for i, lbs := range v.ExtendedLBS {
			log.Printf("[%s]   Extended LBS[%d]: MCC=%d MNC=%d LAC=%d CellID=%d",
				identifier, i, lbs.MCC, lbs.MNC, lbs.LAC, lbs.CellID)
		}

	case *packet.InfoTransferPacket:
		log.Printf("[%s] INFO TRANSFER: %s (0x%02X)", identifier, v.SubProtocol.String(), v.SubProtocol)
		switch v.SubProtocol {
		case protocol.InfoTypeExternalVoltage:
			log.Printf("[%s]   External Voltage: %d (%.2f V)", identifier, v.ExternalVoltage, v.GetExternalVoltageVolts())
		case protocol.InfoTypeICCID:
			log.Printf("[%s]   IMEI: %s", identifier, v.IMEI)
			log.Printf("[%s]   IMSI: %s", identifier, v.IMSI)
			log.Printf("[%s]   ICCID: %s", identifier, v.ICCID)
		case protocol.InfoTypeGPSStatus:
			log.Printf("[%s]   GPS Module Status: %s", identifier, v.GPSStatus.String())
			if v.GPSStatusInfo != nil {
				log.Printf("[%s]   Satellites in Fix: %d", identifier, v.GPSStatusInfo.SatellitesInFix)
				log.Printf("[%s]   Visible Satellites: %d", identifier, v.GPSStatusInfo.VisibleSatellites)
				if v.GPSStatusInfo.BDSModuleStatus != 0 {
					log.Printf("[%s]   BDS Status: %s", identifier, v.GPSStatusInfo.BDSModuleStatus.String())
				}
			}
		case protocol.InfoTypeTerminalSync:
			log.Printf("[%s]   Terminal Sync Data:", identifier)
			if v.TerminalSync != nil {
				if v.TerminalSync.ICCID != "" {
					log.Printf("[%s]     ICCID: %s", identifier, v.TerminalSync.ICCID)
				}
				if v.TerminalSync.IMSI != "" {
					log.Printf("[%s]     IMSI: %s", identifier, v.TerminalSync.IMSI)
				}
				if v.TerminalSync.CenterNumber != "" {
					log.Printf("[%s]     Center Number: %s", identifier, v.TerminalSync.CenterNumber)
				}
				if len(v.TerminalSync.SOSNumbers) > 0 {
					log.Printf("[%s]     SOS Numbers: %v", identifier, v.TerminalSync.SOSNumbers)
				}
				if v.TerminalSync.ALM1 != "" {
					log.Printf("[%s]     Alarm Config: ALM1=%s ALM2=%s ALM3=%s ALM4=%s",
						identifier, v.TerminalSync.ALM1, v.TerminalSync.ALM2, v.TerminalSync.ALM3, v.TerminalSync.ALM4)
				}
				if v.TerminalSync.STA1 != "" {
					log.Printf("[%s]     Status: STA1=%s", identifier, v.TerminalSync.STA1)
				}
				if v.TerminalSync.DYD != "" {
					log.Printf("[%s]     Fuel/Power Cutoff: DYD=%s", identifier, v.TerminalSync.DYD)
				}
				if len(v.TerminalSync.Geofences) > 0 {
					log.Printf("[%s]     Geofences: %d configured", identifier, len(v.TerminalSync.Geofences))
					for _, gf := range v.TerminalSync.Geofences {
						status := "OFF"
						if gf.Enabled {
							status = "ON"
						}
						log.Printf("[%s]       GFENCE%d: %s, Lat=%.6f, Lon=%.6f, Radius=%dm, Dir=%s",
							identifier, gf.ID, status, gf.Latitude, gf.Longitude, gf.Radius, gf.Direction)
					}
				}
			} else {
				// Show raw data as string for debugging
				log.Printf("[%s]     Raw: %s", identifier, v.GetDataAsString())
			}
		case protocol.InfoTypeDoorStatus:
			if v.DoorStatus != nil {
				log.Printf("[%s]   Door: Open=%v, TriggerHigh=%v, IOHigh=%v",
					identifier, v.DoorStatus.DoorOpen, v.DoorStatus.TriggerHigh, v.DoorStatus.IOPortHigh)
			} else {
				log.Printf("[%s]   Door Status Data: %X", identifier, v.Data)
			}
		default:
			// Show raw data for unknown sub-protocols
			if len(v.Data) <= 64 {
				log.Printf("[%s]   Data (%d bytes): %X", identifier, len(v.Data), v.Data)
			} else {
				log.Printf("[%s]   Data (%d bytes): %X... (truncated)", identifier, len(v.Data), v.Data[:64])
			}
			// Also try to show as string if it looks like ASCII
			if isASCII(v.Data) {
				log.Printf("[%s]   As String: %s", identifier, string(v.Data))
			}
		}

	case *packet.GPSAddressRequestPacket:
		log.Printf("[%s] GPS ADDRESS REQUEST", identifier)
		log.Printf("[%s]   Timestamp: %s", identifier, v.DateTime.Time)
		log.Printf("[%s]   Position: %.6f, %.6f", identifier, v.Latitude(), v.Longitude())
		log.Printf("[%s]   Google Maps: https://www.google.com/maps?q=%.6f,%.6f",
			identifier, v.Latitude(), v.Longitude())
		log.Printf("[%s]   Speed: %d km/h | Heading: %d°", identifier, v.Speed, v.Heading())
		log.Printf("[%s]   Satellites: %d | Positioned: %v", identifier, v.Satellites, v.IsPositioned())
		log.Printf("[%s]   Phone Number: %s", identifier, v.PhoneNumber)
		log.Printf("[%s]   Alarm Type: %s | Language: %s", identifier, v.AlarmType, v.Language.String())

	case *packet.TimeCalibrationPacket:
		log.Printf("[%s] TIME CALIBRATION REQUEST", identifier)
		log.Printf("[%s]   Device requested server time synchronization", identifier)

	case *packet.LBSPacket:
		log.Printf("[%s] LBS (Cell Tower)", identifier)
		log.Printf("[%s]   Timestamp: %s", identifier, v.DateTime.Time)
		log.Printf("[%s]   Main Cell: MCC=%d MNC=%d LAC=%d CellID=%d",
			identifier, v.LBSInfo.MCC, v.LBSInfo.MNC, v.LBSInfo.LAC, v.LBSInfo.CellID)
		for i, cell := range v.NeighborCells {
			log.Printf("[%s]   Neighbor[%d]: LAC=%d CellID=%d", identifier, i, cell.LAC, cell.CellID)
		}
		log.Printf("[%s]   Timing Advance: %d", identifier, v.TimingAdvance)
		log.Printf("[%s]   Language: %s", identifier, v.Language.String())
		if v.HasStatus {
			log.Printf("[%s]   Terminal: %s", identifier, v.TerminalInfo)
			log.Printf("[%s]   Voltage: %s | GSM: %s", identifier, v.VoltageLevel.String(), v.GSMSignal.String())
		}

	case *packet.LBS4GPacket:
		log.Printf("[%s] LBS 4G (Cell Tower)", identifier)
		log.Printf("[%s]   Timestamp: %s", identifier, v.DateTime.Time)
		log.Printf("[%s]   Main Cell: MCC=%d MNC=%d LAC=%d CellID=%d",
			identifier, v.LBSInfo.MCC, v.LBSInfo.MNC, v.LBSInfo.LAC, v.LBSInfo.CellID)
		for i, lbs := range v.NeighborCells {
			log.Printf("[%s]   Neighbor[%d]: MCC=%d MNC=%d LAC=%d CellID=%d",
				identifier, i, lbs.MCC, lbs.MNC, lbs.LAC, lbs.CellID)
		}
		log.Printf("[%s]   Terminal: %s", identifier, v.TerminalInfo)
		log.Printf("[%s]   Voltage: %s | GSM: %s", identifier, v.VoltageLevel.String(), v.GSMSignal.String())

	case *packet.CommandResponsePacket:
		log.Printf("[%s] COMMAND RESPONSE", identifier)
		log.Printf("[%s]   Server Flag: 0x%08X", identifier, v.ServerFlag)
		log.Printf("[%s]   Response: %s", identifier, v.Response)

	default:
		log.Printf("[%s] UNKNOWN PACKET", identifier)
		log.Printf("[%s]   Protocol: 0x%02X", identifier, p.ProtocolNumber())
		log.Printf("[%s]   Data: %+v", identifier, p)
	}

	log.Println(strings.Repeat("-", 60))
}

// isASCII checks if byte slice contains printable ASCII
func isASCII(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	for _, b := range data {
		if b < 0x20 || b > 0x7E {
			// Allow common control chars like newline, tab
			if b != '\n' && b != '\r' && b != '\t' {
				return false
			}
		}
	}
	return true
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

// SendCommand sends a command to a device by IMEI
func SendCommand(imei string, serverFlag uint32, command string) error {
	session := GetSession(imei)
	if session == nil {
		return fmt.Errorf("device %s not connected", imei)
	}

	session.mu.Lock()
	defer session.mu.Unlock()

	response := session.encoder.OnlineCommand(1, serverFlag, command)
	session.sendResponse(response)

	log.Printf("[%s] Sent command: %s (flag: 0x%08X)", imei, command, serverFlag)
	return nil
}
