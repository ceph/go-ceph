// +build !luminous,!mimic

package rbd

import (
	"math/rand"
	"testing"

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
