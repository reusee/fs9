package fs9

import (
	"io/fs"
	"time"

	"github.com/reusee/e4"
	"github.com/reusee/it"
)

type MemFSReadBatch struct {
	fs    *MemFS
	ctx   Scope
	root  *DirEntry
	files *FileMap
}

type MemFSWriteBatch struct {
	MemFSReadBatch
}

func (m *MemFS) NewReadBatch() (
	batch *MemFSReadBatch,
	done func(*error),
) {

	m.RLock()
	batch = &MemFSReadBatch{
		fs:    m,
		root:  m.root,
		files: m.files,
	}
	batch.ctx = m.ctx

	done = func(p *error) {
		defer m.RUnlock()
		if !batch.files.Equal(m.files) { // NOCOVER
			panic("should not mutate")
		}
	}

	return
}

func (m *MemFS) NewWriteBatch() (
	batch *MemFSWriteBatch,
	done func(*error),
) {

	m.Lock()
	batch = &MemFSWriteBatch{
		MemFSReadBatch: MemFSReadBatch{
			fs:    m,
			root:  m.root,
			files: m.files,
		},
	}
	batch.ctx = m.ctx

	done = func(p *error) {
		defer m.Unlock()
		if *p != nil {
			return
		}
		if !batch.files.Equal(m.files) {
			m.files = batch.files
		}
	}

	return
}

