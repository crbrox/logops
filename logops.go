package logops

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	allLevel Level = iota
	DebugLevel
	InfoLevel
	WarnLevel
	ErrorLevel
	CriticalLevel
	noneLevel
)

type C map[string]string

var bufferPool = sync.Pool{New: func() interface{} { return &bytes.Buffer{} }}

type Level int

var levelNames = [...]string{
	allLevel:      "ALL",
	DebugLevel:    "DEBUG",
	InfoLevel:     "INFO",
	WarnLevel:     "WARN",
	ErrorLevel:    "ERROR",
	CriticalLevel: "FATAL",
	noneLevel:     "NONE",
}

var (
	timeFormat    string
	formatPrefix  string
	formatField   string
	formatPostfix string
)

var defaultLogger = NewLogger()

func init() {
	format := os.Getenv("LOGOPS_FORMAT")
	if strings.ToLower(format) == "dev" {
		setTextFormat()
	} else {
		setJSONFormat()
	}
}

type Logger struct {
	contextFunc atomic.Value
	context     atomic.Value
	level       int32
	writer      io.Writer
	mu          sync.Mutex
}

func NewLogger() *Logger {
	return NewLoggerWithWriter(io.Writer(os.Stdout))
}
func NewLoggerWithWriter(w io.Writer) *Logger {
	l := &Logger{}
	l.SetContextFunc(nil)
	l.SetContext(nil)
	l.SetLevel(allLevel)
	l.writer = w
	return l
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
	var dynamicContext C

	fmt.Fprintf(buffer, formatPrefix, time.Now().Format(timeFormat), levelNames[level])
	for k, v := range localCx {
		fmt.Fprintf(buffer, formatField, k, v)
	}
	contextFunc := l.contextFunc.Load().(func() C)
	if contextFunc != nil {
		dynamicContext = contextFunc()
		for k, v := range dynamicContext {
			if _, already := localCx[k]; !already {
				fmt.Fprintf(buffer, formatField, k, v)
			}
		}
	}
	loggerContext := l.context.Load().(C)
	for k, v := range loggerContext {
		if _, already := localCx[k]; !already {
			if _, already := dynamicContext[k]; !already {
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
	if Level(atomic.LoadInt32(&l.level)) <= lvl {
		buffer := bufferPool.Get().(*bytes.Buffer)

		l.format(buffer, lvl, context, message, params...)
		l.mu.Lock()
		_, err := l.writer.Write(buffer.Bytes())
		l.mu.Unlock()

		buffer.Reset()
		bufferPool.Put(buffer)

		return err
	}
	return nil
}

func (l *Logger) InfoC(context C, message string, params ...interface{}) {
	l.LogC(InfoLevel, context, message, params)
}

func (l *Logger) Infof(message string, params ...interface{}) {
	l.LogC(InfoLevel, nil, message, params)
}

func (l *Logger) Info(message string) {
	l.LogC(InfoLevel, nil, message, nil)
}

func (l *Logger) SetLevel(lvl Level) {
	atomic.StoreInt32(&l.level, int32(lvl))
}

func (l *Logger) SetContext(c C) {
	l.context.Store(c)

}

func (l *Logger) SetContextFunc(f func() C) {
	l.contextFunc.Store(f)
}

func (l *Logger) SetWriter(w io.Writer) {
	l.mu.Lock()
	l.writer = w
	l.mu.Unlock()
}

// Global logger

func InfoC(context C, message string, params ...interface{}) {
	defaultLogger.LogC(InfoLevel, context, message, params)
}

func Infof(message string, params ...interface{}) {
	defaultLogger.LogC(InfoLevel, nil, message, params)
}

func Info(message string) {
	defaultLogger.LogC(InfoLevel, nil, message, nil)
}

func SetLevel(lvl Level) {
	defaultLogger.SetLevel(lvl)
}

func SetContext(c C) {
	defaultLogger.SetContext(c)
}

func SetContextFunc(f func() C) {
	defaultLogger.SetContextFunc(f)
}

func SetWriter(w io.Writer) {
	defaultLogger.SetWriter(w)
}
