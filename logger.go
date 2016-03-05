package logops

import (
	"bytes"
	"encoding/json"
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

func (l *Logger) format(buffer *bytes.Buffer, lline logLine) {
	var dynamicContext C
	now := time.Now()
	fmt.Fprintf(buffer, prefixFormat, now.Format(timeFormat), levelNames[lline.level])

	if lline.err != nil {
		errMsg := formatError(lline.err)
		fmt.Fprintf(buffer, errorFormat, ErrFieldName, errMsg)
	}

	for k, v := range lline.localCx {
		if lline.err != nil && k == ErrFieldName {
			continue
		}
		fmt.Fprintf(buffer, fieldFormat, k, v)
	}
	contextFunc := l.contextFunc.Load().(func() C)
	if contextFunc != nil {
		dynamicContext = contextFunc()
		for k, v := range dynamicContext {
			if lline.err != nil && k == ErrFieldName {
				continue
			}
			if _, already := lline.localCx[k]; !already {
				fmt.Fprintf(buffer, fieldFormat, k, v)
			}
		}
	}
	loggerContext := l.context.Load().(C)
	for k, v := range loggerContext {
		if lline.err != nil && k == ErrFieldName {
			continue
		}
		if _, already := lline.localCx[k]; !already {
			if _, already := dynamicContext[k]; !already {
				fmt.Fprintf(buffer, fieldFormat, k, v)
			}
		}
	}
	if len(lline.params) == 0 {
		fmt.Fprintf(buffer, postfixFormat, lline.message)
	} else {
		m := fmt.Sprintf(lline.message, lline.params...)
		fmt.Fprintf(buffer, postfixFormat, m)
	}
	fmt.Fprintln(buffer) // newline at the end
}

func formatError(err error) string {
	b := getBuffer()
	defer putBuffer(b)
	errJSON := json.NewEncoder(b).Encode(err)
	if errJSON != nil {
		b.Reset()
		b.WriteString(err.Error())
		b.WriteByte(' ')
		b.WriteByte('(')
		b.WriteString(errJSON.Error())
		b.WriteByte(')')
		return fmt.Sprintf("%q", b.String()) // the string as a valid JSON object
	}
	b.Truncate(b.Len() - 1) // remove trailing newline
	return b.String()
}

func (l *Logger) LogC(ll logLine) error {
	if Level(atomic.LoadInt32(&l.level)) <= ll.level {

		b := getBuffer()

		l.format(b, ll)
		l.mu.Lock()
		_, err := l.writer.Write(b.Bytes())
		l.mu.Unlock()

		putBuffer(b)

		return err
	}
	return nil
}

func (l *Logger) InfoC(context C, message string, params ...interface{}) {
	l.LogC(logLine{level: InfoLevel, localCx: context, message: message, params: params})
}

func (l *Logger) Infof(message string, params ...interface{}) {
	l.LogC(logLine{level: InfoLevel, message: message, params: params})
}

func (l *Logger) Info(message string) {
	l.LogC(logLine{level: InfoLevel, message: message})
}

func (l *Logger) ErrorE(err error, context C, message string, params ...interface{}) {

	l.LogC(logLine{err: err, level: ErrorLevel, localCx: context, message: message, params: params})
}
