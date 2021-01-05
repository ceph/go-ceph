package rbd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGroupCreateRemove(t *testing.T) {
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

	err = GroupCreate(ioctx, "group1")
	assert.NoError(t, err)

	err = GroupRemove(ioctx, "group1")
	assert.NoError(t, err)

	err = GroupRemove(ioctx, "group2")
	assert.NoError(t, err)

	err = GroupCreate(ioctx, "group2")
	assert.NoError(t, err)
	err = GroupCreate(ioctx, "group")
	assert.NoError(t, err)

	err = GroupRemove(ioctx, "group2")
	assert.NoError(t, err)
}

func TestGroupRename(t *testing.T) {
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

	err = GroupCreate(ioctx, "group1")
	assert.NoError(t, err)

	err = GroupRename(ioctx, "group1", "club1")
	assert.NoError(t, err)

	err = GroupRemove(ioctx, "club1")
	assert.NoError(t, err)

	// unlike remove, rename does return an error if the src name
	// doesn't exist
	err = GroupRename(ioctx, "club1", "nowhere")
	assert.Error(t, err)
}
