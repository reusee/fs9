package fs9

import (
	"io/fs"

	"github.com/reusee/dscope"
)

type MemFS struct {
	ctx    Scope
	rootID FileID
	files  *FileMap // FileID -> *File
}

var _ fs.FS = new(MemFS)

//TODO
//var _ FS = new(MemFS)

func NewMemFS() *MemFS {
	return &MemFS{
		ctx:   dscope.New(),
		files: NewFileMap(2, 0),
	}
}

func (m *MemFS) Open(path string) (fs.File, error) {
	return m.OpenHandle(path)
}

func (m *MemFS) OpenHandle(path string) (*MemHandle, error) {
	var root *File
	if _, err := m.files.Mutate(m.ctx, KeyPath{m.rootID}, func(node Node) (Node, error) {
		root = node.(*File)
		return node, nil
	}); err != nil {
		return nil, err
	}
	var id FileID
	keyPath, err := SplitPath(path)
	if err != nil {
		return nil, we(err)
	}
	if _, err := root.Mutate(m.ctx, keyPath, func(node Node) (Node, error) {
		id = node.(NamedFileID).ID
		return node, nil
	}); err != nil {
		return nil, err
	}
	return &MemHandle{
		id: id,
	}, nil
}

func (m *MemFS) stat(id FileID) (info FileInfo, err error) {
	var file *File
	_, err = m.files.Mutate(m.ctx, KeyPath{id}, func(node Node) (Node, error) {
		file = node.(*File)
		return node, nil
	})
	if err != nil {
		return
	}
	info, err = file.Stat()
	return
}
