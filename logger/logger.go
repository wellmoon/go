package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

var debugLog *log.Logger
var traceLog *log.Logger
var errorLog *log.Logger
var fatalLog *log.Logger
var logPath string

const flag int = log.LstdFlags

func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func getLogPath() string {
	if len(logPath) == 0 {

		curPath, _ := os.Getwd()
		executable, _ := os.Executable()
		dir := filepath.Dir(executable)
		var mainPath string
		if strings.HasPrefix(dir, curPath) {
			mainPath = dir
		} else {
			mainPath = curPath
		}

		sep := string(os.PathSeparator)
		logPath = mainPath + sep + "logs" + sep
		if !pathExists(logPath) {
			os.Mkdir(logPath, os.ModePerm)
		}
	}
	return logPath
}

// 日志文件每小时生成一个，判断是否需要切换日志文件
// 如果需要切换，返回true，并且返回切换后的文件
var switchLock sync.Mutex

func switchFile(logFile string) (bool, *os.File) {
	var res bool
	if !pathExists(getLogPath() + logFile + ".log") {
		file, _ := os.Create(getLogPath() + logFile + ".log")
		return false, file
	} else {
		switchLock.Lock()
		defer switchLock.Unlock()
		oldFile, err := os.OpenFile(getLogPath()+logFile+".log", os.O_RDONLY, os.ModePerm)
		if err != nil {
			return false, nil
		}
		fileStat, err := oldFile.Stat()
		if err != nil {
			return false, nil
		}
		modTime := fileStat.ModTime().Format("2006010215")
		curTime := time.Now().Format("2006010215")
		oldFile.Close()
		if curTime != modTime {
			err := os.Rename(getLogPath()+logFile+".log", getLogPath()+logFile+"."+modTime+".log")
			if err != nil {
				fmt.Println("切换日志文件错误：", err)
			}
			res = true
		} else {

			res = false
		}
	}
	newFile, _ := os.OpenFile(getLogPath()+logFile+".log", os.O_CREATE|os.O_RDWR|os.O_APPEND, os.ModeAppend|os.ModePerm)
	return res, newFile
}

// untime/debug.Stack(0x0, 0xc0000ca398, 0x16)
//         /usr/local/go/src/runtime/debug/stack.go:24 +0x9f
// github.com/wellmoon/go/logger.getStack(...)
//         /Users/wenjie/github.com/wellmoon/go/logger/logger.go:94
// github.com/wellmoon/go/logger.print(0xc00008c140, 0x13a3c38, 0x1a, 0xc00006bfa8, 0x1, 0x1)
//         /Users/wenjie/github.com/wellmoon/go/logger/logger.go:106 +0x26
// github.com/wellmoon/go/logger.Debug(0x13a3c38, 0x1a, 0xc00006bfa8, 0x1, 0x1)
//         /Users/wenjie/github.com/wellmoon/go/logger/logger.go:115 +0x165
// go_code/leridge_server/socket.handleMessage(0xc0000ea070, 0x13252e0, 0xc000012d40)
//         /Users/wenjie/go/src/go_code/leridge_server/socket/server.go:136 +0x3cc
// created by go_code/leridge_server/socket.OnMessage
//         /Users/wenjie/go/src/go_code/leridge_server/socket/message.go:138 +0x4c7

