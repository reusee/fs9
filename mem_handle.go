package fs9

import (
	"bytes"
	"io"
	"io/fs"
	"sync"
)

//TODO associate file by id instead of path

type MemHandle struct {
	sync.Mutex
	FS     *MemFS
	Path   []string
	Offset int64
	iter   Src
}

var _ fs.ReadDirFile = new(MemHandle)

func (m *MemHandle) Close() error {
	//TODO
	return nil
}

func (m *MemHandle) Read(buf []byte) (n int, err error) {
	m.Lock()
	defer m.Unlock()
	var eof bool
	b := new(bytes.Buffer)
	b.Grow(len(buf))
	err = m.FS.Apply(
		m.Path,
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
	//TODO
	return 0, nil
}

func (m *MemHandle) Stat() (fs.FileInfo, error) {
	//TODO
	return nil, nil
}

func (m *MemHandle) Write(data []byte) (n int, err error) {
	m.Lock()
	defer m.Unlock()
	err = m.FS.Apply(
		m.Path,
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
	if m.iter == nil {
		err := m.FS.Apply(
			m.Path,
			func(file *File) (*File, error) {
				m.iter = file.Entries.IterFileInfos(nil)
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
		ret = append(ret, v.(fs.DirEntry))
	}
}
