package fs9

import (
	"testing"

	"github.com/reusee/e4"
)

func TestMemFS(t *testing.T) {
	testFS(t, NewMemFS())
}

func TestMemFSOperations(t *testing.T) {
	defer he(nil, e4.WrapStacktrace, e4.TestingFatal(t))

	fs := NewMemFS()

	ce(fs.Apply([]string{"foo"}, Ensure("foo", false, true)))
	eq(
		len(fs.Root.Entries), 1,
		fs.Root.Entries[0].File != nil, true,
		fs.Root.Entries[0].File.Name, "foo",
		fs.Root.Entries[0].File.IsDir, false,
	)

	ce(fs.Apply([]string{"foo"}, Ensure("foo", false, true)))
	eq(
		len(fs.Root.Entries), 1,
		fs.Root.Entries[0].File != nil, true,
		fs.Root.Entries[0].File.Name, "foo",
		fs.Root.Entries[0].File.IsDir, false,
	)

}
