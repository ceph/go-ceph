// +build !luminous,!mimic

package rbd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNamespace(t *testing.T) {
	conn := radosConnect(t)

	poolName := GetUUID()
	err := conn.MakePool(poolName)
	assert.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolName)
	assert.NoError(t, err)

	defer func() {
		ioctx.Destroy()
		conn.DeletePool(poolName)
		conn.Shutdown()
	}()

	t.Run("invalidInputNamespace", func(t *testing.T) {
		// NamespaceCreate.
		err := NamespaceCreate(nil, "someName")
		assert.Error(t, err)
		err = NamespaceCreate(ioctx, "")
		assert.Error(t, err)

		// NamespaceRemove.
		err = NamespaceRemove(nil, "someName")
		assert.Error(t, err)
		err = NamespaceRemove(ioctx, "")
		assert.Error(t, err)

		// NamespaceExists.
		_, err = NamespaceExists(nil, "someName")
		assert.Error(t, err)
		_, err = NamespaceExists(ioctx, "")
		assert.Error(t, err)
	})

	t.Run("CreateNamespace", func(t *testing.T) {
		nameSpace := "myNamespace"
		err := NamespaceCreate(ioctx, nameSpace)
		assert.NoError(t, err)

		// Check whether it exists or not.
		val, err := NamespaceExists(ioctx, nameSpace)
		assert.NoError(t, err)
		assert.Equal(t, val, true)

		// Create again with same name.
		err = NamespaceCreate(ioctx, nameSpace)
		assert.Error(t, err)

		// Remove the namespace.
		err = NamespaceRemove(ioctx, nameSpace)
		assert.NoError(t, err)
	})

	t.Run("NonExistingNamespace", func(t *testing.T) {
		// Try to remove.
		err := NamespaceRemove(ioctx, "someNamespace")
		assert.Error(t, err)

		// Check the existence.
		val, err := NamespaceExists(ioctx, "someNamespace")
		assert.NoError(t, err)
		assert.Equal(t, val, false)
	})
}
