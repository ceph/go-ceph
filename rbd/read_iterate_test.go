package rbd

import (
	"fmt"
	_ "sync"
	"testing"
	_ "time"

	"github.com/ceph/go-ceph/rados"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadIterate(t *testing.T) {
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
		testReadIterateBasic(t, ioctx)
	})
}

func testReadIterateBasic(t *testing.T, ioctx *rados.IOContext) {
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

	_, err = img.WriteAt([]byte("mary had a little lamb"), 0)
	assert.NoError(t, err)
	_, err = img.WriteAt([]byte("it's fleece was white as #FFFFF"), 2048)
	assert.NoError(t, err)

	//_, err = img.Discard(0, 1<<23)
	//assert.NoError(t, err)
	assert.NoError(t, img.Close())
	img, err = OpenImage(ioctx, name, NoSnapshot)

	err = img.ReadIterate(ReadIterateConfig{
		Offset: 0,
		Length: isize,
		Callback: func(o, l uint64, b []byte, d interface{}) int {
			if b == nil {
				fmt.Println("ZZZ", o, l)
				return 0
			}
			fmt.Println("QQQ", o, l, b)
			return 0
		},
	})
	assert.NoError(t, err)
}
