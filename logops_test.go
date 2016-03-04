package logops

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"
)

var stringsForTesting = []string{
	"",
	"a",
	"aaa bbb ccc",
	"España y olé",
	`She said: "I know what it's like to be dead`,
	"{}}"}

var contextForTesting = C{"a": "A", "b": "BB", "España": "olé",
	"She said": `I know what it's like to be dead"`, "{": "}"}

func testFormatJSON(t *testing.T, l *Logger, localCx C, lvlWanted Level, msgWanted, format string, params ...interface{}) {
	var obj map[string]string
	var buffer bytes.Buffer

	start := time.Now()

	l.format(&buffer, lvlWanted, localCx, format, params...)
	res := buffer.Bytes()
	end := time.Now()

	err := json.Unmarshal(res, &obj)
	if err != nil {
		t.Fatal(err)
	}

	timeStr, ok := obj["time"]
	if !ok {
		t.Error("missing 'time' field")
	}
	timeStamp, err := time.Parse(time.RFC3339Nano, timeStr)
	if err != nil {
		t.Error(err)
	}
	if timeStamp.Before(start) {
		t.Error("time before start")
	}
	if timeStamp.After(end) {
		t.Error("time after end")
	}

	lvl, ok := obj["lvl"]
	if !ok {
		t.Error("missing 'lvl' field")
	}
	if lvl != levelNames[lvlWanted] {
		t.Errorf("level: wanted %q, got %q", lvlWanted, lvl)
	}

	msg, ok := obj["msg"]
	if !ok {
		t.Error("missing 'msg' field")
	}
	if msg != msgWanted {
		t.Errorf("level: wanted %q, got %q", msgWanted, msg)
	}

	for k, v := range localCx {
		value, ok := obj[k]
		if !ok {
			t.Errorf("missing local context field %s", k)
		}
		if v != value {
			t.Errorf("value for local context field %s: wanted %s, got %s", k, v, value)
		}
	}

	var funcCx C
	if l.ContextFunc != nil {
		funcCx = l.ContextFunc()
	}
	for k, v := range funcCx {
		if _, ok := localCx[k]; !ok {
			value, ok := obj[k]
			if !ok {
				t.Errorf("missing func context field %s", k)
			}
			if v != value {
				t.Errorf("value for func context field %s: wanted %s, got %s", k, v, value)
			}
		}
	}

	for k, v := range l.Context {
		if _, ok := localCx[k]; !ok {
			if _, ok := funcCx[k]; !ok {
				value, ok := obj[k]
				if !ok {
					t.Errorf("missing logger context field %s", k)
				}
				if v != value {
					t.Errorf("value for logger context field %s: wanted %s, got %s", k, v, value)
				}
			}
		}
	}
}

func TestSimpleMessageJSON(t *testing.T) {
	l := Logger{}
	for lvlWanted := All; lvlWanted < None; lvlWanted++ {
		for _, msgWanted := range stringsForTesting {
			testFormatJSON(t, &l, nil, lvlWanted, msgWanted, msgWanted)
		}
	}
}

func TestComplexMessage(t *testing.T) {
	l := Logger{}
	for lvlWanted := All; lvlWanted < None; lvlWanted++ {
		for i, text := range stringsForTesting {
			format := "%s,%#v,%f"
			cmpl := fmt.Sprintf(format, text, []int{1, i}, float64(i))
			testFormatJSON(t, &l, nil, lvlWanted, cmpl, format, text, []int{1, i}, float64(i))
		}
	}
}

func TestLocalContext(t *testing.T) {
	l := Logger{}
	for lvlWanted := All; lvlWanted < None; lvlWanted++ {
		for _, msgWanted := range stringsForTesting {
			testFormatJSON(t, &l, contextForTesting, lvlWanted, msgWanted, msgWanted)
		}
	}
}

func TestFuncContext(t *testing.T) {
	l := Logger{ContextFunc: func() C { return contextForTesting }}
	for lvlWanted := All; lvlWanted < None; lvlWanted++ {
		for _, msgWanted := range stringsForTesting {
			testFormatJSON(t, &l, nil, lvlWanted, msgWanted, msgWanted)
		}
	}
}

func TestLoggerContext(t *testing.T) {
	l := Logger{Context: contextForTesting}
	for lvlWanted := All; lvlWanted < None; lvlWanted++ {
		for _, msgWanted := range stringsForTesting {
			testFormatJSON(t, &l, nil, lvlWanted, msgWanted, msgWanted)
		}
	}
}

func TestAllContexts(t *testing.T) { t.Skip("to be implemented") }

type (
	contextLogFunc    func(l *Logger, context C, message string, params ...interface{})
	formatLogFunc     func(l *Logger, message string, params ...interface{})
	simpleLogFunction func(l *Logger, message string)
)

func testLevelC(t *testing.T, levelMethod Level, method contextLogFunc) {
	var buffer bytes.Buffer
	l := NewLogger(nil)
	ctx := C{"trying": "something"}
	l.Writer = &buffer
	for loggerLevel := All; loggerLevel <= levelMethod; loggerLevel++ {
		l.Level = loggerLevel
		buffer.Reset()
		method(l, ctx, "a not very long message")
		if buffer.Len() == 0 {
			t.Errorf("log not written for method %s when level %s", levelNames[levelMethod], levelNames[loggerLevel])
		}
	}
	for loggerLevel := levelMethod + 1; loggerLevel < None; loggerLevel++ {
		l.Level = loggerLevel
		buffer.Reset()
		method(l, ctx, "another short message")
		if buffer.Len() > 0 {
			t.Errorf("log written for method %s when level %s", levelNames[levelMethod], levelNames[loggerLevel])
		}
	}
}

func testLevelf(t *testing.T, levelMethod Level, method formatLogFunc) {

	testLevelC(t, levelMethod, func(l *Logger, ctx C, message string, params ...interface{}) {
		method(l, message, params...)
	})
}

func testLevel(t *testing.T, levelMethod Level, method simpleLogFunction) {

	testLevelC(t, levelMethod, func(l *Logger, ctx C, message string, params ...interface{}) {
		method(l, message)
	})
}

func TestInfof(t *testing.T) {
	testLevelf(t, Info, (*Logger).Infof)
}

func TestInfo(t *testing.T) {
	testLevel(t, Info, (*Logger).Info)
}

func TestInfoC(t *testing.T) {
	testLevelC(t, Info, (*Logger).InfoC)
}

type TestingBadWriter struct{}

var TestingBadWriterErr = errors.New("life goes on bra!")

func (TestingBadWriter) Write(b []byte) (int, error) {
	return 0, TestingBadWriterErr
}

func TestWriteErr(t *testing.T) {
	l := Logger{Writer: TestingBadWriter{}}
	for lvlWanted := All; lvlWanted < None; lvlWanted++ {
		for _, msgWanted := range stringsForTesting {
			err := l.LogC(lvlWanted, contextForTesting, msgWanted, nil)
			if err != TestingBadWriterErr {
				t.Errorf("writer error: want %#v, got %#v", TestingBadWriterErr, err)
			}
		}
	}
}
