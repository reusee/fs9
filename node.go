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

type NodeSet struct {
	Nodes   []Node
	MinName string
	MaxName string
}

func NewNodeSet(nodes []Node) *NodeSet {
	set := &NodeSet{
		Nodes: nodes,
	}
	if len(nodes) > 0 {
		r1 := nodes[0].NameRange()
		r2 := nodes[len(nodes)-1].NameRange()
		set.MaxName = r1[0]
		set.MaxName = r2[1]
	}
	return set
}

var _ Node = new(NodeSet)

func (n NodeSet) NameRange() (ret [2]string) {
	return [2]string{n.MinName, n.MaxName}
}

func (n *NodeSet) Mutate(
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

	nodes := n.Nodes

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
			newNodes := make([]Node, len(nodes), len(nodes)+1)
			copy(newNodes, nodes)
			newNodes = append(newNodes, newNode)
			return NewNodeSet(newNodes), nil
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
			newNodes := make([]Node, 0, len(nodes)+1)
			newNodes = append(newNodes, nodes[:i]...)
			newNodes = append(newNodes, newNode)
			newNodes = append(newNodes, nodes[i:]...)
			return NewNodeSet(newNodes), nil
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
			newNodes := make([]Node, 0, len(nodes)-1)
			newNodes = append(newNodes, nodes[:i]...)
			newNodes = append(newNodes, nodes[i+1:]...)
			return NewNodeSet(newNodes), nil

		} else if newNode != nodes[i] {
			// replace
			if newNode.NameRange() != nameRange {
				return nil, we(ErrInvalidName)
			}
			newNodes := make([]Node, 0, len(nodes))
			newNodes = append(newNodes, nodes[:i]...)
			newNodes = append(newNodes, newNode)
			newNodes = append(newNodes, nodes[i+1:]...)
			return NewNodeSet(newNodes), nil
		}

		// not changed
		return n, nil
	}

	panic("impossible")
}
