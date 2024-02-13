package cutil

import (
	"math/rand"
	"runtime"
	"testing"
	"time"
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

	t.Run("keepsReachable", func(t *testing.T) {
		var pgDone, uDone bool
		goPtr := func(b *bool) unsafe.Pointer {
			s := "ok"
			runtime.SetFinalizer(&s, func(_ *string) { *b = true })
			return unsafe.Pointer(&s)
		}
		cPtr := Malloc(PtrSize)
		defer Free(cPtr)
		pg := NewPtrGuard(cPtr, goPtr(&pgDone))
		u := uintptr(goPtr(&uDone))
		runtime.GC()
		assert.Eventually(t, func() bool { return uDone },
			time.Second, 10*time.Millisecond)
		assert.False(t, pgDone)
		pg.Release()
		runtime.GC()
		assert.Eventually(t, func() bool { return pgDone },
			time.Second, 10*time.Millisecond)
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
