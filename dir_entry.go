package fs9

import (
	"fmt"
	"io"
	"io/fs"
	"strings"
)

type DirEntry struct {
	id    FileID
	name  string
	isDir bool
	_type fs.FileMode
	fs    *MemFS
}

var _ fs.DirEntry = DirEntry{}

var _ Node = DirEntry{}

func (d DirEntry) Name() string {
	return d.name
}

func (d DirEntry) IsDir() bool {
	return d.isDir
}

func (d DirEntry) Type() fs.FileMode {
	return d._type
}

func (d DirEntry) Info() (fs.FileInfo, error) {
	return d.fs.stat(d.id)
}

func (d DirEntry) KeyRange() (Key, Key) {
	return d.Name, d.Name
}

func (d DirEntry) Mutate(ctx Scope, path KeyPath, fn func(Node) (Node, error)) (Node, error) {
	if len(path) == 0 {
		return fn(d)
	}
	var get GetFileByID
	ctx.Assign(&get)
	file, err := get(path[0].(FileID))
	if err != nil {
		return nil, err
	}
	return file.Mutate(ctx, path, fn)
}

func (d DirEntry) Dump(w io.Writer, level int) {
	fmt.Fprintf(w, "%sentry: %s %d\n", strings.Repeat(" ", level), d.name, d.id)
}

func (d DirEntry) Walk(cont Src) Src {
	return func() (any, Src, error) {
		return d, cont, nil
	}
}
