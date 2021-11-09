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
	bs := bytes.Repeat([]byte("a"), 4096)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := f.Seek(0, 0)
		ce(err)
		_, err = f.Write(bs)
		ce(err)
		b.SetBytes(int64(len(bs)))
	}
}

func BenchmarkMemFSParallelWrite(b *testing.B) {
	fs := NewMemFS()
	bs := bytes.Repeat([]byte("a"), 4096)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		f, err := fs.OpenHandle("foo", OptCreate(true))
		ce(err)
		defer f.Close()
		for pb.Next() {
			_, err := f.Seek(0, 0)
			ce(err)
			_, err = f.Write(bs)
			ce(err)
			b.SetBytes(int64(len(bs)))
		}
	})
}
