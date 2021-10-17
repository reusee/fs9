package fs9

import (
	"io/fs"

	"github.com/reusee/e4"
)

type MemFS struct {
	Entries []DirEntry
}

var _ FS = new(MemFS)

func (m *MemFS) Open(path string) (fs.File, error) {
	h, err := m.OpenHandle(path)
	if err != nil {
		return nil, err
	}
	return h, nil
}

func (m *MemFS) OpenHandle(path string, opts ...OpenOption) (Handle, error) {
	if !fs.ValidPath(path) {
		return nil, we.With(
			e4.NewInfo("path: %s", path),
		)(ErrInvalidPath)
	}

	handle := &MemHandle{}

	return handle, nil
}
