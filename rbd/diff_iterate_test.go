package rbd

import (
	"sync"
	"testing"
	"time"

	"github.com/ceph/go-ceph/rados"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiffIterate(t *testing.T) {
	conn := radosConnect(t)
	defer conn.Shutdown()

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	assert.NoError(t, err)
	defer conn.DeletePool(poolname)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)
	defer ioctx.Destroy()

	t.Run("basic", func(t *testing.T) {
		testDiffIterateBasic(t, ioctx)
	})
	t.Run("twoAtOnce", func(t *testing.T) {
		testDiffIterateTwoAtOnce(t, ioctx)
	})
	t.Run("earlyExit", func(t *testing.T) {
		testDiffIterateEarlyExit(t, ioctx)
	})
	t.Run("snapshot", func(t *testing.T) {
		testDiffIterateSnapshot(t, ioctx)
	})
	t.Run("callbackData", func(t *testing.T) {
		testDiffIterateCallbackData(t, ioctx)
	})
	t.Run("badImage", func(t *testing.T) {
		var gotCalled int
		img := GetImage(ioctx, "bob")
		err := img.DiffIterate(
			DiffIterateConfig{
				Offset: 0,
				Length: uint64(1 << 22),
				Callback: func(o, l uint64, e int, x interface{}) int {
					gotCalled++
					return 0
				},
			})
		assert.Error(t, err)
		assert.EqualValues(t, 0, gotCalled)
	})
	t.Run("missingCallback", func(t *testing.T) {
		name := GetUUID()
		isize := uint64(1 << 23) // 8MiB
		iorder := 20             // 1MiB
		options := NewRbdImageOptions()
		defer options.Destroy()
		assert.NoError(t,
			options.SetUint64(RbdImageOptionOrder, uint64(iorder)))
		err := CreateImage(ioctx, name, isize, options)
		assert.NoError(t, err)

		img, err := OpenImage(ioctx, name, NoSnapshot)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, img.Close())
			assert.NoError(t, img.Remove())
		}()

		var gotCalled int
		err = img.DiffIterate(
			DiffIterateConfig{
				Offset: 0,
				Length: uint64(1 << 22),
			})
		assert.Error(t, err)
		assert.EqualValues(t, 0, gotCalled)
	})
}

