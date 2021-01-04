package cutil

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestCPtrCSlice(t *testing.T) {
	t.Run("Size", func(t *testing.T) {
		const size = 256
		slice := NewCPtrCSlice(size)
		defer slice.Free()
		assert.Len(t, slice, size)
	})
	t.Run("Fill", func(t *testing.T) {
		const size = 256
		slice := NewCPtrCSlice(size)
		defer slice.Free()
		for i := range slice {
			slice[i] = CPtr(&slice[i])
		}
		for i := 0; i < size; i++ {
			p := (*CPtr)(unsafe.Pointer(uintptr(slice.Ptr()) + uintptr(i)*uintptr(PtrSize)))
			assert.Equal(t, slice[i], *p)
		}
	})
	t.Run("OutOfBound", func(t *testing.T) {
		const size = 1
		slice := NewCPtrCSlice(size)
		defer slice.Free()
		assert.Panics(t, func() { slice[size] = nil })
	})
	t.Run("FreeSetsNil", func(t *testing.T) {
		const size = 1
		slice := NewCPtrCSlice(size)
		slice.Free()
		assert.Nil(t, slice)
	})
	t.Run("EmptySlice", func(t *testing.T) {
		empty := NewCPtrCSlice(0)
		assert.Len(t, empty, 0)
		assert.Nil(t, empty)
		assert.NotPanics(t, func() { empty.Free() })
	})
}

func TestSizeTCSlice(t *testing.T) {
	t.Run("Size", func(t *testing.T) {
		const size = 256
		slice := NewSizeTCSlice(size)
		defer slice.Free()
		assert.Len(t, slice, size)
	})
	t.Run("Fill", func(t *testing.T) {
		const size = 256
		slice := NewSizeTCSlice(size)
		defer slice.Free()
		for i := range slice {
			slice[i] = SizeT(i)
		}
		for i := 0; i < size; i++ {
			p := (*SizeT)(unsafe.Pointer(uintptr(slice.Ptr()) + uintptr(i)*uintptr(PtrSize)))
			assert.Equal(t, slice[i], *p)
		}
	})
	t.Run("OutOfBound", func(t *testing.T) {
		const size = 1
		slice := NewSizeTCSlice(size)
		defer slice.Free()
		assert.Panics(t, func() { slice[size] = 0 })
	})
	t.Run("FreeSetsNil", func(t *testing.T) {
		const size = 1
		slice := NewSizeTCSlice(size)
		slice.Free()
		assert.Nil(t, slice)
	})
	t.Run("EmptySlice", func(t *testing.T) {
		empty := NewSizeTCSlice(0)
		assert.Len(t, empty, 0)
		assert.Nil(t, empty)
		assert.NotPanics(t, func() { empty.Free() })
	})
}
