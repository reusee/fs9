package fs9

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	iofs "io/fs"
	"math/rand"
	pathpkg "path"
	"sort"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
	"time"

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
				if stat.Name() != pathpkg.Base(dir) {
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
			ce(h.Sync())
			ce(h.Close())

			// info
			f, err := fs.Open(path)
			ce(err)
			info, err := f.Stat()
			ce(err)
			if info.Name() != pathpkg.Base(path) {
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
			stat, err := fs.Stat(path)
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

			// truncate
			ce(fs.Truncate(path, 42))
			stat, err = iofs.Stat(fs, path)
			ce(err)
			eq(
				stat.Size(), int64(42),
			)
			data, err := iofs.ReadFile(fs, path)
			ce(err)
			eq(
				len(data), 42,
			)
			ce(h.Truncate(99))
			stat, err = h.Stat()
			ce(err)
			eq(
				stat.Size(), int64(99),
			)
			_, err = h.Seek(0, 0)
			ce(err)
			data, err = io.ReadAll(h)
			ce(err)
			eq(
				len(data), 99,
			)
			ce(h.Truncate(99))
			stat, err = h.Stat()
			ce(err)
			eq(
				stat.Size(), int64(99),
			)

			// change times
			t1 := time.Now().Add(-time.Hour)
			t2 := time.Now().Add(-time.Minute)
			ce(fs.ChangeTimes(path, t1, t2))
			stat, err = iofs.Stat(fs, path)
			ce(err)
			ext = stat.Sys().(ExtFileInfo)
			eq(
				stat.ModTime(), t2,
				ext.AccessTime, t1,
			)
			ce(h.ChangeTimes(t2, t1))
			stat, err = h.Stat()
			ce(err)
			ext = stat.Sys().(ExtFileInfo)
			eq(
				stat.ModTime(), t1,
				ext.AccessTime, t2,
			)

			// bad seek
			_, err = h.Seek(42, 42)
			eq(is(err, ErrBadArgument), true)

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

	t.Run("create", func(t *testing.T) {
		defer he(nil, e4.WrapStacktrace, e4.TestingFatal(t))
		fs := newFS()
		h, err := fs.Create("foo")
		ce(err)
		_, err = h.Write([]byte("foo"))
		ce(err)
		h2, err := fs.Create("foo")
		ce(err)
		stat, err := h2.Stat()
		ce(err)
		eq(
			stat.Size(), int64(0),
		)
		stat, err = h.Stat()
		ce(err)
		eq(
			stat.Size(), int64(0),
		)
	})

	t.Run("link", func(t *testing.T) {
		defer he(nil, e4.WrapStacktrace, e4.TestingFatal(t))
		fs := newFS()
		// new file
		f, err := fs.Create("foo")
		ce(err)
		// write file
		_, err = f.Write([]byte("foo"))
		ce(err)
		ce(f.Close())
		// link to bar
		ce(fs.Link("foo", "bar"))
		// open by bar
		f, err = fs.OpenHandle("bar")
		ce(err)
		defer f.Close()
		// read bar
		content, err := io.ReadAll(f)
		ce(err)
		eq(content, []byte("foo"))
		// remove foo
		ce(fs.Remove("foo"))
		// read bar
		content, err = iofs.ReadFile(fs, "bar")
		ce(err)
		eq(content, []byte("foo"))
		// new dir
		ce(fs.MakeDir("foo"))
		err = fs.Link("foo", "qux")
		eq(is(err, ErrCannotLink), true)
	})

	t.Run("symlink", func(t *testing.T) {
		defer he(nil, e4.WrapStacktrace, e4.TestingFatal(t))
		fs := newFS()
		// create foo
		f, err := fs.Create("foo")
		ce(err)
		// write foo
		_, err = f.Write([]byte("foo"))
		ce(err)
		ce(f.Close())
		// symlink bar to foo
		ce(fs.SymLink("foo", "bar"))
		// read bar link
		link, err := fs.ReadLink("bar")
		ce(err)
		eq(link, "foo")
		// read content by bar
		content, err := iofs.ReadFile(fs, "bar")
		ce(err)
		eq(content, []byte("foo"))
		// change owner by bar
		h, err := fs.OpenHandle("bar")
		ce(err)
		defer h.Close()
		ce(h.ChangeOwner(42, 42))
		// change mode by bar
		ce(fs.ChangeMode("bar", 0777))
		// read stat by foo
		stat, err := iofs.Stat(fs, "foo")
		ce(err)
		ext := stat.Sys().(ExtFileInfo)
		eq(
			stat.Mode(), iofs.FileMode(0777),
			ext.UserID, 42,
			ext.GroupID, 42,
		)
		// change symlink
		ce(fs.ChangeMode("bar", 0777, OptNoFollow(true)))
		ce(fs.ChangeOwner("bar", 42, 99, OptNoFollow(true)))
		now := time.Now()
		ce(fs.ChangeTimes("bar", now, now, OptNoFollow(true)))
		// link stat
		stat, err = fs.LinkStat("bar")
		ce(err)
		ext = stat.Sys().(ExtFileInfo)
		eq(
			stat.Mode(), iofs.FileMode(0777),
			ext.UserID, 42,
			ext.GroupID, 99,
			stat.ModTime(), now,
			ext.AccessTime, now,
		)
		// remove bar
		ce(fs.Remove("bar"))
		// read by foo
		content, err = iofs.ReadFile(fs, "foo")
		ce(err)
		eq(content, []byte("foo"))
	})

	t.Run("rename", func(t *testing.T) {
		defer he(nil, e4.WrapStacktrace, e4.TestingFatal(t))
		fs := newFS()
		f, err := fs.Create("foo")
		ce(err)
		ce(f.Close())
		ce(fs.Rename("foo", "bar"))
		_, err = fs.Open("foo")
		eq(is(err, ErrFileNotFound), true)
		_, err = fs.Open("bar")
		ce(err)
		_, err = fs.Create("qux")
		ce(err)
		err = fs.Rename("qux", "bar")
		eq(is(err, ErrFileExisted), true)
	})

	t.Run("handle", func(t *testing.T) {
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
		ce(err)
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

	t.Run("parallel operations", func(t *testing.T) {
		defer he(nil, e4.WrapStacktrace, e4.TestingFatal(t))
		fs := newFS()

		wg := new(sync.WaitGroup)
		n := 1024
		for i := 0; i < n; i++ {
			wg.Add(1)
			i := i
			go func() {
				defer wg.Done()

				name := fmt.Sprintf("%d", i)
				h, err := fs.Create(name)
				ce(err)
				defer h.Close()

				data := []byte(fmt.Sprintf("%d", i))
				_, err = h.Write(data)
				ce(err)

				h2, err := fs.OpenHandle(name)
				ce(err)
				ce(h2.ChangeOwner(42, 24))

				stat, err := iofs.Stat(fs, name)
				ce(err)
				ext := stat.Sys().(ExtFileInfo)
				eq(
					stat.Name(), name,
					ext.UserID, 42,
					ext.GroupID, 24,
					stat.Size(), int64(len(data)),
				)

				content, err := iofs.ReadFile(fs, name)
				ce(err)
				eq(content, data)
			}()
		}
		wg.Wait()

		paths := make(map[string]bool)
		ce(iofs.WalkDir(fs, ".", func(path string, entry iofs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			paths[path] = true
			return nil
		}))
		eq(len(paths), n+1)
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
		ce(err)
		content, err := io.ReadAll(h1)
		ce(err)
		eq(
			bytes.Equal(content, []byte("foo")), true,
		)

	})

	t.Run("snapshot", func(t *testing.T) {
		defer he(nil, e4.WrapStacktrace, e4.TestingFatal(t))
		fs := newFS()
		var snapshots []FS
		for i := 0; i < 1024; i++ {
			h, err := fs.Create("foo")
			ce(err)
			_, err = h.Write([]byte(fmt.Sprintf("%d", i)))
			ce(err)
			ce(h.Close())
			snapshots = append(snapshots, fs.Snapshot())
		}
		for i, fs := range snapshots {
			content, err := iofs.ReadFile(fs, "foo")
			ce(err)
			eq(content, []byte(fmt.Sprintf("%d", i)))
		}
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

	t.Run("handle close", func(t *testing.T) {
		defer he(nil, e4.WrapStacktrace, e4.TestingFatal(t))
		fs := newFS()
		h, err := fs.Create("foo")
		ce(err)
		h.Close()
		_, err = h.Seek(0, 0)
		eq(is(err, ErrClosed), true)
		_, err = h.Stat()
		eq(is(err, ErrClosed), true)
		err = h.Sync()
		eq(is(err, ErrClosed), true)
		err = h.ChangeMode(0755)
		eq(is(err, ErrClosed), true)
		err = h.ChangeOwner(42, 42)
		eq(is(err, ErrClosed), true)
		err = h.ChangeTimes(time.Now(), time.Now())
		eq(is(err, ErrClosed), true)
		_, err = h.ReadDir(-1)
		eq(is(err, ErrClosed), true)
		_, err = h.ReadAt(nil, 0)
		eq(is(err, ErrClosed), true)
		_, err = h.Write([]byte("foo"))
		eq(is(err, ErrClosed), true)
		err = h.Truncate(0)
		eq(is(err, ErrClosed), true)
	})

	t.Run("file existed", func(t *testing.T) {
		defer he(nil, e4.WrapStacktrace, e4.TestingFatal(t))
		fs := newFS()
		_, err := fs.Create("foo")
		ce(err)
		_, err = fs.Create("yes")
		ce(err)
		err = fs.Link("yes", "foo")
		eq(is(err, ErrFileExisted), true)
		err = fs.SymLink("yes", "foo")
		eq(is(err, ErrFileExisted), true)
		err = fs.MakeDir("")
		eq(is(err, ErrFileExisted), true)
	})

	t.Run("permission", func(t *testing.T) {
		defer he(nil, e4.WrapStacktrace, e4.TestingFatal(t))
		fs := newFS()
		err := fs.Remove("")
		eq(
			is(err, ErrNoPermission), true,
			is(err, ErrCannotRemove), true,
		)
		err = fs.Remove(".")
		eq(
			is(err, ErrNoPermission), true,
			is(err, ErrCannotRemove), true,
		)
	})

	t.Run("file not found", func(t *testing.T) {
		defer he(nil, e4.WrapStacktrace, e4.TestingFatal(t))
		fs := newFS()
		_, err := fs.OpenHandle("foo/bar")
		eq(is(err, ErrFileNotFound), true)
		err = fs.Link("foo", "bar")
		eq(is(err, ErrFileNotFound), true)
		_, err = fs.ReadLink("foo")
		eq(is(err, ErrFileNotFound), true)
		err = fs.Rename("foo", "bar")
		eq(is(err, ErrFileNotFound), true)
	})

}
