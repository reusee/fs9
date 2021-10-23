package fs9

import (
	"io/fs"
	"sync"

	"github.com/reusee/e4"
	"github.com/reusee/pp"
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
	err = m.Apply(pathParts, OperationCtx{}, func(file *File) (*File, error) {
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
				if m.openedIDs[id] == 0 {
					// delete detached if any
					delete(m.detached, id)
					// clear
					delete(m.openedIDs, id)
				}
			},
		},
	}

	return handle, nil
}

type ApplySpec struct {
	Path []string
	Ctx  OperationCtx
	Op   Operation
}

func (m *MemFS) ApplyAll(specs ...ApplySpec) error {
	if len(specs) == 0 {
		return nil
	}

	m.Lock()
	defer m.Unlock()
	m.version++
	root := m.Root

	for _, spec := range specs {
		spec.Ctx.Version = m.version

		// detached
		if spec.Ctx.FileID > 0 && m.detached[spec.Ctx.FileID] != nil {
			detached, ok := m.detached[spec.Ctx.FileID]
			if !ok {
				panic("impossible")
			}
			newFile, err := spec.Op(detached)
			if err != nil {
				return err
			}
			if newFile != detached {
				m.detached[spec.Ctx.FileID] = newFile
			}
			continue
		}

		var err error
		root, err = root.Apply(spec.Path, spec.Ctx, func(file *File) (*File, error) {

			if spec.Ctx.FileID > 0 && (file == nil || file.id != spec.Ctx.FileID) {
				// detached
				detached, ok := m.detached[spec.Ctx.FileID]
				if !ok {
					panic("impossible")
				}
				newFile, err := spec.Op(detached)
				if err != nil {
					return nil, err
				}
				if newFile != detached {
					m.detached[spec.Ctx.FileID] = newFile
				}

				// return origin file to avoid updating the tree
				return file, nil
			}

			return spec.Op(file)
		})
		if err != nil {
			return err
		}
	}

	if root != m.Root {
		m.Root = root
	}

	return nil
}

func (m *MemFS) Apply(path []string, ctx OperationCtx, op Operation) error {
	return m.ApplyAll(
		ApplySpec{
			Path: path,
			Ctx:  ctx,
			Op:   op,
		},
	)
}

func (m *MemFS) MakeDir(path string) error {
	parts, err := SplitPath(path)
	if err != nil {
		return err
	}
	return m.Apply(
		parts,
		OperationCtx{},
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
			OperationCtx{},
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
		OperationCtx{},
		func(file *File) (*File, error) {
			if file.IsDir && len(file.Entries) > 0 && !spec.All {
				return nil, we.With(
					e4.Info("path: %s", path),
				)(ErrDirNotEmpty)
			}
			err := pp.Copy(
				file.IterAllFiles(nil),
				pp.Tap(func(v any) error {
					file := v.(*File)
					if m.openedIDs[file.id] > 0 {
						// add to detached files
						if _, ok := m.detached[file.id]; !ok {
							m.detached[file.id] = file
						}
					}
					return nil
				}),
			)
			if err != nil {
				return nil, err
			}
			return nil, nil
		},
	)
}
