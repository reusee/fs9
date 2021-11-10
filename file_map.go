package fs9

import (
	"fmt"
	"io"
	"strings"

	"github.com/reusee/it"
)

// FileMap is sharded map of FileID to *File
type FileMap struct {
	nodeID   int64
	subs     *NodeSet
	level    int
	shardKey uint8
}

var _ Node = new(FileMap)

func NewFileMap(level int, shardKey uint8) *FileMap {
	return &FileMap{
		nodeID:   it.NewNodeID(),
		subs:     it.NewNodeSet(nil),
		level:    level,
		shardKey: shardKey,
	}
}

func (f *FileMap) Equal(n2 Node) bool {
	switch n2 := n2.(type) {
	case *FileMap:
		return n2.nodeID == f.nodeID
	}
	panic("type mismatch")
}

func (f *FileMap) Clone() *FileMap {
	newMap := *f
	newMap.nodeID = it.NewNodeID()
	return &newMap
}

func (f FileMap) GetPath(id FileID) (path KeyPath) {
	u := id
	for i := 0; i < f.level; i++ {
		path = append(path, uint8(u&0xff))
		u = u >> 8
	}
	path = append(path, id)
	return
}

func (f FileMap) KeyRange() (Key, Key) {
	return f.shardKey, f.shardKey
}

func (f *FileMap) Mutate(
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

	if len(path) == 1 {
		// subs is *File, do not auto create
		newNode, err := f.subs.Mutate(ctx, path, fn)
		if err != nil { // NOCOVER
			return nil, we(err)
		}
		if !newNode.Equal(f.subs) {
			newMap := f.Clone()
			newMap.subs = newNode.(*NodeSet)
			return newMap, nil
		}
		return f, nil
	}

	// subs is *FileMap
	key := path[0].(uint8)
	newNode, err := f.subs.Mutate(ctx, KeyPath{key}, func(node Node) (Node, error) {
		// ensure shard
		if node == nil {
			node = NewFileMap(f.level, key)
		}
		return node, nil
	})
	if err != nil { // NOCOVER
		return nil, we(err)
	}

	subs := newNode.(*NodeSet)
	newNode, err = subs.Mutate(ctx, path, fn)
	if err != nil { // NOCOVER
		return nil, we(err)
	}
	if !newNode.Equal(f.subs) {
		newMap := f.Clone()
		newMap.subs = newNode.(*NodeSet)
		return newMap, nil
	}

	return f, nil
}

func (f FileMap) Dump(w io.Writer, level int) {
	fmt.Fprintf(w, "%sfile map %x\n", strings.Repeat(" ", level), f.shardKey)
	f.subs.Dump(w, level+1)
}

func (f *FileMap) Merge(ctx Scope, node2 Node) (Node, error) {
	map2, ok := node2.(*FileMap)
	if !ok {
		panic(fmt.Errorf("bad merge type %T", node2))
	}
	if map2.Equal(f) {
		// not chnaged
		return f, nil
	}
	if map2.level != f.level || map2.shardKey != f.shardKey {
		panic(fmt.Errorf("cannot merge"))
	}
	// new
	newMap := NewFileMap(map2.level, map2.shardKey)
	newSubsNode, err := f.subs.Merge(ctx, map2.subs)
	if err != nil {
		return nil, err
	}
	if newSubsNode != nil {
		newMap.subs = newSubsNode.(*NodeSet)
	}
	return newMap, nil
}
