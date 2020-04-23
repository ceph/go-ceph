package cutil

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestCommandOutput(t *testing.T) {
	t.Run("newAndFree", func(t *testing.T) {
		co := NewCommandOutput()
		assert.NotNil(t, co)
		co.Free()
	})
	t.Run("setValues", func(t *testing.T) {
		co := NewCommandOutput()
		assert.NotNil(t, co)
		defer co.Free()
		testSetString(co.OutBuf(), co.OutBufLen(), "i got style")
		testSetString(co.Outs(), co.OutsLen(), "i got rhythm")
		b, s := co.GoValues()
		assert.EqualValues(t, []byte("i got style"), b)
		assert.EqualValues(t, "i got rhythm", s)
	})
	t.Run("setOnlyOutBuf", func(t *testing.T) {
		co := NewCommandOutput()
		assert.NotNil(t, co)
		defer co.Free()
		testSetString(co.OutBuf(), co.OutBufLen(), "i got style")
		b, s := co.GoValues()
		assert.EqualValues(t, []byte("i got style"), b)
		assert.EqualValues(t, "", s)
	})
	t.Run("setOnlyOuts", func(t *testing.T) {
		co := NewCommandOutput()
		assert.NotNil(t, co)
		defer co.Free()
		testSetString(co.Outs(), co.OutsLen(), "i got rhythm")
		b, s := co.GoValues()
		assert.Nil(t, b)
		assert.EqualValues(t, "i got rhythm", s)
	})
	t.Run("customFreeFunc", func(t *testing.T) {
		callCount := 0
		co := NewCommandOutput().SetFreeFunc(func(p unsafe.Pointer) {
			callCount++
			free(p)
		})
		assert.NotNil(t, co)
		testSetString(co.OutBuf(), co.OutBufLen(), "i got style")
		testSetString(co.Outs(), co.OutsLen(), "i got rhythm")
		co.Free()
		assert.Equal(t, 2, callCount)
	})
}
