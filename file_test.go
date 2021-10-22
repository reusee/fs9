package fs9

import (
	"fmt"
	"strings"
	"testing"

	"github.com/reusee/e4"
)

func TestFilePersistence(t *testing.T) {
	defer he(nil, e4.TestingFatal(t))

	ctx := OperationCtx{
		Version: 1,
	}
	root := NewFile("root", true)
	var files []*File
	for i := 0; i < 1024; i++ {
		root, err := root.Apply([]string{"foo"}, ctx, Ensure("foo", true, true))
		ce(err)
		ctx.Version++
		root, err = root.Apply([]string{"foo", "bar"}, ctx, Ensure("bar", true, true))
		ce(err)
		ctx.Version++
		file, err := root.Apply(
			[]string{"foo", "bar", "baz"},
			ctx,
			Ensure("baz", false, true))
		ce(err)
		ctx.Version++
		file, err = file.Apply(
			[]string{"foo", "bar", "baz"},
			ctx,
			Write(0, strings.NewReader(fmt.Sprintf("%d", i)), nil))
		ce(err)
		ctx.Version++
		files = append(files, file)
		root = file
	}

	for i, file := range files {
		buf := new(strings.Builder)
		_, err := file.Apply(
			[]string{"foo", "bar", "baz"},
			ctx,
			Read(0, 20, buf, nil, nil))
		ce(err)
		if buf.String() != fmt.Sprintf("%d", i) {
			t.Fatal()
		}
	}

}
