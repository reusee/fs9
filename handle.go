package fs9

import "io/fs"

type Handle interface {
	fs.File
	Seek(offset int64, whence int) (ret int64, err error)
	Write(data []byte) (n int, err error)
}
