//go:build ceph_preview

package rbd

import (
	"bytes"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"

	"github.com/ceph/go-ceph/internal/errutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRebuildObjectMapWithProgress(t *testing.T) {
	invalidMask := FlagObjectMapInvalid | FlagFastDiffInvalid

	t.Run("valid", func(t *testing.T) {
		image, _, _ := makeObjectMapImage(t, true)

		payload := bytes.Repeat([]byte{0x5a}, 4096)
		for _, offset := range []int64{0, 1 << testImageOrder, 2 << testImageOrder} {
			n, err := image.WriteAt(payload, offset)
			require.NoError(t, err)
			require.Equal(t, len(payload), n)
		}

		flags, err := image.GetFlags()
		require.NoError(t, err)
		require.Equal(t, invalidMask, flags&invalidMask,
			"fixture must start with invalid object-map and fast-diff flags")

		calls := 0
		var callbackLock sync.Mutex
		cb := func(objectNumber, objectCount uint64, data interface{}) int {
			callbackLock.Lock()
			defer callbackLock.Unlock()
			calls++
			assert.Equal(t, uint64(42), data)
			assert.Positive(t, objectCount)
			assert.LessOrEqual(t, objectNumber, objectCount)
			return 0
		}

		err = image.RebuildObjectMapWithProgress(cb, uint64(42))
		require.NoError(t, err)

		flags, err = image.GetFlags()
		require.NoError(t, err)
		assert.Zero(t, flags&invalidMask, "successful rebuild must clear both invalid flags")

		callbackLock.Lock()
		defer callbackLock.Unlock()
		assert.GreaterOrEqual(t, calls, 1)
	})

	t.Run("selectedSnapshot", func(t *testing.T) {
		image, ioctx, name := makeObjectMapImage(t, true)

		snapName := GetUUID()
		snapshot, err := image.CreateSnapshot(snapName)
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, snapshot.Remove()) })

		snapImage, err := OpenImage(ioctx, name, snapName)
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, snapImage.Close()) })

		flags, err := snapImage.GetFlags()
		require.NoError(t, err)
		require.Equal(t, invalidMask, flags&invalidMask,
			"snapshot must start with invalid object-map and fast-diff flags")

		var calls atomic.Int32
		err = snapImage.RebuildObjectMapWithProgress(
			func(_, _ uint64, _ interface{}) int {
				calls.Add(1)
				return 0
			}, nil)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, calls.Load(), int32(1))

		flags, err = snapImage.GetFlags()
		require.NoError(t, err)
		assert.Zero(t, flags&invalidMask, "snapshot rebuild must clear both invalid flags")

		flags, err = image.GetFlags()
		require.NoError(t, err)
		assert.Equal(t, invalidMask, flags&invalidMask,
			"snapshot rebuild must leave both HEAD invalid flags set")
	})

	t.Run("abortCallback", func(t *testing.T) {
		image, _, _ := makeObjectMapImage(t, true)

		var calls atomic.Int32
		err := image.RebuildObjectMapWithProgress(
			func(_, _ uint64, _ interface{}) int {
				calls.Add(1)
				return -int(syscall.ECANCELED)
			}, nil)
		require.ErrorIs(t, err,
			errutil.GetError("rbd", -int(syscall.ECANCELED)))
		assert.GreaterOrEqual(t, calls.Load(), int32(1))

		err = image.RebuildObjectMapWithProgress(
			func(_, _ uint64, _ interface{}) int { return 0 }, nil)
		require.NoError(t, err)
	})

	t.Run("objectMapDisabled", func(t *testing.T) {
		image, _, _ := makeObjectMapImage(t, false)

		err := image.RebuildObjectMapWithProgress(
			func(_, _ uint64, _ interface{}) int { return 0 }, nil)
		assert.ErrorIs(t, err,
			errutil.GetError("rbd", -int(syscall.EINVAL)))
	})

	t.Run("closedImage", func(t *testing.T) {
		image, _, _ := makeObjectMapImage(t, true)
		require.NoError(t, image.Close())

		err := image.RebuildObjectMapWithProgress(
			func(_, _ uint64, _ interface{}) int { return 0 }, nil)
		assert.ErrorIs(t, err, ErrImageNotOpen)
	})

	t.Run("readOnlyImage", func(t *testing.T) {
		image, ioctx, name := makeObjectMapImage(t, true)
		require.NoError(t, image.Close())

		readOnlyImage, err := OpenImageReadOnly(ioctx, name, NoSnapshot)
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, readOnlyImage.Close()) })

		err = readOnlyImage.RebuildObjectMapWithProgress(
			func(_, _ uint64, _ interface{}) int { return 0 }, nil)
		assert.ErrorIs(t, err,
			errutil.GetError("rbd", -int(syscall.EROFS)))
	})

	t.Run("nilCallback", func(t *testing.T) {
		image, _, _ := makeObjectMapImage(t, true)

		err := image.RebuildObjectMapWithProgress(nil, nil)
		assert.ErrorIs(t, err,
			errutil.GetError("rbd", -int(syscall.EINVAL)))
	})
}
