package fs9

import (
	"io/fs"
	"time"
)

type FileInfo struct {
	*File
}

func (f FileInfo) Info() (fs.FileInfo, error) {
	return f, nil
}

var _ fs.DirEntry = FileInfo{}

var _ fs.FileInfo = FileInfo{}

func (f FileInfo) IsDir() bool {
	return f.File.IsDir
}

func (f FileInfo) Name() string {
	return f.File.Name
}

func (f FileInfo) ModTime() time.Time {
	return f.File.ModTime
}

func (f FileInfo) Type() fs.FileMode {
	return f.File.Mode & fs.ModeType
}

func (f FileInfo) Mode() fs.FileMode {
	return f.File.Mode
}

func (f FileInfo) Size() int64 {
	return f.File.Size
}

func (f FileInfo) Sys() any {
	return nil
}
