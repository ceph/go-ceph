package rbd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGroupSnapshots(t *testing.T) {
	// tests are done as subtests to avoid creating pools, images, etc
	// over and over again.
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

	// create a group, some images, and add images to the group
	gname := "snapme"
	err = GroupCreate(ioctx, gname)
	assert.NoError(t, err)

	options := NewRbdImageOptions()
	assert.NoError(t,
		options.SetUint64(ImageOptionOrder, uint64(testImageOrder)))

	name1 := GetUUID()
	err = CreateImage(ioctx, name1, testImageSize, options)
	require.NoError(t, err)

	name2 := GetUUID()
	err = CreateImage(ioctx, name2, testImageSize, options)
	require.NoError(t, err)

	err = GroupImageAdd(ioctx, gname, ioctx, name1)
	assert.NoError(t, err)
	err = GroupImageAdd(ioctx, gname, ioctx, name2)
	assert.NoError(t, err)

	t.Run("groupSnapCreateRemove", func(t *testing.T) {
		err := GroupSnapCreate(ioctx, gname, "snap1")
		assert.NoError(t, err)
		err = GroupSnapRemove(ioctx, gname, "snap1")
		assert.NoError(t, err)
	})
	t.Run("groupSnapRename", func(t *testing.T) {
		err := GroupSnapCreate(ioctx, gname, "snap2a")
		assert.NoError(t, err)
		err = GroupSnapRename(ioctx, gname, "fred", "wilma")
		assert.Error(t, err)
		err = GroupSnapRename(ioctx, gname, "snap2a", "snap2b")
		assert.NoError(t, err)
		err = GroupSnapRemove(ioctx, gname, "snap2a")
		assert.Error(t, err, "remove of old name: expect error")
		err = GroupSnapRemove(ioctx, gname, "snap2b")
		assert.NoError(t, err, "remove of current name: expect success")
	})
	t.Run("groupSnappList", func(t *testing.T) {
		err := GroupSnapCreate(ioctx, gname, "snap1")
		assert.NoError(t, err)
		err = GroupSnapCreate(ioctx, gname, "snap2")
		assert.NoError(t, err)
		err = GroupSnapCreate(ioctx, gname, "snap3")
		assert.NoError(t, err)

		gsl, err := GroupSnapList(ioctx, gname)
		assert.NoError(t, err)
		if assert.Len(t, gsl, 3) {
			names := []string{}
			for _, gsi := range gsl {
				assert.Equal(t, GroupSnapStateComplete, gsi.State)
				names = append(names, gsi.Name)
			}
			assert.Contains(t, names, "snap1")
			assert.Contains(t, names, "snap2")
			assert.Contains(t, names, "snap3")
		}

		err = GroupSnapRemove(ioctx, gname, "snap3")
		assert.NoError(t, err)
		err = GroupSnapRemove(ioctx, gname, "snap2")
		assert.NoError(t, err)
		err = GroupSnapRename(ioctx, gname, "snap1", "snap1a")

		gsl, err = GroupSnapList(ioctx, gname)
		assert.NoError(t, err)
		if assert.Len(t, gsl, 1) {
			assert.Equal(t, GroupSnapStateComplete, gsl[0].State)
			assert.Equal(t, "snap1a", gsl[0].Name)
		}

		err = GroupSnapRemove(ioctx, gname, "snap1a")
		assert.NoError(t, err)

		gsl, err = GroupSnapList(ioctx, gname)
		assert.NoError(t, err)
		assert.Len(t, gsl, 0)
	})
	t.Run("groupSnapRollback", func(t *testing.T) {
		img, err := OpenImage(ioctx, name1, NoSnapshot)
		assert.NoError(t, err)
		_, err = img.WriteAt([]byte("HELLO WORLD"), 0)
		assert.NoError(t, err)
		err = img.Close()
		assert.NoError(t, err)

		snapname := "snap1"
		err = GroupSnapCreate(ioctx, gname, snapname)
		assert.NoError(t, err)

		img, err = OpenImage(ioctx, name1, NoSnapshot)
		assert.NoError(t, err)
		_, err = img.WriteAt([]byte("GOODBYE WORLD"), 0)
		assert.NoError(t, err)
		err = img.Close()
		assert.NoError(t, err)

		img, err = OpenImage(ioctx, name2, NoSnapshot)
		assert.NoError(t, err)
		_, err = img.WriteAt([]byte("2222222222222"), 0)
		assert.NoError(t, err)
		err = img.Close()
		assert.NoError(t, err)

		err = GroupSnapRollback(ioctx, gname, snapname)
		assert.NoError(t, err)

		b := make([]byte, 8)
		img, err = OpenImage(ioctx, name1, NoSnapshot)
		assert.NoError(t, err)
		_, err = img.ReadAt(b, 0)
		assert.NoError(t, err)
		err = img.Close()
		assert.NoError(t, err)
		assert.Equal(t, []byte("HELLO WO"), b)

		img, err = OpenImage(ioctx, name2, NoSnapshot)
		assert.NoError(t, err)
		_, err = img.ReadAt(b, 0)
		assert.NoError(t, err)
		err = img.Close()
		assert.NoError(t, err)
		assert.Equal(t, []byte("\x00\x00\x00\x00\x00\x00\x00\x00"), b)

		err = GroupSnapRemove(ioctx, gname, snapname)
		assert.NoError(t, err)
	})
	t.Run("groupSnapRollbackWithProgress", func(t *testing.T) {
		img, err := OpenImage(ioctx, name1, NoSnapshot)
		assert.NoError(t, err)
		_, err = img.WriteAt([]byte("SAY CHEESE"), 0)
		assert.NoError(t, err)
		_, err = img.WriteAt([]byte("AND SMILE_"), 10240)
		assert.NoError(t, err)
		err = img.Close()
		assert.NoError(t, err)

		snapname := "snap2r"
		err = GroupSnapCreate(ioctx, gname, snapname)
		assert.NoError(t, err)

		img, err = OpenImage(ioctx, name1, NoSnapshot)
		assert.NoError(t, err)
		_, err = img.WriteAt([]byte("GOODBYE WORLD"), 0)
		assert.NoError(t, err)
		_, err = img.WriteAt([]byte("_THIS_IS_ALL_"), 10240)
		assert.NoError(t, err)
		_, err = img.WriteAt([]byte("I_HAVE_TO_SAY"), 11240)
		assert.NoError(t, err)
		err = img.Close()
		assert.NoError(t, err)

		img, err = OpenImage(ioctx, name2, NoSnapshot)
		assert.NoError(t, err)
		_, err = img.WriteAt([]byte("3333333333333"), 0)
		assert.NoError(t, err)
		err = img.Close()
		assert.NoError(t, err)

		cc := 0
		cb := func(offset, total uint64, v interface{}) int {
			cc++
			val := v.(int)
			assert.Equal(t, 0, val)
			assert.Equal(t, uint64(1), total)
			return 0
		}
		err = GroupSnapRollbackWithProgress(ioctx, gname, snapname, cb, 0)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, cc, 2)

		b := make([]byte, 8)
		img, err = OpenImage(ioctx, name1, NoSnapshot)
		assert.NoError(t, err)
		_, err = img.ReadAt(b, 0)
		assert.NoError(t, err)
		err = img.Close()
		assert.NoError(t, err)
		assert.Equal(t, []byte("SAY CHEE"), b)

		img, err = OpenImage(ioctx, name2, NoSnapshot)
		assert.NoError(t, err)
		_, err = img.ReadAt(b, 0)
		assert.NoError(t, err)
		err = img.Close()
		assert.NoError(t, err)
		assert.Equal(t, []byte("\x00\x00\x00\x00\x00\x00\x00\x00"), b)

		err = GroupSnapRemove(ioctx, gname, snapname)
		assert.NoError(t, err)
	})
	t.Run("invalidIOContext", func(t *testing.T) {
		assert.Panics(t, func() {
			GroupSnapCreate(nil, gname, "foo")
		})
		assert.Panics(t, func() {
			GroupSnapRemove(nil, gname, "foo")
		})
		assert.Panics(t, func() {
			GroupSnapRename(nil, gname, "foo", "bar")
		})
		assert.Panics(t, func() {
			GroupSnapList(nil, gname)
		})
		assert.Panics(t, func() {
			GroupSnapRollback(nil, gname, "foo")
		})
		assert.Panics(t, func() {
			cb := func(o, t uint64, v interface{}) int { return 0 }
			GroupSnapRollbackWithProgress(nil, gname, "foo", cb, nil)
		})
	})
	t.Run("invalidCallback", func(t *testing.T) {
		err := GroupSnapRollbackWithProgress(ioctx, gname, "foo", nil, nil)
		assert.Error(t, err)
	})
}
