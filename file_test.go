package fs9

import (
	"fmt"
	"strings"
	"testing"

	"github.com/reusee/dscope"
	"github.com/reusee/e4"
)

func TestFilePersistence(t *testing.T) {
	defer he(nil, e4.TestingFatal(t))

	var version Version
	ctx := dscope.New(&version)
	root := NewFile("root", true)
	var files []*NamedFile
	for i := 0; i < 1024; i++ {
		root, err := root.Mutate(ctx, KeyPath{"foo"}, Ensure("foo", true, true))
		ce(err)
		version++
		ctx = ctx.Fork(&version)
		root, err = root.Mutate(ctx, KeyPath{"foo", "bar"}, Ensure("bar", true, true))
		ce(err)
		version++
		ctx = ctx.Fork(&version)
		file, err := root.Mutate(
			ctx,
			KeyPath{"foo", "bar", "baz"},
			Ensure("baz", false, true))
		ce(err)
		version++
		ctx = ctx.Fork(&version)
		file, err = file.Mutate(
			ctx,
			KeyPath{"foo", "bar", "baz"},
			Write(0, strings.NewReader(fmt.Sprintf("%d", i)), nil))
		ce(err)
		version++
		ctx = ctx.Fork(&version)
		files = append(files, file.(*NamedFile))
		root = file
	}

	for i, file := range files {
		buf := new(strings.Builder)
		_, err := file.Mutate(
			ctx,
			KeyPath{"foo", "bar", "baz"},
			Read(0, 20, buf, nil, nil))
		ce(err)
		if buf.String() != fmt.Sprintf("%d", i) {
			t.Fatal()
		}
	}

}
