package cutil

import "testing"

func BenchmarkSyncBuffer(b *testing.B) {
	data := make([]byte, 1024*64)
	var p = Malloc(PtrSize)
	defer Free(p)
	for i := 0; i < b.N; i++ {
		sb := NewSyncBuffer(p, data)
		sb.Sync()
		sb.Release()
	}
}
