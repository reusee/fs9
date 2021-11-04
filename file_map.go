package fs9

import (
	"io"

	"github.com/reusee/it"
)

// FileMap is sharded map of FileID to *File
type FileMap struct {
	subs     *NodeSet
	level    int
	shardKey uint8
}

var _ Node = new(FileMap)

func NewFileMap(level int, shardKey uint8) *FileMap {
	return &FileMap{
		subs:     it.NewNodeSet(nil),
		level:    level,
		shardKey: shardKey,
	}
}

func (f FileMap) KeyRange() (Key, Key) {
	return f.shardKey, f.shardKey
}

func (f FileMap) Mutate(
	ctx Scope,
	path KeyPath,
	fn func(Node) (Node, error),
) (
	retNode Node,
	err error,
) {
	if len(path) == 0 {
		return nil, ErrFileNotFound
	}

	if f.level == 0 {
		// last level, subs is *File
		newNode, err := f.subs.Mutate(ctx, path, fn)
		if err != nil {
			return nil, err
		}
		newSubs := newNode.(*NodeSet)
		if newSubs != f.subs {
			return &FileMap{
				subs:     newSubs,
				level:    f.level,
				shardKey: f.shardKey,
			}, nil
		}
		return f, nil
	}

	// middle level, subs is *FileMap
	id := path[0].(FileID)
	newNode, err := f.subs.Mutate(ctx, path, func(node Node) (Node, error) {
		if node == nil {
			// always create new shard
			shardKey := uint8((id >> (8 * (f.level - 1))) & 0xff)
			node = NewFileMap(f.level-1, shardKey)
			return node.Mutate(ctx, path, fn)
		}
		return node.(*FileMap).Mutate(ctx, path, fn)
	})
	if err != nil {
		return nil, err
	}
	newSubs := newNode.(*NodeSet)

	if newSubs != f.subs {
		return &FileMap{
			subs:     newSubs,
			level:    f.level,
			shardKey: f.shardKey,
		}, nil
	}
	return f, nil
}

func (f FileMap) Dump(w io.Writer, level int) {
	//TODO
}

func (f FileMap) Walk(cont Src) Src {
	//TODO
	return nil
}