func testDiffIterateBasic(t *testing.T, ioctx *rados.IOContext) {
	name := GetUUID()
	isize := uint64(1 << 23) // 8MiB
	iorder := 20             // 1MiB
	options := NewRbdImageOptions()
	defer options.Destroy()
	assert.NoError(t,
		options.SetUint64(RbdImageOptionOrder, uint64(iorder)))
	err := CreateImage(ioctx, name, isize, options)
	assert.NoError(t, err)

	img, err := OpenImage(ioctx, name, NoSnapshot)
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, img.Close())
		assert.NoError(t, img.Remove())
	}()

	type diResult struct {
		offset uint64
		length uint64
	}
	calls := []diResult{}

	err = img.DiffIterate(
		DiffIterateConfig{
			Offset: 0,
			Length: isize,
			Callback: func(o, l uint64, e int, x interface{}) int {
				calls = append(calls, diResult{offset: o, length: l})
				return 0
			},
		})
	assert.NoError(t, err)
	// Image is new, empty. Callback will not be called
	assert.Len(t, calls, 0)

	_, err = img.WriteAt([]byte("sometimes you feel like a nut"), 0)
	assert.NoError(t, err)

	err = img.DiffIterate(
		DiffIterateConfig{
			Offset: 0,
			Length: isize,
			Callback: func(o, l uint64, e int, x interface{}) int {
				calls = append(calls, diResult{offset: o, length: l})
				return 0
			},
		})
	assert.NoError(t, err)
	if assert.Len(t, calls, 1) {
		assert.EqualValues(t, 0, calls[0].offset)
		assert.EqualValues(t, 29, calls[0].length)
	}

	_, err = img.WriteAt([]byte("sometimes you don't"), 32)
	assert.NoError(t, err)

	calls = []diResult{}
	err = img.DiffIterate(
		DiffIterateConfig{
			Offset: 0,
			Length: isize,
			Callback: func(o, l uint64, e int, x interface{}) int {
				calls = append(calls, diResult{offset: o, length: l})
				return 0
			},
		})
	if assert.NoError(t, err) {
		assert.Len(t, calls, 1)
		assert.EqualValues(t, 0, calls[0].offset)
		assert.EqualValues(t, 51, calls[0].length)
	}

	// dirty a 2nd chunk
	newOffset := 3145728 // 3MiB
	_, err = img.WriteAt([]byte("alright, alright, alright"), int64(newOffset))
	assert.NoError(t, err)

	calls = []diResult{}
	err = img.DiffIterate(
		DiffIterateConfig{
			Offset: 0,
			Length: isize,
			Callback: func(o, l uint64, e int, x interface{}) int {
				calls = append(calls, diResult{offset: o, length: l})
				return 0
			},
		})
	assert.NoError(t, err)
	if assert.Len(t, calls, 2) {
		assert.EqualValues(t, 0, calls[0].offset)
		assert.EqualValues(t, 51, calls[0].length)
		assert.EqualValues(t, newOffset, calls[1].offset)
		assert.EqualValues(t, 25, calls[1].length)
	}

	// dirty a 3rd chunk
	newOffset2 := 5242880 + 1024 // 5MiB + 1KiB
	_, err = img.WriteAt([]byte("zowie!"), int64(newOffset2))
	assert.NoError(t, err)

	calls = []diResult{}
	err = img.DiffIterate(
		DiffIterateConfig{
			Offset: 0,
			Length: isize,
			Callback: func(o, l uint64, e int, x interface{}) int {
				calls = append(calls, diResult{offset: o, length: l})
				return 0
			},
		})
	assert.NoError(t, err)
	if assert.Len(t, calls, 3) {
		assert.EqualValues(t, 0, calls[0].offset)
		assert.EqualValues(t, 51, calls[0].length)
		assert.EqualValues(t, newOffset, calls[1].offset)
		assert.EqualValues(t, 25, calls[1].length)
		assert.EqualValues(t, newOffset2-1024, calls[2].offset)
		assert.EqualValues(t, 6+1024, calls[2].length)
	}
}

// testDiffIterateTwoAtOnce aims to verify that multiple DiffIterate
// callbacks can be executed at the same time without error.
func testDiffIterateTwoAtOnce(t *testing.T, ioctx *rados.IOContext) {
	isize := uint64(1 << 23) // 8MiB
	iorder := 20             // 1MiB
	options := NewRbdImageOptions()
	defer options.Destroy()
	assert.NoError(t,
		options.SetUint64(RbdImageOptionOrder, uint64(iorder)))

	name1 := GetUUID()
	err := CreateImage(ioctx, name1, isize, options)
	assert.NoError(t, err)

	img1, err := OpenImage(ioctx, name1, NoSnapshot)
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, img1.Close())
		assert.NoError(t, img1.Remove())
	}()

	name2 := GetUUID()
	err = CreateImage(ioctx, name2, isize, options)
	assert.NoError(t, err)

	img2, err := OpenImage(ioctx, name2, NoSnapshot)
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, img2.Close())
		assert.NoError(t, img2.Remove())
	}()

	type diResult struct {
		offset uint64
		length uint64
	}

	diffTest := func(wg *sync.WaitGroup, inbuf []byte, img *Image) {
		_, err = img.WriteAt(inbuf[0:3], 0)
		assert.NoError(t, err)
		_, err = img.WriteAt(inbuf[3:6], 3145728)
		assert.NoError(t, err)
		_, err = img.WriteAt(inbuf[6:9], 5242880)
		assert.NoError(t, err)

		calls := []diResult{}
		err = img.DiffIterate(
			DiffIterateConfig{
				Offset: 0,
				Length: isize,
				Callback: func(o, l uint64, e int, x interface{}) int {
					time.Sleep(8 * time.Millisecond)
					calls = append(calls, diResult{offset: o, length: l})
					return 0
				},
			})
		assert.NoError(t, err)
		if assert.Len(t, calls, 3) {
			assert.EqualValues(t, 0, calls[0].offset)
			assert.EqualValues(t, 3, calls[0].length)
			assert.EqualValues(t, 3145728, calls[1].offset)
			assert.EqualValues(t, 3, calls[1].length)
			assert.EqualValues(t, 5242880, calls[2].offset)
			assert.EqualValues(t, 3, calls[2].length)
		}

		wg.Done()
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go diffTest(wg, []byte("foobarbaz"), img1)
	wg.Add(1)
	go diffTest(wg, []byte("abcdefghi"), img2)
	wg.Wait()
}

