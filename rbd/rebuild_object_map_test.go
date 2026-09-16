//go:build ceph_preview

package rbd

import (
	"sync"
	"syscall"
	"testing"

	"github.com/ceph/go-ceph/internal/errutil"
	"github.com/ceph/go-ceph/rados"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeObjectMapImage(t *testing.T, enableObjectMap bool) (*Image, *rados.IOContext, string) {
	t.Helper()

	conn := radosConnect(t)
	t.Cleanup(conn.Shutdown)

	poolName := GetUUID()
	require.NoError(t, conn.MakePool(poolName))
	t.Cleanup(func() { require.NoError(t, conn.DeletePool(poolName)) })

	ioctx, err := conn.OpenIOContext(poolName)
	require.NoError(t, err)
	t.Cleanup(ioctx.Destroy)

	name := GetUUID()
	_, err = Create(ioctx, name, testImageSize, testImageOrder,
		FeatureExclusiveLock)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, RemoveImage(ioctx, name)) })

	image, err := OpenImage(ioctx, name, NoSnapshot)
	require.NoError(t, err)
	t.Cleanup(func() {
		if image.image != nil {
			require.NoError(t, image.Close())
		}
	})

	if enableObjectMap {
		require.NoError(t, image.UpdateFeatures(FeatureObjectMap, true))
	}

	return image, ioctx, name
}

func TestRebuildObjectMapWithProgress(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		image, _, _ := makeObjectMapImage(t, true)

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

		err := image.RebuildObjectMapWithProgress(cb, uint64(42))
		assert.NoError(t, err)
		callbackLock.Lock()
		defer callbackLock.Unlock()
		assert.GreaterOrEqual(t, calls, 1)
	})

	t.Run("abortCallback", func(t *testing.T) {
		image, _, _ := makeObjectMapImage(t, true)

		calls := 0
		err := image.RebuildObjectMapWithProgress(
			func(_, _ uint64, _ interface{}) int {
				calls++
				return -1
			}, nil)
		assert.Error(t, err)
		assert.GreaterOrEqual(t, calls, 1)

		err = image.RebuildObjectMapWithProgress(
			func(_, _ uint64, _ interface{}) int { return 0 }, nil)
		assert.NoError(t, err)
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
