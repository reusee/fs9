package fs9

import (
	"fmt"

	"github.com/reusee/e4"
	"github.com/reusee/pp"
)

var (
	he = e4.Handle
	ce = e4.Check
	we = e4.Wrap
	pt = fmt.Printf
)

type (
	any  = interface{}
	Src  = pp.Src
	Sink = pp.Sink
)
