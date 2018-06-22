package elog

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	colorRed = iota + 91
	colorGreen
	colorYellow
	colorBlue
	colorMagenta
)

const (
	NoticeLog = iota
	DebugLog
	InfoLog
	WarnLog
	ErrorLog
	MaxLevelLog
)

type Logger interface {
	Notice(a ...interface{})
	Debug(a ...interface{})
	Info(a ...interface{})
	Warn(a ...interface{})
	Error(a ...interface{})
	Fatal(a ...interface{})
	GetLogger() *log.Logger
	SetLogLevel(level int) error
	GetLogLevel() int
}

type loggerModule struct {
	logger *log.Logger
	name   string
	level  int
}

func fileOpen(path string) (*os.File, error) {
	if fi, err := os.Stat(path); err == nil {
		if !fi.IsDir() {
			return nil, fmt.Errorf("open %s: not a directory", path)
		}
	} else if os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0766); err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	var currentTime = time.Now().Format("2006-01-02_15.04")
	logfile, err := os.OpenFile(path+currentTime+"_LOG.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return logfile, nil
}

func NewLogger(moduleName string, level int) Logger {
	logger := log.New(os.Stderr, "", log.Ldate|log.Lmicroseconds|log.LstdFlags)
	module := loggerModule{logger, moduleName, level}
	return &module
}

func NewFileLogger(moduleName string, level int) Logger {
	InitFile()
	logger := log.New(fileAndStdoutWrite, "", log.Ldate|log.Lmicroseconds|log.LstdFlags)
	module := loggerModule{logger, moduleName, level}
	return &module
}

var fileAndStdoutWrite io.Writer

func InitFile() {
	var logFile *os.File
	var writers = []io.Writer{os.Stdout}
	logFile, err := fileOpen("./Log/")
	if err != nil {
		fmt.Println("error: open log file failed")
		os.Exit(1)
	}
	writers = append(writers, logFile)
	fileAndStdoutWrite = io.MultiWriter(writers...)
}

func (l *loggerModule) GetLogger() *log.Logger {
	return l.logger
}

func (l *loggerModule) SetLogLevel(level int) error {
	if level > MaxLevelLog || level < 0 {
		return errors.New("invalid log level")
	}
	l.level = level
	return nil
}

func (l *loggerModule) GetLogLevel() int {
	return l.level
}

func GetGID() uint64 {
	var buf [64]byte
	b := buf[:runtime.Stack(buf[:], false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

func getFunctionName() string {
	pc := make([]uintptr, 10)
	runtime.Callers(3, pc)
	f := runtime.FuncForPC(pc[0])

	file, line := f.FileLine(pc[0])
	fileName := filepath.Base(file)

	nameFull := f.Name()
	nameEnd := filepath.Ext(nameFull)

	funcName := strings.TrimPrefix(nameEnd, ".")


	return fileName + ":" + strconv.Itoa(line) + "-" + funcName


}

func (l *loggerModule) Notice(a ...interface{}) {
	if l.level > NoticeLog {
		return
	}
	prefix := []interface{}{"\x1b[" + strconv.Itoa(colorGreen) + "m" + "▶ NOTI " + "[" + l.name + "] " + getFunctionName() + "():" + "\x1b[0m "}
	a = append(prefix, a...)

	l.logger.Output(2, fmt.Sprintln(a...))
}

func (l *loggerModule) Debug(a ...interface{}) {
	if l.level > DebugLog {
		return
	}
	prefix := []interface{}{"\x1b[" + strconv.Itoa(colorBlue) + "m" + "▶ DEBU " + "[" + l.name + "] " +  getFunctionName() + "():" + "\x1b[0m "}
	a = append(prefix, a...)
	l.logger.Output(2, fmt.Sprintln(a...))
}

func (l *loggerModule) Info(a ...interface{}) {
	if l.level > InfoLog {
		return
	}
	prefix := []interface{}{"\x1b[" + strconv.Itoa(colorYellow) + "m" + "▶ INFO " + "[" + l.name + "] " +  getFunctionName() + "():" + "\x1b[0m "}
	a = append(prefix, a...)
	l.logger.Output(2, fmt.Sprintln(a...))
}

func (l *loggerModule) Warn(a ...interface{}) {
	if l.level > WarnLog {
		return
	}
	prefix := []interface{}{"\x1b[" + strconv.Itoa(colorMagenta) + "m" + "▶ WARN " + "[" + l.name + "] " +  getFunctionName() + "():" + "\x1b[0m "}
	a = append(prefix, a...)
	l.logger.Output(2, fmt.Sprintln(a...))
}

func (l *loggerModule) Error(a ...interface{}) {
	if l.level > ErrorLog {
		return
	}
	prefix := []interface{}{"\x1b[" + strconv.Itoa(colorRed) + "m" + "▶ ERRO " + "[" + l.name + "] " +  getFunctionName() + "():" + "\x1b[0m "}
	a = append(prefix, a...)
	l.logger.Output(2, fmt.Sprintln(a...))
}

func (l *loggerModule) Fatal(a ...interface{}) {
	if l.level > ErrorLog {
		return
	}
	prefix := []interface{}{"\x1b[" + strconv.Itoa(colorRed) + "m" + "▶ FATAL " + "[" + l.name + "] " +  getFunctionName() + "():" + "\x1b[0m "}
	a = append(prefix, a...)
	l.logger.Fatal(a...)
}
