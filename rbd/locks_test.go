//go:build !nautilus && ceph_preview
// +build !nautilus,ceph_preview

package rbd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocking(t *testing.T) {
	conn := radosConnect(t)
	require.NotNil(t, conn)
	defer conn.Shutdown()

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	require.NoError(t, err)
	defer conn.DeletePool(poolname)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)
	defer ioctx.Destroy()

	// Connect another cluster
	conn1 := radosConnect(t)
	require.NotNil(t, conn1)
	defer conn1.Shutdown()

	ioctx1, err := conn1.OpenIOContext(poolname)
	require.NoError(t, err)
	defer ioctx1.Destroy()

	options := NewRbdImageOptions()
	defer options.Destroy()
	assert.NoError(t, options.SetUint64(ImageOptionOrder, uint64(testImageOrder)))

	name := GetUUID()
	err = CreateImage(ioctx, name, testImageSize, options)
	assert.NoError(t, err)

	t.Run("acquireLock", func(t *testing.T) {
		img, err := OpenImage(ioctx, name, NoSnapshot)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, img.Close())
		}()

		// Shared lock mode is not supported
		// Ref: Check lock_acquire() logic in ceph codebase at https://github.com/ceph/ceph/blob/main/src/librbd/internal.cc
		err = img.LockAcquire(LockModeShared)
		assert.Error(t, err)

		err = img.LockAcquire(LockModeExclusive)
		assert.NoError(t, err)

		isOwner, err := img.LockIsExclusiveOwner()
		assert.NoError(t, err)
		assert.True(t, isOwner)

		err = img.LockBreak(LockModeExclusive, "not owner")
		assert.Error(t, err)

		err = img.LockRelease()
		assert.NoError(t, err)
	})

	t.Run("listLock", func(t *testing.T) {
		img, err := OpenImage(ioctx, name, NoSnapshot)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, img.Close())
		}()

		img1, err := OpenImage(ioctx1, name, NoSnapshot)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, img1.Close())
		}()

		err = img1.LockAcquire(LockModeExclusive)
		assert.NoError(t, err)

		err = img.LockAcquire(LockModeExclusive)
		assert.Error(t, err)

		isOwner, err := img.LockIsExclusiveOwner()
		assert.NoError(t, err)
		assert.False(t, isOwner)

		// This logic of fetching the lock owners and breaking the lock using a
		// different image is borrowed from the TCs in ceph codebase, check
		// the TEST_F(TestLibRBD, BreakLock) at https://github.com/ceph/ceph/blob/main/src/test/librbd/test_librbd.cc
		locksList, err := img.LockGetOwners()
		assert.NoError(t, err)
		assert.Equal(t, len(locksList), 1)
		assert.Equal(t, locksList[0].Mode, LockModeExclusive)

		err = img.LockBreak(LockModeExclusive, locksList[0].Owner)
		assert.NoError(t, err)

		err = img.LockAcquire(LockModeExclusive)
		assert.NoError(t, err)

		err = img.LockRelease()
		assert.NoError(t, err)
	})

	t.Run("emptyListLock", func(t *testing.T) {
		img, err := OpenImage(ioctx, name, NoSnapshot)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, img.Close())
		}()

		locksList, err := img.LockGetOwners()
		assert.NoError(t, err)
		assert.Equal(t, len(locksList), 0)
	})

	t.Run("closedImage", func(t *testing.T) {
		img, err := OpenImage(ioctx, name, NoSnapshot)
		assert.NoError(t, err)
		assert.NoError(t, img.Close())
		defer func() {
			assert.NoError(t, img.Remove())
		}()

		err = img.LockAcquire(LockModeExclusive)
		assert.Error(t, err)

		err = img.LockBreak(LockModeExclusive, "")
		assert.Error(t, err)

		_, err = img.LockGetOwners()
		assert.Error(t, err)

		_, err = img.LockIsExclusiveOwner()
		assert.Error(t, err)

		err = img.LockRelease()
		assert.Error(t, err)
	})
}
