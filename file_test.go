package fs9

import (
	"fmt"
	"strings"
	"testing"

	"github.com/reusee/e4"
)

func TestFilePersistence(t *testing.T) {
	defer he(nil, e4.TestingFatal(t))

	root := NewFile("root", true)
	var files []*File
	for i := 0; i < 1024; i++ {
		root, err := root.Apply([]string{"foo"}, Ensure("foo", true, true))
		ce(err)
		root, err = root.Apply([]string{"foo", "bar"}, Ensure("bar", true, true))
		ce(err)
		file, err := root.Apply([]string{"foo", "bar", "baz"},
			Ensure("baz", false, true))
		ce(err)
		file, err = file.Apply([]string{"foo", "bar", "baz"},
			Write(0, strings.NewReader(fmt.Sprintf("%d", i)), nil))
		ce(err)
		files = append(files, file)
		root = file
	}

	for i, file := range files {
		buf := new(strings.Builder)
		_, err := file.Apply([]string{"foo", "bar", "baz"},
			Read(0, 20, buf, nil, nil))
		ce(err)
		if buf.String() != fmt.Sprintf("%d", i) {
			t.Fatal()
		}
	}

}
