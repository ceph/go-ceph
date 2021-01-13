package cutil

import (
	"math/rand"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestPtrGuard(t *testing.T) {
	t.Run("storeAndRelease", func(t *testing.T) {
		s := "string"
		goPtr := (unsafe.Pointer)(&s)
		cPtr := Malloc(PtrSize)
		defer Free(cPtr)
		pg := NewPtrGuard(cPtr, goPtr)
		assert.Equal(t, *(*unsafe.Pointer)(cPtr), goPtr)
		pg.Release()
		assert.Zero(t, *(*unsafe.Pointer)(cPtr))
	})

	t.Run("multiRelease", func(t *testing.T) {
		s := "string"
		goPtr := (unsafe.Pointer)(&s)
		cPtr := Malloc(PtrSize)
		defer Free(cPtr)
		pg := NewPtrGuard(cPtr, goPtr)
		assert.Equal(t, *(*unsafe.Pointer)(cPtr), goPtr)
		pg.Release()
		pg.Release()
		pg.Release()
		pg.Release()
		assert.Zero(t, *(*unsafe.Pointer)(cPtr))
	})

	t.Run("stressTest", func(t *testing.T) {
		const N = 1000
		const M = 10000
		var ptrGuards [N]*PtrGuard
		cPtrArr := (*[N]CPtr)(unsafe.Pointer(Malloc(N * PtrSize)))
		defer Free(CPtr(&cPtrArr[0]))
		for n := 0; n < M; n++ {
			i := uintptr(rand.Intn(N))
			if ptrGuards[i] == nil {
				goPtr := unsafe.Pointer(&(struct{ byte }{42}))
				cPtrPtr := CPtr(&cPtrArr[i])
				ptrGuards[i] = NewPtrGuard(cPtrPtr, goPtr)
				assert.Equal(t, (unsafe.Pointer)(cPtrArr[i]), goPtr)
			} else {
				ptrGuards[i].Release()
				ptrGuards[i] = nil
				assert.Zero(t, cPtrArr[i])
			}
		}
		for i := range ptrGuards {
			if ptrGuards[i] != nil {
				ptrGuards[i].Release()
				ptrGuards[i] = nil
			}
		}
		for i := uintptr(0); i < N; i++ {
			assert.Zero(t, cPtrArr[i])
		}
	})
}
