package fs9

import (
	"fmt"
	"io"
	"io/fs"
	"strings"
	"time"
)

type File struct {
	IsDir   bool
	Name    string
	Size    int64
	Mode    fs.FileMode
	ModTime time.Time
	Subs    *NodeSet // NamedFileID
}

var _ Node = new(File)

func (f File) KeyRange() (Key, Key) {
	return f.Name, f.Name
}

func (f *File) Mutate(
	ctx Scope,
	path KeyPath,
	fn func(Node) (Node, error),
) (
	retNode Node,
	err error,
) {
	if len(path) == 0 {
		return fn(f)
	}
	newNode, err := f.Subs.Mutate(ctx, path, fn)
	if err != nil {
		return nil, err
	}
	newSubs := newNode.(*NodeSet)
	if newSubs != f.Subs {
		newFile := *f
		newFile.Subs = newSubs
		return &newFile, nil
	}
	return f, nil
}

func (f File) Dump(w io.Writer, level int) {
	fmt.Fprintf(w, "%sfile: %+v", strings.Repeat(" ", level), f)
	//TODO dump subs
}

func (f *File) Walk(cont Src) Src {
	return func() (any, Src, error) {
		return f, f.Subs.Walk(cont), nil
	}
}

func (f File) Stat() (FileInfo, error) {
	return FileInfo{
		name:    f.Name,
		size:    f.Size,
		mode:    f.Mode,
		modTime: f.ModTime,
		isDir:   f.IsDir,
	}, nil
}
