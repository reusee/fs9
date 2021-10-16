package fs9

import (
	"os"
	"time"
)

type File struct {
	IsDir   bool
	Name    string
	Entries []Entry
	Size    int64
	Mode    os.FileMode
	ModTime time.Time
}
