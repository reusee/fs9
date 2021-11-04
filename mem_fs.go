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
	m := &MemFS{
		files: NewFileMap(2, 0),
	}
	m.ctx = dscope.New(
		func() GetFileByID {
			return m.GetFileByID
		},
	)
	return m
}

func (m *MemFS) Open(path string) (fs.File, error) {
	return m.OpenHandle(path)
}

func (m *MemFS) OpenHandle(path string) (*MemHandle, error) {
	rootFile, err := m.GetFileByID(m.rootID)
	if err != nil {
		return nil, err
	}
	var id FileID
	keyPath, err := SplitPath(path)
	if err != nil {
		return nil, we(err)
	}
	if _, err := rootFile.Mutate(m.ctx, keyPath, func(node Node) (Node, error) {
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
	file, err = m.GetFileByID(id)
	if err != nil {
		return
	}
	info, err = file.Stat()
	return
}

type GetFileByID func(id FileID) (*File, error)

func (m *MemFS) GetFileByID(id FileID) (*File, error) {
	var file *File
	_, err := m.files.Mutate(m.ctx, KeyPath{id}, func(node Node) (Node, error) {
		file = node.(*File)
		return node, nil
	})
	if err != nil {
		return nil, err
	}
	return file, nil
}
