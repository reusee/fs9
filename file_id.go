package fs9

import "io"

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
	//TODO get file from id to descend
	return fn(n)
}

func (n NamedFileID) Dump(w io.Writer, level int) {
	//TODO
}

func (n NamedFileID) Walk(cont Src) Src {
	//TODO
	return nil
}
