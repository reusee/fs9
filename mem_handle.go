package fs9

import (
	"io"
	"io/fs"
)

type MemHandle struct {
}

func (m *MemHandle) Close() error {
	//TODO
	return nil
}

func (m *MemHandle) Read(buf []byte) (n int, err error) {
	//TODO
	return 0, io.EOF
}

func (m *MemHandle) Seek(offset int64, whence int) (int64, error) {
	//TODO
	return 0, nil
}

func (m *MemHandle) Stat() (fs.FileInfo, error) {
	//TODO
	return nil, nil
}

func (m *MemHandle) Write(data []byte) (int, error) {
	//TODO
	return 0, nil
}
