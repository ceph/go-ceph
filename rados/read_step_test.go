package rados

import (
	"runtime"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"

	"github.com/ceph/go-ceph/internal/cutil"
)

// cgoAccepts reports whether C may be handed Go memory holding p, which
// cgocheck allows only while p is pinned.
func cgoAccepts(p *byte) (ok bool) {
	holder := &struct{ p *byte }{p}
	dst := cutil.Malloc(cutil.PtrSize)
	defer cutil.Free(dst)
	defer func() { ok = recover() == nil }()
	cutil.Memcpy(dst, cutil.CPtr(unsafe.Pointer(holder)), cutil.PtrSize)
	return
}

func TestReadStepPinsBuffer(t *testing.T) {
	b := make([]byte, 16)
	if cgoAccepts(&b[0]) {
		t.Skip("cgocheck is disabled")
	}
	op := CreateReadOp()
	op.Read(0, b)
	assert.True(t, cgoAccepts(&b[0]), "buffer not pinned")
	op.Release()
	assert.False(t, cgoAccepts(&b[0]), "buffer still pinned after Release")
}

func TestReadStepUnreleased(_ *testing.T) {
	// the finalizer of a step that was never released must unpin its buffer
	func() {
		_ = newReadStep(make([]byte, 16), 0)
	}()
	for range 3 {
		runtime.GC()
	}
}
