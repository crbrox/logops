package logops

import "io"

// Global logger
var defaultLogger = NewLogger()

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
