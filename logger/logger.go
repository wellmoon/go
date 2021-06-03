package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
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
func switchFile(logFile string) (bool, *os.File) {
	var res bool
	if !pathExists(getLogPath() + logFile + ".log") {
		file, _ := os.Create(getLogPath() + logFile + ".log")
		return false, file
	} else {
		oldFile, _ := os.OpenFile(getLogPath()+logFile+".log", os.O_RDONLY, os.ModePerm)
		fileStat, _ := oldFile.Stat()
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

func Debug(format string, v ...interface{}) {
	switchFile, logFile := switchFile("detail")
	if debugLog == nil || switchFile {
		logWriter := io.MultiWriter(logFile, os.Stdout)
		debugLog = log.New(logWriter, "[DEBUG] ", flag)
	}
	debugLog.Printf(format, v...)
}

func Trace(format string, v ...interface{}) {
	switchFile, logFile := switchFile("detail")
	if traceLog == nil || switchFile {
		traceLog = log.New(logFile, "[TRACE] ", flag)
	}
	traceLog.Printf(format, v...)
}

func Error(format string, v ...interface{}) {
	switchFile, logFile := switchFile("detail")
	if errorLog == nil || switchFile {
		logWriter := io.MultiWriter(logFile, os.Stdout)
		errorLog = log.New(logWriter, "[ERROR] ", flag)
	}
	errorLog.Printf(format, v...)
}

func Fatal(format string, v ...interface{}) {
	switchFile, logFile := switchFile("detail")
	if fatalLog == nil || switchFile {
		logWriter := io.MultiWriter(logFile, os.Stdout)
		fatalLog = log.New(logWriter, "[ERROR] ", flag)
	}
	fatalLog.Fatalf(format, v...)
}
