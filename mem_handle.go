package fs9

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"sync"
)

type MemHandle struct {
	sync.Mutex
	FS          *MemFS
	id          FileID
	Path        KeyPath
	Offset      int64
	iterStarted bool
	iter        Src
	closed      bool
	onClose     []func()
}

var _ fs.ReadDirFile = new(MemHandle)

func (m *MemHandle) Close() error {
	m.Lock()
	defer m.Unlock()
	if m.closed {
		return nil
	}
	for _, fn := range m.onClose {
		fn()
	}
	m.closed = true
	return nil
}

func (m *MemHandle) Read(buf []byte) (n int, err error) {
	m.Lock()
	defer m.Unlock()
	if m.closed {
		return 0, ErrClosed
	}
	var eof bool
	b := new(bytes.Buffer)
	b.Grow(len(buf))
	err = m.FS.Apply(
		m.Path,
		[]any{
			&m.id,
		},
		Read(m.Offset, int64(len(buf)), b, &n, &eof),
	)
	if err != nil {
		return 0, err
	}
	copy(buf[:n], b.Bytes())
	if eof {
		err = io.EOF
	}
	m.Offset += int64(n)
	return
}

func (m *MemHandle) Seek(offset int64, whence int) (int64, error) {
	m.Lock()
	defer m.Unlock()
	if m.closed {
		return 0, ErrClosed
	}
	switch whence {
	case 0:
		m.Offset = offset
	case 1:
		m.Offset += offset
	case 2:
		if err := m.FS.Apply(
			m.Path,
			[]any{
				&m.id,
			},
			func(node Node) (Node, error) {
				file := node.(*NamedFile)
				m.Offset = file.Size + offset
				return file, nil
			},
		); err != nil {
			return m.Offset, err
		}
	default:
		return m.Offset, fmt.Errorf("bad whence")
	}
	return m.Offset, nil
}

func (m *MemHandle) Stat() (info fs.FileInfo, err error) {
	m.Lock()
	defer m.Unlock()
	if m.closed {
		err = ErrClosed
		return
	}
	if err := m.FS.Apply(
		m.Path,
		[]any{
			&m.id,
		},
		func(node Node) (Node, error) {
			file := node.(*NamedFile)
			info = file.Info()
			return file, nil
		},
	); err != nil {
		return nil, err
	}
	return
}

func (m *MemHandle) Write(data []byte) (n int, err error) {
	m.Lock()
	defer m.Unlock()
	if m.closed {
		return 0, ErrClosed
	}
	err = m.FS.Apply(
		m.Path,
		[]any{
			&m.id,
		},
		Write(m.Offset, bytes.NewReader(data), &n),
	)
	if err != nil {
		return 0, err
	}
	m.Offset += int64(n)
	return
}

func (m *MemHandle) ReadDir(n int) (ret []fs.DirEntry, err error) {
	m.Lock()
	defer m.Unlock()
	if m.closed {
		return nil, ErrClosed
	}
	if !m.iterStarted {
		err := m.FS.Apply(
			m.Path,
			[]any{
				&m.id,
			},
			func(node Node) (Node, error) {
				file := node.(*NamedFile)
				m.iter = file.Entries.Range(nil)
				m.iterStarted = true
				return file, nil
			},
		)
		if err != nil {
			return nil, err
		}
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
		ret = append(ret, v.(*NamedFile).Info())
	}
}
