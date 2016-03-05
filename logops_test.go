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

func testFormatJSON(t *testing.T, l *Logger, contextFromArg C, lvlWanted Level, msgWanted, format string, params ...interface{}) {
	var obj map[string]string
	var buffer bytes.Buffer

	start := time.Now()

	l.format(&buffer, lvlWanted, contextFromArg, format, params...)
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

	for k, v := range contextFromArg {
		value, ok := obj[k]
		if !ok {
			t.Errorf("missing local context field %s", k)
		}
		if v != value {
			t.Errorf("value for local context field %s: wanted %s, got %s", k, v, value)
		}
	}

	var contextFromFunction C
	var function = l.contextFunc.Load().(func() C)
	if function != nil {
		contextFromFunction = function()
	}
	for k, v := range contextFromFunction {
		if _, ok := contextFromArg[k]; !ok {
			value, ok := obj[k]
			if !ok {
				t.Errorf("missing func context field %s", k)
			}
			if v != value {
				t.Errorf("value for func context field %s: wanted %s, got %s", k, v, value)
			}
		}
	}

	for k, v := range l.context.Load().(C) {
		if _, ok := contextFromArg[k]; !ok {
			if _, ok := contextFromFunction[k]; !ok {
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
	l := NewLogger()
	for lvlWanted := allLevel; lvlWanted < noneLevel; lvlWanted++ {
		for _, msgWanted := range stringsForTesting {
			testFormatJSON(t, l, nil, lvlWanted, msgWanted, msgWanted)
		}
	}
}

func TestComplexMessage(t *testing.T) {
	l := NewLogger()
	for lvlWanted := allLevel; lvlWanted < noneLevel; lvlWanted++ {
		for i, text := range stringsForTesting {
			format := "%s,%#v,%f"
			cmpl := fmt.Sprintf(format, text, []int{1, i}, float64(i))
			testFormatJSON(t, l, nil, lvlWanted, cmpl, format, text, []int{1, i}, float64(i))
		}
	}
}

func TestLocalContext(t *testing.T) {
	l := NewLogger()
	for lvlWanted := allLevel; lvlWanted < noneLevel; lvlWanted++ {
		for _, msgWanted := range stringsForTesting {
			testFormatJSON(t, l, contextForTesting, lvlWanted, msgWanted, msgWanted)
		}
	}
}

func TestFuncContext(t *testing.T) {
	l := NewLogger()
	l.SetContextFunc(func() C { return contextForTesting })
	for lvlWanted := allLevel; lvlWanted < noneLevel; lvlWanted++ {
		for _, msgWanted := range stringsForTesting {
			testFormatJSON(t, l, nil, lvlWanted, msgWanted, msgWanted)
		}
	}
}

func TestLoggerContext(t *testing.T) {
	l := NewLogger()
	l.SetContext(contextForTesting)
	for lvlWanted := allLevel; lvlWanted < noneLevel; lvlWanted++ {
		for _, msgWanted := range stringsForTesting {
			testFormatJSON(t, l, nil, lvlWanted, msgWanted, msgWanted)
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
	l := NewLoggerWithWriter(&buffer)
	ctx := C{"trying": "something"}
	for loggerLevel := allLevel; loggerLevel <= levelMethod; loggerLevel++ {
		l.SetLevel(loggerLevel)
		buffer.Reset()
		method(l, ctx, "a not very long message")
		if buffer.Len() == 0 {
			t.Errorf("log not written for method %s when level %s", levelNames[levelMethod], levelNames[loggerLevel])
		}
	}
	for loggerLevel := levelMethod + 1; loggerLevel < noneLevel; loggerLevel++ {
		l.SetLevel(loggerLevel)
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
	testLevelf(t, InfoLevel, (*Logger).Infof)
}

func TestInfo(t *testing.T) {
	testLevel(t, InfoLevel, (*Logger).Info)
}

func TestInfoC(t *testing.T) {
	testLevelC(t, InfoLevel, (*Logger).InfoC)
}

type TestingBadWriter struct{}

var TestingBadWriterErr = errors.New("life goes on bra!")

func (TestingBadWriter) Write(b []byte) (int, error) {
	return 0, TestingBadWriterErr
}

func TestWriteErr(t *testing.T) {
	l := NewLoggerWithWriter(TestingBadWriter{})
	for lvlWanted := allLevel; lvlWanted < noneLevel; lvlWanted++ {
		for _, msgWanted := range stringsForTesting {
			err := l.LogC(lvlWanted, contextForTesting, msgWanted, nil)
			if err != TestingBadWriterErr {
				t.Errorf("writer error: want %#v, got %#v", TestingBadWriterErr, err)
			}
		}
	}
}
