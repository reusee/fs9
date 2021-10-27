package fs9

import (
	"fmt"
	"io"
	"io/fs"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/reusee/e4"
)

type NamedFile struct {
	Name string
	File
}

type File struct {
	id      FileID
	IsDir   bool
	Entries *NodeSet
	Size    int64
	Mode    os.FileMode
	ModTime time.Time
	Bytes   []byte
}

type FileID int64

func NewFile(name string, isDir bool) *NamedFile {
	var mode fs.FileMode
	if isDir {
		mode |= fs.ModeDir
	}
	return &NamedFile{
		Name: name,
		File: File{
			id:      FileID(rand.Int63()),
			IsDir:   isDir,
			ModTime: time.Now(),
			Mode:    mode,
		},
	}
}

var _ Node = new(NamedFile)

func (f *NamedFile) KeyRange() (Key, Key) {
	return f.Name, f.Name
}

func (f *NamedFile) Mutate(ctx Scope, path KeyPath, fn func(Node) (Node, error)) (Node, error) {

	if len(path) == 0 {
		newNode, err := fn(f)
		if err != nil {
			return nil, we(err)
		}
		if newNode != nil {
			if err := f.checkNewFile(newNode.(*NamedFile)); err != nil {
				return nil, we(err)
			}
		}
		return newNode, nil
	}

	if f == nil {
		return nil, we(ErrFileNotFound)
	}

	if !f.IsDir {
		return nil, we(ErrFileNotFound)
	}

	// descend
	newNodeSet, err := f.Entries.Mutate(ctx, path, fn)
	if err != nil {
		return nil, we(err)
	}
	if newNodeSet == nil {
		// delete
		newFile := *f
		newFile.Entries = nil
		newFile.ModTime = time.Now()
		return &newFile, nil
	} else if newNodeSet.(*NodeSet) != f.Entries {
		// replace
		newFile := *f
		newFile.Entries = newNodeSet.(*NodeSet)
		newFile.ModTime = time.Now()
		return &newFile, nil
	}

	return f, nil
}

func (f *NamedFile) Walk(cont Src) Src {
	return func() (any, Src, error) {
		if f.Entries == nil {
			return f, cont, nil
		}
		return f, f.Entries.Walk(cont), nil
	}
}

func (f *NamedFile) Info() FileInfo {
	return FileInfo{
		NamedFile: f,
	}
}

func (f *NamedFile) Dump(w io.Writer, level int) {
	if w == nil {
		w = os.Stdout
	}
	fmt.Fprintf(w, "%s%s\n", strings.Repeat(".", level), f.Name)
	if f.Entries != nil {
		f.Entries.Dump(w, level+1)
	}
}

func (f *NamedFile) wrapDump() e4.WrapFunc {
	buf := new(strings.Builder)
	f.Dump(buf, 0)
	return e4.Info("%s", buf.String())
}

func (f *NamedFile) checkNewFile(newFile *NamedFile) error {
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
