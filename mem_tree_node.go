package fs9

import (
	"fmt"
	"io"
	"strings"
)

type MemTreeNode struct {
	ID FileID
}

var _ Node = MemTreeNode{}

func (n MemTreeNode) Mutate(ctx Scope, path KeyPath, fn func(Node) (Node, error)) (Node, error) {
	return n, nil
}

func (n MemTreeNode) KeyRange() (Key, Key) {
	return n.ID, n.ID
}

func (n MemTreeNode) Dump(w io.Writer, level int) {
	fmt.Fprintf(w, "%sfile: %d", strings.Repeat(" ", level), n.ID)
}

func (n MemTreeNode) Walk(cont Src) Src {
	return func() (any, Src, error) {
		return n, cont, nil
	}
}
