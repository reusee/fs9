package fs9

import (
	"io"
	"io/fs"
)

type Handle interface {
	fs.File
	fs.ReadDirFile
	io.Seeker
	io.Writer
	io.ReaderAt

	//TODO
	ChangeMode(mode fs.FileMode) error
	ChangeOwner(uid, gid int) error
	Name() string
	//Sync() error
	//Truncate(size int64) error
}
