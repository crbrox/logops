package logops

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

func init() {
	format := os.Getenv("LOGOPS_FORMAT")
	if strings.ToLower(format) == "dev" {
		setTextFormat()
	} else {
		setJSONFormat()
	}
}

var pool = sync.Pool{New: func() interface{} { return &bytes.Buffer{} }}

type Level int

const (
	All Level = iota
	Debug
	Info
	Warn
	Error
	Fatal
	None
)

var levelNames = [...]string{
	All:   "ALL",
	Debug: "DEBUG",
	Info:  "INFO",
	Warn:  "WARN",
	Error: "ERROR",
	Fatal: "FATAL",
	None:  "NONE",
}

var (
	timeFormat    string
	formatPrefix  string
	formatField   string
	formatPostfix string
)

type C map[string]string

type Logger struct {
	ContextFunc func() C
	Context     C
	Level       Level
	mu          sync.Mutex
	Writer      io.Writer
}

func NewLogger(context C) *Logger {
	return &Logger{Writer: os.Stdout, Context: context}
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

func (l *Logger) format(buffer *bytes.Buffer, level Level, localCx C, message string, params ...interface{}) {
	var dynCx C

	fmt.Fprintf(buffer, formatPrefix, time.Now().Format(timeFormat), levelNames[level])
	for k, v := range localCx {
		fmt.Fprintf(buffer, formatField, k, v)
	}
	if l.ContextFunc != nil {
		dynCx = l.ContextFunc()
		for k, v := range dynCx {
			if _, already := localCx[k]; !already {
				fmt.Fprintf(buffer, formatField, k, v)
			}
		}
	}
	for k, v := range l.Context {
		if _, already := localCx[k]; !already {
			if _, already := dynCx[k]; !already {
				fmt.Fprintf(buffer, formatField, k, v)
			}
		}
	}
	if len(params) == 0 {
		fmt.Fprintf(buffer, formatPostfix, message)
	} else {
		m := fmt.Sprintf(message, params...)
		fmt.Fprintf(buffer, formatPostfix, m)
	}
	fmt.Fprintln(buffer)
}

func (l *Logger) LogC(lvl Level, context C, message string, params []interface{}) error {
	if l.Level <= lvl {
		buffer := pool.Get().(*bytes.Buffer)

		l.format(buffer, lvl, context, message, params...)
		l.mu.Lock()
		_, err := l.Writer.Write(buffer.Bytes())
		l.mu.Unlock()

		buffer.Reset()
		pool.Put(buffer)

		return err
	}
	return nil
}

func (l *Logger) InfoC(context C, message string, params ...interface{}) {
	l.LogC(Info, context, message, params)
}

func (l *Logger) Infof(message string, params ...interface{}) {
	l.LogC(Info, nil, message, params)
}

func (l *Logger) Info(message string) {
	l.LogC(Info, nil, message, nil)
}
