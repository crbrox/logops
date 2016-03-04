package main

import (
	"fmt"

	"github.com/crbrox/logops"
)

func main() {
	l := logops.NewLogger()
	l.Infof("%d y %d son %d", 2, 2, 4)
	l.Info("y ocho dieciséis")
	l.SetContext(logops.C{"prefix": "prefijo"})
	l.InfoC(logops.C{"local": "España y olé"}, "%d y %d son %d", 2, 2, 4)
	l.InfoC(logops.C{"local": `{"json":"pompón"}`}, "y ocho dieciséis")

	fmt.Println("Con funciones del paquete")
	logops.Infof("%d y %d son %d", 2, 2, 4)
	logops.Info("y ocho dieciséis")
	logops.SetContext(logops.C{"prefix": "prefijo"})
	logops.InfoC(logops.C{"local": "España y olé"}, "%d y %d son %d", 2, 2, 4)
	logops.InfoC(logops.C{"local": `{"json":"pompón"}`}, "y ocho dieciséis")

}
