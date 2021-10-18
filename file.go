package fs9

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
	"time"

	"github.com/reusee/e4"
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

func NewFile(name string, isDir bool) *File {
	var mode fs.FileMode
	if isDir {
		mode |= fs.ModeDir
	}
	return &File{
		IsDir:   isDir,
		Name:    name,
		ModTime: time.Now(),
		Mode:    mode,
	}
}

func (f *File) Apply(path []string, op Operation) (newFile *File, err error) {
	//ce(f.verifyStructure())
	//defer func() {
	//	if newFile != nil {
	//		ce(newFile.verifyStructure())
	//	}
	//}()

	if len(path) == 0 {
		newFile, err := op(f)
		if err != nil {
			return nil, err
		}
		if err := f.checkNewFile(newFile); err != nil {
			return nil, we(err)
		}
		return newFile, nil
	}

	if f == nil {
		return nil, we(ErrInvalidPath)
	}

	if !f.IsDir {
		return nil, we(ErrInvalidPath)
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
	if w == nil {
		w = os.Stdout
	}
	fmt.Fprintf(w, "%s%s\n", strings.Repeat(".", level), f.Name)
	f.Entries.Dump(w, level+1)
}

func (f *File) wrapDump() e4.WrapFunc {
	buf := new(strings.Builder)
	f.Dump(buf, 0)
	return e4.NewInfo("%s", buf.String())
}

func (f *File) checkNewFile(newFile *File) error {
	if f == nil || newFile == nil {
		return nil
	}
	// cannot chagne IsDir and Name
	if newFile.IsDir != f.IsDir {
		return ErrTypeMismatch
	}
	if newFile.Name != f.Name {
		return ErrNameMismatch
	}
	return nil
}

func (f *File) verifyStructure() error {
	if f == nil {
		return nil
	}
	return f.Entries.verifyStructure()
}
