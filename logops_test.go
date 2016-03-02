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
	for _, lvlWanted := range []Level{
		Debug,
		Info,
		Warn,
		Error,
		Fatal} {
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
	for _, lvlWanted := range []Level{
		Debug,
		Info,
		Warn,
		Error,
		Fatal} {
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
