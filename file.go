package fs9

import (
	"io"
	"io/fs"
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
	//TODO

	return
}

func (f File) Dump(w io.Writer, level int) {
	//TODO
}

func (f File) Walk(cont Src) Src {
	//TODO
	return nil
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
