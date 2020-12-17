package cutil

import (
	"math/rand"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

const ptrSize = unsafe.Sizeof(unsafe.Pointer(nil))

func TestPtrGuard(t *testing.T) {
	t.Run("storeAndRelease", func(t *testing.T) {
		s := "string"
		goPtr := (unsafe.Pointer)(&s)
		cPtr := (*unsafe.Pointer)(cMalloc(unsafe.Sizeof(goPtr)))
		defer cFree(unsafe.Pointer(cPtr))
		pg := NewPtrGuard((*unsafe.Pointer)(cPtr), goPtr)
		assert.Equal(t, *cPtr, goPtr)
		pg.Release()
		assert.Zero(t, *cPtr)
	})

	t.Run("multiRelease", func(t *testing.T) {
		s := "string"
		goPtr := (unsafe.Pointer)(&s)
		cPtr := (*unsafe.Pointer)(cMalloc(unsafe.Sizeof(goPtr)))
		defer cFree(unsafe.Pointer(cPtr))
		pg := NewPtrGuard((*unsafe.Pointer)(cPtr), goPtr)
		assert.Equal(t, *cPtr, goPtr)
		pg.Release()
		pg.Release()
		pg.Release()
		pg.Release()
		assert.Zero(t, *cPtr)
	})

	t.Run("stressTest", func(t *testing.T) {
		const N = 1000
		const M = 10000
		var ptrGuards [N]*PtrGuard
		cPtrArr := cMalloc(N * ptrSize)
		defer cFree(cPtrArr)
		for n := 0; n < M; n++ {
			i := uintptr(rand.Intn(N))
			if ptrGuards[i] == nil {
				goPtr := unsafe.Pointer(&(struct{ byte }{42}))
				cPtr := (*unsafe.Pointer)(unsafe.Pointer(uintptr(cPtrArr) + i*ptrSize))
				pg := NewPtrGuard((*unsafe.Pointer)(cPtr), goPtr)
				ptrGuards[i] = pg
				assert.Equal(t, *cPtr, goPtr)
			} else {
				ptrGuards[i].Release()
				ptrGuards[i] = nil
				cPtr := (*unsafe.Pointer)(unsafe.Pointer(uintptr(cPtrArr) + i*ptrSize))
				assert.Zero(t, *cPtr)
			}
		}
		for _, pg := range ptrGuards {
			if pg != nil {
				pg.Release()
				pg = nil
			}
		}
		for i := uintptr(0); i < N; i++ {
			cPtr := (*unsafe.Pointer)(unsafe.Pointer(uintptr(cPtrArr) + i*ptrSize))
			assert.Zero(t, *cPtr)
		}
	})
}
