package fs9

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/reusee/e4"
)

func TestMemFS(t *testing.T) {
	testFS(t, func() FS {
		return NewMemFS()
	})
}

func TestMemFSOperations(t *testing.T) {
	defer he(nil, e4.WrapStacktrace, e4.TestingFatal(t))

	fs := NewMemFS()

	ce(fs.Apply(KeyPath{"foo"}, nil, Ensure("foo", false, true)))
	file := fs.Root.Entries.Nodes[0].(*NamedFile)
	eq(
		len(fs.Root.Entries.Nodes), 1,
		file != nil, true,
		file.Name, "foo",
		file.IsDir, false,
	)

	ce(fs.Apply(KeyPath{"foo"}, nil, Ensure("foo", false, true)))
	file = fs.Root.Entries.Nodes[0].(*NamedFile)
	eq(
		len(fs.Root.Entries.Nodes), 1,
		file != nil, true,
		file.Name, "foo",
		file.IsDir, false,
	)

	ce(fs.Apply(KeyPath{"a"}, nil, Ensure("a", false, true)))
	file = fs.Root.Entries.Nodes[0].(*NamedFile)
	eq(
		len(fs.Root.Entries.Nodes), 2,
		file != nil, true,
		file.Name, "a",
		file.IsDir, false,
	)

	ce(fs.Apply(KeyPath{"b"}, nil, Ensure("b", true, true)))
	file = fs.Root.Entries.Nodes[1].(*NamedFile)
	eq(
		len(fs.Root.Entries.Nodes), 3,
		file != nil, true,
		file.Name, "b",
		file.IsDir, true,
	)

	fooErr := errors.New("foo")
	err := fs.Apply(KeyPath{"a"}, nil, func(node Node) (Node, error) {
		return nil, fooErr
	})
	eq(
		len(fs.Root.Entries.Nodes), 3,
		err != nil, true,
		is(err, fooErr), true,
	)

	err = fs.Apply(KeyPath{"a"}, nil, func(node Node) (Node, error) {
		file := node.(*NamedFile)
		newFile := *file
		newFile.Name = "foo"
		return &newFile, nil
	})
	eq(
		len(fs.Root.Entries.Nodes), 3,
		err != nil, true,
		is(err, ErrNameMismatch), true,
	)

	err = fs.Apply(KeyPath{"a"}, nil, func(node Node) (Node, error) {
		file := node.(*NamedFile)
		newFile := *file
		newFile.IsDir = true
		return &newFile, nil
	})
	eq(
		len(fs.Root.Entries.Nodes), 3,
		err != nil, true,
		is(err, ErrTypeMismatch), true,
	)

	err = fs.Apply(KeyPath{"a", "foo"}, nil, func(node Node) (Node, error) {
		file := node.(*NamedFile)
		return file, nil
	})
	eq(
		len(fs.Root.Entries.Nodes), 3,
		err != nil, true,
		is(err, ErrFileNotFound), true,
	)

	var info os.FileInfo
	ce(fs.Apply(KeyPath{"a"}, nil, func(node Node) (Node, error) {
		file := node.(*NamedFile)
		info = file.Info()
		return file, nil
	}))
	eq(
		len(fs.Root.Entries.Nodes), 3,
		info != nil, true,
		info.Name(), "a",
	)

	buf := new(strings.Builder)
	fs.Root.Dump(buf, 0)

}
