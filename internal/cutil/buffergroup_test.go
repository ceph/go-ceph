package cutil

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestBufferGroup(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		s := NewBufferGroupStrings(nil)
		assert.NotPanics(t, func() { s.Free() })
	})
	t.Run("NotEmpty", func(t *testing.T) {
		s := NewBufferGroupStrings([]string{"hello"})
		assert.NotPanics(t, func() { s.Free() })
	})
	t.Run("FreeSetsNil", func(t *testing.T) {
		s := NewBufferGroupStrings([]string{"hello"})
		s.Free()
		assert.Nil(t, s.Buffers)
		assert.Nil(t, s.Lengths)
	})
	t.Run("DoubleFree", func(t *testing.T) {
		s := NewBufferGroupStrings([]string{"hello"})
		assert.NotPanics(t, func() { s.Free() })
		assert.NotPanics(t, func() { s.Free() })
	})
	t.Run("ValidPtrs", func(t *testing.T) {
		s := NewBufferGroupStrings([]string{"hello"})
		defer s.Free()
		assert.Equal(t, unsafe.Pointer(&s.Buffers[0]), unsafe.Pointer(s.BuffersPtr()))
		assert.Equal(t, unsafe.Pointer(&s.Lengths[0]), unsafe.Pointer(s.LengthsPtr()))
	})
	t.Run("ValidContents", func(t *testing.T) {
		values := []string{
			"1", "12", "123", "世界", "abc\x00", "ab\x00c",
		}

		s := NewBufferGroupStrings(values)
		defer s.Free()

		assert.Equal(t, len(values), len(s.Buffers))
		assert.Equal(t, len(values), len(s.Lengths))

		for i := range values {
			actualStr, actualLen := testBufferGroupGet(s, i)
			assert.Equal(t, values[i], actualStr)
			assert.Equal(t, len(values[i]), actualLen)
		}
	})
}
