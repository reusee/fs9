package fs9

import (
	"io/fs"
	"sync"

	"github.com/reusee/e4"
)

type MemFS struct {
	sync.RWMutex
	version   int64
	Root      *File
	openedIDs map[int64]int
	detached  map[int64]*File
}

var _ FS = new(MemFS)

func NewMemFS() *MemFS {
	return &MemFS{
		Root:      NewFile("_root_", true),
		openedIDs: make(map[int64]int),
		detached:  make(map[int64]*File),
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

	var id int64
	pathParts, err := SplitPath(path)
	if err != nil {
		return nil, err
	}
	err = m.Apply(pathParts, 0, func(file *File) (*File, error) {
		ret, err := Ensure(
			pathParts[len(pathParts)-1],
			false,
			spec.Create,
		)(file)
		if err != nil {
			return nil, err
		}
		id = ret.id
		return ret, nil
	})
	if err != nil {
		return nil, err
	}
	if id == 0 {
		panic("impossible")
	}

	m.openedIDs[id]++
	handle := &MemHandle{
		FS:   m,
		Path: pathParts,
		id:   id,
		onClose: []func(){
			func() {
				m.Lock()
				defer m.Unlock()
				m.openedIDs[id]--
			},
		},
	}

	return handle, nil
}

func (m *MemFS) Apply(path []string, id int64, op Operation) error {
	m.Lock()
	defer m.Unlock()
	m.version++

	newRoot, err := m.Root.Apply(m.version, path, func(file *File) (*File, error) {

		if id > 0 && (file == nil || file.id != id) {
			// detached
			detached, ok := m.detached[id]
			if !ok {
				panic("impossible")
			}
			newFile, err := op(detached)
			if err != nil {
				return nil, err
			}
			if newFile != detached {
				m.detached[id] = newFile
			}

			// return origin file to avoid updating the tree
			return file, nil
		}

		return op(file)
	})
	if err != nil {
		return err
	}

	if newRoot != m.Root {
		m.Root = newRoot
	}

	return nil
}

func (m *MemFS) MakeDir(path string) error {
	parts, err := SplitPath(path)
	if err != nil {
		return err
	}
	return m.Apply(
		parts,
		0,
		Ensure(
			parts[len(parts)-1],
			true,
			true,
		),
	)
}

func (m *MemFS) MakeDirAll(path string) error {
	if path == "" {
		return nil
	}
	parts, err := SplitPath(path)
	if err != nil {
		return err
	}
	for i := 0; i < len(parts); i++ {
		dir := parts[:i+1]
		name := parts[i]
		if err := m.Apply(
			dir,
			0,
			Ensure(
				name,
				true,
				true,
			),
		); err != nil {
			return err
		}
	}
	return nil
}

func (m *MemFS) Remove(path string, options ...RemoveOption) error {
	if path == "" {
		return nil
	}
	var spec removeSpec
	for _, fn := range options {
		fn(&spec)
	}
	parts, err := SplitPath(path)
	if err != nil {
		return err
	}
	return m.Apply(
		parts,
		0,
		func(file *File) (*File, error) {
			if file.IsDir && len(file.Entries) > 0 && !spec.All {
				return nil, we.With(
					e4.Info("path: %s", path),
				)(ErrDirNotEmpty)
			}
			if m.openedIDs[file.id] > 0 {
				// add to detached files
				if _, ok := m.detached[file.id]; !ok {
					m.detached[file.id] = file
					//TODO delete entry
				}
			}
			return nil, nil
		},
	)
}
