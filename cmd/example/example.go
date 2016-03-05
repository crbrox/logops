package main

import (
	"errors"
	"fmt"

	"github.com/crbrox/logops"
)

type complexErr struct {
	Text  string
	Cause *complexErr
}

func (ce complexErr) Error() string { return "uno complejito" }

type NotJSONableNError struct{}

func (NotJSONableNError) Error() string { return "a very strange error" }
func (NotJSONableNError) MarshalJSON() ([]byte, error) {
	return nil, errors.New("JSON not supported")
}

func main() {
	l := logops.NewLogger()

	l.Infof("%d y %d son %d", 2, 2, 4)
	l.Info("y ocho dieciséis")
	l.SetContext(logops.C{"prefix": "prefijo"})
	l.InfoC(logops.C{"local": "España y olé"}, "%d y %d son %d", 2, 2, 4)
	l.InfoC(logops.C{"local": `{"json":"pompón"}`}, "y ocho dieciséis")

	ce := complexErr{"1", &complexErr{"2", &complexErr{"3", nil}}}
	l.ErrorE(ce, nil, "esta sí que es buena")
	nj := NotJSONableNError{}
	l.ErrorE(nj, nil, "otro mejor")

	fmt.Println()

	fmt.Println("Con funciones del paquete")
	logops.Infof("%d y %d son %d", 2, 2, 4)
	logops.Info("y ocho dieciséis")
	logops.SetContext(logops.C{"prefix": "prefijo"})
	logops.InfoC(logops.C{"local": "España y olé"}, "%d y %d son %d", 2, 2, 4)
	logops.InfoC(logops.C{"local": `{"json":"pompón"}`}, "y ocho dieciséis")

}
