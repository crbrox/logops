package logops

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

var pool = sync.Pool{New: func() interface{} { return &bytes.Buffer{} }}

type Level int

const (
	None = iota
	Debug
	Info
	Warn
	Error
	Fatal
)

var levelNames = [...]string{
	None:  "NONE",
	Debug: "DEBUG",
	Info:  "INFO",
	Warn:  "WARN",
	Error: "ERROR",
	Fatal: "FATAL",
}

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

func (l *Logger) formatJSON(buffer *bytes.Buffer, level Level, localCx C, message string, params ...interface{}) {
	var dynCx C

	fmt.Fprintf(buffer, `{"time":%q, "lvl":%q`, time.Now().Format(time.RFC3339Nano), levelNames[level])
	for k, v := range localCx {
		fmt.Fprintf(buffer, ",%q:%q", k, v)
	}
	if l.ContextFunc != nil {
		dynCx = l.ContextFunc()
		for k, v := range dynCx {
			if _, already := localCx[k]; !already {
				fmt.Fprintf(buffer, ",%q:%q", k, v)
			}
		}
	}
	for k, v := range l.Context {
		if _, already := localCx[k]; !already {
			if _, already := dynCx[k]; !already {
				fmt.Fprintf(buffer, ",%q:%q", k, v)
			}
		}
	}
	if len(params) == 0 {
		fmt.Fprintf(buffer, `,"msg":%q}`, message)
	} else {
		m := fmt.Sprintf(message, params...)
		fmt.Fprintf(buffer, `,"msg":%q}`, m)
	}
	fmt.Fprintln(buffer)
}

func (l *Logger) logC(lvl Level, context C, message string, params []interface{}) {
	buffer := pool.Get().(*bytes.Buffer)

	l.formatJSON(buffer, lvl, context, message, params...)
	l.mu.Lock()
	_, err := l.Writer.Write(buffer.Bytes())
	l.mu.Unlock()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	buffer.Reset()
	pool.Put(buffer)
}
func (l *Logger) InfoC(context C, message string, params ...interface{}) {
	l.logC(Info, context, message, params)
}
func (l *Logger) Info(message string, params ...interface{}) {
	l.InfoC(nil, message, params...)
}
