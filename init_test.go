package fs9

import "github.com/reusee/e4"

func init() {
	ce = ce.With(e4.WrapStacktrace)
	we = we.With(e4.WrapStacktrace)
}
