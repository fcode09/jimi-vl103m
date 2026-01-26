package protocol

// Protocol numbers for different packet types as defined in JM-VL03 specification v1.1.2
const (
	// Login and Authentication
	ProtocolLogin = 0x01 // Login packet - device authentication with IMEI

	// Status and Monitoring
	ProtocolHeartbeat = 0x13 // Heartbeat packet - keep-alive with battery/signal info

	// GPS Location (2G/3G)
	ProtocolGPSLocation = 0x22 // GPS location packet (UTC, 2G/3G base station)

	// LBS (Location Based Service) - 2G/3G
	ProtocolLBSMultiBase = 0x28 // LBS multi-base extended information packet (2G/3G)

	// Alarm Packets - 2G/3G
	ProtocolAlarm             = 0x26 // Alarm data packet (UTC, single geofence)
	ProtocolAlarmMultiFence   = 0x27 // Alarm data packet (UTC, multiple geofences)
	ProtocolGPSAddressRequest = 0x2A // GPS address request packet (UTC)

	// Command and Control
	ProtocolOnlineCommand      = 0x80 // Online command from server to terminal
	ProtocolCommandResponse    = 0x21 // Response to online command by terminal (universal version)
	ProtocolCommandResponseOld = 0x15 // Response to online command (old version)

	// Time Synchronization
	ProtocolTimeCalibration = 0x8A // Time calibration packet (UTC)

	// Information Transfer
	ProtocolInfoTransfer = 0x94 // Information transfer packet (various device info)

	// Address Response
	ProtocolAddressResponseChinese = 0x17 // Chinese address packet from server
	ProtocolAddressResponseEnglish = 0x97 // English address packet from server

	// 4G Base Station Data
	ProtocolGPSLocation4G     = 0xA0 // GPS location packet (UTC, 4G base station data)
	ProtocolLBSMultiBase4G    = 0xA1 // LBS multi-base extended information packet (4G)
	ProtocolAlarmMultiFence4G = 0xA4 // Multi-fence alarm packet (4G)
)

// Start bit markers for packet framing
const (
	// StartBitShort indicates a packet with 1-byte length field (max 255 bytes content)
	StartBitShort = 0x7878

	// StartBitLong indicates a packet with 2-byte length field (for larger packets)
	StartBitLong = 0x7979
)

// Stop bit marker (always 2 bytes)
const (
	// StopBit marks the end of every packet
	StopBit = 0x0D0A
)

// Packet structure field sizes (in bytes)
const (
	StartBitSize  = 2 // Start bit size
	StopBitSize   = 2 // Stop bit size
	CRCSize       = 2 // CRC checksum size
	SerialNumSize = 2 // Information serial number size

	// Length field sizes
	LengthFieldSizeShort = 1 // For packets starting with 0x7878
	LengthFieldSizeLong  = 2 // For packets starting with 0x7979

	// Protocol number field size
	ProtocolNumSize = 1
)

// Minimum packet sizes
const (
	// MinPacketSize is the absolute minimum size for a valid packet
	// Start(2) + Length(1) + Protocol(1) + Serial(2) + CRC(2) + Stop(2) = 10 bytes
	MinPacketSize = 10

	// MinPacketSizeLong is minimum for packets with 2-byte length field
	// Start(2) + Length(2) + Protocol(1) + Serial(2) + CRC(2) + Stop(2) = 11 bytes
	MinPacketSizeLong = 11
)

// Maximum packet sizes
const (
	// MaxPacketSizeShort is max size for short packets (1-byte length field)
	// Max length byte value is 255, but this includes Protocol + Content + Serial + CRC
	MaxPacketSizeShort = 255 + StartBitSize + LengthFieldSizeShort + StopBitSize

	// MaxPacketSizeLong is max size for long packets (2-byte length field)
	// Max length value is 65535
	MaxPacketSizeLong = 65535 + StartBitSize + LengthFieldSizeLong + StopBitSize
)

// Language codes
const (
	LanguageChinese = 0x01 // Chinese language
	LanguageEnglish = 0x02 // English language
)

// TimeZone constants
const (
	// TimeZoneMultiplier is used to calculate timezone value
	// Timezone value = (hours * 100 + minutes) * TimeZoneMultiplier
	// Example: GMT+8:00 = 8 * 100 = 800 = 0x0320
	TimeZoneMultiplier = 100
)

// ACC (Ignition) status
const (
	ACCOff = 0x00 // ACC off (engine off)
	ACCOn  = 0x01 // ACC on (engine on)
)

// GPS positioning status
const (
	GPSNotPositioned = 0x00 // GPS not positioned
	GPSPositioned    = 0x01 // GPS positioned (has valid fix)
)

// Data upload modes (for location packets)
const (
	UploadModeInterval         = 0x00 // Upload at fixed interval
	UploadModeDistance         = 0x01 // Upload at fixed distance
	UploadModeTurningPoint     = 0x02 // Upload at turning point
	UploadModeACCChange        = 0x03 // Upload when ACC status changes
	UploadModeStopToStill      = 0x04 // Upload last fix after moving to still
	UploadModeNetworkReconnect = 0x05 // Upload last fix before network interruption and reconnection
	UploadModeRefreshEphem     = 0x06 // Force upload when refreshing ephemeris
	UploadModeKeyPress         = 0x07 // Upload when key pressed
	UploadModePowerOn          = 0x08 // Upload location on power on
	UploadModeAfterStill       = 0x0A // Upload last position after becoming still
	UploadModeWiFi             = 0x0B // Parse WiFi uploaded packet
	UploadModeImmediate        = 0x0C // Upload by LJDW command (immediate position)
	UploadModeLastStill        = 0x0D // Upload last position after still
	UploadModeGPSDUP           = 0x0E // Upload GPSDUP (fixed interval in still state)
	UploadModeExitTracking     = 0x0F // Exit tracking mode
)

// Information Transfer sub-protocols (Type field in 0x94 packets)
const (
	InfoTypeExternalVoltage = 0x00 // External battery voltage
	InfoTypeTerminalSync    = 0x04 // Terminal state synchronization
	InfoTypeDoorStatus      = 0x05 // Door status
	InfoTypeSelfCheck       = 0x08 // Self-check parameters
	InfoTypeGPSStatus       = 0x09 // GPS satellite status info
	InfoTypeICCID           = 0x0A // ICCID information
	InfoTypeRFID            = 0x1B // RFID information
)

// Fence numbers for multi-fence alarms
const (
	FenceInvalid = 0xFF // Invalid fence number
)

// Protocol version info
const (
	ProtocolVersion = "1.1.2"
	ProtocolName    = "JM-VL03"
)
