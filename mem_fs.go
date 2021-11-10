package fs9

import (
	"io/fs"
	"sync"
	"time"

	"github.com/reusee/dscope"
	"github.com/reusee/it"
)

type MemFS struct {
	sync.RWMutex
	ctx   Scope
	root  *DirEntry
	files *FileMap // FileID -> *File
}

var _ fs.FS = new(MemFS)

var _ FS = new(MemFS)

func NewMemFS() *MemFS {
	m := &MemFS{
		files: NewFileMap(2, 0),
	}

	// ctx
	m.ctx = dscope.New()

	// root file
	rootFile := NewFile(true)
	newNode, err := m.files.Mutate(m.ctx, m.files.GetPath(rootFile.ID), func(node Node) (Node, error) {
		return rootFile, nil
	})
	if err != nil {
		panic(err)
	}
	m.files = newNode.(*FileMap)
	m.root = &DirEntry{
		nodeID: it.NewNodeID(),
		id:     rootFile.ID,
		name:   ".",
		isDir:  rootFile.IsDir,
		_type:  rootFile.Mode,
		fs:     m,
	}

	return m
}

func (m *MemFS) Snapshot() FS {
	m.RLock()
	defer m.RUnlock()
	return &MemFS{
		ctx:   m.ctx,
		root:  m.root,
		files: m.files,
	}
}

func (m *MemFS) Open(path string) (fs.File, error) {
	return m.OpenHandle(path)
}

func (m *MemFS) OpenHandle(name string, options ...OpenOption) (handle Handle, err error) {
	batch, done := m.NewWriteBatch()
	defer done(&err)
	return batch.OpenHandle(name, options...)
}

func (m *MemFS) MakeDir(p string) (err error) {
	batch, done := m.NewWriteBatch()
	defer done(&err)
	return batch.MakeDir(p)
}

func (m *MemFS) MakeDirAll(p string) (err error) {
	batch, done := m.NewWriteBatch()
	defer done(&err)
	return batch.MakeDirAll(p)
}

func (m *MemFS) Remove(name string, options ...RemoveOption) (err error) {
	batch, done := m.NewWriteBatch()
	defer done(&err)
	return batch.Remove(name, options...)
}

func (m *MemFS) ChangeMode(name string, mode fs.FileMode, options ...ChangeOption) (err error) {
	batch, done := m.NewWriteBatch()
	defer done(&err)
	return batch.ChangeMode(name, mode, options...)
}

func (m *MemFS) ChangeOwner(name string, uid, gid int, options ...ChangeOption) (err error) {
	batch, done := m.NewWriteBatch()
	defer done(&err)
	return batch.ChangeOwner(name, uid, gid, options...)
}

func (m *MemFS) Truncate(name string, size int64) (err error) {
	batch, done := m.NewWriteBatch()
	defer done(&err)
	return batch.Truncate(name, size)
}

func (m *MemFS) ChangeTimes(name string, atime, mtime time.Time, options ...ChangeOption) (err error) {
	batch, done := m.NewWriteBatch()
	defer done(&err)
	return batch.ChangeTimes(name, atime, mtime, options...)
}

func (m *MemFS) Create(name string) (handle Handle, err error) {
	batch, done := m.NewWriteBatch()
	defer done(&err)
	return batch.Create(name)
}

func (m *MemFS) Link(oldname, newname string) (err error) {
	batch, done := m.NewWriteBatch()
	defer done(&err)
	return batch.Link(oldname, newname)
}

func (m *MemFS) SymLink(oldname, newname string) (err error) {
	batch, done := m.NewWriteBatch()
	defer done(&err)
	return batch.SymLink(oldname, newname)
}

func (m *MemFS) stat(name string, id FileID) (info FileInfo, err error) {
	batch, done := m.NewReadBatch()
	defer done(&err)
	return batch.stat(name, id)
}

func (m *MemFS) Stat(name string) (info fs.FileInfo, err error) {
	batch, done := m.NewReadBatch()
	defer done(&err)
	return batch.Stat(name)
}

func (m *MemFS) ReadLink(name string) (link string, err error) {
	batch, done := m.NewReadBatch()
	defer done(&err)
	return batch.ReadLink(name)
}
func (m *MemFS) Rename(oldname, newname string) (err error) {
	batch, done := m.NewWriteBatch()
	defer done(&err)
	return batch.Rename(oldname, newname)
}
