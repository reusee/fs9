package fs9

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/reusee/e4"
	"github.com/reusee/pp"
)

type File struct {
	IsDir   bool
	Name    string
	Entries DirEntries
	Size    int64
	Mode    os.FileMode
	ModTime time.Time
	Bytes   []byte
}

func (f *File) Apply(path []string, op Operation) (*File, error) {
	we := we.With(
		e4.NewInfo("path: %s", strings.Join(path, "/")),
	)

	if len(path) == 0 {
		return op(f)
	}

	if !f.IsDir {
		return nil, we(ErrFileNotFound)
	}

	// descend
	newEntries, err := f.Entries.Apply(path, op)
	if err != nil {
		return nil, err
	}
	if newEntries != nil {
		newFile := *f
		newFile.Entries = *newEntries
		newFile.ModTime = time.Now()
		return &newFile, nil
	}

	return f, nil
}

func (f *File) Info() FileInfo {
	return FileInfo{
		File: f,
	}
}

func (f *File) Dump(w io.Writer, level int) {
	fmt.Fprintf(w, "%s%s\n", strings.Repeat(" ", level), f.Name)
	ce(pp.Copy(
		f.Entries.IterFiles(nil),
		pp.Tap(func(v any) error {
			v.(*File).Dump(w, level+1)
			return nil
		}),
	))
}
