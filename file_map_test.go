package fs9

import (
	"strings"
	"testing"

	"github.com/reusee/dscope"
	"github.com/reusee/e4"
)

func TestFileMap(t *testing.T) {
	defer he(nil, e4.TestingFatal(t))
	ctx := dscope.New()
	m := NewFileMap(3, 0)

	// new path
	id := FileID(0xdeadbeef1234)
	newNode, err := m.Mutate(ctx, m.GetPath(id), func(node Node) (Node, error) {
		eq(node == nil, true)
		return &File{
			ID:   id,
			Name: "foo",
		}, nil
	})
	ce(err)
	m = newNode.(*FileMap)

	eq(
		len(m.subs.Nodes), 1,
		len(m.subs.Nodes[0].(*FileMap).subs.Nodes), 1,
		len(m.subs.Nodes[0].(*FileMap).subs.Nodes[0].(*FileMap).subs.Nodes), 1,
		len(m.subs.Nodes[0].(*FileMap).subs.Nodes[0].(*FileMap).subs.Nodes[0].(*FileMap).subs.Nodes), 1,
		m.subs.Nodes[0].(*FileMap).subs.Nodes[0].(*FileMap).subs.Nodes[0].(*FileMap).subs.Nodes[0].(*File).Name, "foo",
	)

	buf := new(strings.Builder)
	m.Dump(buf, 0)
	//pt("%s\n", buf.String())
	eq(buf.String(), `file map 0
 file map 34
  file map 12
   file map ef
    file: {ID:244837814047284 IsDir:false Name:foo Size:0 Mode:---------- ModTime:0001-01-01 00:00:00 +0000 UTC Subs:<nil> Content:[]}
`)

	// tap
	var file *File
	newNode, err = m.Mutate(ctx, m.GetPath(id), func(node Node) (Node, error) {
		file = node.(*File)
		return node, nil
	})
	ce(err)
	m2 := newNode.(*FileMap)
	eq(
		m2 == m, true,
		file != nil, true,
		file.Name, "foo",
	)

	// delete
	newNode, err = m.Mutate(ctx, m.GetPath(id), func(node Node) (Node, error) {
		return nil, nil
	})
	ce(err)
	m = newNode.(*FileMap)
	eq(
		len(m.subs.Nodes), 1,
		len(m.subs.Nodes[0].(*FileMap).subs.Nodes), 1,
		len(m.subs.Nodes[0].(*FileMap).subs.Nodes[0].(*FileMap).subs.Nodes), 1,
		len(m.subs.Nodes[0].(*FileMap).subs.Nodes[0].(*FileMap).subs.Nodes[0].(*FileMap).subs.Nodes), 0,
	)

}
