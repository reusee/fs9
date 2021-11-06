package fs9

import "testing"

func BenchmarkMemFSWrite(b *testing.B) {
	fs := NewMemFS()
	f, err := fs.OpenHandle("foo", OptCreate(true))
	ce(err)
	defer f.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := f.Seek(0, 0)
		ce(err)
		_, err = f.Write([]byte("foo"))
		ce(err)
	}
}
