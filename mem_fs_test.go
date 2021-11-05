package fs9

import "testing"

func TestMemFS(t *testing.T) {
	t.Skip() //TODO
	testFS(t, func() FS {
		return NewMemFS()
	})
}
