package fs9

type Tree interface {
	Mutate(mutation Mutation) (Tree, error)
	Open(path string) (Handle, error)
}
