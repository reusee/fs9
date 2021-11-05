package fs9

import (
	"io/fs"
)

type MemHandle struct {
	fs *MemFS
	id FileID
}

var _ Handle = new(MemHandle)

//TODO
//var _ fs.ReadDirFile = new(MemHandle)
//var _ io.ReaderAt = new(MemHandle)

func (m MemHandle) Stat() (fs.FileInfo, error) {
	return m.fs.stat(m.id)
}

func (m *MemHandle) Read(buf []byte) (n int, err error) {
	//TODO
	return
}

func (m *MemHandle) Close() error {
	//TODO
	return nil
}

func (m *MemHandle) Seek(offset int64, whence int) (int64, error) {
	//TODO
	return 0, nil
}

func (m *MemHandle) Write(data []byte) (int, error) {
	//TODO
	return 0, nil
}
