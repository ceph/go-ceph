package cutil

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestCSlice(t *testing.T) {
	t.Run("CPtrSlice", func(t *testing.T) {
		const size = 256
		slice := NewCPtrSlice(size)
		for i := range slice {
			slice[i] = CPtr(&slice[i])
		}
		for i := 0; i < size; i++ {
			p := (*CPtr)(unsafe.Pointer(uintptr(slice.Ptr()) + uintptr(i)*uintptr(ptrSize)))
			assert.Equal(t, CPtr(p), *p)
		}
		assert.Panics(t, func() { slice[size] = nil })
		slice.Free()
		assert.Len(t, slice, 0)
		assert.Nil(t, slice)
		empty := NewCPtrSlice(0)
		assert.Len(t, empty, 0)
		assert.Nil(t, empty)
		empty.Free()
	})
	t.Run("CSizeSlice", func(t *testing.T) {
		const size = 256
		slice := NewCSizeSlice(size)
		for i := range slice {
			slice[i] = CSize(i)
		}
		for i := 0; i < size; i++ {
			p := (*CSize)(unsafe.Pointer(uintptr(slice.Ptr()) + uintptr(i)*unsafe.Sizeof(CSize(0))))
			assert.Equal(t, CSize(i), *p)
		}
		assert.Panics(t, func() { slice[size] = 0 })
		slice.Free()
		assert.Len(t, slice, 0)
		assert.Nil(t, slice)
		empty := NewCSizeSlice(0)
		assert.Len(t, empty, 0)
		assert.Nil(t, empty)
		empty.Free()
	})
}
