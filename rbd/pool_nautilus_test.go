// +build !luminous,!mimic

package rbd

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPoolMetadata(t *testing.T) {
	conn := radosConnect(t)
	poolName := GetUUID()
	err := conn.MakePool(poolName)
	assert.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolName)
	assert.NoError(t, err)
	assert.NotNil(t, ioctx)

	defer func() {
		ioctx.Destroy()
		conn.DeletePool(poolName)
		conn.Shutdown()
	}()

	t.Run("GetWrongKey", func(t *testing.T) {
		_, err := GetPoolMetadata(ioctx, "someKey")
		assert.Error(t, err)
	})

	t.Run("NullKey", func(t *testing.T) {
		_, err := GetPoolMetadata(ioctx, "")
		assert.Error(t, err)
		err = SetPoolMetadata(ioctx, "", "")
		assert.NoError(t, err)
		err = RemovePoolMetadata(ioctx, "")
		assert.NoError(t, err)
	})

	t.Run("NullIOContext", func(t *testing.T) {
		_, err := GetPoolMetadata(nil, "someKey")
		assert.Error(t, err)
		err = SetPoolMetadata(nil, "someKey", "someValue")
		assert.Error(t, err)
		err = RemovePoolMetadata(nil, "someKey")
		assert.Error(t, err)
	})

	t.Run("SetGetValues", func(t *testing.T) {
		var (
			key1 = "key1"
			val1 = "val1"
		)

		err := SetPoolMetadata(ioctx, key1, val1)
		assert.NoError(t, err)
		metaVal, err := GetPoolMetadata(ioctx, key1)
		assert.NoError(t, err)
		assert.Equal(t, val1, metaVal)
	})

	t.Run("largeSetValue", func(t *testing.T) {
		keyLen := 5004
		var charValues = "_:$%&/()"

		bytes := make([]byte, keyLen)
		for i := 0; i < keyLen; i++ {
			bytes[i] = charValues[rand.Intn(len(charValues))]
		}
		myKey := "myKey"
		err := SetPoolMetadata(ioctx, myKey, string(bytes))
		assert.NoError(t, err)

		myVal, err := GetPoolMetadata(ioctx, myKey)
		assert.NoError(t, err)
		assert.Equal(t, keyLen, len(myVal))
	})

	t.Run("removeNonExistingKey", func(t *testing.T) {
		err := RemovePoolMetadata(ioctx, "someKey")
		assert.Error(t, err)
	})

	t.Run("removeExistingKey", func(t *testing.T) {
		var (
			myKey = "myKey"
			myVal = "myVal"
		)
		assert.NoError(t, SetPoolMetadata(ioctx, myKey, myVal))
		_, err := GetPoolMetadata(ioctx, myKey)
		assert.NoError(t, err)

		// Remove the key.
		err = RemovePoolMetadata(ioctx, myKey)
		assert.NoError(t, err)
		_, err = GetPoolMetadata(ioctx, myKey)
		assert.Error(t, err)
	})
}

func TestPoolInit(t *testing.T) {
	conn := radosConnect(t)
	poolName := GetUUID()
	err := conn.MakePool(poolName)
	assert.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolName)
	assert.NoError(t, err)
	assert.NotNil(t, ioctx)

	defer func() {
		ioctx.Destroy()
		conn.DeletePool(poolName)
		conn.Shutdown()
	}()

	t.Run("NullIOContext", func(t *testing.T) {
		err := PoolInit(nil, true)
		assert.Error(t, err)
	})

	t.Run("PoolInitWithForce", func(t *testing.T) {
		err := PoolInit(ioctx, true)
		assert.NoError(t, err)
	})

	t.Run("PoolInitWithoutForce", func(t *testing.T) {
		err := PoolInit(ioctx, false)
		assert.NoError(t, err)
	})
}

func TestGetAllPoolStat(t *testing.T) {
	conn := radosConnect(t)
	poolName := GetUUID()
	err := conn.MakePool(poolName)
	assert.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolName)
	assert.NoError(t, err)
	assert.NotNil(t, ioctx)

	defer func() {
		ioctx.Destroy()
		conn.DeletePool(poolName)
		conn.Shutdown()
	}()

	poolstats := poolStatsCreate()
	defer func() {
		poolstats.destroy()
	}()

	var imageName string
	size := uint64(2 << 20)
	var expectedSize uint64

	for idx := 0; idx < 3; idx++ {
		imageName = GetUUID()
		image, err := Create(ioctx, imageName, size, testImageOrder)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, image.Remove())
		}()
		expectedSize += size
	}

	imageName = GetUUID()
	_, err = Create(ioctx, imageName, size, testImageOrder)
	assert.NoError(t, err)
	expectedSize += size
	img, err := OpenImage(ioctx, imageName, NoSnapshot)
	assert.NoError(t, err)

	mySnap, err := img.CreateSnapshot("mySnap")
	assert.NoError(t, err)

	t.Run("NullIOContext", func(t *testing.T) {
		_, err = GetAllPoolStats(nil)
		assert.Error(t, err)
	})

	t.Run("CheckPoolStatOption", func(t *testing.T) {
		err := img.Resize(size)
		assert.NoError(t, err)
		assert.NoError(t, mySnap.Remove())
		assert.NoError(t, img.Close())
		err = img.Trash(time.Hour)
		assert.NoError(t, err)
		defer func() {
			trashList, err := GetTrashList(ioctx)
			assert.NoError(t, err)
			assert.NoError(t, TrashRemove(ioctx, trashList[0].Id, true))
		}()

		omap, err := GetAllPoolStats(ioctx)
		assert.NoError(t, err)

		assert.Equal(t, uint64(3), omap[PoolStatOptionImages])
		assert.Equal(t, (expectedSize - size), omap[PoolStatOptionImageProvisionedBytes])
		assert.Equal(t, expectedSize-size, omap[PoolStatOptionImageMaxProvisionedBytes])
		assert.Equal(t, uint64(0), omap[PoolStatOptionImageSnapshots])
		assert.Equal(t, uint64(0), omap[PoolStatOptionTrashSnapshots])
		assert.Equal(t, uint64(1), omap[PoolStatOptionTrashImages])
		assert.Equal(t, size, omap[PoolStatOptionTrashProvisionedBytes])
		assert.Equal(t, size, omap[PoolStatOptionTrashMaxProvisionedBytes])
	})
}
