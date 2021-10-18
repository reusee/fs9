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

	ce(fs.Apply([]string{"foo"}, Ensure("foo", false, true)))
	eq(
		len(fs.Root.Entries), 1,
		fs.Root.Entries[0].File != nil, true,
		fs.Root.Entries[0].File.Name, "foo",
		fs.Root.Entries[0].File.IsDir, false,
	)

	ce(fs.Apply([]string{"foo"}, Ensure("foo", false, true)))
	eq(
		len(fs.Root.Entries), 1,
		fs.Root.Entries[0].File != nil, true,
		fs.Root.Entries[0].File.Name, "foo",
		fs.Root.Entries[0].File.IsDir, false,
	)

	ce(fs.Apply([]string{"a"}, Ensure("a", false, true)))
	eq(
		len(fs.Root.Entries), 2,
		fs.Root.Entries[0].File != nil, true,
		fs.Root.Entries[0].File.Name, "a",
		fs.Root.Entries[0].File.IsDir, false,
	)

	ce(fs.Apply([]string{"b"}, Ensure("b", true, true)))
	eq(
		len(fs.Root.Entries), 3,
		fs.Root.Entries[1].File != nil, true,
		fs.Root.Entries[1].File.Name, "b",
		fs.Root.Entries[1].File.IsDir, true,
	)

	err := fs.Apply([]string{"a"}, func(file *File) (*File, error) {
		return nil, fmt.Errorf("foo")
	})
	eq(
		len(fs.Root.Entries), 3,
		err != nil, true,
		err.Error(), "foo",
	)

	err = fs.Apply([]string{"a"}, func(file *File) (*File, error) {
		newFile := *file
		newFile.Name = "foo"
		return &newFile, nil
	})
	eq(
		len(fs.Root.Entries), 3,
		err != nil, true,
		is(err, ErrNameMismatch), true,
	)

	err = fs.Apply([]string{"a"}, func(file *File) (*File, error) {
		newFile := *file
		newFile.IsDir = true
		return &newFile, nil
	})
	eq(
		len(fs.Root.Entries), 3,
		err != nil, true,
		is(err, ErrTypeMismatch), true,
	)

	err = fs.Apply([]string{"a", "foo"}, func(file *File) (*File, error) {
		return file, nil
	})
	eq(
		len(fs.Root.Entries), 3,
		err != nil, true,
		is(err, ErrInvalidPath), true,
	)

	var info os.FileInfo
	ce(fs.Apply([]string{"a"}, func(file *File) (*File, error) {
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
