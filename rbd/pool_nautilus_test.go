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
}
