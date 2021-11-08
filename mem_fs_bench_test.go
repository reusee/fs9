package fs9

import (
	"bytes"
	"testing"
)

func BenchmarkMemFSWrite(b *testing.B) {
	fs := NewMemFS()
	f, err := fs.OpenHandle("foo", OptCreate(true))
	ce(err)
	defer f.Close()
	bs := bytes.Repeat([]byte("a"), 1*1024*1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := f.Seek(0, 0)
		ce(err)
		_, err = f.Write(bs)
		ce(err)
		b.SetBytes(int64(len(bs)))
	}
}
