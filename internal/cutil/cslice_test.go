package cutil

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestCSlice(t *testing.T) {
	t.Run("Size", func(t *testing.T) {
		const size = 256
		slice := NewElementTypeCSlice(size)
		defer slice.Free()
		assert.Len(t, slice, size)
	})
	t.Run("Fill", func(t *testing.T) {
		const size = 256
		slice := NewElementTypeCSlice(size)
		defer slice.Free()
		for i := range slice {
			slice[i] = ElementType(i)
		}
		for i := 0; i < size; i++ {
			p := (*ElementType)(unsafe.Pointer(uintptr(slice.Ptr()) + uintptr(i)*uintptr(ElementTypeSize)))
			assert.Equal(t, slice[i], *p)
		}
	})
	t.Run("OutOfBound", func(t *testing.T) {
		const size = 1
		slice := NewElementTypeCSlice(size)
		defer slice.Free()
		assert.Panics(t, func() { slice[size] = nil })
	})
	t.Run("FreeSetsNil", func(t *testing.T) {
		const size = 1
		slice := NewElementTypeCSlice(size)
		slice.Free()
		assert.Nil(t, slice)
	})
	t.Run("EmptySlice", func(t *testing.T) {
		empty := NewElementTypeCSlice(0)
		assert.Len(t, empty, 0)
		assert.Nil(t, empty)
		assert.NotPanics(t, func() { empty.Free() })
	})
}
