package main

import (
	"github.com/crbrox/logops"
)

func main() {
	l := logops.NewLogger(nil)
	l.Infof("%d y %d son %d", 2, 2, 4)
	l.Info("y ocho dieciséis")
	l.Context = logops.C{"prefix": "prefijo"}
	l.InfoC(logops.C{"local": "España y olé"}, "%d y %d son %d", 2, 2, 4)
	l.InfoC(logops.C{"local": `{"json":"pompón"}`}, "y ocho dieciséis")

}
