//go:build ceph_preview

package rbd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeCloneForFlatten is a test helper that creates a parent image, protects a
// snapshot, and returns an open clone image ready to be flattened, along with
// a cleanup function that tears down every resource in the correct order.
func makeCloneForFlatten(t *testing.T) (clone *Image, cleanup func()) {
	t.Helper()

	conn := radosConnect(t)

	poolname := GetUUID()
	require.NoError(t, conn.MakePool(poolname))

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)

	// Create and open the parent image.
	parentName := GetUUID()
	options := NewRbdImageOptions()
	require.NoError(t, options.SetUint64(ImageOptionOrder, uint64(testImageOrder)))
	require.NoError(t, options.SetUint64(ImageOptionFeatures, FeatureLayering))
	err = CreateImage(ioctx, parentName, testImageSize, options)
	require.NoError(t, err)

	parent, err := OpenImage(ioctx, parentName, NoSnapshot)
	require.NoError(t, err)

	// Create and protect a snapshot so the parent can be cloned.
	snapName := GetUUID()
	snap, err := parent.CreateSnapshot(snapName)
	require.NoError(t, err)
	require.NoError(t, snap.Protect())

	// Clone the snapshot into a new child image.
	childName := GetUUID()
	_, err = parent.Clone(snapName, ioctx, childName, 1, testImageOrder)
	require.NoError(t, err)

	// Open the child image.
	child, err := OpenImage(ioctx, childName, NoSnapshot)
	require.NoError(t, err)

	cleanup = func() {
		_ = child.Close()
		_ = RemoveImage(ioctx, childName)

		// Unprotect and remove the snapshot before the parent can be removed.
		p, err := OpenImage(ioctx, parentName, NoSnapshot)
		if err == nil {
			s := p.GetSnapshot(snapName)
			_ = s.Unprotect()
			_ = s.Remove()
			_ = p.Close()
		}
		_ = RemoveImage(ioctx, parentName)

		ioctx.Destroy()
		conn.DeletePool(poolname)
		conn.Shutdown()
	}

	require.NoError(t, parent.Close())
	return child, cleanup
}

func TestFlattenWithProgress(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		child, cleanup := makeCloneForFlatten(t)
		defer cleanup()

		cc := 0
		cb := func(_, _ uint64, v interface{}) int {
			cc++
			val := v.(int)
			assert.Equal(t, 42, val)
			return 0
		}

		err := child.FlattenWithProgress(cb, 42)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, cc, 1)
	})

	t.Run("abortCallback", func(t *testing.T) {
		child, cleanup := makeCloneForFlatten(t)
		defer cleanup()

		cb := func(_, _ uint64, _ interface{}) int {
			return -1 // signal abort
		}

		err := child.FlattenWithProgress(cb, nil)
		assert.Error(t, err)
	})

	t.Run("closedImage", func(t *testing.T) {
		conn := radosConnect(t)
		defer conn.Shutdown()

		poolname := GetUUID()
		err := conn.MakePool(poolname)
		require.NoError(t, err)
		defer conn.DeletePool(poolname)

		ioctx, err := conn.OpenIOContext(poolname)
		require.NoError(t, err)
		defer ioctx.Destroy()

		name := GetUUID()
		err = quickCreate(ioctx, name, testImageSize, testImageOrder)
		require.NoError(t, err)
		defer func() { assert.NoError(t, RemoveImage(ioctx, name)) }()

		img, err := OpenImage(ioctx, name, NoSnapshot)
		require.NoError(t, err)
		require.NoError(t, img.Close())

		cb := func(_, _ uint64, _ interface{}) int { return 0 }
		err = img.FlattenWithProgress(cb, nil)
		assert.Error(t, err)
		assert.Equal(t, ErrImageNotOpen, err)
	})

	t.Run("nilCallback", func(t *testing.T) {
		conn := radosConnect(t)
		defer conn.Shutdown()

		poolname := GetUUID()
		err := conn.MakePool(poolname)
		require.NoError(t, err)
		defer conn.DeletePool(poolname)

		ioctx, err := conn.OpenIOContext(poolname)
		require.NoError(t, err)
		defer ioctx.Destroy()

		name := GetUUID()
		err = quickCreate(ioctx, name, testImageSize, testImageOrder)
		require.NoError(t, err)
		defer func() { assert.NoError(t, RemoveImage(ioctx, name)) }()

		img, err := OpenImage(ioctx, name, NoSnapshot)
		require.NoError(t, err)
		defer func() { assert.NoError(t, img.Close()) }()

		err = img.FlattenWithProgress(nil, nil)
		assert.Error(t, err)
	})
}
