package fs9

import (
	"fmt"
	"strings"
	"testing"

	"github.com/reusee/e4"
)

func TestFilePersistence(t *testing.T) {
	defer he(nil, e4.TestingFatal(t))

	version := int64(1)
	root := NewFile("root", true)
	var files []*File
	for i := 0; i < 1024; i++ {
		root, err := root.Apply(version, []string{"foo"}, Ensure("foo", true, true))
		ce(err)
		version++
		root, err = root.Apply(version, []string{"foo", "bar"}, Ensure("bar", true, true))
		ce(err)
		version++
		file, err := root.Apply(version, []string{"foo", "bar", "baz"},
			Ensure("baz", false, true))
		ce(err)
		version++
		file, err = file.Apply(version, []string{"foo", "bar", "baz"},
			Write(0, strings.NewReader(fmt.Sprintf("%d", i)), nil))
		ce(err)
		version++
		files = append(files, file)
		root = file
	}

	for i, file := range files {
		buf := new(strings.Builder)
		_, err := file.Apply(version, []string{"foo", "bar", "baz"},
			Read(0, 20, buf, nil, nil))
		ce(err)
		if buf.String() != fmt.Sprintf("%d", i) {
			t.Fatal()
		}
	}

}
