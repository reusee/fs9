package fs9

import (
	"io"
	"os"
	"sort"
)

type Node interface {
	NameRange() (string, string)
	Mutate(
		ctx Scope,
		path []string,
		fn func(Node) (Node, error),
	) (
		retNode Node,
		err error,
	)
	Dump(w io.Writer, level int)
	Walk(cont Src) Src
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
		set.MinName, _ = nodes[0].NameRange()
		_, set.MaxName = nodes[len(nodes)-1].NameRange()
	}
	return set
}

var _ Node = new(NodeSet)

func (n NodeSet) NameRange() (string, string) {
	return n.MinName, n.MaxName
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

	var nodes []Node
	if n != nil {
		nodes = n.Nodes
	}

	// search
	i := sort.Search(len(nodes), func(i int) bool {
		min, _ := nodes[i].NameRange()
		return min >= name
	})

	if i == len(nodes) {
		// not found
		if len(path) > 1 {
			return nil, we(ErrNodeNotFound)
		}
		newNode, err := fn(nil)
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

	minName, maxName := nodes[i].NameRange()
	if name < minName {
		// not found
		if len(path) > 1 {
			return nil, we(ErrNodeNotFound)
		}
		newNode, err := fn(nil)
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

	} else if minName <= name && name <= maxName {
		// in range
		var newNode Node
		if minName == name && maxName == name {
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
			newMinName, newMaxName := newNode.NameRange()
			if newMinName != minName || newMaxName != maxName {
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

func (n NodeSet) Dump(w io.Writer, level int) {
	if w == nil {
		w = os.Stdout
	}
	for _, node := range n.Nodes {
		node.Dump(w, level)
	}
}

func (n NodeSet) Walk(cont Src) Src {
	i := 0
	var src Src
	src = func() (any, Src, error) {
		if i == len(n.Nodes) {
			return nil, cont, nil
		}
		i++
		return n.Nodes[i-1].Walk(src)()
	}
	return src
}

func (n NodeSet) Range(cont Src) Src {
	i := 0
	var src Src
	src = func() (any, Src, error) {
		if i == len(n.Nodes) {
			return nil, cont, nil
		}
		i++
		switch node := n.Nodes[i-1].(type) {
		case *NodeSet:
			return node.Walk(src)()
		default:
			return node, src, nil
		}
	}
	return src
}
