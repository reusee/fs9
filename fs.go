package fs9

import (
	"io/fs"
	"time"
)

type FS interface {
	fs.FS

	ChangeMode(name string, mode fs.FileMode, options ...ChangeOption) error
	ChangeOwner(name string, uid, gid int, options ...ChangeOption) error
	ChangeTimes(name string, atime time.Time, mtime time.Time, options ...ChangeOption) error
	Create(name string) (Handle, error)
	Link(oldname, newname string) error
	MakeDir(path string) error
	MakeDirAll(path string) error
	OpenHandle(path string, options ...OpenOption) (Handle, error)
	ReadLink(name string) (string, error)
	Remove(path string, options ...RemoveOption) error
	Rename(oldpath, newpath string) error
	SymLink(oldname, newname string) error
	Truncate(name string, size int64) error
	Stat(name string) (fs.FileInfo, error)
	LinkStat(name string) (fs.FileInfo, error)

	Snapshot() FS
}

type OpenOption func(*openSpec)

type openSpec struct {
	Create bool
}

func OptCreate(b bool) OpenOption {
	return func(spec *openSpec) {
		spec.Create = b
	}
}

type RemoveOption func(*removeSpec)

type removeSpec struct {
	All bool
}

func OptAll(b bool) RemoveOption {
	return func(spec *removeSpec) {
		spec.All = b
	}
}

type changeSpec struct {
	NoFollow bool
}

type ChangeOption func(*changeSpec)

func OptNoFollow(b bool) ChangeOption {
	return func(spec *changeSpec) {
		spec.NoFollow = b
	}
}
