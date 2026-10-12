package rados

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadStepRelease(t *testing.T) {
	op := CreateReadOp()
	s := op.Read(0, make([]byte, 16))
	assert.NotNil(t, s)
	op.Release()
	// Release unpins the buffer, so collecting it must not panic.
	runtime.GC()
}

func TestReadStepUnreleased(_ *testing.T) {
	// The runtime panics if it collects a Pinner that still pins memory.
	// The step's finalizer must unpin the buffer of an op that was never
	// released.
	func() {
		_ = newReadStep(make([]byte, 16), 0)
	}()
	for range 3 {
		runtime.GC()
	}
}
