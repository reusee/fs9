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

type File struct {
	id      FileID
	IsDir   bool
	Name    string
	Entries *NodeSet
	Size    int64
	Mode    os.FileMode
	ModTime time.Time
	Bytes   []byte
}

type FileID int64

func NewFile(name string, isDir bool) *File {
	var mode fs.FileMode
	if isDir {
		mode |= fs.ModeDir
	}
	return &File{
		id:      FileID(rand.Int63()),
		IsDir:   isDir,
		Name:    name,
		ModTime: time.Now(),
		Mode:    mode,
	}
}

var _ Node = new(File)

func (f *File) NameRange() (string, string) {
	return f.Name, f.Name
}

func (f *File) Mutate(ctx Scope, path []string, fn func(Node) (Node, error)) (Node, error) {

	if len(path) == 0 {
		newNode, err := fn(f)
		if err != nil {
			return nil, we(err)
		}
		if newNode != nil {
			if err := f.checkNewFile(newNode.(*File)); err != nil {
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

func (f *File) Walk(cont Src) Src {
	return func() (any, Src, error) {
		if f.Entries == nil {
			return f, cont, nil
		}
		return f, f.Entries.Walk(cont), nil
	}
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
	if f.Entries != nil {
		f.Entries.Dump(w, level+1)
	}
}

func (f *File) wrapDump() e4.WrapFunc {
	buf := new(strings.Builder)
	f.Dump(buf, 0)
	return e4.Info("%s", buf.String())
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
