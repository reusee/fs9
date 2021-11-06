package fs9

import (
	"fmt"

	"github.com/reusee/sb"
)

func eq(args ...any) {
	if len(args)%2 != 0 {
		panic("bad args count")
	}
	for i := 0; i < len(args); i += 2 {
		if sb.MustCompare(
			sb.Marshal(args[i]),
			sb.Marshal(args[i+1]),
		) != 0 {
			ce(fmt.Errorf(
				"pair %d / %d not equal\nleft: %T: %+v\nright: %T: %+v",
				i,
				i+1,
				args[i],
				args[i],
				args[i+1],
				args[i+1],
			))
		}
	}
}
