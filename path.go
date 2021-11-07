package fs9

import (
	"io/fs"
	"strings"

	"github.com/reusee/e4"
)

func NameToPath(name string) (ret []string, err error) {
	if name == "" || name == "." {
		return
	}
	if !fs.ValidPath(name) {
		return nil, we.With(
			e4.Info("path: %q", name),
		)(ErrInvalidPath)
	}
	return strings.Split(name, "/"), nil
}
