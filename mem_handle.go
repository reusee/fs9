package fs9

import (
	"fmt"
	"io"
	"io/fs"
	"path"
	"sync"
	"time"
)

type MemHandle struct {
	sync.Mutex
	fs          *MemFS
	id          FileID
	name        string
	offset      int64
	closed      bool
	iter        Src
	iterStarted bool
	//TODO read/write permission
}

var _ Handle = new(MemHandle)

func (m *MemHandle) Name() string {
	return m.name
}

func (m *MemHandle) Stat() (fs.FileInfo, error) {
	m.Lock()
	if m.closed {
		return nil, ErrClosed
	}
	m.Unlock()
	return m.fs.stat(path.Base(m.name), m.id)
}

func (m *MemHandle) Read(buf []byte) (n int, err error) {
	m.Lock()
	defer m.Unlock()
	if m.closed {
		return 0, ErrClosed
	}
	batch, apply := m.fs.NewBatch()
	defer apply(&err)
	file, err := batch.GetFileByID(m.id)
	if err != nil {
		return 0, err
	}
	n, err = file.ReadAt(buf, m.offset)
	m.offset += int64(n)
	return n, err
}

func (m *MemHandle) ReadAt(buf []byte, offset int64) (n int, err error) {
	m.Lock()
	if m.closed {
		return 0, ErrClosed
	}
	m.Unlock()
	batch, apply := m.fs.NewBatch()
	defer apply(&err)
	file, err := batch.GetFileByID(m.id)
	if err != nil {
		return 0, err
	}
	return file.ReadAt(buf, offset)
}

func (m *MemHandle) Close() error {
	m.Lock()
	defer m.Unlock()
	m.closed = true
	return nil
}

func (m *MemHandle) Seek(offset int64, whence int) (n int64, err error) {
	m.Lock()
	defer m.Unlock()
	if m.closed {
		return 0, ErrClosed
	}
	switch whence {
	case 0:
		m.offset = offset
	case 1:
		m.offset += offset
	case 2:
		batch, apply := m.fs.NewBatch()
		defer apply(&err)
		file, err := batch.GetFileByID(m.id)
		if err != nil {
			return 0, err
		}
		m.offset = file.Size + offset
	default:
		return m.offset, fmt.Errorf("bad whence")
	}
	return m.offset, nil
}

func (m *MemHandle) Write(data []byte) (n int, err error) {
	m.Lock()
	defer m.Unlock()
	if m.closed {
		return 0, ErrClosed
	}
	batch, apply := m.fs.NewBatch()
	defer apply(&err)
	file, err := batch.GetFileByID(m.id)
	if err != nil {
		return 0, err
	}
	var newFile *File
	newFile, n, err = file.WriteAt(data, m.offset)
	if err != nil {
		return 0, err
	}
	m.offset += int64(n)
	if err := batch.updateFile(newFile); err != nil {
		return 0, err
	}
	return
}

func (m *MemHandle) ReadDir(n int) (ret []fs.DirEntry, err error) {
	m.Lock()
	defer m.Unlock()
	if m.closed {
		return nil, ErrClosed
	}
	if !m.iterStarted {
		batch, apply := m.fs.NewBatch()
		defer apply(&err)
		file, err := batch.GetFileByID(m.id)
		if err != nil {
			return nil, err
		}
		m.iter = file.Subs.Range(nil)
		m.iterStarted = true
	}
	for {
		if n > 0 && len(ret) == n {
			return
		}
		var v any
		v, err = m.iter.Next()
		if err != nil {
			return nil, err
		}
		if v == nil {
			if n > 0 {
				err = io.EOF
			}
			return
		}
		entry := v.(DirEntry)
		entry.fs = m.fs
		ret = append(ret, entry)
	}
}

func (h *MemHandle) ChangeMode(mode fs.FileMode) (err error) {
	batch, apply := h.fs.NewBatch()
	defer apply(&err)
	return batch.changeFileByID(h.id, true, fileChangeMode(mode))
}

func (h *MemHandle) ChangeOwner(uid, gid int) (err error) {
	batch, apply := h.fs.NewBatch()
	defer apply(&err)
	return batch.changeFileByID(h.id, true, fileChagneOwner(uid, gid))
}

func (h *MemHandle) Sync() error {
	//TODO materialize
	return nil
}

func (h *MemHandle) Truncate(size int64) (err error) {
	batch, apply := h.fs.NewBatch()
	defer apply(&err)
	return batch.changeFileByID(h.id, true, fileTruncate(size))
}

func (h *MemHandle) ChangeTimes(atime, mtime time.Time) (err error) {
	batch, apply := h.fs.NewBatch()
	defer apply(&err)
	return batch.changeFileByID(h.id, true, fileChangeTimes(atime, mtime))
}
