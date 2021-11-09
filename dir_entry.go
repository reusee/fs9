package fs9

import (
	"fmt"
	"io"
	"io/fs"
	"strings"

	"github.com/reusee/it"
)

type DirEntry struct {
	nodeID int64
	id     FileID
	name   string
	isDir  bool
	_type  fs.FileMode
	fs     *MemFS
}

var _ fs.DirEntry = DirEntry{}

var _ Node = DirEntry{}

func (d DirEntry) NodeID() int64 {
	return d.nodeID
}

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
	info, err := d.fs.stat(d.name, d.id)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func (d DirEntry) KeyRange() (Key, Key) {
	return d.name, d.name
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

func (d DirEntry) Merge(ctx Scope, node2 Node) (Node, error) {
	entry2, ok := node2.(DirEntry)
	if !ok {
		panic(fmt.Errorf("bad merge type %T", node2))
	}
	if entry2.nodeID == d.nodeID {
		// not changed
		return d, nil
	}
	if entry2.name != d.name {
		panic(fmt.Errorf("cannot merge different name"))
	}
	// new
	newNode := entry2
	newNode.nodeID = it.NewNodeID()
	return newNode, nil
}
