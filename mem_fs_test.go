package fs9

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/reusee/e4"
)

func TestMemFS(t *testing.T) {
	testFS(t, NewMemFS())
}

func TestMemFSOperations(t *testing.T) {
	defer he(nil, e4.WrapStacktrace, e4.TestingFatal(t))

	fs := NewMemFS()

	ce(fs.Apply([]string{"foo"}, 0, Ensure("foo", false, true)))
	file := fs.Root.Entries[0].Latest().(*File)
	eq(
		len(fs.Root.Entries), 1,
		file != nil, true,
		file.Name, "foo",
		file.IsDir, false,
	)

	ce(fs.Apply([]string{"foo"}, 0, Ensure("foo", false, true)))
	file = fs.Root.Entries[0].Latest().(*File)
	eq(
		len(fs.Root.Entries), 1,
		file != nil, true,
		file.Name, "foo",
		file.IsDir, false,
	)

	ce(fs.Apply([]string{"a"}, 0, Ensure("a", false, true)))
	file = fs.Root.Entries[0].Latest().(*File)
	eq(
		len(fs.Root.Entries), 2,
		file != nil, true,
		file.Name, "a",
		file.IsDir, false,
	)

	ce(fs.Apply([]string{"b"}, 0, Ensure("b", true, true)))
	file = fs.Root.Entries[1].Latest().(*File)
	eq(
		len(fs.Root.Entries), 3,
		file != nil, true,
		file.Name, "b",
		file.IsDir, true,
	)

	err := fs.Apply([]string{"a"}, 0, func(file *File) (*File, error) {
		return nil, fmt.Errorf("foo")
	})
	eq(
		len(fs.Root.Entries), 3,
		err != nil, true,
		err.Error(), "foo",
	)

	err = fs.Apply([]string{"a"}, 0, func(file *File) (*File, error) {
		newFile := *file
		newFile.Name = "foo"
		return &newFile, nil
	})
	eq(
		len(fs.Root.Entries), 3,
		err != nil, true,
		is(err, ErrNameMismatch), true,
	)

	err = fs.Apply([]string{"a"}, 0, func(file *File) (*File, error) {
		newFile := *file
		newFile.IsDir = true
		return &newFile, nil
	})
	eq(
		len(fs.Root.Entries), 3,
		err != nil, true,
		is(err, ErrTypeMismatch), true,
	)

	err = fs.Apply([]string{"a", "foo"}, 0, func(file *File) (*File, error) {
		return file, nil
	})
	eq(
		len(fs.Root.Entries), 3,
		err != nil, true,
		is(err, ErrInvalidPath), true,
	)

	var info os.FileInfo
	ce(fs.Apply([]string{"a"}, 0, func(file *File) (*File, error) {
		info = file.Info()
		return file, nil
	}))
	eq(
		len(fs.Root.Entries), 3,
		info != nil, true,
		info.Name(), "a",
	)

	buf := new(strings.Builder)
	fs.Root.Dump(buf, 0)

}
