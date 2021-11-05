package fs9

import (
	"fmt"
	"io"
	"io/fs"
	"strings"
	"time"
)

type File struct {
	ID      FileID
	IsDir   bool
	Name    string
	Size    int64
	Mode    fs.FileMode
	ModTime time.Time
	Subs    *NodeSet // name -> NamedFileID
	Content []byte
}

type FileID uint64

var _ Node = new(File)

func (f File) KeyRange() (Key, Key) {
	return f.ID, f.ID
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
	fmt.Fprintf(w, "%sfile: %+v\n", strings.Repeat(" ", level), f)
	if f.Subs != nil {
		f.Subs.Dump(w, level+1)
	}
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

func (f File) ReadAt(buf []byte, offset int64) (n int, err error) {
	if int(offset) > len(f.Content) {
		offset = int64(len(f.Content))
	}
	end := int(offset) + len(buf)
	if end > len(f.Content) {
		end = len(f.Content)
	}
	n = copy(buf, f.Content[offset:end])
	if n < len(buf) {
		err = io.EOF
	}
	return
}

func (f *File) WriteAt(data []byte, offset int64) (*File, int, error) {
	newFile := *f
	if l := int(offset) + len(data); l > len(f.Content) {
		newFile.Content = make([]byte, l)
	} else {
		newFile.Content = make([]byte, len(f.Content))
	}
	copy(newFile.Content, f.Content)
	copy(newFile.Content[offset:], data)
	return &newFile, len(data), nil
}
