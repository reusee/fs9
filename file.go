package fs9

import (
	"fmt"
	"io"
	"io/fs"
	"strings"
	"time"

	"github.com/reusee/it"
)

type File struct {
	nodeID     int64
	ID         FileID
	IsDir      bool
	Size       int64
	Mode       fs.FileMode
	ModTime    time.Time
	Subs       *NodeSet // name -> NamedFileID
	Symlink    string
	Content    []byte
	UserID     int
	GroupID    int
	AccessTime time.Time
}

type FileID uint64

func (f FileID) Cmp(i any) int {
	b := i.(FileID)
	if f < b {
		return -1
	} else if f > b {
		return 1
	}
	return 0
}

func NewFile(isDir bool) *File {
	var mode fs.FileMode
	if isDir {
		mode = fs.ModeDir
	}
	f := &File{
		nodeID:  it.NewNodeID(),
		ID:      FileID(it.NewNodeID()),
		IsDir:   isDir,
		Mode:    mode,
		ModTime: time.Now(),
	}
	if isDir {
		f.Subs = it.NewNodeSet(nil)
	}
	return f
}

func (f *File) Clone() *File {
	newFile := *f
	now := time.Now()
	if now.Equal(newFile.ModTime) {
		now = now.Add(time.Nanosecond)
	}
	newFile.ModTime = now
	newFile.nodeID = it.NewNodeID()
	return &newFile
}

var _ Node = new(File)

func (f File) Equal(n2 Node) bool {
	switch n2 := n2.(type) {
	case *File:
		return n2.nodeID == f.nodeID
	}
	panic("type mismatch")
}

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
	if !newNode.Equal(f.Subs) {
		newFile := f.Clone()
		newFile.Subs = newNode.(*NodeSet)
		return newFile, nil
	}
	return f, nil
}

func (f File) Dump(w io.Writer, level int) {
	fmt.Fprintf(w, "%sfile: %+v\n", strings.Repeat(" ", level), f)
	if f.Subs != nil {
		f.Subs.Dump(w, level+1)
	}
}

func (f File) Stat() (FileInfo, error) {
	return FileInfo{
		size:    f.Size,
		mode:    f.Mode,
		modTime: f.ModTime,
		isDir:   f.IsDir,
		ext: ExtFileInfo{
			UserID:     f.UserID,
			GroupID:    f.GroupID,
			AccessTime: f.AccessTime,
		},
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
	newFile := f.Clone()
	if l := int(offset) + len(data); l > len(f.Content) {
		newFile.Content = make([]byte, l)
	} else {
		newFile.Content = make([]byte, len(f.Content))
	}
	copy(newFile.Content, f.Content)
	copy(newFile.Content[offset:], data)
	newFile.Size = int64(len(newFile.Content))
	return newFile, len(data), nil
}

func (f *File) Merge(ctx Scope, node2 Node) (Node, error) {
	file2, ok := node2.(*File)
	if !ok {
		panic(fmt.Errorf("bad merge type: %T", node2))
	}
	if node2.Equal(f) {
		// not changed
		return f, nil
	}
	if file2.ID != f.ID {
		panic(fmt.Errorf("cannot mrege different id"))
	}

	// new
	newFile := *file2
	newFile.nodeID = it.NewNodeID()
	newSubsNode, err := f.Subs.Merge(ctx, file2.Subs)
	if err != nil {
		return nil, err
	}
	if newSubsNode != nil {
		newFile.Subs = newSubsNode.(*NodeSet)
	}

	return &newFile, nil
}
