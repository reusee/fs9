package fs9

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/reusee/e4"
)

func testTree(
	t *testing.T,
	tree Tree,
) {
	defer he(nil, e4.TestingFatal(t))

	// write
	h, err := tree.OpenHandle("foo", OptCreate(true))
	ce(err)
	_, err = fmt.Fprintf(h, "foo")
	ce(err)
	ce(h.Close())

	// read
	f, err := tree.Open("foo")
	ce(err)
	content, err := io.ReadAll(f)
	ce(err)
	ce(f.Close())
	if !bytes.Equal(content, []byte("foo")) {
		t.Fatal()
	}

}
