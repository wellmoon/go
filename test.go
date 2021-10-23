package main

import (
	"fmt"
	"time"

	"github.com/wellmoon/go/pool"
)

func main() {
	b := []byte{0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00}
	fmt.Println(string(b))
}

func testPool() {
	zpool := pool.New()

	for i := 0; i < 100; i++ {
		zpool.Run(testFunc, "test", i)
	}
	for i := 0; i < 30; i++ {
		fmt.Println("curNum:", zpool.GetCurNum(), "waitNum:", zpool.GetWaitNum())
		time.Sleep(time.Duration(3) * time.Second)
	}
}

func testFunc(name string, num int) {
	fmt.Println("name:", name, ", num:", num)
	time.Sleep(time.Duration(10) * time.Second)
}
