package logops

import (
	"bytes"
	"os"
	"strings"
	"sync"
	"time"
)

type C map[string]string

type Level int

const (
	allLevel Level = iota
	DebugLevel
	InfoLevel
	WarnLevel
	ErrorLevel
	CriticalLevel
	noneLevel
)

var levelNames = [...]string{
	allLevel:      "ALL",
	DebugLevel:    "DEBUG",
	InfoLevel:     "INFO",
	WarnLevel:     "WARN",
	ErrorLevel:    "ERROR",
	CriticalLevel: "FATAL",
	noneLevel:     "NONE",
}

var bufferPool = sync.Pool{New: func() interface{} { return &bytes.Buffer{} }}

var (
	timeFormat    string
	formatPrefix  string
	formatField   string
	formatPostfix string
)

func init() {
	format := os.Getenv("LOGOPS_FORMAT")
	if strings.ToLower(format) == "dev" {
		setTextFormat()
	} else {
		setJSONFormat()
	}
}

func setJSONFormat() {
	timeFormat = time.RFC3339Nano
	formatPrefix = `{"time":%q, "lvl":%q`
	formatField = ",%q:%q"
	formatPostfix = `,"msg":%q}`
}

func setTextFormat() {
	timeFormat = "15:04:05.000"
	formatPrefix = "%s %s\t" // time and level
	formatField = " [%s=%s]" // key and value
	formatPostfix = " %s"    // message
}
