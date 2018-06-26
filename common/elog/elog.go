// Copyright 2018 The go-ecoball Authors
// This file is part of the go-ecoball library.
//
// The go-ecoball library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ecoball library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ecoball library. If not, see <http://www.gnu.org/licenses/>.

package elog

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/ecoball/go-ecoball/common/config"
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
	FatalLog
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
	InitFile()
	logger := log.New(fileAndStdoutWrite, "", log.Ldate|log.Lmicroseconds|log.LstdFlags)
	module := loggerModule{logger, moduleName, level}
	return &module
}

var fileAndStdoutWrite io.Writer

func InitFile() {
	//get configured output
	var output io.Writer = os.Stdout
	if !config.OutputToTerminal {
		output = ioutil.Discard
	}

	//get configured log directory
	logDir := "./Log/"
	if config.LogDir != "" && config.LogDir != logDir {
		logDir = config.LogDir
	}

	logFile, err := fileOpen(logDir)
	if err != nil {
		fmt.Println("open log file failed: ", err)
		os.Exit(1)
	}

	var writers = []io.Writer{output, logFile}
	fileAndStdoutWrite = io.MultiWriter(writers...)
}

func checkPrint(printLevel int) bool {
	if printLevel < config.LogLevel {
		return false
	}
	return true
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
	if l.level > NoticeLog || !checkPrint(NoticeLog) {
		return
	}
	prefix := []interface{}{"\x1b[" + strconv.Itoa(colorGreen) + "m" + "▶ NOTI " + "[" + l.name + "] " + getFunctionName() + "():" + "\x1b[0m "}
	a = append(prefix, a...)

	l.logger.Output(2, fmt.Sprintln(a...))
}

func (l *loggerModule) Debug(a ...interface{}) {
	if l.level > DebugLog || !checkPrint(DebugLog) {
		return
	}
	prefix := []interface{}{"\x1b[" + strconv.Itoa(colorBlue) + "m" + "▶ DEBU " + "[" + l.name + "] " + getFunctionName() + "():" + "\x1b[0m "}
	a = append(prefix, a...)
	l.logger.Output(2, fmt.Sprintln(a...))
}

func (l *loggerModule) Info(a ...interface{}) {
	if l.level > InfoLog || !checkPrint(InfoLog) {
		return
	}
	prefix := []interface{}{"\x1b[" + strconv.Itoa(colorYellow) + "m" + "▶ INFO " + "[" + l.name + "] " + getFunctionName() + "():" + "\x1b[0m "}
	a = append(prefix, a...)
	l.logger.Output(2, fmt.Sprintln(a...))
}

func (l *loggerModule) Warn(a ...interface{}) {
	if l.level > WarnLog || !checkPrint(WarnLog) {
		return
	}
	prefix := []interface{}{"\x1b[" + strconv.Itoa(colorMagenta) + "m" + "▶ WARN " + "[" + l.name + "] " + getFunctionName() + "():" + "\x1b[0m "}
	a = append(prefix, a...)
	l.logger.Output(2, fmt.Sprintln(a...))
}

func (l *loggerModule) Error(a ...interface{}) {
	if l.level > ErrorLog || !checkPrint(ErrorLog) {
		return
	}
	prefix := []interface{}{"\x1b[" + strconv.Itoa(colorRed) + "m" + "▶ ERRO " + "[" + l.name + "] " + getFunctionName() + "():" + "\x1b[0m "}
	a = append(prefix, a...)
	l.logger.Output(2, fmt.Sprintln(a...))
}

func (l *loggerModule) Fatal(a ...interface{}) {
	if l.level > FatalLog || !checkPrint(FatalLog) {
		return
	}
	prefix := []interface{}{"\x1b[" + strconv.Itoa(colorRed) + "m" + "▶ FATAL " + "[" + l.name + "] " + getFunctionName() + "():" + "\x1b[0m "}
	a = append(prefix, a...)
	l.logger.Fatal(a...)
}
