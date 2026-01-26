# GPS Tracker Communication Protocol - JM-VL03

**Manufacturer:** Shenzhen Concox Information Technology Co., Ltd.  
**Version:** V1.1.2  
**Security Level:** Confidential

---

## 📋 Table of Contents

1. [Change History](#change-history)
2. [General Packet Format](#general-packet-format)
3. [Protocol Numbers](#protocol-numbers)
4. [Packet Details](#packet-details)
   - [Login Packet (0x01)](#1-login-packet-0x01)
   - [Heartbeat Packet (0x13)](#2-heartbeat-packet-0x13)
   - [GPS Location Packet (0x22)](#3-gps-location-packet-0x22)
   - [LBS Multi-Base Extended (0x28)](#4-lbs-multi-base-extended-information-packet-0x28)
   - [Alarm Packet (0x26/0x27)](#5-alarm-packet-0x260x27)
   - [GPS Address Request (0x2A)](#6-gps-address-request-packet-0x2a)
   - [Online Command (0x80)](#7-online-command-0x80)
   - [Time Calibration (0x8A)](#8-time-calibration-packet-0x8a)
   - [Information Transfer (0x94)](#9-information-transfer-packet-0x94)
   - [GPS Location 4G (0xA0)](#10-gps-location-packet-4g---0xa0)
   - [LBS Extended 4G (0xA1)](#11-lbs-multi-base-extended-4g---0xa1)
   - [Multi-fence Alarm 4G (0xA4)](#12-multi-fence-alarm-packet-4g---0xa4)
5. [CRC-ITU Algorithm](#crc-itu-algorithm)
6. [Service Flow Diagram](#service-flow-diagram)
7. [Reference Tables](#reference-tables)

---

## Change History

| Author | Date | Version | Description |
|--------|------|---------|-------------|
| Bian Yutao | Dec. 9, 2015 | 1.0.0 | Initial version |
| Bian Yutao | Mar. 3, 2017 | 1.1.0 | Added transparent protocol for plug-in module |
| Bian Yutao | Apr. 14, 2017 | 1.1.1 | Added description for online command responses |
| Bian Yutao | Oct. 14, 2017 | 1.1.2 | Synchronized audio recording protocol |

---

## General Packet Format

### Packet Structure

| Field | Length (Bytes) | Description |
|-------|----------------|-------------|
| **Start Bit** | 2 | `0x78 0x78` (1 byte length) or `0x79 0x79` (2 bytes length) |
| **Packet Length** | 1 (or 2) | Length = Protocol Number + Information Content + Information SN + CRC |
| **Protocol Number** | 1 | Indicates the type of transfer packet |
| **Information Content** | N | Determined by different applications and their protocol numbers |
| **Information SN** | 2 | Automatically increments by "1" for each data transmission after power on |
| **CRC** | 2 | CRC-ITU value from "Packet Length" to "Information SN" |
| **Stop Bit** | 2 | Fixed at `0x0D 0x0A` |

### Important Notes
- If the receiver receives a packet with a CRC error, it ignores the error and discards the packet
- The Start Bit determines if the Packet Length is 1 or 2 bytes:
  - `0x78 0x78` → Packet Length is 1 byte
  - `0x79 0x79` → Packet Length is 2 bytes

---

## Protocol Numbers

| Packet Type | Protocol Number |
|-------------|-----------------|
| Login Packet | `0x01` |
| GPS Location Packet (UTC) | `0x22` |
| Heartbeat Packet | `0x13` |
| Response to Online Command by Terminal | `0x21` |
| Alarm Data (UTC) | `0x26` |
| Alarm Data Multiple Geofences | `0x27` |
| LBS Multi-base Extended Information Packet | `0x28` |
| GPS Address Request Packet (UTC) | `0x2A` |
| Online Command | `0x80` |
| Time Calibration Packet | `0x8A` |
| Information Transfer Packet | `0x94` |
| Chinese Address Packet | `0x17` |
| English Address Packet | `0x97` |
| GPS Location Packet (UTC, 4G Base Station Data) | `0xA0` |
| LBS Multi-base Extended Information Packet (4G) | `0xA1` |
| Multi-fence Alarm Packet (4G) | `0xA4` |

---

## Packet Details

### 1. Login Packet (0x01)

#### Description
- Used to establish connection between terminal and platform
- Contains terminal information
- When GPRS link is established, the terminal sends a login packet to the server
- If a response is received within 5 seconds, the link is established
- If no response is received in 5 seconds, it is considered a timeout
- If the timeout counter reaches 3, the terminal enables scheduled reboot

#### Login Packet Structure (Terminal → Server)

| Field | Length | Description |
|-------|--------|-------------|
| Start Bit | 2 | `0x78 0x78` |
| Packet Length | 1 | Length = Protocol Number + Information Content + Information SN + CRC |
| Protocol Number | 1 | `0x01` |
| **Terminal ID** | 8 | Device IMEI. Ex: IMEI `123456789123456` → `0x01 0x23 0x45 0x67 0x89 0x12 0x34 0x56` |
| **Type Identity Code** | 2 | Used to identify the terminal type |
| **Time Zone/Language** | 2 | See detail table below |
| Information SN | 2 | Sequence number |
| CRC | 2 | CRC-ITU |
| Stop Bit | 2 | `0x0D 0x0A` |

**Example:** `78781101075253367890024270003201000512790D0A`

#### Time Zone/Language Details

| Bits | Description |
|------|-------------|
| Bit15-Bit4 | Value calculated by expanding time zone by 100 |
| Bit3 | 0 = East Time, 1 = West Time |
| Bit2 | Undefined |
| Bit1 | Language select bit 1 |
| Bit0 | Language select bit 0 |

**Examples:**
- `0x32 0x00` = GMT+8:00 → `8*100=800` → `0x0320`
- `0x4D 0xD8` = GMT-12:45 → `12.45*100=1245` → `0x04DD`

#### Server Response (Server → Terminal)

| Field | Length | Description |
|-------|--------|-------------|
| Start Bit | 2 | `0x78 0x78` |
| Packet Length | 1 | Length |
| Protocol Number | 1 | `0x01` |
| Information SN | 2 | Sequence number |
| CRC | 2 | CRC-ITU |
| Stop Bit | 2 | `0x0D 0x0A` |

**Example:** `7878050100059FF80D0A`

---

### 2. Heartbeat Packet (0x13)

#### Description
- Used to maintain GPRS link connectivity
- The terminal sends heartbeat to the server
- If a response is received within 5 seconds, the link is active
- If timeout reaches 3, scheduled reboot is enabled

#### Heartbeat Structure (Terminal → Server)

| Field | Length | Description |
|-------|--------|-------------|
| Start Bit | 2 | `0x78 0x78` |
| Packet Length | 1 | Length |
| Protocol Number | 1 | `0x13` |
| **Terminal Information Content** | 1 | See bit table below |
| **Voltage Level** | 1 | Battery level (0x00-0x06) |
| **GSM Signal Strength** | 1 | Signal strength (0x00-0x04) |
| **Language/Extended Port Status** | 2 | 0x01=Chinese, 0x02=English |
| Information SN | 2 | Sequence number |
| CRC | 2 | CRC-ITU |
| Stop Bit | 2 | `0x0D 0x0A` |

**Example:** `78780A134004040001000FDCEE0D0A`

#### Voltage Level

| Value | Description |
|-------|-------------|
| `0x00` | No power (shutdown) |
| `0x01` | Extremely low battery (cannot make calls/SMS) |
| `0x02` | Very low battery (low battery alert will be triggered) |
| `0x03` | Low battery (normal use) |
| `0x04` | Medium battery |
| `0x05` | High battery |
| `0x06` | Extremely high battery |

#### GSM Signal Strength

| Value | Description |
|-------|-------------|
| `0x00` | No signal |
| `0x01` | Extremely weak signal |
| `0x02` | Weak signal |
| `0x03` | Good signal |
| `0x04` | Strong signal |

#### Terminal Information Content (Bits)

| Bit | Description |
|-----|-------------|
| Bit7 | 1: Cut fuel/power, 0: Restore fuel/power |
| Bit6 | 1: GPS positioned, 0: Not positioned |
| Bit3-Bit5 | Extended bit |
| Bit2 | 1: Charge with power connected, 0: Charge without power connected |
| Bit1 | 1: ACC on, 0: ACC off |
| Bit0 | 1: Defense on, 0: Defense off |

#### Server Response

**Example:** `78 78 05 13 01 00 E1 A0 0D 0A`

---

### 3. GPS Location Packet (0x22)

#### Description
- Contains terminal location data
- After GPS module is positioned and connection established, terminal uploads fix data according to preset rules
- Also uploads cached fixes when connection is established

#### Location Packet Structure (Terminal → Server)

| Field | Length | Description |
|-------|--------|-------------|
| Start Bit | 2 | `0x78 0x78` |
| Packet Length | 1 | Length |
| Protocol Number | 1 | `0x22` (UTC) |
| **Date and Time** | 6 | Year, Month, Day, Hour, Minute, Second (1 byte each, convert to decimal) |
| **Number of Satellites** | 1 | First char: GPS Info Length; Second: Number of satellites |
| **Latitude** | 4 | Decimal value divided by 1,800,000 |
| **Longitude** | 4 | Decimal value divided by 1,800,000 |
| **Speed** | 1 | Value in decimal (km/h) |
| **Course and Status** | 2 | 16-bit binary (see details below) |
| **MCC** | 2 | Mobile Country Code (decimal) |
| **MNC** | 1 | Mobile Network Code (decimal) |
| **LAC** | 2 | Location Area Code (decimal) |
| **CellID** | 3 | Cell Tower ID (decimal) |
| **ACC** | 1 | 0x00=ACC off, 0x01=ACC on |
| **Data Upload Mode** | 1 | GPS data upload mode |
| **GPS Data Re-upload** | 1 | 0x00=Real-time, 0x01=Re-upload |
| **Mileage Statistics** | 4 | Convert to decimal (if product has this feature) |
| Information SN | 2 | Sequence number |
| CRC | 2 | CRC-ITU |
| Stop Bit | 2 | `0x0D 0x0A` |

**Example:** `787822220F0C1D023305C9027AC8180C46586000140001CC00287D001F71000001000820860D0A`

#### Data Upload Mode

| Value | Description |
|-------|-------------|
| `0x00` | Upload by fixed interval |
| `0x01` | Upload by fixed distance |
| `0x02` | Upload by cornering |
| `0x03` | Upload by ACC state change |
| `0x04` | Upload last fix after switching from moving to static |
| `0x05` | Upload last valid fix before network interruption and reconnection |
| `0x06` | Force upload GPS fix when refreshing ephemeris |
| `0x07` | Upload fix on key press |
| `0x08` | Upload location info on power on |
| `0x09` | Not used |
| `0x0A` | Upload last long/lat and update time after becoming static |
| `0x0B` | Parse long/lat packet uploaded via WiFi |
| `0x0C` | Upload by LJDW command (immediate position) |
| `0x0D` | Upload last long/lat after becoming static |
| `0x0E` | Upload GPSDUP (fixed interval in static state) |
| `0x0F` | Exit tracking mode |

#### Course and Status (2 Bytes Details)

Occupies 2 bytes to indicate the terminal's movement direction. Range: 0-360°. North = 0° and counts clockwise.

**BYTE_1:**

| Bit | Description |
|-----|-------------|
| Bit7 | 0 |
| Bit6 | 0 |
| Bit5 | GPS Real-time/Differential Positioning |
| Bit4 | Positioned or Not (1=positioned, 0=not positioned) |
| Bit3 | East/West longitude (0=East, 1=West) |
| Bit2 | South/North latitude (0=South, 1=North) |
| Bit1-Bit0 | Course (high bits) |

**BYTE_2:**

| Bit | Description |
|-----|-------------|
| Bit7-Bit0 | Course (low bits) |

**Decoding Example:**
- Value: `0x15 0x4C` → Binary: `00010101 01001100`
- BYTE_1 Bit5 = 0 → GPS real-time
- BYTE_1 Bit4 = 1 → GPS positioned
- BYTE_1 Bit3 = 0 → East Longitude
- BYTE_1 Bit2 = 1 → North Latitude
- Course = `0101001100` = 332° (binary to decimal)

#### Server Response
No server response packet required.

---

### 4. LBS Multi-Base Extended Information Packet (0x28)

#### Description
- Used to transmit location info when terminal has no GPS fix

#### Packet Structure (Terminal → Server)

| Field | Length | Description |
|-------|--------|-------------|
| Start Bit | 2 | `0x78 0x78` |
| Packet Length | 1 | Length |
| Protocol Number | 1 | `0x28` |
| **UTC** | 6 | Date and time |
| **MCC** | 2 | Mobile Country Code |
| **MNC** | 1 | Mobile Network Code |
| **LAC** | 2 | Location Area Code |
| **CI** | 3 | Cell Tower ID |
| **RSSI** | 1 | Signal strength (0x00-0xFF) |
| **NLAC1** | 2 | Neighbor LAC 1 |
| **NCI1** | 3 | Neighbor Cell ID 1 |
| **NRSSI1** | 1 | Neighbor RSSI 1 |
| **NLAC2-6, NCI2-6, NRSSI2-6** | - | Same for neighbors 2-6 |
| **Timing Advance** | 1 | Signal time difference |
| **Language** | 2 | 0x01=Chinese, 0x02=English |
| Information SN | 2 | Sequence number |
| CRC | 2 | CRC-ITU |
| Stop Bit | 2 | `0x0D 0x0A` |

**Example:** `78783B2810010D02020201CC00287D001F713E287D001F7231287D001E232D287D001F4018000000000000000000000000000000000000FF00020005B14B0D0A`

#### RSSI (Signal Strength)
- `0x00` = Weakest signal
- `0xFF` = Strongest signal

#### Server Response
No response packet required.

---

### 5. Alarm Packet (0x26/0x27)

#### Description
- Used to transmit terminal-defined alarm content
- Server responds to alarm content and sends parsed long/lat address to terminal
- Terminal sends received address to preset SOS number

#### Alarm Packet Structure - Single Geofence (0x26)

| Field | Length | Description |
|-------|--------|-------------|
| Start Bit | 2 | `0x78 0x78` |
| Packet Length | 1 | Length |
| Protocol Number | 1 | `0x26` (UTC) |
| **Date and Time** | 6 | Date and time |
| **Number of Satellites** | 1 | GPS Info |
| **Latitude** | 4 | Divide by 1,800,000 |
| **Longitude** | 4 | Divide by 1,800,000 |
| **Speed** | 1 | Speed in decimal |
| **Course and Status** | 2 | Direction and status |
| **LBS Length** | 1 | Total length of LBS info |
| **MCC** | 2 | Mobile Country Code |
| **MNC** | 1 | Mobile Network Code |
| **LAC** | 2 | Location Area Code |
| **CellID** | 3 | Cell Tower ID |
| **Terminal Information** | 1 | See bit table |
| **Voltage Level** | 1 | Battery level |
| **GSM Signal Strength** | 1 | Signal strength |
| **Alert and Language** | 2 | Alert type and language |
| **Mileage Statistics** | 4 | Mileage statistics |
| Information SN | 2 | Sequence number |
| CRC | 2 | CRC-ITU |
| Stop Bit | 2 | `0x0D 0x0A` |

**Example:** `787825260F0C1D030B26C9027AC8180C4658600004000901CC00287D001F718004041302000C472A0D0A`

#### Alarm Packet Structure - Multiple Geofences (0x27)

Similar to 0x26 but includes:
- **Fence No.** (1 byte): Valid for geofence alerts. 0=Fence 1, 1=Fence 2, ..., FF=Invalid

**Example:** `78 78 26 27 10 04 19 09 2D 07 C5 02 7A C9 1C 0C 46 58 00 00 05 37 09 00 00 00 00 00 00 00 00 80 02 00 0C 01 FF 00 00 4D F6 0D 0A`

#### Terminal Information (Bit Details)

| Bit | Description |
|-----|-------------|
| Bit7 | 1: Cut fuel/power, 0: Restore |
| Bit6 | 1: GPS positioned, 0: Not positioned |
| Bit3-Bit5 | 100: SOS, 011: Low battery alert, 010: Power cut, 001: Vibration alert, 000: Normal |
| Bit2 | 1: Charge with power, 0: Charge without power |
| Bit1 | 1: ACC on, 0: ACC off |
| Bit0 | 1: Defense on, 0: Defense off |

#### Alert and Language (Byte 1 - Alert Types)

| Value | Description |
|-------|-------------|
| `0x00` | Normal |
| `0x01` | SOS alert |
| `0x02` | Power cut alert |
| `0x03` | Vibrating alert |
| `0x04` | Entered fence alert |
| `0x05` | Left fence alert |
| `0x06` | Speed alert |
| `0x09` | Tow/theft alert |
| `0x0A` | Entered GPS blind spot alert |
| `0x0B` | Left GPS blind spot alert |
| `0x0C` | Powered on alert |
| `0x0D` | GPS first fix alert |
| `0x0E` | Low external battery alert |
| `0x0F` | External battery low voltage protection alert |
| `0x10` | SIM changed alert |
| `0x11` | Powered off alert |
| `0x12` | Airplane mode on following external battery low voltage protection |
| `0x13` | Tamper alert |
| `0x14` | Door alert |
| `0x15` | Powered off due to low battery |
| `0x16` | Sound-control alert |
| `0x17` | Rogue base station detected alert |
| `0x18` | Cover removed alert |
| `0x19` | Low internal battery alert |
| `0x20` | Entered deep sleep mode alert |
| `0x21` | Reserved |
| `0x22` | Reserved |
| `0x23` | Fall alert |
| `0x29` | Harsh acceleration |
| `0x2A` | Sharp left cornering alert |
| `0x2B` | Sharp right cornering alert |
| `0x2C` | Collision alert |
| `0x30` | Harsh braking |
| `0x32` | Device unplugged alert |
| `0xFF` | ACC OFF |
| `0xFE` | ACC ON |

#### Alert and Language (Byte 2 - Language)

| Value | Description |
|-------|-------------|
| `0x01` | Chinese |
| `0x02` | English |
| `0x00` | No platform response required |

#### Server Response (0x26)

**Example:** `78780526001C9D860D0A`

#### Server Returns Chinese Address (0x17)

| Field | Length | Description |
|-------|--------|-------------|
| Start Bit | 2 | `0x78 0x78` |
| Packet Length | 1 | Length |
| Protocol Number | 1 | `0x17` |
| **Length** | 1 | Data length between server flag and SN |
| **Server Flag Bit** | 4 | Used by server to mark alert |
| **ALARMSMS** | 8 | Alarm code flag (ASCII) |
| **&&** | 2 | Separator (ASCII) |
| **Address Content** | M | Parsed address (UNICODE) |
| **&&** | 2 | Separator (ASCII) |
| **Phone Number** | 21 | "0" for uploaded alarm packets (ASCII) |
| **##** | 2 | Separator (ASCII) |
| Information SN | 2 | Sequence number |
| CRC | 2 | CRC-ITU |
| Stop Bit | 2 | `0x0D 0x0A` |

#### Server Returns English Address (0x97)

Similar to 0x17 but:
- Start Bit: `0x79 0x79`
- Packet Length: 2 bytes
- Protocol Number: `0x97`

---

### 6. GPS Address Request Packet (0x2A)

#### Description
- User sends address request command to terminal
- Terminal sends address request packet to server to parse address
- Terminal sends parsed address returned by server to user

#### Request Packet Structure (Terminal → Server)

| Field | Length | Description |
|-------|--------|-------------|
| Start Bit | 2 | `0x78 0x78` |
| Packet Length | 1 | Length |
| Protocol Number | 1 | `0x2A` |
| **Date and Time** | 6 | Date and time |
| **Number of Satellites** | 1 | GPS Info |
| **Latitude** | 4 | Divide by 1,800,000 |
| **Longitude** | 4 | Divide by 1,800,000 |
| **Speed** | 1 | Speed |
| **Course and Status** | 2 | Direction |
| **Phone Number** | 21 | Phone number |
| **Alert and Language** | 2 | 0x01=Chinese, 0x02=English |
| Information SN | 2 | Sequence number |
| CRC | 2 | CRC-ITU |
| Stop Bit | 2 | `0x0D 0x0A` |

**Example:** `78782E2A0F0C1D071139CA027AC8000C4658000014D8313235323031333533323137373037390000000000000100 2A6ECE0D0A`

---

### 7. Online Command (0x80)

#### Description
- Assigned by server to control terminal and execute tasks
- Terminal responds to server with execution results

#### Server Sends Online Command (Server → Terminal)

| Field | Length | Description |
|-------|--------|-------------|
| Start Bit | 2 | `0x78 0x78` |
| Packet Length | 1 | Length |
| Protocol Number | 1 | `0x80` |
| **Length** | 1 | Server flag bit + command content length |
| **Server Flag Bit** | 4 | Reserved for server acknowledgement |
| **Command Content** | M | Character string in ASCII (compatible with SMS command) |
| **Language** | 2 | 0x01=Chinese, 0x02=English |
| Information SN | 2 | Sequence number |
| CRC | 2 | CRC-ITU |
| Stop Bit | 2 | `0x0D 0x0A` |

**Example:** `78780E800800000000736F732300016D6A0D0A`

#### Terminal Response (Terminal → Server) - Universal Version (0x21)

| Field | Length | Description |
|-------|--------|-------------|
| Start Bit | 2 | `0x79 0x79` |
| Packet Length | 2 | Length |
| Protocol Number | 1 | `0x21` |
| **Server Flag Bit** | 4 | Received data returned as is |
| **Code** | 1 | 0x01=ASCII, 0x02=UTF 16-BE |
| **Content** | M | Data to send (by encoding format) |
| Information SN | 2 | Sequence number |
| CRC | 2 | CRC-ITU |
| Stop Bit | 2 | `0x0D 0x0A` |

#### Terminal Response - Old Version (0x15)

| Field | Length | Description |
|-------|--------|-------------|
| Start Bit | 2 | `0x78 0x78` |
| Packet Length | 1 | Length |
| Protocol Number | 1 | `0x15` |
| **Length** | 1 | Server flag bit + command content length |
| **Server Flag Bit** | 4 | Received data returned |
| **Command Content** | M | String returned in ASCII |
| **Language** | 2 | Chinese: 0x00 0x01, English: 0x00 0x02 |
| Information SN | 2 | Sequence number |
| CRC | 2 | CRC-ITU |
| Stop Bit | 2 | `0x0D 0x0A` |

---

### 8. Time Calibration Packet (0x8A)

#### Description
- Sent by terminal to server on power-up to request time synchronization
- Solves time error problem when terminal is not positioned
- Server responds with correct UTC in correct format

#### Calibration Request (Terminal → Server)

| Field | Length | Description |
|-------|--------|-------------|
| Start Bit | 2 | `0x78 0x78` |
| Packet Length | 1 | Length |
| Protocol Number | 1 | `0x8A` |
| Information SN | 2 | Sequence number |
| CRC | 2 | CRC-ITU |
| Stop Bit | 2 | `0x0D 0x0A` |

**Example:** `7878058A000688290D0A`

#### Server Response (Server → Terminal)

| Field | Length | Description |
|-------|--------|-------------|
| Start Bit | 2 | `0x78 0x78` |
| Packet Length | 1 | Length |
| Protocol Number | 1 | `0x8A` (UTC) |
| **Date and Time** | 6 | Year, Month, Day, Hour, Minute, Second |
| Information SN | 2 | Sequence number |
| CRC | 2 | CRC-ITU |
| Stop Bit | 2 | `0x0D 0x0A` |

**Example:** `78780B8A0F0C1D0000150006F0860D0A`

---

### 9. Information Transfer Packet (0x94)

#### Description
- Used to transmit all types of non-location data

#### Packet Structure (Terminal → Server)

| Field | Length | Description |
|-------|--------|-------------|
| Start Bit | 2 | `0x79 0x79` |
| Packet Length | 2 | Length |
| Protocol Number | 1 | `0x94` |
| **Information Type** | 1 | Sub-protocol (see table) |
| **Data Content** | N | Content according to info type |
| Information SN | 2 | Sequence number |
| CRC | 2 | CRC-ITU |
| Stop Bit | 2 | `0x0D 0x0A` |

#### Information Types (Sub-protocol)

| Value | Description |
|-------|-------------|
| `0x00` | External battery voltage |
| `0x01-0x03` | Custom |
| `0x04` | Terminal status synchronization |
| `0x05` | Door status |
| `0x08` | Self-check parameters |
| `0x09` | Visible satellite info |
| `0x0A` | ICCID Information |
| `0x1B` | RFID |

#### Type 0x00 - External Battery Voltage
- 2 bytes hex converted to decimal and divided by 100
- Example: `0x04 0x9F` = 1183 decimal = 11.83V

#### Type 0x04 - Terminal Status Synchronization

**Content IDs:**

| Definition | ID |
|------------|-----|
| Alarm byte 1 | ALM1 |
| Alarm byte 2 | ALM2 |
| Alarm byte 3 | ALM3 |
| Alarm byte 4 | ALM4 |
| Status byte 1 | STA1 |
| SOS number | SOS |
| Center number | CENTER |
| Geofence | FENCE |
| Fuel/power cutoff status | DYD |
| Mode | MODE |

**ALM1 (Status):**

| Bit | Definition |
|-----|------------|
| bit7 | Vibrating alert (1=ON, 0=OFF) |
| bit6 | Alert via GPRS |
| bit5 | Alert via call |
| bit4 | Alert via SMS |
| bit3 | Tow/theft alert |
| bit2 | Alert via GPRS |
| bit1 | Alert via call |
| bit0 | Alert via SMS |

**ALM2 (Status):**

| Bit | Definition |
|-----|------------|
| bit7 | Low internal battery alert |
| bit6 | Alert via GPRS |
| bit5 | Alert via call |
| bit4 | Alert via SMS |
| bit3 | Low external battery alert |
| bit2 | Alert via GPRS |
| bit1 | Alert via call |
| bit0 | Alert via SMS |

**ALM3 (Status):**

| Bit | Definition |
|-----|------------|
| bit7 | Speed alert |
| bit6 | Alert via GPRS |
| bit5 | Alert via call |
| bit4 | Alert via SMS |
| bit3 | Power cut alert |
| bit2 | Alert via GPRS |
| bit1 | Alert via call |
| bit0 | Alert via SMS |

**ALM4 (Status):**

| Bit | Definition |
|-----|------------|
| bit7 | SOS alert |
| bit6 | Alert via GPRS |
| bit5 | Alert via call |
| bit4 | Alert via SMS |
| bit3 | Voice control alert |
| bit2 | Alert via GPRS |
| bit1 | Alert via call |
| bit0 | Alert via SMS |

**STA1 (Status):**

| Bit | Definition |
|-----|------------|
| bit7 | Defense status (1=ON, 0=OFF) |
| bit6 | Auto defense |
| bit5 | Manual defense |
| bit4 | Remote cancellation of defense |
| bit3 | To be defined |
| bit2 | To be defined |
| bit1 | Tamper switch (1=Close, 0=Open) |
| bit0 | Tamper alert |

**Fuel/Power Cutoff Status:**

| Bit | Definition |
|-----|------------|
| bit7-bit4 | Undefined |
| bit3 | Delay execution because speed is too high |
| bit2 | Delay execution because terminal is not positioned |
| bit1 | Cut fuel/power |
| bit0 | Connect fuel/power |

**Example format:**
```
ALM1=FF;ALM2=FF;ALM3=FF;STA1=CO;DYD=01;SOS=12345,2345,5678;CENTER=987654;
FENCE=FENCE,ON,0,-22.277120,-113.516763,5,IN,1;MODE=MODE,1,20,500
```

#### Type 0x05 - Door Status

| Bit | Definition |
|-----|------------|
| bit7-bit3 | TBD |
| bit2 | I/O port status (1=High, 0=Low) |
| bit1 | Trigger status (1=Level high, 0=Level low) |
| bit0 | Door status (1=ON, 0=OFF) |

#### Type 0x09 - GPS Status Info

| Field | Bytes | Description |
|-------|-------|-------------|
| GPS module status | 1 | 0x00=No feature, 0x01=Searching, 0x02=2D, 0x03=3D, 0x04=Sleeping |
| Number of satellites in fix | 1 | Base for strength transmission amount |
| GPS1-N strength | 1 each | Strength of satellites in fix |
| Number of visible GPS satellites | 1 | Visible satellites not in fix |
| Visible GPS1-N strength | 1 each | Strength of visible satellites |
| BDS module status | 1 | Same as GPS |
| BDS satellites info | - | Similar to GPS |
| Extended bit length | 1 | Reserved for expansion |
| Extended bit | N | Change according to extended bit length |

#### Type 0x0A - ICCID Information

| Field | Bytes | Description |
|-------|-------|-------------|
| IMEI | 8 | Example: IMEI 123456789123456 → 0x01 0x23 0x45 0x67 0x89 0x12 0x34 0x56 |
| IMSI | 8 | Same format |
| ICCID | 10 | Example: ICCID 12345123456789123456 → 0x12 0x34 0x51 0x23 0x45 0x67 0x89 0x12 0x34 0x56 |

#### Type 0x1B - RFID Information

| Field | Bytes | Description |
|-------|-------|-------------|
| RFID | 8 | Example: RFID 2345678912 → 0x23 0x45 0x67 0x89 0x12 |

#### Server Response
No response required.

---

### 10. GPS Location Packet 4G - 0xA0

#### Description
Similar to packet 0x22 but with 4G base station support.

#### Main Differences from 0x22

| Field | Length 2G | Length 4G |
|-------|-----------|-----------|
| MNC | 1 | 1 or 2 |
| LAC | 2 | 4 |
| Cell ID | 3 | 8 |

#### MCC Bits (To determine MNC length)

| Bit | Description |
|-----|-------------|
| Bit15 | 1: MNC is 2 bytes, 0: MNC is 1 byte |
| Bit0-Bit14 | MCC Information |

**Note:** For devices already shipped, Bit15 is "0" by default; for new devices, Bit15 is "1".

---

### 11. LBS Multi-Base Extended 4G - 0xA1

#### Description
Similar to packet 0x28 but with 4G base station support.

#### Main Differences from 0x28

| Field | Length 2G | Length 4G |
|-------|-----------|-----------|
| MNC | 1 | 1 or 2 |
| LAC | 2 | 4 |
| CI | 3 | 8 |
| NLAC1-6 | 2 | 4 |
| NCI1-6 | 3 | 8 |

---

### 12. Multi-fence Alarm Packet 4G - 0xA4

#### Description
Similar to packet 0x27 but with 4G base station support.

#### Main Differences from 0x27

| Field | Length 2G | Length 4G |
|-------|-----------|-----------|
| MNC | 1 | 1 or 2 |
| LAC | 2 | 4 |
| Cell ID | 3 | 8 |

---

## CRC-ITU Algorithm

### C Implementation

```c
static const U16 crctab16[] = {
    0X0000, 0X1189, 0X2312, 0X329B, 0X4624, 0X57AD, 0X6536, 0X74BF,
    0X8C48, 0X9DC1, 0XAF5A, 0XBED3, 0XCA6C, 0XDBE5, 0XE97E, 0XF8F7,
    0X1081, 0X0108, 0X3393, 0X221A, 0X56A5, 0X472C, 0X75B7, 0X643E,
    0X9CC9, 0X8D40, 0XBFDB, 0XAE52, 0XDAED, 0XCB64, 0XF9FF, 0XE876,
    0X2102, 0X308B, 0X0210, 0X1399, 0X6726, 0X76AF, 0X4434, 0X55BD,
    0XAD4A, 0XBCC3, 0X8E58, 0X9FD1, 0XEB6E, 0XFAE7, 0XC87C, 0XD9F5,
    0X3183, 0X200A, 0X1291, 0X0318, 0X77A7, 0X662E, 0X54B5, 0X453C,
    0XBDCB, 0XAC42, 0X9ED9, 0X8F50, 0XFBEF, 0XEA66, 0XD8FD, 0XC974,
    0X4204, 0X538D, 0X6116, 0X709F, 0X0420, 0X15A9, 0X2732, 0X36BB,
    0XCE4C, 0XDFC5, 0XED5E, 0XFCD7, 0X8868, 0X99E1, 0XAB7A, 0XBAF3,
    0X5285, 0X430C, 0X7197, 0X601E, 0X14A1, 0X0528, 0X37B3, 0X263A,
    0XDECD, 0XCF44, 0XFDDF, 0XEC56, 0X98E9, 0X8960, 0XBBFB, 0XAA72,
    0X6306, 0X728F, 0X4014, 0X519D, 0X2522, 0X34AB, 0X0630, 0X17B9,
    0XEF4E, 0XFEC7, 0XCC5C, 0XDDD5, 0XA96A, 0XB8E3, 0X8A78, 0X9BF1,
    0X7387, 0X620E, 0X5095, 0X411C, 0X35A3, 0X242A, 0X16B1, 0X0738,
    0XFFCF, 0XEE46, 0XDCDD, 0XCD54, 0XB9EB, 0XA862, 0X9AF9, 0X8B70,
    0X8408, 0X9581, 0XA71A, 0XB693, 0XC22C, 0XD3A5, 0XE13E, 0XF0B7,
    0X0840, 0X19C9, 0X2B52, 0X3ADB, 0X4E64, 0X5FED, 0X6D76, 0X7CFF,
    0X9489, 0X8500, 0XB79B, 0XA612, 0XD2AD, 0XC324, 0XF1BF, 0XE036,
    0X18C1, 0X0948, 0X3BD3, 0X2A5A, 0X5EE5, 0X4F6C, 0X7DF7, 0X6C7E,
    0XA50A, 0XB483, 0X8618, 0X9791, 0XE32E, 0XF2A7, 0XC03C, 0XD1B5,
    0X2942, 0X38CB, 0X0A50, 0X1BD9, 0X6F66, 0X7EEF, 0X4C74, 0X5DFD,
    0XB58B, 0XA402, 0X9699, 0X8710, 0XF3AF, 0XE226, 0XD0BD, 0XC134,
    0X39C3, 0X284A, 0X1AD1, 0X0B58, 0X7FE7, 0X6E6E, 0X5CF5, 0X4D7C,
    0XC60C, 0XD785, 0XE51E, 0XF497, 0X8028, 0X91A1, 0XA33A, 0XB2B3,
    0X4A44, 0X5BCD, 0X6956, 0X78DF, 0X0C60, 0X1DE9, 0X2F72, 0X3EFB,
    0XD68D, 0XC704, 0XF59F, 0XE416, 0X90A9, 0X8120, 0XB3BB, 0XA232,
    0X5AC5, 0X4B4C, 0X79D7, 0X685E, 0X1CE1, 0X0D68, 0X3FF3, 0X2E7A,
    0XE70E, 0XF687, 0XC41C, 0XD595, 0XA12A, 0XB0A3, 0X8238, 0X93B1,
    0X6B46, 0X7ACF, 0X4854, 0X59DD, 0X2D62, 0X3CEB, 0X0E70, 0X1FF9,
    0XF78F, 0XE606, 0XD49D, 0XC514, 0XB1AB, 0XA022, 0X92B9, 0X8330,
    0X7BC7, 0X6A4E, 0X58D5, 0X495C, 0X3DE3, 0X2C6A, 0X1EF1, 0X0F78,
    0XFFCF, 0XEE46, 0XDCDD, 0XCD54, 0XB9EB, 0XA862, 0X9AF9, 0X8B70,
    0X8408, 0X9581, 0XA71A, 0XB693, 0XC22C, 0XD3A5, 0XE13E, 0XF0B7
};

// Calculate 16-bit CRC of given length data
U16 GetCrc16(const U8* pData, int nLength)
{
    U16 fcs = 0xffff;  // Initialize
    while(nLength > 0) {
        fcs = (fcs >> 8) ^ crctab16[(fcs ^ *pData) & 0xff];
        nLength--;
        pData++;
    }
    return ~fcs;  // Negate
}
```

### CRC Notes
- Calculated from "Packet Length" to "Information SN"
- If receiver gets packet with CRC error, ignore error and discard packet
- Initial value: 0xFFFF
- Final result: negated (~fcs)

---

## Service Flow Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        Terminal boot & reboot                           │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
                    ┌───────────────────────────────┐
                    │   Establish GPRS connection?  │
                    └───────────────────────────────┘
                           │                  │
                       successful            fail
                           │                  │
                           ▼                  ▼
              ┌──────────────────┐    ┌─────────────────┐
              │ Send login packet│    │ Reconnection    │
              │    to server     │    │     time?       │
              └──────────────────┘    └─────────────────┘
                           │               │         │
                           ▼           <20min    >20min
              ┌──────────────────┐         │         │
              │ Reply from server│    reconnect   reboot
              │    correct?      │
              └──────────────────┘
                    │         │
                   Yes        No → reconnect/reboot
                    │
                    ▼
         ┌─────────────────────┐
         │ Connection successful│
         └─────────────────────┘
                    │
        ┌───────────┼───────────┐
        │           │           │
        ▼           ▼           ▼
   ┌─────────┐ ┌─────────┐ ┌─────────┐
   │ Alarm   │ │Location │ │Heartbeat│
   │ status  │ │  data   │ │ packet  │
   │         │ │ packet  │ │         │
   └─────────┘ └─────────┘ └─────────┘
        │           │           │
        ▼           ▼           ▼
   Send alarm  Upload      Send heartbeat
   packet to   regularly   to server
   server                       │
                                ▼
                    ┌───────────────────────┐
                    │ Response from server  │
                    │      normal?          │
                    └───────────────────────┘
                           │         │
                          Yes        No
                           │         │
                    Continue     Fail to receive
                    uploading    within 5min →
                                 reconnect/reboot
```

---

## Reference Tables

### Coordinate Conversion

**Formula:**
```
Real Latitude/Longitude = Hex converted to decimal / 1,800,000
```

**Example:**
- Hex value: `0x02 0x7A 0xC8 0x18`
- Decimal: 41,625,624
- Coordinate: 41,625,624 / 1,800,000 = 23.125346°

### IMEI to Terminal ID Conversion

**Example:**
- IMEI: `123456789123456`
- Terminal ID: `0x01 0x23 0x45 0x67 0x89 0x12 0x34 0x56`

Each pair of IMEI digits is converted to a hexadecimal byte.

### Timeouts and Retries

| Operation | Timeout | Retries | Action after failure |
|-----------|---------|---------|----------------------|
| Login | 5 seconds | 3 | Scheduled reboot |
| Heartbeat | 5 seconds | 3 | Scheduled reboot |
| Reconnection | - | - | <20min: Reconnect, >20min: Reboot |

### 2G vs 4G Comparison

| Feature | 2G (0x22, 0x28, 0x26) | 4G (0xA0, 0xA1, 0xA4) |
|---------|------------------------|------------------------|
| MNC Length | 1 byte | 1 or 2 bytes (based on MCC Bit15) |
| LAC Length | 2 bytes | 4 bytes |
| Cell ID Length | 3 bytes | 8 bytes |

---

## Additional Notes

### About Alerts
- Alerts can accumulate
- When alert byte is `0x00`, alarm content in terminal info can be determined
- Alert bytes in terminal info have priority over alarm byte

### About Language
- `0x01` = Chinese
- `0x02` = English
- `0x00` = No platform response required

### About Start Bit
- `0x78 0x78` = 1-byte Packet Length (max 255 content bytes)
- `0x79 0x79` = 2-byte Packet Length (for larger packets)

### About Unanswered Packets
The following packets do not require a server response:
- GPS Location Packet (0x22)
- LBS Multi-Base Extended (0x28)
- Information Transfer Packet (0x94)
- GPS Location 4G (0xA0)
- LBS Extended 4G (0xA1)

---

## License and Confidentiality

This document contains confidential information of Shenzhen Concox Information Technology Co., Ltd. Unauthorized distribution or use is prohibited.

---

*Documentation generated from official JM-VL03 protocol version V1.1.2*