// testDiffIterateEarlyExit checks that returning an error from the callback
// function triggers the DiffIterate call to stop.
func testDiffIterateEarlyExit(t *testing.T, ioctx *rados.IOContext) {
	isize := uint64(1 << 23) // 8MiB
	iorder := 20             // 1MiB
	options := NewRbdImageOptions()
	defer options.Destroy()
	assert.NoError(t,
		options.SetUint64(RbdImageOptionOrder, uint64(iorder)))

	name := GetUUID()
	err := CreateImage(ioctx, name, isize, options)
	assert.NoError(t, err)

	img, err := OpenImage(ioctx, name, NoSnapshot)
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, img.Close())
		assert.NoError(t, img.Remove())
	}()

	type diResult struct {
		offset uint64
		length uint64
	}

	// "damage" the image
	inbuf := []byte("xxxyyyzzz")
	_, err = img.WriteAt(inbuf[0:3], 0)
	assert.NoError(t, err)
	_, err = img.WriteAt(inbuf[3:6], 3145728)
	assert.NoError(t, err)
	_, err = img.WriteAt(inbuf[6:9], 5242880)
	assert.NoError(t, err)

	// if the offset is less than zero the callback will return an "error" and
	// that will abort the DiffIterate call early and it will return the error
	// code our callback used.
	calls := []diResult{}
	err = img.DiffIterate(
		DiffIterateConfig{
			Offset: 0,
			Length: isize,
			Callback: func(o, l uint64, e int, x interface{}) int {
				if o > 1 {
					return -5
				}
				calls = append(calls, diResult{offset: o, length: l})
				return 0
			},
		})
	assert.Error(t, err)
	if rbderr, ok := err.(RBDError); assert.True(t, ok) {
		assert.EqualValues(t, -5, int(rbderr))
	}
	if assert.Len(t, calls, 1) {
		assert.EqualValues(t, 0, calls[0].offset)
		assert.EqualValues(t, 3, calls[0].length)
	}
}

