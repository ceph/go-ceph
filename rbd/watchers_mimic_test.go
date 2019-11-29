// +build !luminous
//
// Ceph Mimic is the first version that supports watchers through librbd.

package rbd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListWatchers(t *testing.T) {
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

	name := GetUUID()
	options := NewRbdImageOptions()
	err = CreateImage(ioctx, name, 1<<22, options)
	require.NoError(t, err)
	defer func() { assert.NoError(t, RemoveImage(ioctx, name)) }()

	t.Run("imageNotOpen", func(t *testing.T) {
		image, err := OpenImageReadOnly(ioctx, name, NoSnapshot)
		require.NoError(t, err)
		require.NotNil(t, image)

		err = image.Close()
		require.NoError(t, err)

		_, err = image.ListWatchers()
		assert.Equal(t, ErrImageNotOpen, err)
	})

	t.Run("noWatchers", func(t *testing.T) {
		// open image read-only, as OpenImage() automatically adds a watcher
		image, err := OpenImageReadOnly(ioctx, name, NoSnapshot)
		require.NoError(t, err)
		require.NotNil(t, image)
		defer func() { assert.NoError(t, image.Close()) }()

		watchers, err := image.ListWatchers()
		assert.NoError(t, err)
		assert.Equal(t, 0, len(watchers))
	})

	t.Run("addWatchers", func(t *testing.T) {
		// open image read-only, as OpenImage() automatically adds a watcher
		image, err := OpenImageReadOnly(ioctx, name, NoSnapshot)
		require.NoError(t, err)
		require.NotNil(t, image)
		defer func() { assert.NoError(t, image.Close()) }()

		watchers, err := image.ListWatchers()
		assert.NoError(t, err)
		assert.Equal(t, 0, len(watchers))

		// opening an image writable adds a watcher automatically
		image2, err := OpenImage(ioctx, name, NoSnapshot)
		require.NoError(t, err)
		require.NotNil(t, image2)

		watchers, err = image.ListWatchers()
		assert.NoError(t, err)
		assert.Equal(t, 1, len(watchers))

		watchers, err = image2.ListWatchers()
		assert.NoError(t, err)
		assert.Equal(t, 1, len(watchers))

		image3, err := OpenImage(ioctx, name, NoSnapshot)
		require.NoError(t, err)
		require.NotNil(t, image3)

		watchers, err = image.ListWatchers()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(watchers))

		watchers, err = image2.ListWatchers()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(watchers))

		watchers, err = image3.ListWatchers()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(watchers))

		// closing an image removes the watchers
		err = image3.Close()
		require.NoError(t, err)

		watchers, err = image.ListWatchers()
		assert.NoError(t, err)
		assert.Equal(t, 1, len(watchers))

		watchers, err = image2.ListWatchers()
		assert.NoError(t, err)
		assert.Equal(t, 1, len(watchers))

		err = image2.Close()
		require.NoError(t, err)

		watchers, err = image.ListWatchers()
		assert.NoError(t, err)
		assert.Equal(t, 0, len(watchers))
	})
}
