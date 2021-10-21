package fs9

import (
	"io/fs"
	"strings"

	"github.com/reusee/e4"
)

func SplitPath(path string) ([]string, error) {
	if !fs.ValidPath(path) {
		return nil, we.With(
			e4.Info("path: %s", path),
		)(ErrInvalidPath)
	}
	return strings.Split(path, "/"), nil
}
