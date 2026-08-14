package retry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrier(t *testing.T) {
	t.Run("bigEnoughFirstTry", func(t *testing.T) {
		tcount := 0
		WithTries(1024, 2, func(size int) Hint {
			tcount++
			if size == 1024 {
				return nil
			}
			assert.Fail(t, "invalid size")
			return nil
		})
		assert.Equal(t, 1, tcount)
	})

	t.Run("idealSizeHint", func(t *testing.T) {
		tcount := 0
		WithTries(1024, 2, func(size int) Hint {
			tcount++
			if size == 1024 {
				return Size(12345)
			}
			if size == 12345 {
				return nil
			}
			assert.Fail(t, "invalid size")
			return nil
		})
		assert.Equal(t, 2, tcount)
	})

	t.Run("neverBigEnough", func(t *testing.T) {
		tcount := 0
		WithTries(1024, 8, func(size int) Hint {
			tcount++
			// imagine 12345 was our target size, but could never get there
			return Size(size + 1) // poorly chosen size
		})
		assert.Equal(t, 8, tcount)
	})

	t.Run("doubleTrouble", func(t *testing.T) {
		tcount := 0
		WithTries(1024, 16, func(size int) Hint {
			tcount++
			// imagine 12345 was our target size, but could never get there
			assert.GreaterOrEqual(t, size, 1024)
			return Size(1) // poorly chosen size
		})
		assert.Equal(t, 16, tcount)
	})
}
