package fs9

import (
	"io/fs"
	"math/rand"
	"time"
)

type MemFSBatch struct {
	fs    *MemFS
	ctx   Scope
	root  *DirEntry
	files *FileMap
}

func (m *MemFS) NewBatch() (
	batch *MemFSBatch,
	apply func(*error),
) {
	batch = &MemFSBatch{
		fs:    m,
		root:  m.root,
		files: m.files,
	}
	getFileByID := GetFileByID(batch.GetFileByID)
	batch.ctx = m.ctx.Fork(
		&getFileByID,
	)
	apply = func(p *error) {
		if *p != nil {
			return
		}
		if batch.files.NodeID() != m.files.NodeID() {
			m.files = batch.files
		}
	}
	return
}

func (m *MemFSBatch) OpenHandle(name string, options ...OpenOption) (handle Handle, err error) {
	path, err := NameToPath(name)
	if err != nil {
		return nil, we(err)
	}

	var spec openSpec
	for _, option := range options {
		option(&spec)
	}

	id, err := m.GetFileIDByPath(path)
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

func (m *MemFSBatch) mutateDirEntry(
	path []string,
	fn func(fileMap **FileMap, node Node) (Node, error),
) error {

	parentID, err := m.GetFileIDByPath(path[:len(path)-1])
	if err != nil {
		return we(err)
	}
	parentFile, err := m.GetFileByID(parentID)
	if err != nil {
		return we(err)
	}

	fileMap := m.files
	name := path[len(path)-1]
	newParentNode, err := parentFile.Mutate(m.ctx, KeyPath{name}, func(node Node) (Node, error) {
		newNode, err := fn(&fileMap, node)
		return newNode, err
	})
	if err != nil {
		return we(err)
	}

	if newParentNode.NodeID() != parentFile.NodeID() {
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

func (m *MemFSBatch) GetFileIDByPath(path []string) (FileID, error) {
	entry, err := m.GetDirEntryByPath(nil, path)
	if err != nil {
		return 0, err
	}
	return entry.id, nil
}

func (m *MemFSBatch) GetDirEntryByPath(parent *DirEntry, path []string) (entry *DirEntry, err error) {
	if parent == nil {
		parent = m.root
	}
	if len(path) == 0 {
		return parent, nil
	}
	file, err := m.GetFileByID(parent.id)
	if err != nil {
		return nil, we(err)
	}
	name := path[0]
	_, err = file.Subs.Mutate(m.ctx, KeyPath{name}, func(node Node) (Node, error) {
		if node == nil {
			return nil, we(ErrFileNotFound)
		}
		e := node.(DirEntry)
		entry = &e
		return node, nil
	})
	if err != nil {
		return nil, we(err)
	}
	return m.GetDirEntryByPath(entry, path[1:])
}

func (m *MemFSBatch) GetFileByID(id FileID) (*File, error) {
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

func (m *MemFSBatch) GetFileByName(name string) (*File, error) {
	path, err := NameToPath(name)
	if err != nil {
		return nil, err
	}
	id, err := m.GetFileIDByPath(path)
	if err != nil {
		return nil, err
	}
	file, err := m.GetFileByID(id)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (m *MemFSBatch) Link(oldname, newname string) error {
	oldpath, err := NameToPath(oldname)
	if err != nil {
		return err
	}
	entry, err := m.GetDirEntryByPath(nil, oldpath)
	if err != nil {
		return err
	}
	if entry.isDir {
		return ErrCannotLink
	}

	path, err := NameToPath(newname)
	if err != nil {
		return err
	}

	if err := m.mutateDirEntry(path,
		func(fileMap **FileMap, node Node) (Node, error) {
			if node != nil {
				// existed
				return node, ErrFileExisted
			}
			return DirEntry{
				nodeID: rand.Int63(),
				id:     entry.id,
				name:   path[len(path)-1],
				isDir:  entry.isDir,
				_type:  entry._type,
				fs:     m.fs,
			}, nil
		},
	); err != nil {
		return err
	}

	return nil
}

func (m *MemFSBatch) stat(id FileID) (info FileInfo, err error) {
	var file *File
	file, err = m.GetFileByID(id)
	if err != nil {
		return
	}
	info, err = file.Stat()
	return
}

func (m *MemFSBatch) ensureFile(
	path []string,
	isDir bool,
) (
	fileID FileID,
	created bool,
	err error,
) {

	name := path[len(path)-1]
	if err = m.mutateDirEntry(path,
		func(fileMap **FileMap, node Node) (Node, error) {
			if node != nil {
				// existed
				fileID = node.(DirEntry).id
				return node, nil
			}

			// add new file
			file := NewFile(name, isDir)
			fileID = file.ID
			created = true
			newNode, err := (*fileMap).Mutate(m.ctx, m.files.GetPath(file.ID), func(node Node) (Node, error) {
				if node != nil {
					panic("impossible")
				}
				return file, nil
			})
			if err != nil {
				return nil, we(err)
			}
			*fileMap = newNode.(*FileMap)

			return DirEntry{
				nodeID: rand.Int63(),
				id:     file.ID,
				name:   file.Name,
				isDir:  file.IsDir,
				_type:  file.Mode & fs.ModeType,
				fs:     m.fs,
			}, nil
		},
	); err != nil {
		return
	}

	return
}

func (m *MemFSBatch) MakeDir(p string) error {
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

func (m *MemFSBatch) MakeDirAll(p string) error {
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

func (m *MemFSBatch) Remove(name string, options ...RemoveOption) error {

	path, err := NameToPath(name)
	if err != nil {
		return we(err)
	}
	if len(path) == 0 {
		return nil
	}

	var spec removeSpec
	for _, option := range options {
		option(&spec)
	}

	if err := m.mutateDirEntry(path,
		func(fileMap **FileMap, node Node) (Node, error) {
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
		},
	); err != nil {
		return err
	}

	return nil
}

func (m *MemFSBatch) changeFile(name string, followSymlink bool, fn func(*File) error) error {
	file, err := m.GetFileByName(name)
	if err != nil {
		return err
	}
	if followSymlink {
		if file.Mode&fs.ModeSymlink > 0 {
			return m.changeFile(file.Symlink, true, fn)
		}
	}
	newFile := file.Clone()
	if err := fn(newFile); err != nil {
		return err
	}
	return m.updateFile(newFile)
}

func (m *MemFSBatch) changeFileByID(id FileID, followSymlink bool, fn func(*File) error) error {
	file, err := m.GetFileByID(id)
	if err != nil {
		return err
	}
	if followSymlink {
		if file.Mode&fs.ModeSymlink > 0 {
			return m.changeFile(file.Symlink, true, fn)
		}
	}
	newFile := file.Clone()
	if err := fn(newFile); err != nil {
		return err
	}
	return m.updateFile(newFile)
}

func (m *MemFSBatch) ChangeMode(name string, mode fs.FileMode, options ...ChangeOption) error {
	var spec changeSpec
	for _, fn := range options {
		fn(&spec)
	}
	return m.changeFile(name, !spec.NoFollow, fileChangeMode(mode))
}

func (m *MemFSBatch) updateFile(file *File) error {
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

func (m *MemFSBatch) ChangeOwner(name string, uid, gid int, options ...ChangeOption) error {
	var spec changeSpec
	for _, fn := range options {
		fn(&spec)
	}
	return m.changeFile(name, !spec.NoFollow, fileChagneOwner(uid, gid))
}

func (m *MemFSBatch) Truncate(name string, size int64) error {
	return m.changeFile(name, true, fileTruncate(size))
}

func (m *MemFSBatch) ChangeTimes(name string, atime, mtime time.Time, options ...ChangeOption) error {
	var spec changeSpec
	for _, fn := range options {
		fn(&spec)
	}
	return m.changeFile(name, !spec.NoFollow, fileChangeTimes(atime, mtime))
}

func (m *MemFSBatch) Create(name string) (Handle, error) {
	path, err := NameToPath(name)
	if err != nil {
		return nil, err
	}
	id, created, err := m.ensureFile(path, false)
	if err != nil {
		return nil, err
	}
	if !created {
		if err := m.Truncate(name, 0); err != nil {
			return nil, err
		}
	} else {
		if err := m.ChangeMode(name, 0666); err != nil {
			return nil, err
		}
	}
	return m.NewHandle(name, id), nil
}

func (m *MemFSBatch) NewHandle(name string, id FileID) *MemHandle {
	return &MemHandle{
		name: name,
		fs:   m.fs,
		id:   id,
	}
}
