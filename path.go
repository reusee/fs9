package fs9

import (
	"io/fs"
	"strings"

	"github.com/reusee/e4"
)

func PathToSlice(path string) (ret []string, err error) {
	if !fs.ValidPath(path) {
		return nil, we.With(
			e4.Info("path: %s", path),
		)(ErrInvalidPath)
	}
	return strings.Split(path, "/"), nil
}

func PathToKeyPath(path string) (ret KeyPath, err error) {
	if !fs.ValidPath(path) {
		return nil, we.With(
			e4.Info("path: %s", path),
		)(ErrInvalidPath)
	}
	for _, part := range strings.Split(path, "/") {
		ret = append(ret, part)
	}
	return
}