func testDiffIterateSnapshot(t *testing.T, ioctx *rados.IOContext) {
	name := GetUUID()
	isize := uint64(1 << 23) // 8MiB
	iorder := 20             // 1MiB
	options := NewRbdImageOptions()
	defer options.Destroy()
	assert.NoError(t,
		options.SetUint64(RbdImageOptionOrder, uint64(iorder)))
	err := CreateImage(ioctx, name, isize, options)
	assert.NoError(t, err)

	img, err := OpenImage(ioctx, name, NoSnapshot)
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, img.Close())
		assert.NoError(t, img.Remove())
	}()

	type diResult struct {
		offset uint64
		length uint64
	}
	calls := []diResult{}

	err = img.DiffIterate(
		DiffIterateConfig{
			Offset: 0,
			Length: isize,
			Callback: func(o, l uint64, e int, x interface{}) int {
				calls = append(calls, diResult{offset: o, length: l})
				return 0
			},
		})
	assert.NoError(t, err)
	// Image is new, empty. Callback will not be called
	assert.Len(t, calls, 0)

	_, err = img.WriteAt([]byte("sometimes you feel like a nut"), 0)
	assert.NoError(t, err)

	calls = []diResult{}
	err = img.DiffIterate(
		DiffIterateConfig{
			Offset: 0,
			Length: isize,
			Callback: func(o, l uint64, e int, x interface{}) int {
				calls = append(calls, diResult{offset: o, length: l})
				return 0
			},
		})
	assert.NoError(t, err)
	if assert.Len(t, calls, 1) {
		assert.EqualValues(t, 0, calls[0].offset)
		assert.EqualValues(t, 29, calls[0].length)
	}

	ss1, err := img.CreateSnapshot("ss1")
	assert.NoError(t, err)
	defer func() { assert.NoError(t, ss1.Remove()) }()

	// there should be no differences between "now" and "ss1"
	calls = []diResult{}
	err = img.DiffIterate(
		DiffIterateConfig{
			SnapName: "ss1",
			Offset:   0,
			Length:   isize,
			Callback: func(o, l uint64, e int, x interface{}) int {
				calls = append(calls, diResult{offset: o, length: l})
				return 0
			},
		})
	assert.NoError(t, err)
	assert.Len(t, calls, 0)

	// this next check was shamelessly cribbed from the pybind
	// tests for diff_iterate out of the ceph tree
	// it discards the current image, makes a 2nd snap, and compares
	// the diff between snapshots 1 & 2.
	_, err = img.Discard(0, isize)
	assert.NoError(t, err)

	ss2, err := img.CreateSnapshot("ss2")
	assert.NoError(t, err)
	defer func() { assert.NoError(t, ss2.Remove()) }()
	err = ss2.Set() // caution: this side-effects img!
	assert.NoError(t, err)

	calls = []diResult{}
	err = img.DiffIterate(
		DiffIterateConfig{
			SnapName: "ss1",
			Offset:   0,
			Length:   isize,
			Callback: func(o, l uint64, e int, x interface{}) int {
				calls = append(calls, diResult{offset: o, length: l})
				return 0
			},
		})
	assert.NoError(t, err)
	if assert.Len(t, calls, 1) {
		assert.EqualValues(t, 0, calls[0].offset)
		assert.EqualValues(t, 29, calls[0].length)
	}
}

func testDiffIterateCallbackData(t *testing.T, ioctx *rados.IOContext) {
	name := GetUUID()
	isize := uint64(1 << 23) // 8MiB
	iorder := 20             // 1MiB
	options := NewRbdImageOptions()
	defer options.Destroy()
	assert.NoError(t,
		options.SetUint64(RbdImageOptionOrder, uint64(iorder)))
	err := CreateImage(ioctx, name, isize, options)
	assert.NoError(t, err)

	img, err := OpenImage(ioctx, name, NoSnapshot)
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, img.Close())
		assert.NoError(t, img.Remove())
	}()

	type diResult struct {
		offset uint64
		length uint64
	}
	calls := []diResult{}

	_, err = img.WriteAt([]byte("sometimes you feel like a nut"), 0)
	assert.NoError(t, err)

	err = img.DiffIterate(
		DiffIterateConfig{
			Offset: 0,
			Length: isize,
			Callback: func(o, l uint64, e int, x interface{}) int {
				if v, ok := x.(int); ok {
					assert.EqualValues(t, 77, v)
				} else {
					t.Fatalf("incorrect type")
				}
				calls = append(calls, diResult{offset: o, length: l})
				return 0
			},
			Data: 77,
		})
	assert.NoError(t, err)
	if assert.Len(t, calls, 1) {
		assert.EqualValues(t, 0, calls[0].offset)
		assert.EqualValues(t, 29, calls[0].length)
	}

	calls = []diResult{}
	err = img.DiffIterate(
		DiffIterateConfig{
			Offset: 0,
			Length: isize,
			Callback: func(o, l uint64, e int, x interface{}) int {
				if v, ok := x.(string); ok {
					assert.EqualValues(t, "bob", v)
				} else {
					t.Fatalf("incorrect type")
				}
				calls = append(calls, diResult{offset: o, length: l})
				return 0
			},
			Data: "bob",
		})
	assert.NoError(t, err)
	if assert.Len(t, calls, 1) {
		assert.EqualValues(t, 0, calls[0].offset)
		assert.EqualValues(t, 29, calls[0].length)
	}
}
