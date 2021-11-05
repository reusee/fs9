package fs9

import (
	"io/fs"
	"strings"

	"github.com/reusee/e4"
)

func PathToSlice(path string) (ret []string, err error) {
	if path == "" || path == "." {
		return
	}
	if !fs.ValidPath(path) {
		return nil, we.With(
			e4.Info("path: %q", path),
		)(ErrInvalidPath)
	}
	return strings.Split(path, "/"), nil
}

func PathToKeyPath(path string) (ret KeyPath, err error) {
	if path == "" || path == "." {
		return
	}
	if !fs.ValidPath(path) {
		return nil, we.With(
			e4.Info("path: %q", path),
		)(ErrInvalidPath)
	}
	for _, part := range strings.Split(path, "/") {
		ret = append(ret, part)
	}
	return
}