func (m *MemFSWriteBatch) OpenHandle(name string, options ...OpenOption) (handle Handle, err error) {
	path, err := NameToPath(name)
	if err != nil {
		return nil, we(err)
	}

	var spec openSpec
	for _, option := range options {
		option(&spec)
	}

	id, err := m.GetFileIDByPath(path, true)
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

func (m *MemFSWriteBatch) mutateDirEntry(
	path []string,
	fn func(node Node) (Node, error),
) error {

	parentID, err := m.GetFileIDByPath(path[:len(path)-1], false)
	if err != nil {
		return we(err)
	}
	parentFile, err := m.GetFileByID(parentID)
	if err != nil {
		return we(err)
	}

	name := path[len(path)-1]
	newParentNode, err := parentFile.Mutate(m.ctx, KeyPath{name}, func(node Node) (Node, error) {
		newNode, err := fn(node)
		return newNode, err
	})
	if err != nil {
		return we(err)
	}

	if !newParentNode.Equal(parentFile) {
		newMapNode, err := m.files.Mutate(m.ctx, m.files.GetPath(parentID), func(node Node) (Node, error) {
			return newParentNode, nil
		})
		if err != nil {
			return err
		}
		if !newMapNode.Equal(m.files) {
			m.files = newMapNode.(*FileMap)
		}
	}

	return nil
}

func (m *MemFSReadBatch) GetFileIDByPath(path []string, followSymlink bool) (FileID, error) {
	entry, err := m.GetDirEntryByPath(nil, path, followSymlink)
	if err != nil {
		return 0, err
	}
	return entry.id, nil
}

func (m *MemFSReadBatch) GetDirEntryByPath(parent *DirEntry, path []string, followSymlink bool) (entry *DirEntry, err error) {
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
	if followSymlink && entry._type&fs.ModeSymlink > 0 {
		file, err := m.GetFileByID(entry.id)
		if err != nil {
			return nil, err
		}
		path, err := NameToPath(file.Symlink)
		if err != nil {
			return nil, err
		}
		return m.GetDirEntryByPath(nil, path, followSymlink)
	}
	return m.GetDirEntryByPath(entry, path[1:], followSymlink)
}

func (m *MemFSReadBatch) GetFileByID(id FileID) (*File, error) {
	var file *File
	_, err := m.files.Mutate(
		m.ctx,
		m.files.GetPath(id),
		func(node Node) (Node, error) {
			if node == nil { // NOCOVER
				return nil, ErrFileNotFound
			}
			file = node.(*File)
			return node, nil
		},
	)
	if err != nil {
		return nil, we(err)
	}
	//TODO check permission
	return file, nil
}

func (m *MemFSReadBatch) GetFileByName(name string, followSymlink bool) (*File, error) {
	path, err := NameToPath(name)
	if err != nil {
		return nil, err
	}
	id, err := m.GetFileIDByPath(path, followSymlink)
	if err != nil {
		return nil, err
	}
	file, err := m.GetFileByID(id)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (m *MemFSWriteBatch) Link(oldname, newname string) error {
	oldpath, err := NameToPath(oldname)
	if err != nil {
		return err
	}
	entry, err := m.GetDirEntryByPath(nil, oldpath, true)
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
		func(node Node) (Node, error) {
			if node != nil {
				// existed
				return node, ErrFileExisted
			}
			return DirEntry{
				nodeID: it.NewNodeID(),
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

func (m *MemFSReadBatch) stat(name string, id FileID) (info FileInfo, err error) {
	var file *File
	file, err = m.GetFileByID(id)
	if err != nil {
		return
	}
	info, err = file.Stat()
	info.name = name
	return
}

func (m *MemFSWriteBatch) ensureFile(
	path []string,
	isDir bool,
) (
	fileID FileID,
	created bool,
	err error,
) {

	if err = m.mutateDirEntry(path,
		func(node Node) (Node, error) {
			if node != nil {
				// existed
				fileID = node.(DirEntry).id
				return node, nil
			}

			// add new file
			file := NewFile(isDir)
			fileID = file.ID
			created = true
			newNode, err := m.files.Mutate(m.ctx, m.files.GetPath(file.ID), func(node Node) (Node, error) {
				if node != nil { // NOCOVER
					panic("impossible")
				}
				return file, nil
			})
			if err != nil {
				return nil, we(err)
			}
			if !newNode.Equal(m.files) {
				m.files = newNode.(*FileMap)
			}

			name := path[len(path)-1]
			return DirEntry{
				nodeID: it.NewNodeID(),
				id:     file.ID,
				name:   name,
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

func (m *MemFSWriteBatch) MakeDir(p string) error {
	parts, err := NameToPath(p)
	if err != nil {
		return we(err)
	}
	if len(parts) == 0 {
		return ErrFileExisted
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

func (m *MemFSWriteBatch) MakeDirAll(p string) error {
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

func (m *MemFSWriteBatch) Remove(name string, options ...RemoveOption) error {

	path, err := NameToPath(name)
	if err != nil {
		return we(err)
	}
	if len(path) == 0 {
		return we.With(
			e4.With(ErrCannotRemove),
		)(ErrNoPermission)
	}

	var spec removeSpec
	for _, option := range options {
		option(&spec)
	}

	if err := m.mutateDirEntry(path,
		func(node Node) (Node, error) {
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

func (m *MemFSWriteBatch) changeFile(name string, followSymlink bool, fn func(*File) error) error {
	file, err := m.GetFileByName(name, followSymlink)
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

func (m *MemFSWriteBatch) changeFileByID(id FileID, followSymlink bool, fn func(*File) error) error {
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

func (m *MemFSWriteBatch) ChangeMode(name string, mode fs.FileMode, options ...ChangeOption) error {
	var spec changeSpec
	for _, fn := range options {
		fn(&spec)
	}
	return m.changeFile(name, !spec.NoFollow, fileChangeMode(mode))
}

func (m *MemFSWriteBatch) updateFile(file *File) error {
	newMapNode, err := m.files.Mutate(m.ctx, m.files.GetPath(file.ID), func(node Node) (Node, error) {
		return file, nil
	})
	if err != nil {
		return err
	}
	if !newMapNode.Equal(m.files) {
		m.files = newMapNode.(*FileMap)
	}
	return nil
}

func (m *MemFSWriteBatch) ChangeOwner(name string, uid, gid int, options ...ChangeOption) error {
	var spec changeSpec
	for _, fn := range options {
		fn(&spec)
	}
	return m.changeFile(name, !spec.NoFollow, fileChagneOwner(uid, gid))
}

func (m *MemFSWriteBatch) Truncate(name string, size int64) error {
	return m.changeFile(name, true, fileTruncate(size))
}

func (m *MemFSWriteBatch) ChangeTimes(name string, atime, mtime time.Time, options ...ChangeOption) error {
	var spec changeSpec
	for _, fn := range options {
		fn(&spec)
	}
	return m.changeFile(name, !spec.NoFollow, fileChangeTimes(atime, mtime))
}

func (m *MemFSWriteBatch) Create(name string) (Handle, error) {
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

func (m *MemFSReadBatch) NewHandle(name string, id FileID) *MemHandle {
	return &MemHandle{
		name: name,
		fs:   m.fs,
		id:   id,
	}
}

func (m *MemFSWriteBatch) SymLink(oldname, newname string) error {
	path, err := NameToPath(newname)
	if err != nil {
		return err
	}
	if err := m.mutateDirEntry(path,
		func(node Node) (Node, error) {
			if node != nil {
				// existed
				return node, ErrFileExisted
			}

			file := NewFile(false)
			file.Mode = file.Mode | fs.ModeSymlink
			file.Symlink = oldname
			newNode, err := m.files.Mutate(m.ctx, m.files.GetPath(file.ID), func(node Node) (Node, error) {
				if node != nil { // NOCOVER
					panic("impossible")
				}
				return file, nil
			})
			if err != nil {
				return nil, err
			}
			if !newNode.Equal(m.files) {
				m.files = newNode.(*FileMap)
			}

			name := path[len(path)-1]
			return DirEntry{
				nodeID: it.NewNodeID(),
				id:     file.ID,
				name:   name,
				isDir:  file.IsDir,
				_type:  file.Mode & fs.ModeType,
				fs:     m.fs,
			}, nil
		},
	); err != nil {
		return err
	}
	return nil
}

func (m *MemFSReadBatch) ReadLink(name string) (link string, err error) {
	file, err := m.GetFileByName(name, false)
	if err != nil {
		return "", err
	}
	return file.Symlink, nil
}

func (m *MemFSWriteBatch) Rename(oldname string, newname string) error {
	oldpath, err := NameToPath(oldname)
	if err != nil {
		return err
	}

	entry, err := m.GetDirEntryByPath(nil, oldpath, true)
	if err != nil {
		return err
	}

	path, err := NameToPath(oldname)
	if err != nil {
		return err
	}
	if err := m.mutateDirEntry(path,
		func(node Node) (Node, error) {
			if node == nil {
				return nil, ErrFileNotFound
			}
			return nil, nil
		},
	); err != nil {
		return err
	}

	path, err = NameToPath(newname)
	if err != nil {
		return err
	}
	if err := m.mutateDirEntry(path,
		func(node Node) (Node, error) {
			if node != nil {
				return node, ErrFileExisted
			}
			entry.name = path[len(path)-1]
			return entry, nil
		},
	); err != nil {
		return err
	}

	return nil
}
