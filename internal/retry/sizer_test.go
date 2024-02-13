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
		t.Run(fmt.Sprintf("bcopy_%d", i), func(t *testing.T) {
			var out []byte
			var err error
			WithSizes(1, 4096, func(size int) Hint {
				out, err = bcopy(b, size)
				return DoubleSize.If(err == tooLong)
			})
			assert.Nil(t, err)
			assert.Equal(t, b, out[:len(b)])
		})
	}

	for i, b := range src {
		t.Run(fmt.Sprintf("bcopy_hint_%d", i), func(t *testing.T) {
			var tries int
			var err error
			var out []byte
			WithSizes(1, 4096, func(size int) Hint {
				tries++
				out, err = bcopy(b, size)
				return Size(len(b)).If(err == tooLong)
			})
			assert.Nil(t, err)
			assert.Equal(t, b, out)
			assert.Equal(t, 2, tries)
		})
	}

	t.Run("exceedsMax", func(t *testing.T) {
		var tries int
		var err error
		WithSizes(1, 1024, func(_ int) Hint {
			tries++
			err = errors.New("foo")
			return DoubleSize
		})
		assert.Error(t, err)
		assert.Equal(t, 11, tries)
	})

	t.Run("hintExceedsMax", func(t *testing.T) {
		var tries int
		var lastSize int
		WithSizes(1, 1024, func(size int) Hint {
			tries++
			lastSize = size
			return Size(1025)
		})
		assert.Equal(t, 1024, lastSize)
		assert.Equal(t, 2, tries)
	})

	t.Run("weirdSizeAndMax", func(t *testing.T) {
		var tries int
		var lastSize int
		WithSizes(3, 1022, func(size int) Hint {
			tries++
			lastSize = size
			return DoubleSize
		})
		assert.Equal(t, 10, tries)
		assert.Equal(t, 1022, lastSize)
	})

	t.Run("sizeExceedsMax", func(t *testing.T) {
		var lastSize int
		WithSizes(1023, 1022, func(size int) Hint {
			lastSize = size
			return DoubleSize
		})
		assert.Equal(t, 0, lastSize)
	})

}
