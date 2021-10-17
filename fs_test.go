package fs9

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/reusee/e4"
)

func testFS(
	t *testing.T,
	fs FS,
) {
	defer he(nil, e4.TestingFatal(t))

	// make dir
	err := fs.MakeDir("foo")
	ce(err)

	// write
	h, err := fs.OpenHandle("foo/bar", OptCreate(true))
	ce(err)
	_, err = fmt.Fprintf(h, "foo")
	ce(err)
	ce(h.Close())

	// read
	f, err := fs.Open("foo/bar")
	ce(err)
	content, err := io.ReadAll(f)
	ce(err)
	ce(f.Close())
	if !bytes.Equal(content, []byte("foo")) {
		t.Fatal()
	}

	// fstest TODO
	//ce(fstest.TestFS(fs, "foo"))

}
