package fs9

import (
	"io/fs"
)

type FS interface {
	fs.FS

	//TODO
	ChangeMode(name string, mode fs.FileMode) error
	ChangeOwner(name string, uid, gid int) error
	//ChangeTimes(name string, atime time.Time, mtime time.Time) error
	//Create(name string) (Handle, error)
	//Link(oldname, newname string) error
	//LinkChangeOwner(name string, uid, gid int) error
	MakeDir(path string) error
	MakeDirAll(path string) error
	OpenHandle(path string, options ...OpenOption) (Handle, error)
	//ReadLink(name string) (string, error)
	Remove(path string, options ...RemoveOption) error
	//Rename(oldpath, newpath string) error
	//SymLink(oldname, newname string) error
	//Truncate(name string, size int64) error

	//TODO snapshot
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
