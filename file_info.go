package fs9

import (
	"io/fs"
	"time"
)

type FileInfo struct {
	*NamedFile
}

func (f FileInfo) Info() (fs.FileInfo, error) {
	return f, nil
}

var _ fs.DirEntry = FileInfo{}

var _ fs.FileInfo = FileInfo{}

func (f FileInfo) IsDir() bool {
	return f.NamedFile.IsDir
}

func (f FileInfo) Name() string {
	return f.NamedFile.Name
}

func (f FileInfo) ModTime() time.Time {
	return f.NamedFile.ModTime
}

func (f FileInfo) Type() fs.FileMode {
	return f.NamedFile.Mode & fs.ModeType
}

func (f FileInfo) Mode() fs.FileMode {
	return f.NamedFile.Mode
}

func (f FileInfo) Size() int64 {
	return f.NamedFile.Size
}

func (f FileInfo) Sys() any {
	return nil
}
