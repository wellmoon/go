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
