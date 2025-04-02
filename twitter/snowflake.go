package twitter

import (
	"XDMArchiver/logger"
	"strconv"
	"time"
)

type Snowflake struct {
	Timestamp             time.Time
	MachineID             uint16
	MachineSequenceNumber uint16
}

const (
	TwitterEpochMs     int64  = 1288834974657
	sequenceNumberMask uint64 = 0xFFF // 12 bits (0-11)
	machineIDMask      uint64 = 0x3FF // 10 bits (12-21)
	timestampBitShift  uint64 = 22    // Timestamp starts at bit 22
)

func DecodeSnowflake(snowflakeStr string) Snowflake {
	id, err := strconv.ParseUint(snowflakeStr, 10, 64)
	if err != nil {
		logger.MediaLogger.Fatal("Failed to parse snowflake id")
	}

	msTimestamp := (id >> timestampBitShift) + uint64(TwitterEpochMs)
	machineID := uint16((id >> 12) & machineIDMask)
	sequenceNumber := uint16(id & sequenceNumberMask)

	return Snowflake{
		Timestamp:             time.UnixMilli(int64(msTimestamp)),
		MachineID:             machineID,
		MachineSequenceNumber: sequenceNumber,
	}
}

func EncodeFakeSnowflakeFromTimestamp(timestamp time.Time) string {
	timestampMs := timestamp.UnixMilli() - TwitterEpochMs
	var id uint64

	id = id | uint64(0&0xFFF)
	id = id | (uint64(0&0x3FF) << 12)
	id = id | (uint64(timestampMs) << 22)

	return strconv.FormatUint(id, 10)
}
