package fs9

import (
	"io"
	"testing"

	"github.com/reusee/e4"
)

func TestFileRead(t *testing.T) {
	defer he(nil, e4.TestingFatal(t))

	file := &File{
		Content: []byte("foo"),
	}

	buf := make([]byte, 1)
	n, err := file.ReadAt(buf, 0)
	ce(err)
	eq(
		n, 1,
		buf, []byte("f"),
	)

	buf = make([]byte, 1)
	n, err = file.ReadAt(buf, 9)
	eq(
		is(err, io.EOF), true,
		n, 0,
	)

	buf = make([]byte, 2)
	n, err = file.ReadAt(buf, 1)
	eq(
		err == nil, true,
		n, 2,
		buf, []byte("oo"),
	)

}

func TestFileWrite(t *testing.T) {
	defer he(nil, e4.TestingFatal(t))

	file := &File{}

	newFile, n, err := file.WriteAt([]byte("foo"), 0)
	ce(err)
	eq(
		n, 3,
		file.Content, []byte(""),
		newFile.Content, []byte("foo"),
	)

	newFile, n, err = newFile.WriteAt([]byte("bar"), 0)
	ce(err)
	eq(
		n, 3,
		newFile.Content, []byte("bar"),
	)

	newFile, n, err = newFile.WriteAt([]byte("oo"), 1)
	ce(err)
	eq(
		n, 2,
		newFile.Content, []byte("boo"),
	)

}
