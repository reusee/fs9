package fs9

import (
	"fmt"
	"io"
	"io/fs"
	"sync"
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
	return m.fs.stat(m.id)
}

func (m *MemHandle) Read(buf []byte) (n int, err error) {
	m.Lock()
	defer m.Unlock()
	if m.closed {
		return 0, ErrClosed
	}
	file, err := m.fs.GetFileByID(m.id)
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
	file, err := m.fs.GetFileByID(m.id)
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

func (m *MemHandle) Seek(offset int64, whence int) (int64, error) {
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
		file, err := m.fs.GetFileByID(m.id)
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
	file, err := m.fs.GetFileByID(m.id)
	if err != nil {
		return 0, err
	}
	var newFile *File
	newFile, n, err = file.WriteAt(data, m.offset)
	if err != nil {
		return 0, err
	}
	m.offset += int64(n)
	if err := m.fs.updateFile(newFile); err != nil {
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
		file, err := m.fs.GetFileByID(m.id)
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

func (h *MemHandle) ChangeMode(mode fs.FileMode) error {
	return h.fs.changeFileByID(h.id, fileChangeMode(mode))
}

func (h *MemHandle) ChangeOwner(uid, gid int) error {
	return h.fs.changeFileByID(h.id, fileChagneOwner(uid, gid))
}

func (h *MemHandle) Sync() error {
	//TODO materialize
	return nil
}

func (h *MemHandle) Truncate(size int64) error {
	return h.fs.changeFileByID(h.id, fileTruncate(size))
}