// goroutine 6 [running]:
// runtime/debug.Stack(0x10, 0x11, 0x1)
//
//	/usr/local/go/src/runtime/debug/stack.go:24 +0x9f
//
// github.com/wellmoon/go/logger.getStack(0xc0000f99f0, 0xc000016610)
//
//	/Users/wenjie/go/pkg/mod/github.com/wellmoon/go@v0.0.0-20210905020657-0e9656f9d933/logger/logger.go:99 +0x34
//
// github.com/wellmoon/go/logger.print(0xc00008c1e0, 0xc0000f99f0, 0x10, 0xc0000f9ac8, 0x1, 0x1)
//
//	/Users/wenjie/go/pkg/mod/github.com/wellmoon/go@v0.0.0-20210905020657-0e9656f9d933/logger/logger.go:122 +0xd7
//
// github.com/wellmoon/go/logger.Debug(0x12bc5e0, 0xf, 0xc0000f9ac8, 0x1, 0x1)
//
//	/Users/wenjie/go/pkg/mod/github.com/wellmoon/go@v0.0.0-20210905020657-0e9656f9d933/logger/logger.go:131 +0x165
//
// main.MusicFind(0x1313350, 0xc0000e00e0, 0xc0000ea000)
//
//	/Users/wenjie/github.com/wellmoon/MusicFinder/main.go:25 +0x147
//
// net/http.HandlerFunc.ServeHTTP(0x12d0fc0, 0x1313350, 0xc0000e00e0, 0xc0000ea000)
//
//	/usr/local/go/src/net/http/server.go:2069 +0x44
//
// net/http.(*ServeMux).ServeHTTP(0x14788a0, 0x1313350, 0xc0000e00e0, 0xc0000ea000)
//
//	/usr/local/go/src/net/http/server.go:2448 +0x1ad
//
// net/http.serverHandler.ServeHTTP(0xc0000e0000, 0x1313350, 0xc0000e00e0, 0xc0000ea000)
//
//	/usr/local/go/src/net/http/server.go:2887 +0xa3
//
// net/http.(*conn).serve(0xc0000a8960, 0x1313860, 0xc000080240)
//
//	/usr/local/go/src/net/http/server.go:1952 +0x8cd
//
// created by net/http.(*Server).Serve
//
//	/usr/local/go/src/net/http/server.go:3013 +0x39b
func getStack() string {
	sep := string(os.PathSeparator)
	var res string
	var arr []string
	res = string(debug.Stack())
	// fmt.Println("=====================")
	// fmt.Println(res)
	// fmt.Println("=====================")
	arr = strings.Split(res, "github.com/wellmoon/go/logger.")
	res = arr[len(arr)-1]
	arr = strings.Split(res, "\n")
	res = arr[3]
	res = strings.TrimSpace(res)
	res = strings.Split(res, " ")[0]
	arr = strings.Split(res, sep)
	res = arr[len(arr)-1]
	res = "[" + res + "] "
	return res
}

func print(tarLog *log.Logger, s string, v ...interface{}) {
	if strings.Contains(s, "{}") {
		s = strings.ReplaceAll(s, "{}", "%v")
	}
	if !strings.HasSuffix(s, "\n") {
		s = s + "\n"
	}
	tarLog.Printf(getStack()+s, v...)
}

func Debug(format string, v ...interface{}) {
	switchFile, logFile := switchFile("detail")
	if debugLog == nil || switchFile {
		logWriter := io.MultiWriter(logFile, os.Stdout)
		debugLog = log.New(logWriter, "[DEBUG] ", flag)
	}
	print(debugLog, format, v...)
}

func Trace(format string, v ...interface{}) {
	switchFile, logFile := switchFile("detail")
	if traceLog == nil || switchFile {
		logWriter := io.MultiWriter(logFile)
		traceLog = log.New(logWriter, "[TRACE] ", flag)
	}
	print(traceLog, format, v...)
}

func MaskSensitiveInfo(logMessage string) string {
	// Regular expression to match the bot token pattern
	tokenRegex := regexp.MustCompile(`bot\d+:[a-zA-Z0-9_-]+`)

	// Replace the matched bot token with a masked version
	maskedLog := tokenRegex.ReplaceAllString(logMessage, "bot[REDACTED]")

	return maskedLog
}

func Error(format string, v ...interface{}) {
	switchFile, logFile := switchFile("error")
	if errorLog == nil || switchFile {
		logWriter := io.MultiWriter(logFile, os.Stdout)
		errorLog = log.New(logWriter, "[ERROR] ", flag)
	}
	print(errorLog, format, v...)
}

func Fatal(format string, v ...interface{}) {
	switchFile, logFile := switchFile("detail")
	if fatalLog == nil || switchFile {
		logWriter := io.MultiWriter(logFile, os.Stdout)
		fatalLog = log.New(logWriter, "[ERROR] ", flag)
	}
	print(fatalLog, format, v...)
}
