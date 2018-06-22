package elog_test

import (
	"github.com/ecoball/go-ecoball/common/elog"
	"testing"
)

func TestLogger_P(t *testing.T) {
	l := elog.NewLogger("Module", elog.NoticeLog)
	l.Notice("Test")
	l.Debug("Test")
	l.Info("Test")
	l.Warn("Test")
	l.Error("Test")

	l2 := elog.NewFileLogger("Module", elog.DebugLog)
	l2.Notice("example")
	l2.Debug("example")
	l2.Info("example")
	l2.Warn("example")
	l2.Error("example")
	ll := l2.GetLogger()
	ll.Println("Test")
}
