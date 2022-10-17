package rbd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRenameSnapshot(t *testing.T) {
	conn := radosConnect(t)
	poolName := GetUUID()
	err := conn.MakePool(poolName)
	require.NoError(t, err)
	ioctx, err := conn.OpenIOContext(poolName)
	require.NoError(t, err)

	name := GetUUID()
	err = CreateImage(ioctx, name, 1<<22, NewRbdImageOptions())
	require.NoError(t, err)

	// create snapshot
	img, err := OpenImage(ioctx, name, NoSnapshot)
	require.NoError(t, err)
	snapshotName := "mysnap"
	snapshot, err := img.CreateSnapshot(snapshotName)
	require.NoError(t, err)

	// verify snapshot opens
	snapImg, err := OpenImage(ioctx, name, snapshotName)
	require.NoError(t, err)
	err = snapImg.Close()
	require.NoError(t, err)

	// rename snapshot
	newSnapshotName := "myrenamedsnap"
	err = snapshot.Rename(newSnapshotName)
	require.NoError(t, err)

	// verify snapshot still opens
	snapImg, err = OpenImage(ioctx, name, newSnapshotName)
	require.NoError(t, err)
	err = snapImg.Close()
	require.NoError(t, err)

	err = snapshot.Remove()
	require.NoError(t, err)

	err = img.Close()
	require.NoError(t, err)
	err = img.Remove()
	require.NoError(t, err)

	ioctx.Destroy()
	err = conn.DeletePool(poolName)
	require.NoError(t, err)
	conn.Shutdown()
}
