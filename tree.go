package fs9

import "io/fs"

type Tree interface {
	fs.FS
	Mutate(mutation Mutation) (Tree, error)
	OpenHandle(path string) (Handle, error)
}
