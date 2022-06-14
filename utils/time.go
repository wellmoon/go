package utils

import (
	"strconv"
	"sync/atomic"
	"time"
)

var randomInt32 int32

func GenerateCmdIdx() string {
	nano := strconv.Itoa(time.Now().Nanosecond())
	s := atomic.AddInt32(&randomInt32, 1)
	return nano + strconv.Itoa(int(s))
}

func GetLongTime() int64 {
	return time.Now().UnixNano() / 1000000000
}

func GetTimeMillis() int64 {
	return time.Now().UnixNano() / 1000000
}

func GetTimeStr() string {
	return time.Now().Format("20060102150405")
}

func GetCurTimeStr(formatter string) string {
	return time.Now().Format(formatter)
}

// formatter : "2006-01-02 15:04:05"
func GetTimeFromStr(timeStr string, formatter string) *time.Time {
	loc, _ := time.LoadLocation("Local")

	theTime, err := time.ParseInLocation(formatter, timeStr, loc)
	if err == nil {
		return &theTime
	}
	return nil
}

func Sleepms(ms int) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

func Sleeps(s int) {
	time.Sleep(time.Duration(s) * time.Second)
}
