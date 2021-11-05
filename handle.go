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
}
