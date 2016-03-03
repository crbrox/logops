package logops

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func testFormatJSON(t *testing.T, l *Logger, localCx C, lvlWanted Level, msgWanted, format string, params ...interface{}) {
	var obj map[string]string
	var buffer bytes.Buffer

	start := time.Now()

	l.formatJSON(&buffer, lvlWanted, localCx, format, params...)
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

}

func TestSimpleMessageJSON(t *testing.T) {
	l := Logger{}
	for lvlWanted := All; lvlWanted < None; lvlWanted++ {
		for _, msgWanted := range []string{
			"",
			"a",
			"aaa bbb ccc",
			"España y olé",
			`She said: "I know what it's like to be dead`,
			"{}}"} {
			testFormatJSON(t, &l, nil, lvlWanted, msgWanted, msgWanted)
		}
	}
}

func TestComplexMessage(t *testing.T) {
	l := Logger{}
	for lvlWanted := All; lvlWanted < None; lvlWanted++ {
		for i, text := range []string{
			"",
			"a",
			"aaa bbb ccc",
			"España y olé",
			`She said: "I know what it's like to be dead`,
			"{}}"} {
			format := "%s,%#v,%f"
			cmpl := fmt.Sprintf(format, text, []int{1, i}, float64(i))
			testFormatJSON(t, &l, nil, lvlWanted, cmpl, format, text, []int{1, i}, float64(i))
		}
	}
}

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
