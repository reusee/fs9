package fs9

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	iofs "io/fs"
	"math/rand"
	"sort"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/reusee/e4"
)

func testFS(
	t *testing.T,
	newFS func() FS,
) {
	defer he(nil, e4.WrapStacktrace, e4.TestingFatal(t))

	t.Run("basic operations", func(t *testing.T) {
		defer he(nil, e4.WrapStacktrace, e4.TestingFatal(t))
		fs := newFS()

		dirNames := []string{"foo", "bar", "baz", "qux", "quux"}
		randomPath := func() []string {
			slice := make([]string, rand.Intn(len(dirNames)))
			for i := range slice {
				slice[i] = dirNames[rand.Intn(len(dirNames))]
			}
			slice = append(slice, fmt.Sprintf("%d", rand.Int63()))
			return slice
		}

		var allPaths []string

		for i := 0; i < 1024; i++ {
			parts := randomPath()
			name := parts[len(parts)-1]
			dir := strings.Join(parts[:len(parts)-1], "/")
			path := strings.Join(parts, "/")
			allPaths = append(allPaths, path)

			// make dir
			err := fs.MakeDirAll(dir)
			ce(err)

			for i := 1; i < len(parts)-1; i++ {
				dir := strings.Join(parts[:i], "/")
				file, err := fs.Open(dir)
				if err != nil {
					t.Fatal(err)
				}
				stat, err := file.Stat()
				if err != nil {
					t.Fatal(err)
				}
				if !stat.IsDir() {
					t.Fatal()
				}
				if stat.Name() != parts[i-1] {
					t.Fatal()
				}
				if stat.Mode()&iofs.ModeDir == 0 {
					t.Fatal()
				}
			}

			_, err = fs.Open(dir)
			if err != nil {
				t.Fatal(err)
			}

			// write
			h, err := fs.OpenHandle(path, OptCreate(true))
			ce(err)
			_, err = fmt.Fprintf(h, "foo")
			ce(err)
			ce(h.Close())

			// info
			f, err := fs.Open(path)
			ce(err)
			info, err := f.Stat()
			ce(err)
			if info.Name() != name {
				t.Fatal()
			}
			if info.Size() != 3 {
				t.Fatalf("got %d", info.Size())
			}

			// read
			content, err := io.ReadAll(f)
			ce(err)
			ce(f.Close())
			if !bytes.Equal(content, []byte("foo")) {
				t.Fatalf("got %s", content)
			}

			h, err = fs.OpenHandle(path)
			ce(err)
			defer h.Close()

			// name
			if h.Name() != path {
				t.Fatal()
			}

			// chagne mode
			ce(fs.ChangeMode(path, 0755))
			stat, err := iofs.Stat(fs, path)
			ce(err)
			eq(
				stat.Mode()&iofs.ModePerm, iofs.FileMode(0755),
			)
			ce(h.ChangeMode(0644))
			stat, err = h.Stat()
			ce(err)
			eq(
				stat.Mode()&iofs.ModePerm, iofs.FileMode(0644),
			)

			// change owner
			ce(fs.ChangeOwner(path, 42, 2))
			stat, err = iofs.Stat(fs, path)
			ce(err)
			ext := stat.Sys().(ExtFileInfo)
			eq(
				ext.UserID, 42,
				ext.GroupID, 2,
			)
			ce(h.ChangeOwner(42, 2))
			stat, err = h.Stat()
			ce(err)
			ext = stat.Sys().(ExtFileInfo)
			eq(
				ext.UserID, 42,
				ext.GroupID, 2,
			)

		}

		ce(iofs.WalkDir(fs, ".", func(path string, entry iofs.DirEntry, err error) error {
			return err
		}))

		// fstest
		ce(fstest.TestFS(fs, allPaths...))

		for _, path := range allPaths {
			ce(fs.Remove(path))
		}

		var dirPaths []string
		ce(iofs.WalkDir(fs, ".", func(path string, entry iofs.DirEntry, err error) error {
			if err != nil {
				return we.With(
					e4.Info("path %s", path),
				)(err)
			}
			if path == "." {
				return nil
			}
			if !entry.IsDir() {
				return nil
			}
			dirPaths = append(dirPaths, path)
			return nil
		}))
		sort.Slice(dirPaths, func(i, j int) bool {
			return rand.Intn(2) == 0
		})
		var deleted []string
	loop_paths:
		for _, path := range dirPaths {
			for _, d := range deleted {
				if strings.HasPrefix(path, d) {
					continue loop_paths
				}
			}
			ce(fs.Remove(path, OptAll(true)))
			deleted = append(deleted, path)
		}
	})

	t.Run("concurrent handle", func(t *testing.T) {
		defer he(nil, e4.WrapStacktrace, e4.TestingFatal(t))
		fs := newFS()

		// write
		handle1, err := fs.OpenHandle("foo", OptCreate(true))
		ce(err)
		_, err = handle1.Write([]byte("foo"))
		ce(err)

		// same file
		handle2, err := fs.OpenHandle("foo", OptCreate(true))
		ce(err)
		content, err := io.ReadAll(handle2)
		ce(err)
		eq(
			bytes.Equal(content, []byte("foo")), true,
		)

		// remove
		ce(fs.Remove("foo"))

		// read deleted
		_, err = handle1.Seek(0, 0)
		ce(err)
		content, err = io.ReadAll(handle1)
		ce(err)
		eq(
			bytes.Equal(content, []byte("foo")), true,
		)

		// new file
		handle3, err := fs.OpenHandle("foo", OptCreate(true))
		ce(err)
		_, err = handle3.Write([]byte("bar"))
		ce(err)

		// read deleted
		_, err = handle2.Seek(0, 0)
		ce(err)
		content, err = io.ReadAll(handle2)
		ce(err)
		eq(
			bytes.Equal(content, []byte("foo")), true,
		)

		// write deleted
		_, err = handle1.Seek(0, 0)
		ce(err)
		_, err = handle1.Write([]byte("FOO"))
		ce(err)
		// read written
		_, err = handle2.Seek(0, 0)
		ce(err)
		content, err = io.ReadAll(handle2)
		ce(err)
		eq(
			bytes.Equal(content, []byte("FOO")), true,
		)

		// read new file
		_, err = handle3.Seek(0, 0)
		content, err = io.ReadAll(handle3)
		ce(err)
		eq(
			bytes.Equal(content, []byte("bar")), true,
		)

		// close new file
		ce(handle3.Close())

		// read deleted
		_, err = handle2.Seek(0, 0)
		ce(err)
		content, err = io.ReadAll(handle2)
		ce(err)
		eq(
			bytes.Equal(content, []byte("FOO")), true,
		)

		// new file again
		handle3, err = fs.OpenHandle("foo", OptCreate(true))
		ce(err)
		_, err = handle3.Write([]byte("bar"))
		ce(err)

		// read deleted
		_, err = handle2.Seek(0, 0)
		ce(err)
		content, err = io.ReadAll(handle2)
		ce(err)
		eq(
			bytes.Equal(content, []byte("FOO")), true,
		)

		// close all
		ce(handle1.Close())
		ce(handle2.Close())
		_, err = io.ReadAll(handle2)
		eq(
			is(err, ErrClosed), true,
		)

	})

	t.Run("delete non-empty dir", func(t *testing.T) {
		defer he(nil, e4.WrapStacktrace, e4.TestingFatal(t))
		fs := newFS()
		ce(fs.MakeDirAll("foo/bar/baz"))
		err := fs.Remove("foo")
		eq(
			is(err, ErrDirNotEmpty), true,
		)
		ce(fs.Remove("foo", OptAll(true)))
	})

	t.Run("test parent delete", func(t *testing.T) {
		defer he(nil, e4.WrapStacktrace, e4.TestingFatal(t))
		fs := newFS()

		ce(fs.MakeDirAll("foo/bar/baz"))

		h1, err := fs.OpenHandle("foo/bar/baz/qux", OptCreate(true))
		ce(err)
		defer h1.Close()

		h2, err := fs.OpenHandle("foo/bar/qux", OptCreate(true))
		ce(err)
		defer h2.Close()

		h3, err := fs.OpenHandle("foo/qux", OptCreate(true))
		ce(err)
		defer h3.Close()

		ce(fs.Remove("foo", OptAll(true)))

		_, err = h1.Write([]byte("foo"))
		ce(err)
		_, err = h2.Write([]byte("foo"))
		ce(err)
		_, err = h3.Write([]byte("foo"))
		ce(err)

		_, err = h1.Seek(0, 0)
		content, err := io.ReadAll(h1)
		ce(err)
		eq(
			bytes.Equal(content, []byte("foo")), true,
		)

	})

	t.Run("snapshot persistence", func(t *testing.T) {
		defer he(nil, e4.WrapStacktrace, e4.TestingFatal(t))
		//TODO
	})

	t.Run("mod time", func(t *testing.T) {
		defer he(nil, e4.WrapStacktrace, e4.TestingFatal(t))
		s := newFS()
		info, err := fs.Stat(s, ".")
		ce(err)
		modTime1 := info.ModTime()
		ce(s.MakeDir("foo"))
		info, err = fs.Stat(s, ".")
		ce(err)
		modTime2 := info.ModTime()
		eq(
			modTime2.After(modTime1), true,
		)
	})

}
