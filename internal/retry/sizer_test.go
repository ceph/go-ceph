package retry

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSizer(t *testing.T) {
	tooLong := errors.New("too long")

	src := [][]byte{
		[]byte("foobarbaz"),
		[]byte("gondwandaland"),
		[]byte("longer and longer still, not quite done"),
		[]byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."),
	}

	// mimic a complex-ish data copy call
	bcopy := func(src []byte, size int) ([]byte, error) {
		if size < len(src) {
			return nil, tooLong
		}
		dst := make([]byte, size)
		copy(dst, src)
		return dst, nil
	}

	for i, b := range src {
		t.Run(fmt.Sprintf("update_%d", i), func(t *testing.T) {
			var err error
			var out []byte
			for sizer := NewSizerEV(1, 4096, tooLong); sizer.Continue(); {
				out, err = bcopy(b, sizer.Size())
				err = sizer.Update(err)
			}
			assert.Equal(t, b, out[:len(b)])
		})
	}

	for i, b := range src {
		t.Run(fmt.Sprintf("updateWants_%d", i), func(t *testing.T) {
			var tries int
			var err error
			var out []byte
			for sizer := NewSizerEV(1, 4096, tooLong); sizer.Continue(); {
				tries++
				out, err = bcopy(b, sizer.Size())
				err = sizer.UpdateWants(err, len(b))
			}
			assert.Equal(t, b, out[:len(b)])
			assert.Equal(t, len(b), len(out))
			assert.Equal(t, 2, tries)
		})
	}

	t.Run("exceedsMax", func(t *testing.T) {
		var tries int
		var err error
		for sizer := NewSizerEV(1, 1024, tooLong); sizer.Continue(); {
			tries++
			err = sizer.Update(tooLong)
		}
		assert.Error(t, err)
		assert.Equal(t, tooLong, err)
		assert.Equal(t, 11, tries)
	})

	t.Run("otherError", func(t *testing.T) {
		var tries int
		var err error
		oops := errors.New("foo")
		for sizer := NewSizerEV(1, 1024, tooLong); sizer.Continue(); {
			tries++
			err = sizer.Update(oops)
		}
		assert.Error(t, err)
		assert.Equal(t, oops, err)
		assert.Equal(t, 1, tries)
	})
}
