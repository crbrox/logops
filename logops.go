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
	prefixFormat  = `{"time":%q, "lvl":%q`
	fieldFormat   = ",%q:%q"
	postfixFormat = `,"msg":%q}`
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

func (l *Logger) formatJSON(buffer *bytes.Buffer, level Level, localCx C, message string, params ...interface{}) {
	var dynCx C

	fmt.Fprintf(buffer, prefixFormat, time.Now().Format(time.RFC3339Nano), levelNames[level])
	for k, v := range localCx {
		fmt.Fprintf(buffer, fieldFormat, k, v)
	}
	if l.ContextFunc != nil {
		dynCx = l.ContextFunc()
		for k, v := range dynCx {
			if _, already := localCx[k]; !already {
				fmt.Fprintf(buffer, fieldFormat, k, v)
			}
		}
	}
	for k, v := range l.Context {
		if _, already := localCx[k]; !already {
			if _, already := dynCx[k]; !already {
				fmt.Fprintf(buffer, fieldFormat, k, v)
			}
		}
	}
	if len(params) == 0 {
		fmt.Fprintf(buffer, postfixFormat, message)
	} else {
		m := fmt.Sprintf(message, params...)
		fmt.Fprintf(buffer, postfixFormat, m)
	}
	fmt.Fprintln(buffer)
}

func (l *Logger) LogC(lvl Level, context C, message string, params []interface{}) error {
	if l.Level <= lvl {
		buffer := pool.Get().(*bytes.Buffer)

		l.formatJSON(buffer, lvl, context, message, params...)
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
