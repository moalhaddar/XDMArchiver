package utils

import (
	"dmarchiver/logger"
	"fmt"
	"os"
	"strconv"
	"time"
)

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}

func UnixTimestampStringToTime(unixTimestampStr string, isMs bool) (*time.Time, error) {
	unixTimestamp, err := strconv.ParseInt(unixTimestampStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing timestamp: %w", err)
	}

	var t time.Time
	if isMs {
		t = time.UnixMilli(unixTimestamp)
	} else {
		t = time.Unix(unixTimestamp, 0)
	}

	return &t, nil
}

func FormatUnixTimestamp(unixTimestampStr string, isMs bool) (string, error) {
	t, err := UnixTimestampStringToTime(unixTimestampStr, isMs)
	if err != nil {
		return "", err
	}
	formattedTime := t.Format(time.DateTime)
	return formattedTime, nil
}

func SleepUntil(wakeupTime time.Time) {
	now := time.Now()
	sleepDuration := wakeupTime.Sub(now)
	logger.MediaLogger.Printf("Sleeping until %s\n", wakeupTime.Local().Format(time.DateTime))
	time.Sleep(sleepDuration)
}
