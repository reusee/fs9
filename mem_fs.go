package fs9

import (
	"io/fs"
	"strings"
	"sync"
	"time"

	"github.com/reusee/e4"
)

type MemFS struct {
	sync.RWMutex
	Root *File
}

var _ FS = new(MemFS)

func NewMemFS() *MemFS {
	return &MemFS{
		Root: &File{
			IsDir:   true,
			Name:    "_root_",
			ModTime: time.Now(),
		},
	}
}

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

	var spec openSpec
	for _, opt := range opts {
		opt(&spec)
	}

	if path == "." {
		// root path
		return &MemHandle{
			FS:   m,
			Path: []string{},
		}, nil
	}

	pathParts := strings.Split(path, "/")
	err := m.Apply(pathParts, Ensure(pathParts[len(pathParts)-1], spec.Create))
	if err != nil {
		return nil, err
	}

	handle := &MemHandle{
		FS:   m,
		Path: pathParts,
	}

	return handle, nil
}

func (m *MemFS) Apply(path []string, op Operation) error {
	newRoot, err := m.Root.Apply(path, op)
	if err != nil {
		return err
	}
	if newRoot != nil {
		m.Lock()
		m.Root = newRoot
		m.Unlock()
	}
	return nil
}
