package fs9

import (
	"io"
	"io/fs"
)

type Handle interface {
	fs.File
	io.Seeker
	io.Writer
}
