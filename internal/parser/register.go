package parser

// This file is imported by the decoder to trigger parser registration.
// All parsers register themselves in their init() functions.
//
// Registered parsers:
// - Login (0x01)
// - Heartbeat (0x13)
// - Command Response Old (0x15)
// - Chinese Address Response (0x17)
// - Command Response (0x21)
// - GPS Location (0x22)
// - Alarm (0x26)
// - Alarm Multi-Fence (0x27)
// - LBS Multi-Base (0x28)
// - GPS Address Request (0x2A)
// - Online Command (0x80)
// - Time Calibration (0x8A)
// - Information Transfer (0x94)
// - English Address Response (0x97)
// - GPS Location 4G (0xA0)
// - LBS Multi-Base 4G (0xA1)
// - Alarm 4G (0xA4)

// Ensure all parser files are compiled by referencing something from each.
// This is a compile-time check that all parsers are properly linked.
var _ = NewLoginParser
var _ = NewHeartbeatParser
var _ = NewCommandResponseOldParser
var _ = NewChineseAddressParser
var _ = NewCommandResponseParser
var _ = NewLocationParser
var _ = NewAlarmParser
var _ = NewAlarmMultiFenceParser
var _ = NewLBSParser
var _ = NewGPSAddressRequestParser
var _ = NewOnlineCommandParser
var _ = NewTimeCalibrationParser
var _ = NewInfoTransferParser
var _ = NewEnglishAddressParser
var _ = NewLocation4GParser
var _ = NewLBS4GParser
var _ = NewAlarm4GParser
