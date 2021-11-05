package fs9

import (
	"io/fs"
	"strings"

	"github.com/reusee/e4"
)

func SplitPath[T ~[]string](path string, target *T) (err error) {
	if !fs.ValidPath(path) {
		return we.With(
			e4.Info("path: %s", path),
		)(ErrInvalidPath)
	}
	for _, part := range strings.Split(path, "/") {
		*target = append(*target, part)
	}
	return
}
