package fs9

import (
	"io/fs"
	iofs "io/fs"
	"testing"

	"github.com/reusee/e4"
)

func TestMemFS(t *testing.T) {
	testFS(t, func() FS {
		return NewMemFS()
	})
}

func TestMemFS2(t *testing.T) {
	defer he(nil, e4.TestingFatal(t))

	s := NewMemFS()

	// root file
	root, err := s.Open(".")
	ce(err)
	defer root.Close()
	stat, err := root.Stat()
	ce(err)
	eq(
		stat.Name(), ".",
		stat.Size(), int64(0),
		stat.Mode()&fs.ModeDir, fs.ModeDir,
		!stat.ModTime().IsZero(), true,
		stat.IsDir(), true,
	)

	// new file
	foo, err := s.OpenHandle("foo", OptCreate(true))
	ce(err)
	defer foo.Close()
	stat, err = foo.Stat()
	ce(err)
	eq(
		stat.Name(), "foo",
		stat.Size(), int64(0),
		!stat.ModTime().IsZero(), true,
		stat.IsDir(), false,
	)

	// read root
	rootDir := root.(fs.ReadDirFile)
	entries, err := rootDir.ReadDir(-1)
	ce(err)
	eq(
		len(entries), 1,
		entries[0].Name(), "foo",
		entries[0].IsDir(), false,
	)
	stat, err = entries[0].Info()
	ce(err)
	eq(
		stat.Name(), "foo",
		stat.Size(), int64(0),
		!stat.ModTime().IsZero(), true,
		stat.IsDir(), false,
	)

	// make dir, existed
	err = s.MakeDir("foo")
	eq(
		is(err, ErrFileExisted), true,
	)

	// make dir
	err = s.MakeDir("bar")
	ce(err)
	stat, err = fs.Stat(s, "bar")
	ce(err)
	eq(
		stat.Name(), "bar",
		stat.Size(), int64(0),
		stat.Mode()&fs.ModeDir, fs.ModeDir,
		!stat.ModTime().IsZero(), true,
		stat.IsDir(), true,
	)

	// make all dir
	err = s.MakeDirAll("bar/baz/quux")
	ce(err)
	stat, err = fs.Stat(s, "bar/baz")
	ce(err)
	eq(
		stat.Name(), "baz",
		stat.Mode()&fs.ModeDir, fs.ModeDir,
		stat.IsDir(), true,
	)
	stat, err = fs.Stat(s, "bar/baz/quux")
	ce(err)
	eq(
		stat.Name(), "quux",
		stat.Mode()&fs.ModeDir, fs.ModeDir,
		stat.IsDir(), true,
	)

	// remove
	err = s.Remove("nonexist")
	eq(
		is(err, ErrFileNotFound), true,
	)
	err = s.Remove("foo")
	ce(err)
	_, err = fs.Stat(s, "foo")
	eq(
		is(err, ErrFileNotFound), true,
	)

}

func TestMemFSBatchMerge(t *testing.T) {
	defer he(nil, e4.TestingFatal(t))
	fs := NewMemFS()

	batch1, apply1 := fs.NewBatch()
	batch2, apply2 := fs.NewBatch()
	_, err := batch1.Create("foo")
	ce(err)
	_, err = batch2.Create("bar")
	ce(err)
	_, err = iofs.Stat(fs, "foo")
	eq(is(err, ErrFileNotFound), true)
	err = nil
	apply1(&err)
	_, err = iofs.Stat(fs, "foo")
	ce(err)
	_, err = iofs.Stat(fs, "bar")
	eq(is(err, ErrFileNotFound), true)
	err = nil
	apply2(&err)
	_, err = iofs.Stat(fs, "bar")
	ce(err)

}
