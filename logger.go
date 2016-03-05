package logops

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

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

func (l *Logger) format(buffer *bytes.Buffer, level Level, localCx C, message string, params []interface{}) {
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

		l.format(buffer, lvl, context, message, params)
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
