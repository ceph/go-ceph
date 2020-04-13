package cutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandInput(t *testing.T) {
	t.Run("newAndFree", func(t *testing.T) {
		ci := NewCommandInput(
			[][]byte{[]byte("foobar")},
			nil)
		ci.Free()
	})
	t.Run("cmd", func(t *testing.T) {
		ci := NewCommandInput(
			[][]byte{[]byte("foobar")},
			nil)
		defer ci.Free()
		assert.Len(t, ci.cmd, 1)
		assert.EqualValues(t, 1, ci.CmdLen())
		assert.NotNil(t, ci.Cmd())
	})
	t.Run("cmd2", func(t *testing.T) {
		ci := NewCommandInput(
			[][]byte{[]byte("foobar"), []byte("snarf")},
			nil)
		defer ci.Free()
		assert.Len(t, ci.cmd, 2)
		assert.EqualValues(t, 2, ci.CmdLen())
		assert.NotNil(t, ci.Cmd())
	})
	t.Run("noInBuf", func(t *testing.T) {
		ci := NewCommandInput(
			[][]byte{[]byte("foobar")},
			nil)
		defer ci.Free()
		assert.EqualValues(t, 0, ci.InBufLen())
		assert.Equal(t, CharPtr(nil), ci.InBuf())
	})
	t.Run("hasInBuf", func(t *testing.T) {
		ci := NewCommandInput(
			[][]byte{[]byte("foobar")},
			[]byte("original oregano"))
		defer ci.Free()
		assert.EqualValues(t, 16, ci.InBufLen())
		assert.NotEqual(t, CharPtr(nil), ci.InBuf())
	})
}
