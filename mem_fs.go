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
		func() GetFileIDByPath {
			return m.GetFileIDByPath
		},
	)
	return m
}

func (m *MemFS) Open(path string) (fs.File, error) {
	return m.OpenHandle(path)
}

func (m *MemFS) OpenHandle(path string) (*MemHandle, error) {
	parts, err := PathToSlice(path)
	if err != nil {
		return nil, err
	}
	id, err := m.GetFileIDByPath(m.rootID, parts)
	if err != nil {
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

type GetFileIDByPath func(root FileID, path []string) (FileID, error)

func (m *MemFS) GetFileIDByPath(root FileID, path []string) (FileID, error) {
	if len(path) == 0 {
		return root, nil
	}
	file, err := m.GetFileByID(root)
	if err != nil {
		return 0, err
	}
	name := path[0]
	var id FileID
	_, err = file.Subs.Mutate(m.ctx, KeyPath{name}, func(node Node) (Node, error) {
		if node == nil {
			return nil, ErrFileNotFound
		}
		id = node.(NamedFileID).ID
		return node, nil
	})
	if err != nil {
		return 0, err
	}
	return m.GetFileIDByPath(id, path[1:])
}
