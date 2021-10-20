package fs9

import "io/fs"

type FS interface {
	fs.FS
	OpenHandle(path string, options ...OpenOption) (Handle, error)
	MakeDir(path string) error
	MakeDirAll(path string) error
	Remove(path string, options ...RemoveOption) error
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
