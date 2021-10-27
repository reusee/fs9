package fs9

import (
	"io/fs"
	"sync"

	"github.com/reusee/dscope"
	"github.com/reusee/e4"
	"github.com/reusee/pp"
)

type MemFS struct {
	sync.RWMutex
	ctx     Scope
	version Version
	Root    *NamedFile
	handles map[FileID]*MemHandlesInfo
}

type MemHandlesInfo struct {
	Count    int
	Detached *NamedFile
}

type Version int64

var _ FS = new(MemFS)

func NewMemFS() *MemFS {
	return &MemFS{
		ctx: dscope.New(
			func() FileID {
				return 0
			},
		),
		Root:    NewFile("_root_", true),
		handles: make(map[FileID]*MemHandlesInfo),
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
			Path: KeyPath{},
		}, nil
	}

	var id FileID
	pathParts, err := SplitPath(path)
	if err != nil {
		return nil, err
	}
	err = m.Apply(pathParts, nil, func(node Node) (Node, error) {
		ret, err := Ensure(
			pathParts[len(pathParts)-1].(string),
			false,
			spec.Create,
		)(node)
		if err != nil {
			return nil, err
		}
		id = ret.(*NamedFile).id
		return ret, nil
	})
	if err != nil {
		return nil, err
	}
	if id == 0 {
		panic("impossible")
	}

	info, ok := m.handles[id]
	if !ok {
		info = new(MemHandlesInfo)
		m.handles[id] = info
	}

	info.Count++
	handle := &MemHandle{
		FS:   m,
		Path: pathParts,
		id:   id,
		onClose: []func(){
			func() {
				m.Lock()
				defer m.Unlock()
				info.Count--
				if info.Count == 0 {
					delete(m.handles, id)
				}
			},
		},
	}

	return handle, nil
}

type ApplySpec struct {
	Path KeyPath
	Defs []any
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
	ctx := m.ctx.Fork(&m.version)

	for _, spec := range specs {
		ctx := ctx.Fork(spec.Defs...)

		var fileID FileID
		ctx.Assign(&fileID)

		if fileID > 0 {
			if info := m.handles[fileID]; info != nil {
				// detached
				if info.Detached != nil {
					newNode, err := spec.Op(info.Detached)
					if err != nil {
						return err
					}
					newFile := newNode.(*NamedFile)
					if newFile != info.Detached {
						info.Detached = newFile
					}
					continue
				}
			}
		}

		var err error
		newNode, err := root.Mutate(ctx, spec.Path, func(node Node) (Node, error) {

			if fileID > 0 && (node == nil || node.(*NamedFile).id != fileID) {
				if info := m.handles[fileID]; info != nil {
					if info.Detached != nil {
						newNode, err := spec.Op(info.Detached)
						if err != nil {
							return nil, err
						}
						newFile := newNode.(*NamedFile)
						if newFile != info.Detached {
							info.Detached = newFile
						}

						// return origin file to avoid updating the tree
						return node, nil
					}
				}
			}

			return spec.Op(node)
		})
		if err != nil {
			return err
		}
		root = newNode.(*NamedFile)
	}

	if root != m.Root {
		m.Root = root
	}

	return nil
}

func (m *MemFS) Apply(path KeyPath, defs []any, op Operation) error {
	return m.ApplyAll(
		ApplySpec{
			Path: path,
			Defs: defs,
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
		nil,
		Ensure(
			parts[len(parts)-1].(string),
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
		name := parts[i].(string)
		if err := m.Apply(
			dir,
			nil,
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
		nil,
		func(node Node) (Node, error) {
			file := node.(*NamedFile)
			if file.IsDir && len(file.Entries.Nodes) > 0 && !spec.All {
				return nil, we.With(
					e4.Info("path: %s", path),
				)(ErrDirNotEmpty)
			}
			if err := pp.Copy(
				file.Walk(nil),
				pp.Tap(func(v any) error {
					file := v.(*NamedFile)
					if info := m.handles[file.id]; info != nil {
						// add to detached files
						if info.Detached == nil {
							info.Detached = file
						}
					}
					return nil
				}),
			); err != nil {
				return nil, err
			}
			return nil, nil
		},
	)
}
