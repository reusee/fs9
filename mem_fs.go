package fs9

import (
	"io/fs"
	"math/rand"
	"os"
	"time"

	"github.com/reusee/dscope"
)

type MemFS struct {
	ctx    Scope
	rootID FileID
	files  *FileMap // FileID -> *File
}

var _ fs.FS = new(MemFS)

var _ FS = new(MemFS)

type GetFileByID func(id FileID) (*File, error)

type GetFileIDByPath func(root FileID, path []string) (FileID, error)

func NewMemFS() *MemFS {
	m := &MemFS{
		files: NewFileMap(2, 0),
	}

	// ctx
	m.ctx = dscope.New(
		func() GetFileByID {
			return m.GetFileByID
		},
		func() GetFileIDByPath {
			return m.GetFileIDByPath
		},
	)

	// root file
	rootFile := &File{
		ID:      FileID(rand.Int63()),
		IsDir:   true,
		Name:    "root",
		Mode:    fs.ModeDir,
		ModTime: time.Now(),
	}
	newNode, err := m.files.Mutate(m.ctx, m.files.GetPath(rootFile.ID), func(node Node) (Node, error) {
		return rootFile, nil
	})
	if err != nil {
		panic(err)
	}
	m.files = newNode.(*FileMap)
	m.rootID = rootFile.ID

	return m
}

func (m *MemFS) Open(path string) (fs.File, error) {
	return m.OpenHandle(path)
}

func (m *MemFS) OpenHandle(path string, options ...OpenOption) (Handle, error) {
	parts, err := PathToSlice(path)
	if err != nil {
		return nil, we(err)
	}

	var spec openSpec
	for _, option := range options {
		option(&spec)
	}

	id, err := m.GetFileIDByPath(m.rootID, parts)
	if err != nil {

		if is(err, ErrFileNotFound) && spec.Create {
			// try create
			file, err := m.makeFile(parts, false)
			if err != nil {
				return nil, we(err)
			}
			id = file.ID

		} else {
			return nil, we(err)
		}
	}

	return &MemHandle{
		fs: m,
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

func (m *MemFS) GetFileByID(id FileID) (*File, error) {
	var file *File
	_, err := m.files.Mutate(
		m.ctx,
		m.files.GetPath(id),
		func(node Node) (Node, error) {
			file = node.(*File)
			return node, nil
		},
	)
	if err != nil {
		return nil, we(err)
	}
	return file, nil
}

func (m *MemFS) GetFileIDByPath(root FileID, path []string) (FileID, error) {
	if len(path) == 0 {
		return root, nil
	}
	file, err := m.GetFileByID(root)
	if err != nil {
		return 0, we(err)
	}
	name := path[0]
	var id FileID
	_, err = file.Subs.Mutate(m.ctx, KeyPath{name}, func(node Node) (Node, error) {
		if node == nil {
			return nil, we(ErrFileNotFound)
		}
		id = node.(NamedFileID).ID
		return node, nil
	})
	if err != nil {
		return 0, we(err)
	}
	return m.GetFileIDByPath(id, path[1:])
}

func (m *MemFS) MakeDir(p string) error {
	parts, err := PathToSlice(p)
	if err != nil {
		return we(err)
	}
	if len(parts) == 0 {
		return nil
	}
	_, err = m.makeFile(parts, true)
	if err != nil {
		return we(err)
	}
	return nil
}

func (m *MemFS) makeFile(
	parts []string,
	isDir bool,
) (*File, error) {

	// get parent
	parentID, err := m.GetFileIDByPath(m.rootID, parts[:len(parts)-1])
	if err != nil {
		return nil, we(err)
	}
	parentFile, err := m.GetFileByID(parentID)
	if err != nil {
		return nil, we(err)
	}

	// add to parent file
	fileMap := m.files
	name := parts[len(parts)-1]
	var file *File
	newParentNode, err := parentFile.Mutate(m.ctx, KeyPath{name}, func(node Node) (Node, error) {
		if node != nil {
			// already exists
			return node, nil
		}

		// add new file
		var mode fs.FileMode
		if isDir {
			mode |= fs.ModeDir
		}
		file = &File{
			ID:      FileID(rand.Int63()),
			IsDir:   isDir,
			Name:    name,
			Mode:    mode,
			ModTime: time.Now(),
		}
		newNode, err := fileMap.Mutate(m.ctx, m.files.GetPath(file.ID), func(node Node) (Node, error) {
			if node != nil {
				panic("impossible")
			}
			return file, nil
		})
		if err != nil {
			return nil, we(err)
		}
		fileMap = newNode.(*FileMap)

		return NamedFileID{
			Name: file.Name,
			ID:   file.ID,
		}, nil
	})
	if err != nil {
		return nil, we(err)
	}

	// update parent and map
	newMapNode, err := fileMap.Mutate(m.ctx, fileMap.GetPath(parentID), func(node Node) (Node, error) {
		return newParentNode, nil
	})
	m.files = newMapNode.(*FileMap)

	return file, nil
}

func (m *MemFS) MakeDirAll(p string) error {
	parts, err := PathToSlice(p)
	if err != nil {
		return we(err)
	}
	if len(parts) == 0 {
		return nil
	}
	for i := 1; i < len(parts); i++ {
		if _, err := m.makeFile(parts[:i], true); err != nil {
			return we(err)
		}
	}
	m.files.Dump(os.Stdout, 0)
	return nil
}

func (m *MemFS) Remove(p string, options ...RemoveOption) error {

	parts, err := PathToSlice(p)
	if err != nil {
		return we(err)
	}
	if len(parts) == 0 {
		return nil
	}

	// get parent
	parentID, err := m.GetFileIDByPath(m.rootID, parts[:len(parts)-1])
	if err != nil {
		return we(err)
	}
	parentFile, err := m.GetFileByID(parentID)
	if err != nil {
		return we(err)
	}

	// remove from parent file
	fileMap := m.files
	name := parts[len(parts)-1]
	newParentNode, err := parentFile.Mutate(m.ctx, KeyPath{name}, func(node Node) (Node, error) {
		if node == nil {
			return nil, we(ErrFileNotFound)
		}
		return nil, nil
	})
	if err != nil {
		return we(err)
	}

	// update parent and map
	newMapNode, err := fileMap.Mutate(m.ctx, fileMap.GetPath(parentID), func(node Node) (Node, error) {
		return newParentNode, nil
	})
	m.files = newMapNode.(*FileMap)

	return nil
}
