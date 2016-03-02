package main

import (
	"github.com/crbrox/logops"
)

func main() {
	l := logops.NewLogger(nil)
	l.Info("%d y %d son %d", 2, 2, 4)
	l.Info("y ocho dieciséis")
	l.Context = logops.C{"prefix": "prefijo"}
	l.InfoC(logops.C{"local": "llamada 1"}, "%d y %d son %d", 2, 2, 4)
	l.InfoC(logops.C{"local": "llamada 2"}, "y ocho dieciséis")

}
