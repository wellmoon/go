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

func Sleepms(ms int) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

func Sleeps(s int) {
	time.Sleep(time.Duration(s) * time.Second)
}
