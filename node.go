package fs9

type Node interface {
	ID() int64
	Mutate(mutation Mutation) (Node, error)
}
