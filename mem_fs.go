package fs9

import (
	"io/fs"
	"math/rand"
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
	rootFile := NewFile("root", true)
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

func (m *MemFS) OpenHandle(name string, options ...OpenOption) (Handle, error) {
	path, err := NameToPath(name)
	if err != nil {
		return nil, we(err)
	}

	var spec openSpec
	for _, option := range options {
		option(&spec)
	}

	id, err := m.GetFileIDByPath(m.rootID, path)
	if err != nil {

		if is(err, ErrFileNotFound) && spec.Create {
			// try create
			fileID, _, err := m.ensureFile(path, false)
			if err != nil {
				return nil, we(err)
			}
			id = fileID

		} else {
			return nil, we(err)
		}
	}

	return m.NewHandle(name, id), nil
}

func (m *MemFS) NewHandle(name string, id FileID) *MemHandle {
	return &MemHandle{
		name: name,
		fs:   m,
		id:   id,
	}
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
		id = node.(DirEntry).id
		return node, nil
	})
	if err != nil {
		return 0, we(err)
	}
	return m.GetFileIDByPath(id, path[1:])
}

func (m *MemFS) GetFileByName(name string) (*File, error) {
	path, err := NameToPath(name)
	if err != nil {
		return nil, err
	}
	id, err := m.GetFileIDByPath(m.rootID, path)
	if err != nil {
		return nil, err
	}
	file, err := m.GetFileByID(id)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (m *MemFS) MakeDir(p string) error {
	parts, err := NameToPath(p)
	if err != nil {
		return we(err)
	}
	if len(parts) == 0 {
		return nil
	}
	_, created, err := m.ensureFile(parts, true)
	if err != nil {
		return we(err)
	}
	if !created {
		return ErrFileExisted
	}
	return nil
}

func (m *MemFS) ensureFile(
	path []string,
	isDir bool,
) (
	fileID FileID,
	created bool,
	err error,
) {

	// get parent
	parentID, err := m.GetFileIDByPath(m.rootID, path[:len(path)-1])
	if err != nil {
		return 0, false, we(err)
	}
	parentFile, err := m.GetFileByID(parentID)
	if err != nil {
		return 0, false, we(err)
	}

	// ensure in parent file
	fileMap := m.files
	name := path[len(path)-1]
	newParentNode, err := parentFile.Mutate(m.ctx, KeyPath{name}, func(node Node) (Node, error) {
		if node != nil {
			// existed
			fileID = node.(DirEntry).id
			return node, nil
		}

		// add new file
		file := NewFile(name, isDir)
		fileID = file.ID
		created = true
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

		return DirEntry{
			nodeID: rand.Int63(),
			id:     file.ID,
			name:   file.Name,
			isDir:  file.IsDir,
			_type:  file.Mode & fs.ModeType,
			fs:     m,
		}, nil
	})
	if err != nil {
		return 0, false, we(err)
	}

	if newParentNode.NodeID() != parentFile.NodeID() {
		// update map
		newMapNode, err := fileMap.Mutate(m.ctx, fileMap.GetPath(parentID), func(node Node) (Node, error) {
			return newParentNode, nil
		})
		if err != nil {
			return 0, false, err
		}
		if newMapNode.NodeID() != m.files.NodeID() {
			m.files = newMapNode.(*FileMap)
		}
	}

	return
}

func (m *MemFS) updateFile(file *File) error {
	newMapNode, err := m.files.Mutate(m.ctx, m.files.GetPath(file.ID), func(node Node) (Node, error) {
		return file, nil
	})
	if err != nil {
		return err
	}
	if newMapNode.NodeID() != m.files.NodeID() {
		m.files = newMapNode.(*FileMap)
	}
	return nil
}

func (m *MemFS) MakeDirAll(p string) error {
	parts, err := NameToPath(p)
	if err != nil {
		return we(err)
	}
	if len(parts) == 0 {
		return nil
	}
	for i := 1; i < len(parts)+1; i++ {
		if _, _, err := m.ensureFile(parts[:i], true); err != nil {
			return we(err)
		}
	}
	return nil
}

func (m *MemFS) Remove(p string, options ...RemoveOption) error {

	parts, err := NameToPath(p)
	if err != nil {
		return we(err)
	}
	if len(parts) == 0 {
		return nil
	}

	var spec removeSpec
	for _, option := range options {
		option(&spec)
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

		if !spec.All {
			// check empty
			entry := node.(DirEntry)
			if entry.IsDir() {
				file, err := m.GetFileByID(entry.id)
				if err != nil {
					return nil, err
				}
				iter := file.Subs.Range(nil)
				v, err := iter.Next()
				if err != nil {
					return nil, err
				}
				if v != nil {
					// not empty
					return nil, ErrDirNotEmpty
				}
			}
		}

		return nil, nil
	})
	if err != nil {
		return we(err)
	}

	if newParentNode.NodeID() != parentFile.NodeID() {
		// update parent and map
		newMapNode, err := fileMap.Mutate(m.ctx, fileMap.GetPath(parentID), func(node Node) (Node, error) {
			return newParentNode, nil
		})
		if err != nil {
			return err
		}
		if newMapNode.NodeID() != m.files.NodeID() {
			m.files = newMapNode.(*FileMap)
		}
	}

	return nil
}

func (m *MemFS) changeFile(name string, fn func(*File) error) error {
	file, err := m.GetFileByName(name)
	if err != nil {
		return err
	}
	if file.Mode&fs.ModeSymlink > 0 {
		return m.changeFile(file.Symlink, fn)
	}
	newFile := file.Clone()
	if err := fn(newFile); err != nil {
		return err
	}
	return m.updateFile(newFile)
}

func (m *MemFS) changeFileByID(id FileID, fn func(*File) error) error {
	file, err := m.GetFileByID(id)
	if err != nil {
		return err
	}
	if file.Mode&fs.ModeSymlink > 0 {
		return m.changeFile(file.Symlink, fn)
	}
	newFile := file.Clone()
	if err := fn(newFile); err != nil {
		return err
	}
	return m.updateFile(newFile)
}

func (m *MemFS) ChangeMode(name string, mode fs.FileMode) error {
	return m.changeFile(name, fileChangeMode(mode))
}

func (m *MemFS) ChangeOwner(name string, uid, gid int) error {
	return m.changeFile(name, fileChagneOwner(uid, gid))
}

func (m *MemFS) Truncate(name string, size int64) error {
	return m.changeFile(name, fileTruncate(size))
}

func (m *MemFS) ChangeTimes(name string, atime, mtime time.Time) error {
	return m.changeFile(name, fileChangeTimes(atime, mtime))
}

func (m *MemFS) Create(name string) (Handle, error) {
	path, err := NameToPath(name)
	if err != nil {
		return nil, err
	}
	id, created, err := m.ensureFile(path, false)
	if err != nil {
		return nil, err
	}
	handle := m.NewHandle(name, id)
	if !created {
		if err := handle.Truncate(0); err != nil {
			return nil, err
		}
	} else {
		if err := handle.ChangeMode(0666); err != nil {
			return nil, err
		}
	}
	return handle, nil
}
