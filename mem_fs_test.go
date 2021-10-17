package fs9

import "testing"

func TestMemFS(t *testing.T) {
	testFS(t, NewMemFS())
}
