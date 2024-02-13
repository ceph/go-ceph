package rbd

import (
	"sync"
	"testing"
	"time"

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

func TestWatch(t *testing.T) {
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

	startSize := uint64(1 << 21)
	name := GetUUID()
	options := NewRbdImageOptions()
	err = CreateImage(ioctx, name, startSize, options)
	require.NoError(t, err)
	defer func() { assert.NoError(t, RemoveImage(ioctx, name)) }()

	t.Run("imageNotOpen", func(t *testing.T) {
		image, err := OpenImageReadOnly(ioctx, name, NoSnapshot)
		require.NoError(t, err)
		require.NotNil(t, image)

		err = image.Close()
		require.NoError(t, err)

		_, err = image.UpdateWatch(func(_ interface{}) {
		}, nil)
		assert.Equal(t, ErrImageNotOpen, err)
	})

	t.Run("simpleWatch", func(t *testing.T) {
		image, err := OpenImage(ioctx, name, NoSnapshot)
		require.NoError(t, err)
		require.NotNil(t, image)

		defer func() {
			assert.NoError(t, image.Close())
		}()

		cc := 0
		w, err := image.UpdateWatch(func(_ interface{}) {
			cc++
		}, nil)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, w.Unwatch())
		}()

		x := make(chan int)
		defer close(x)
		go func() {
			for i := 0; i < 5; i++ {
				i1, err := OpenImage(ioctx, name, NoSnapshot)
				err = i1.Resize(startSize * uint64(1+i))
				assert.NoError(t, err)
				err = i1.Close()
				assert.NoError(t, err)
				time.Sleep(5 * time.Millisecond)
			}
			x <- 0
		}()
		<-x

		assert.Equal(t, 5, cc)
	})

	t.Run("badWatch", func(t *testing.T) {
		w := &Watch{}
		err := w.Unwatch()
		assert.Error(t, err)

		i1, err := OpenImage(ioctx, name, NoSnapshot)
		assert.NoError(t, err)
		assert.NoError(t, i1.Close())
		w.image = i1
		err = w.Unwatch()
		assert.Error(t, err)
	})
}

func TestWatchMultiChannels(t *testing.T) {
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

	startSize := uint64(1 << 21)
	images := map[string]uint64{}
	for i := 0; i < 4; i++ {
		name := GetUUID()
		images[name] = startSize
		options := NewRbdImageOptions()
		err = CreateImage(ioctx, name, startSize, options)
		require.NoError(t, err)
		defer func() {
			assert.NoError(t, RemoveImage(ioctx, name))
		}()
	}

	var wg sync.WaitGroup
	watchMe := func(n string, mon chan<- string, done <-chan bool) {
		img, err := OpenImage(ioctx, n, NoSnapshot)
		require.NoError(t, err)
		defer func() {
			assert.NoError(t, img.Close())
		}()
		w, err := img.UpdateWatch(func(_ interface{}) {
			mon <- n
		}, nil)
		require.NoError(t, err)
		defer func() {
			assert.NoError(t, w.Unwatch())
		}()

		wg.Done() // our watch is set up and active
		<-done
	}

	mon := make(chan string)
	defer close(mon)
	done := make(chan bool, len(images))
	defer close(done)
	wg.Add(len(images))
	for n := range images {
		go watchMe(n, mon, done)
	}
	wg.Wait()

	x := make(chan bool)
	go func() {
		inames := []string{}
		for n := range images {
			inames = append(inames, n)
		}
		for i := 0; i < 12; i++ {
			n := inames[i%len(inames)]
			images[n] *= uint64(2)
			i1, err := OpenImage(ioctx, n, NoSnapshot)
			err = i1.Resize(images[n])
			assert.NoError(t, err)
			err = i1.Close()
			assert.NoError(t, err)
			time.Sleep(5 * time.Millisecond)
		}
		time.Sleep(5 * time.Millisecond)
		for range inames {
			done <- true
		}
		x <- true
	}()

	ncount := map[string]int{}
	for ok := true; ok; {
		select {
		case n := <-mon:
			ncount[n]++
		case <-x:
			ok = false
			break
		}
	}

	assert.Len(t, ncount, 4)
	for _, v := range ncount {
		assert.Equal(t, 3, v)
	}
}
