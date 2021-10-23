package fs9

import (
	"reflect"
	"sort"
)

type Node interface {
	NameRange() [2]string
	Mutate(
		ctx Scope,
		path []string,
		fn func(Node) (Node, error),
	) (
		retNode Node,
		err error,
	)
}

type Nodes []Node

var _ Node = new(Nodes)

func (n Nodes) NameRange() (ret [2]string) {
	if len(n) == 0 {
		return
	}
	r1 := n[0].NameRange()
	r2 := n[len(n)-1].NameRange()
	ret[0] = r1[0]
	ret[1] = r2[1]
	return
}

func (n *Nodes) Mutate(
	ctx Scope,
	path []string,
	fn func(Node) (Node, error),
) (
	retNode Node,
	err error,
) {

	if len(path) == 0 {
		return nil, we(ErrNodeNotFound)
	}

	name := path[0]
	if name == "" || name == "." || name == ".." {
		return nil, we(ErrInvalidPath)
	}

	nodes := *n

	// search
	i := sort.Search(len(nodes), func(i int) bool {
		return nodes[i].NameRange()[0] >= name
	})

	if i == len(nodes) {
		// not found
		node := reflect.New(reflect.TypeOf(fn).In(0)).Elem().Interface().(Node)
		newNode, err := node.Mutate(ctx, path[1:], fn)
		if err != nil {
			return nil, we(err)
		}
		if newNode != nil {
			// append
			newNodes := make(Nodes, len(nodes), len(nodes)+1)
			copy(newNodes, nodes)
			newNodes = append(newNodes, newNode)
			return &newNodes, nil
		}
		// not changed
		return n, nil
	}

	nameRange := nodes[i].NameRange()
	if name < nameRange[0] {
		// not found
		node := reflect.New(reflect.TypeOf(fn).In(0)).Elem().Interface().(Node)
		newNode, err := node.Mutate(ctx, path[1:], fn)
		if err != nil {
			return nil, we(err)
		}
		if newNode != nil {
			// insert
			newNodes := make(Nodes, 0, len(nodes)+1)
			newNodes = append(newNodes, nodes[:i]...)
			newNodes = append(newNodes, newNode)
			newNodes = append(newNodes, nodes[i:]...)
			return &newNodes, nil
		}
		// not changed
		return n, nil

	} else if nameRange[0] <= name && name <= nameRange[1] {
		// in range
		var newNode Node
		if nameRange[0] == name && nameRange[1] == name {
			// exactly
			newNode, err = nodes[i].Mutate(ctx, path[1:], fn)
		} else {
			// descend
			newNode, err = nodes[i].Mutate(ctx, path, fn)
		}
		if err != nil {
			return nil, we(err)
		}

		if newNode == nil {
			// delete
			newNodes := make(Nodes, 0, len(nodes)-1)
			newNodes = append(newNodes, nodes[:i]...)
			newNodes = append(newNodes, nodes[i+1:]...)
			return &newNodes, nil

		} else if newNode != nodes[i] {
			// replace
			if newNode.NameRange() != nameRange {
				return nil, we(ErrInvalidName)
			}
			newNodes := make(Nodes, 0, len(nodes))
			newNodes = append(newNodes, nodes[:i]...)
			newNodes = append(newNodes, newNode)
			newNodes = append(newNodes, nodes[i+1:]...)
			return &newNodes, nil
		}

		// not changed
		return n, nil
	}

	panic("impossible")
}
