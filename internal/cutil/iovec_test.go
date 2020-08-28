package cutil

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestIovec(t *testing.T) {
	t.Run("newAndFree", func(t *testing.T) {
		iov := NewIovec(3)
		iov.Free()
	})
	t.Run("setBufs", func(t *testing.T) {
		b1 := []byte("foo")
		b2 := []byte("barbar")
		b3 := []byte("bazbazbaz")
		iov := NewIovec(3)
		iov.Set(0, b1)
		iov.Set(1, b2)
		iov.Set(2, b3)
		iov.Free()
		// free also unsets internal values
		assert.Equal(t, unsafe.Pointer(nil), iov.cvec)
		assert.Equal(t, 0, iov.length)
	})
	t.Run("testGetters", func(t *testing.T) {
		b1 := []byte("foo")
		b2 := []byte("barbar")
		b3 := []byte("bazbazbaz")
		b4 := []byte("zonk")
		iov := NewIovec(4)
		defer iov.Free()
		iov.Set(0, b1)
		iov.Set(1, b2)
		iov.Set(2, b3)
		iov.Set(3, b4)

		assert.NotNil(t, iov.Pointer())
		assert.Equal(t, 4, iov.Len())
	})
}

func TestByteSlicesToIovec(t *testing.T) {
	d := [][]byte{
		[]byte("ramekin"),
		[]byte("shuffleboard"),
		[]byte("tranche"),
		[]byte("phycobilisomes"),
	}
	iov := ByteSlicesToIovec(d)
	defer iov.Free()

	assert.NotNil(t, iov.Pointer())
	assert.Equal(t, 4, iov.Len())
}
