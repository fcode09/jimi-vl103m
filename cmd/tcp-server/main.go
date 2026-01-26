package main

import (
	"io"
	"log"
	"net"

	"github.com/fcode09/jimi-vl103m/pkg/jimi"
	"github.com/fcode09/jimi-vl103m/pkg/jimi/encoder"
	"github.com/fcode09/jimi-vl103m/pkg/jimi/packet"
	"github.com/fcode09/jimi-vl103m/pkg/jimi/protocol"
)

const (
	Port = "5023"
)

func main() {
	listener, err := net.Listen("tcp", ":"+Port)
	if err != nil {
		log.Fatalf("Error starting TCP server: %v", err)
	}
	defer listener.Close()
	log.Printf("TCP server listening on port %s", Port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	log.Printf("New connection from: %s", conn.RemoteAddr().String())

	decoder := jimi.NewDecoder()
	enc := encoder.New()
	buffer := make([]byte, 0, 4096)
	readBuf := make([]byte, 1024)

	for {
		n, err := conn.Read(readBuf)
		if err != nil {
			if err != io.EOF {
				log.Printf("Error reading from connection: %v", err)
			} else {
				log.Printf("Client %s disconnected", conn.RemoteAddr().String())
			}
			return
		}

		buffer = append(buffer, readBuf[:n]...)

		packets, residue, err := decoder.DecodeStream(buffer)
		if err != nil {
			log.Printf("Decode error: %v, buffer will be cleared", err)
			buffer = nil // Clear buffer on error
			continue
		}

		buffer = residue

		for _, p := range packets {
			logPacket(p)

			var response []byte
			switch p.ProtocolNumber() {
			case protocol.ProtocolLogin:
				response = enc.LoginResponse(p.SerialNumber())
			case protocol.ProtocolHeartbeat:
				response = enc.HeartbeatResponse(p.SerialNumber())
			case protocol.ProtocolAlarm, protocol.ProtocolAlarmMultiFence, protocol.ProtocolAlarmMultiFence4G:
				response = enc.AlarmResponse(p.SerialNumber())
			case protocol.ProtocolTimeCalibration:
				response = enc.TimeCalibrationResponseNow(p.SerialNumber())
			}

			if response != nil {
				if _, err := conn.Write(response); err != nil {
					log.Printf("Error sending response to %s: %v", conn.RemoteAddr().String(), err)
				}
			}
		}
	}
}

func logPacket(p packet.Packet) {
	log.Println("-----------------------------------------------------")
	log.Printf("Received Packet: %s (Proto: 0x%02X, SN: %d)", p.Type(), p.ProtocolNumber(), p.SerialNumber())
	log.Printf("  Raw Hex: %X", p.Raw())

	switch v := p.(type) {
	case *packet.LoginPacket:
		log.Printf("  IMEI: %s", v.GetIMEI())
		log.Printf("  ModelID: 0x%04X", v.ModelID)
		log.Printf("  Timezone: %s (Offset: %d mins, Lang: %s)", v.Timezone.String(), v.Timezone.OffsetMinutes, v.Timezone.LanguageString())
	case *packet.HeartbeatPacket:
		log.Printf("  Terminal Info:")
		log.Printf("    ACC: %v, Charging: %v, GPS Tracking: %v, Armed: %v, Power Cut: %v",
			v.TerminalInfo.ACCOn(), v.TerminalInfo.IsCharging(), v.TerminalInfo.GPSTrackingEnabled(), v.TerminalInfo.IsArmed(), v.TerminalInfo.OilElectricityDisconnected())
		log.Printf("  Voltage: %s (%d%%)", v.VoltageLevel.String(), v.VoltageLevel.Percentage())
		log.Printf("  GSM Signal: %s (%d bars)", v.GSMSignal.String(), v.GSMSignal.Bars())
		if v.HasExtended {
			log.Printf("  Extended Info: 0x%04X", v.ExtendedInfo)
		}
	case *packet.LocationPacket:
		log.Printf("  Timestamp: %s", v.DateTime.Time)
		log.Printf("  Latitude: %.6f", v.Coordinates.SignedLatitude())
		log.Printf("  Longitude: %.6f", v.Coordinates.SignedLongitude())
		log.Printf("  Google Maps: https://www.google.com/maps?q=%.6f,%.6f", v.Coordinates.SignedLatitude(), v.Coordinates.SignedLongitude())
		log.Printf("  Satellites: %d", v.Satellites)
		log.Printf("  Speed: %d km/h", v.Speed)
		log.Printf("  Course Status:")
		log.Printf("    Course: %d°, Realtime: %v, Positioned: %v, East lon: %v, North lat: %v",
			v.CourseStatus.Course, v.CourseStatus.IsGPSRealtime, v.CourseStatus.IsPositioned, v.CourseStatus.IsEastLongitude, v.CourseStatus.IsNorthLatitude)
		log.Printf("  LBS Info: MCC: %d, MNC: %d, LAC: %d, CellID: %d", v.LBSInfo.MCC, v.LBSInfo.MNC, v.LBSInfo.LAC, v.LBSInfo.CellID)
		if v.HasStatus {
			log.Printf("  Terminal Info: ACC: %v", v.TerminalInfo.ACCOn())
		}
		log.Printf("  Upload Mode: %s", v.UploadMode.String())
		log.Printf("  Is Re-upload: %v", v.IsReupload)
		log.Printf("  Mileage: %d", v.Mileage)
	case *packet.Location4GPacket:
		log.Printf("  Timestamp: %s", v.DateTime.Time)
		log.Printf("  Latitude: %.6f", v.Coordinates.SignedLatitude())
		log.Printf("  Longitude: %.6f", v.Coordinates.SignedLongitude())
		log.Printf("  Google Maps: https://www.google.com/maps?q=%.6f,%.6f", v.Coordinates.SignedLatitude(), v.Coordinates.SignedLongitude())
		log.Printf("  Satellites: %d", v.Satellites)
		log.Printf("  Speed: %d km/h", v.Speed)
		log.Printf("  Course Status:")
		log.Printf("    Course: %d°, Realtime: %v, Positioned: %v, East lon: %v, North lat: %v",
			v.CourseStatus.Course, v.CourseStatus.IsGPSRealtime, v.CourseStatus.IsPositioned, v.CourseStatus.IsEastLongitude, v.CourseStatus.IsNorthLatitude)
		log.Printf("  LBS Info: MCC: %d, MNC: %d, LAC: %d, CellID: %d", v.LBSInfo.MCC, v.LBSInfo.MNC, v.LBSInfo.LAC, v.LBSInfo.CellID)
		if v.HasStatus {
			log.Printf("  Terminal Info: ACC: %v", v.TerminalInfo.ACCOn())
		}
		log.Printf("  Upload Mode: %s", v.UploadMode.String())
		log.Printf("  Is Re-upload: %v", v.IsReupload)
		log.Printf("  Mileage: %d", v.Mileage)
		log.Printf("  4G MCC/MNC: %d", v.MCCMNC)
		for i, lbs := range v.ExtendedLBS {
			log.Printf("  Extended LBS %d: MCC: %d, MNC: %d, LAC: %d, CellID: %d", i+1, lbs.MCC, lbs.MNC, lbs.LAC, lbs.CellID)
		}
	case *packet.AlarmPacket:
		log.Printf("  Alarm Type: %s (Critical: %v)", v.AlarmType.String(), v.AlarmType.IsCritical())
		log.Printf("  Timestamp: %s", v.DateTime.Time)
		log.Printf("  Latitude: %.6f", v.Coordinates.SignedLatitude())
		log.Printf("  Longitude: %.6f", v.Coordinates.SignedLongitude())
		log.Printf("  Google Maps: https://www.google.com/maps?q=%.6f,%.6f", v.Coordinates.SignedLatitude(), v.Coordinates.SignedLongitude())
		log.Printf("  Satellites: %d", v.Satellites)
		log.Printf("  Speed: %d km/h", v.Speed)
		log.Printf("  Course Status: Course: %d°, Positioned: %v", v.CourseStatus.Course, v.CourseStatus.IsPositioned)
		log.Printf("  LBS Info: MCC: %d, MNC: %d, LAC: %d, CellID: %d", v.LBSInfo.MCC, v.LBSInfo.MNC, v.LBSInfo.LAC, v.LBSInfo.CellID)
		log.Printf("  Terminal Info: ACC: %v, Charging: %v", v.TerminalInfo.ACCOn(), v.TerminalInfo.IsCharging())
		log.Printf("  Voltage: %s", v.VoltageLevel.String())
		log.Printf("  GSM Signal: %s", v.GSMSignal.String())
		log.Printf("  Mileage: %d", v.Mileage)
	case *packet.AlarmMultiFencePacket:
		log.Printf("  Alarm Type: %s", v.AlarmType.String())
		log.Printf("  Fence ID: %d", v.FenceID)
		log.Printf("  Timestamp: %s", v.DateTime.Time)
		log.Printf("  Google Maps: https://www.google.com/maps?q=%.6f,%.6f", v.Coordinates.SignedLatitude(), v.Coordinates.SignedLongitude())
	case *packet.Alarm4GPacket:
		log.Printf("  Alarm Type: %s", v.AlarmType.String())
		log.Printf("  MCC/MNC: %d", v.MCCMNC)
		log.Printf("  Timestamp: %s", v.DateTime.Time)
		log.Printf("  Google Maps: https://www.google.com/maps?q=%.6f,%.6f", v.Coordinates.SignedLatitude(), v.Coordinates.SignedLongitude())
		log.Printf("  Satellites: %d", v.Satellites)
		log.Printf("  Speed: %d km/h", v.Speed)
		log.Printf("  Course Status: Course: %d°, Positioned: %v", v.CourseStatus.Course, v.CourseStatus.IsPositioned)
		log.Printf("  LBS Info: MCC: %d, MNC: %d, LAC: %d, CellID: %d", v.LBSInfo.MCC, v.LBSInfo.MNC, v.LBSInfo.LAC, v.LBSInfo.CellID)
		log.Printf("  Terminal Info: ACC: %v, Charging: %v", v.TerminalInfo.ACCOn(), v.TerminalInfo.IsCharging())
		log.Printf("  Voltage: %s", v.VoltageLevel.String())
		log.Printf("  GSM Signal: %s", v.GSMSignal.String())
		log.Printf("  Mileage: %d", v.Mileage)
		log.Printf("  4G MCC/MNC: %d", v.MCCMNC)
		for i, lbs := range v.ExtendedLBS {
			log.Printf("  Extended LBS %d: MCC: %d, MNC: %d, LAC: %d, CellID: %d", i+1, lbs.MCC, lbs.MNC, lbs.LAC, lbs.CellID)
		}
	case *packet.InfoTransferPacket:
		log.Printf("  Sub-protocol: %s", v.SubProtocol.String())
		switch v.SubProtocol {
		case protocol.InfoTypeExternalVoltage:
			log.Printf("  External Voltage: %.2f V", v.GetExternalVoltageVolts())
		case protocol.InfoTypeICCID:
			log.Printf("  ICCID: %s", v.ICCID)
		case protocol.InfoTypeGPSStatus:
			log.Printf("  GPS Module Status: %s", v.GPSStatus.String())
		default:
			log.Printf("  Data: %X", v.Data)
		}
	case *packet.GPSAddressRequestPacket:
		log.Printf("  Coordinates: %s", v.Coordinates)
		log.Printf("  Language: %s", v.Language.String())
	case *packet.TimeCalibrationPacket:
		log.Printf("  Device requested server time synchronization.")
	case *packet.LBSPacket:
		log.Printf("  Timestamp: %s", v.DateTime.Time)
		log.Printf("  Main Cell: MCC: %d, MNC: %d, LAC: %d, CellID: %d", v.LBSInfo.MCC, v.LBSInfo.MNC, v.LBSInfo.LAC, v.LBSInfo.CellID)
		for i, cell := range v.NeighborCells {
			log.Printf("  Neighbor Cell %d: LAC: %d, CellID: %d", i+1, cell.LAC, cell.CellID)
		}
		log.Printf("  Timing Advance: %d", v.TimingAdvance)
		log.Printf("  Language: %s", v.Language.String())
		if v.HasStatus {
			log.Printf("  Terminal Info: ACC: %v", v.TerminalInfo.ACCOn())
			log.Printf("  Voltage: %s", v.VoltageLevel.String())
			log.Printf("  GSM Signal: %s", v.GSMSignal.String())
		}
	case *packet.LBS4GPacket:
		log.Printf("  Timestamp: %s", v.DateTime.Time)
		log.Printf("  LBS Info: MCC: %d, MNC: %d, LAC: %d, CellID: %d", v.LBSInfo.MCC, v.LBSInfo.MNC, v.LBSInfo.LAC, v.LBSInfo.CellID)
		for i, lbs := range v.NeighborCells {
			log.Printf("  Neighbor Cell %d: MCC: %d, MNC: %d, LAC: %d, CellID: %d", i+1, lbs.MCC, lbs.MNC, lbs.LAC, lbs.CellID)
		}
	default:
		log.Printf("  Unknown Packet Data: %+v\n", v)
	}
	log.Println("-----------------------------------------------------")
}
