package fs9

import (
	"io"
	"io/fs"
	"time"
)

type Handle interface {
	fs.File
	fs.ReadDirFile
	io.Seeker
	io.Writer
	io.ReaderAt

	ChangeMode(mode fs.FileMode) error
	ChangeOwner(uid, gid int) error
	ChangeTimes(atime time.Time, mtime time.Time) error
	Name() string
	Sync() error
	Truncate(size int64) error
}
