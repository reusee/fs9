package fs9

import "io/fs"

type Tree interface {
	fs.FS
	Mutate(mutation Mutation) (Tree, error)
	OpenHandle(path string, options ...OpenOption) (Handle, error)
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
