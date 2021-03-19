package cutil

import (
	"math/rand"
	"runtime"
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

	t.Run("uintptrescapesTest", func(t *testing.T) {
		// This test assures that the special //go:uintptrescapes comment before
		// the storeUntilRelease() function works as intended, that is the
		// garbage collector doesn't touch the object referenced by the uintptr
		// until the function returns after Release() is called. The test will
		// fail if the //go:uintptrescapes comment is disabled (removed) or
		// stops working in future versions of go.
		var pg_done, u_done bool
		var goPtr = func(b *bool) unsafe.Pointer {
			s := "ok"
			runtime.SetFinalizer(&s, func(p *string) { *b = true })
			return unsafe.Pointer(&s)
		}
		cPtr := Malloc(PtrSize)
		defer Free(cPtr)
		pg := NewPtrGuard(cPtr, goPtr(&pg_done))
		u := uintptr(goPtr(&u_done))
		runtime.GC()
		assert.True(t, u_done)
		assert.False(t, pg_done)
		pg.Release()
		runtime.GC()
		assert.True(t, pg_done)
		assert.NotZero(t, u) // avoid "unused" error
	})

	t.Run("stressTest", func(t *testing.T) {
		// Because the default thread limit of the Go runtime is 10000, creating
		// 20000 parallel PtrGuards asserts, that Go routines of PtrGuards don't
		// create threads.
		const N = 20000  // Number of parallel PtrGuards
		const M = 100000 // Number of loops
		var ptrGuards [N]*PtrGuard
		cPtrArr := (*[N]CPtr)(unsafe.Pointer(Malloc(N * PtrSize)))
		defer Free(CPtr(&cPtrArr[0]))
		toggle := func(i int) {
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
			toggle(i)
		}
		for n := 0; n < M; n++ {
			i := rand.Intn(N)
			toggle(i)
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
