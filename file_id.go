package fs9

import (
	"fmt"
	"io"
	"strings"
)

type FileID uint64

type NamedFileID struct {
	Name string
	ID   FileID
}

var _ Node = new(NamedFileID)

func (n NamedFileID) KeyRange() (Key, Key) {
	return n.Name, n.Name
}

func (n NamedFileID) Mutate(ctx Scope, path KeyPath, fn func(Node) (Node, error)) (Node, error) {
	if len(path) == 0 {
		return fn(n)
	}
	var get GetFileByID
	ctx.Assign(&get)
	file, err := get(path[0].(FileID))
	if err != nil {
		return nil, err
	}
	return file.Mutate(ctx, path, fn)
}

func (n NamedFileID) Dump(w io.Writer, level int) {
	fmt.Fprintf(w, "%snamed file: %s %d", strings.Repeat(" ", level), n.Name, n.ID)
	//TODO dump *File of the ID
}

func (n NamedFileID) Walk(cont Src) Src {
	return func() (any, Src, error) {
		return n, cont, nil
	}
}
