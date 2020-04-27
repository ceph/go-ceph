package rbd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"testing"
	"time"

	"github.com/ceph/go-ceph/rados"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testImageSize  = uint64(1 << 22)
	testImageOrder = 22
)

func GetUUID() string {
	return uuid.Must(uuid.NewV4()).String()
}

func TestVersion(t *testing.T) {
	var major, minor, patch = Version()
	assert.False(t, major < 0 || major > 1000, "invalid major")
	assert.False(t, minor < 0 || minor > 1000, "invalid minor")
	assert.False(t, patch < 0 || patch > 1000, "invalid patch")
}

func radosConnect(t *testing.T) *rados.Conn {
	conn, err := rados.NewConn()
	require.NoError(t, err)
	err = conn.ReadDefaultConfigFile()
	require.NoError(t, err)

	timeout := time.After(time.Second * 5)
	ch := make(chan error)
	go func(conn *rados.Conn) {
		ch <- conn.Connect()
	}(conn)
	select {
	case err = <-ch:
	case <-timeout:
		err = fmt.Errorf("timed out waiting for connect")
	}
	require.NoError(t, err)
	return conn
}

func TestImageCreate(t *testing.T) {
	conn := radosConnect(t)

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	assert.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)

	name := GetUUID()
	image, err := Create(ioctx, name, testImageSize, testImageOrder)
	assert.NoError(t, err)
	err = image.Remove()
	assert.NoError(t, err)

	name = GetUUID()
	image, err = Create(ioctx, name, testImageSize, testImageOrder,
		FeatureLayering|FeatureStripingV2)
	assert.NoError(t, err)
	err = image.Remove()
	assert.NoError(t, err)

	name = GetUUID()
	image, err = Create(ioctx, name, testImageSize, testImageOrder,
		FeatureLayering|FeatureStripingV2, 4096, 2)
	assert.NoError(t, err)
	err = image.Remove()
	assert.NoError(t, err)

	// invalid order
	name = GetUUID()
	_, err = Create(ioctx, name, testImageSize, -1)
	assert.Error(t, err)

	// too many arguments
	_, err = Create(ioctx, name, testImageSize, testImageOrder,
		FeatureLayering|FeatureStripingV2, 4096, 2, 123)
	assert.Error(t, err)

	ioctx.Destroy()
	conn.DeletePool(poolname)
	conn.Shutdown()
}

func TestImageCreate2(t *testing.T) {
	conn := radosConnect(t)

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	assert.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolname)
	assert.NoError(t, err)

	name := GetUUID()
	image, err := Create2(ioctx, name, testImageSize,
		FeatureLayering|FeatureStripingV2, testImageOrder)
	assert.NoError(t, err)
	err = image.Remove()
	assert.NoError(t, err)

	ioctx.Destroy()
	conn.DeletePool(poolname)
	conn.Shutdown()
}

func TestImageCreate3(t *testing.T) {
	conn := radosConnect(t)

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	assert.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolname)
	assert.NoError(t, err)

	name := GetUUID()
	image, err := Create3(ioctx, name, testImageSize,
		FeatureLayering|FeatureStripingV2, testImageOrder, 4096, 2)
	assert.NoError(t, err)
	err = image.Remove()
	assert.NoError(t, err)

	ioctx.Destroy()
	conn.DeletePool(poolname)
	conn.Shutdown()
}

func TestCreateImageWithOptions(t *testing.T) {
	conn := radosConnect(t)

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	assert.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolname)
	assert.NoError(t, err)

	// nil options, causes a panic if not handled correctly
	name := GetUUID()
	err = CreateImage(ioctx, name, testImageSize, nil)
	assert.Error(t, err)

	options := NewRbdImageOptions()

	// empty/default options
	name = GetUUID()
	err = CreateImage(ioctx, name, testImageSize, options)
	assert.NoError(t, err)
	err = RemoveImage(ioctx, name)
	assert.NoError(t, err)

	// create image with RbdImageOptionOrder
	err = options.SetUint64(ImageOptionOrder, uint64(testImageOrder))
	assert.NoError(t, err)
	name = GetUUID()
	err = CreateImage(ioctx, name, testImageSize, options)
	assert.NoError(t, err)
	err = RemoveImage(ioctx, name)
	assert.NoError(t, err)
	options.Clear()

	// create image with a different data pool
	datapool := GetUUID()
	err = conn.MakePool(datapool)
	assert.NoError(t, err)
	err = options.SetString(ImageOptionDataPool, datapool)
	assert.NoError(t, err)
	name = GetUUID()
	err = CreateImage(ioctx, name, testImageSize, options)
	assert.NoError(t, err)
	err = RemoveImage(ioctx, name)
	assert.NoError(t, err)
	conn.DeletePool(datapool)

	// cleanup
	options.Destroy()
	ioctx.Destroy()
	conn.DeletePool(poolname)
	conn.Shutdown()
}

func TestGetImageNames(t *testing.T) {
	conn := radosConnect(t)

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	assert.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)

	createdList := []string{}
	for i := 0; i < 10; i++ {
		name := GetUUID()
		err = quickCreate(ioctx, name, testImageSize, testImageOrder)
		assert.NoError(t, err)
		createdList = append(createdList, name)
	}

	imageNames, err := GetImageNames(ioctx)
	assert.NoError(t, err)

	sort.Strings(createdList)
	sort.Strings(imageNames)
	assert.Equal(t, createdList, imageNames)

	for _, name := range createdList {
		img := GetImage(ioctx, name)
		err := img.Remove()
		assert.NoError(t, err)
	}

	ioctx.Destroy()
	conn.DeletePool(poolname)
	conn.Shutdown()
}

func TestDeprecatedImageOpen(t *testing.T) {
	conn := radosConnect(t)

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	assert.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)

	name := GetUUID()
	image, err := Create(ioctx, name, testImageSize, testImageOrder)
	assert.NoError(t, err)

	// an integer is not a valid argument
	err = image.Open(123)
	assert.Error(t, err)

	// open read-write
	err = image.Open()
	assert.NoError(t, err)
	err = image.Close()
	assert.NoError(t, err)

	// open read-only
	err = image.Open(true)
	assert.NoError(t, err)

	bytesIn := []byte("input data")
	_, err = image.Write(bytesIn)
	// writing should fail in read-only mode
	assert.Error(t, err)

	err = image.Close()
	assert.NoError(t, err)

	err = image.Remove()
	assert.NoError(t, err)

	ioctx.Destroy()
	conn.DeletePool(poolname)
	conn.Shutdown()
}

func TestImageResize(t *testing.T) {
	conn := radosConnect(t)

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	assert.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)

	name := GetUUID()
	reqSize := uint64(1024 * 1024 * 4) // 4MB
	err = quickCreate(ioctx, name, reqSize, testImageOrder)
	assert.NoError(t, err)

	image, err := OpenImage(ioctx, name, NoSnapshot)
	assert.NoError(t, err)

	size, err := image.GetSize()
	assert.NoError(t, err)
	assert.Equal(t, size, reqSize)

	err = image.Resize(reqSize * 2)
	assert.NoError(t, err)

	size, err = image.GetSize()
	assert.NoError(t, err)
	assert.Equal(t, size, reqSize*2)

	err = image.Close()
	assert.NoError(t, err)

	err = image.Resize(reqSize * 2)
	assert.Equal(t, err, ErrImageNotOpen)

	err = image.Remove()
	assert.NoError(t, err)

	ioctx.Destroy()
	conn.DeletePool(poolname)
	conn.Shutdown()
}

func TestImageProperties(t *testing.T) {
	conn := radosConnect(t)

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	require.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)

	name := GetUUID()
	reqSize := uint64(1024 * 1024 * 4) // 4MB
	_, err = Create3(ioctx, name, reqSize,
		FeatureLayering|FeatureStripingV2, testImageOrder, 4096, 2)
	require.NoError(t, err)

	img, err := OpenImage(ioctx, name, NoSnapshot)
	require.NoError(t, err)

	format, err := img.IsOldFormat()
	assert.NoError(t, err)
	assert.Equal(t, format, false)

	size, err := img.GetSize()
	assert.NoError(t, err)
	assert.Equal(t, size, reqSize)

	features, err := img.GetFeatures()
	assert.NoError(t, err)
	// compare features with the two requested ones
	assert.Equal(t, features&(FeatureLayering|FeatureStripingV2),
		FeatureLayering|FeatureStripingV2)

	stripeUnit, err := img.GetStripeUnit()
	assert.NoError(t, err)
	assert.Equal(t, stripeUnit, uint64(4096))

	stripeCount, err := img.GetStripeCount()
	assert.NoError(t, err)
	assert.Equal(t, stripeCount, uint64(2))

	_, err = img.GetOverlap()
	assert.NoError(t, err)

	err = img.Close()
	assert.NoError(t, err)

	err = img.Remove()
	assert.NoError(t, err)

	ioctx.Destroy()
	conn.DeletePool(poolname)
	conn.Shutdown()
}

func TestImageRename(t *testing.T) {
	conn := radosConnect(t)

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	assert.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)

	name := GetUUID()
	err = quickCreate(ioctx, name, testImageSize, testImageOrder)
	assert.NoError(t, err)

	img := GetImage(ioctx, name)
	err = img.Rename(name)
	assert.Error(t, err)

	err = img.Rename(GetUUID())
	assert.NoError(t, err)

	err = img.Remove()
	assert.NoError(t, err)

	ioctx.Destroy()
	conn.DeletePool(poolname)
	conn.Shutdown()
}

func TestImageSeek(t *testing.T) {
	conn := radosConnect(t)

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	assert.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)

	name := GetUUID()
	err = quickCreate(ioctx, name, testImageSize, testImageOrder)
	assert.NoError(t, err)

	img, err := OpenImage(ioctx, name, NoSnapshot)
	assert.NoError(t, err)

	_, err = img.Seek(0, SeekSet)
	assert.NoError(t, err)

	bytesIn := []byte("input data")
	nIn, err := img.Write(bytesIn)
	assert.NoError(t, err)
	assert.Equal(t, nIn, len(bytesIn))

	pos, err := img.Seek(0, SeekCur)
	assert.NoError(t, err)
	assert.Equal(t, pos, int64(nIn))

	pos, err = img.Seek(0, SeekSet)
	assert.NoError(t, err)
	assert.Equal(t, pos, int64(0))

	bytesOut := make([]byte, len(bytesIn))
	nOut, err := img.Read(bytesOut)
	assert.NoError(t, err)
	assert.Equal(t, nOut, len(bytesOut))
	assert.Equal(t, bytesIn, bytesOut)

	pos, err = img.Seek(0, SeekCur)
	assert.NoError(t, err)
	assert.Equal(t, pos, int64(nOut))

	pos, err = img.Seek(0, SeekSet)
	assert.NoError(t, err)
	assert.Equal(t, pos, int64(0))

	pos, err = img.Seek(0, SeekEnd)
	assert.NoError(t, err)
	assert.Equal(t, pos, int64(testImageSize))

	_, err = img.Seek(0, -1)
	assert.Error(t, err)

	err = img.Close()
	assert.NoError(t, err)

	_, err = img.Seek(0, SeekEnd)
	assert.Equal(t, err, ErrImageNotOpen)

	err = img.Remove()
	assert.NoError(t, err)

	ioctx.Destroy()
	conn.DeletePool(poolname)
	conn.Shutdown()
}

func TestImageDiscard(t *testing.T) {
	conn := radosConnect(t)

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	assert.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)

	name := GetUUID()
	err = quickCreate(ioctx, name, testImageSize, testImageOrder)
	assert.NoError(t, err)

	img, err := OpenImage(ioctx, name, NoSnapshot)
	assert.NoError(t, err)

	n, err := img.Discard(0, 1<<16)
	assert.NoError(t, err)
	assert.Equal(t, n, 1<<16)

	err = img.Close()
	assert.NoError(t, err)

	img, err = OpenImageReadOnly(ioctx, name, NoSnapshot)
	assert.NoError(t, err)

	// when read-only, discard should fail
	_, err = img.Discard(0, 1<<16)
	assert.Error(t, err)

	err = img.Close()
	assert.NoError(t, err)

	_, err = img.Discard(0, 1<<16)
	assert.Equal(t, err, ErrImageNotOpen)

	err = img.Remove()
	assert.NoError(t, err)

	ioctx.Destroy()
	conn.DeletePool(poolname)
	conn.Shutdown()
}

func TestIOReaderWriter(t *testing.T) {
	conn := radosConnect(t)

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	assert.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)

	name := GetUUID()
	err = quickCreate(ioctx, name, testImageSize, testImageOrder)
	assert.NoError(t, err)

	img, err := OpenImage(ioctx, name, NoSnapshot)
	assert.NoError(t, err)

	stats, err := img.Stat()
	assert.NoError(t, err)

	encoder := json.NewEncoder(img)
	assert.NoError(t, encoder.Encode(stats))

	err = img.Flush()
	assert.NoError(t, err)

	_, err = img.Seek(0, SeekSet)
	assert.NoError(t, err)

	var stats2 *ImageInfo
	decoder := json.NewDecoder(img)
	assert.NoError(t, decoder.Decode(&stats2))

	assert.Equal(t, &stats, &stats2)

	_, err = img.Seek(0, SeekSet)
	assert.NoError(t, err)

	bytesIn := []byte("input data")
	_, err = img.Write(bytesIn)
	assert.NoError(t, err)

	_, err = img.Seek(0, SeekSet)
	assert.NoError(t, err)

	// reading 0 bytes should succeed
	nilBytes := make([]byte, 0)
	nOut, err := img.Read(nilBytes)
	assert.Equal(t, nOut, 0)
	assert.NoError(t, err)

	bytesOut := make([]byte, len(bytesIn))
	nOut, err = img.Read(bytesOut)

	assert.NoError(t, err)
	assert.Equal(t, nOut, len(bytesIn))
	assert.Equal(t, bytesIn, bytesOut)

	// write some data at the end of the image
	offset := int64(stats.Size) - int64(len(bytesIn))

	_, err = img.Seek(offset, SeekSet)
	assert.NoError(t, err)

	nOut, err = img.Write(bytesIn)
	assert.Equal(t, len(bytesIn), nOut)
	assert.NoError(t, err)

	_, err = img.Seek(offset, SeekSet)
	assert.NoError(t, err)

	bytesOut = make([]byte, len(bytesIn))
	nOut, err = img.Read(bytesOut)
	assert.Equal(t, nOut, len(bytesIn))
	assert.Equal(t, bytesIn, bytesOut)
	assert.NoError(t, err)

	// reading after EOF (needs to be large enough to hit EOF)
	_, err = img.Seek(offset, SeekSet)
	assert.NoError(t, err)

	bytesIn = make([]byte, len(bytesOut)+256)
	nOut, err = img.Read(bytesIn)
	assert.Equal(t, nOut, len(bytesOut))
	assert.Equal(t, bytesIn[0:len(bytesOut)], bytesOut)
	assert.Equal(t, io.EOF, err)

	err = img.Close()
	assert.NoError(t, err)

	err = img.Remove()
	assert.NoError(t, err)

	ioctx.Destroy()
	conn.DeletePool(poolname)
	conn.Shutdown()
}

func TestReadAt(t *testing.T) {
	conn := radosConnect(t)

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	assert.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)

	name := GetUUID()
	options := NewRbdImageOptions()
	defer options.Destroy()
	err = options.SetUint64(ImageOptionOrder, uint64(testImageOrder))
	assert.NoError(t, err)
	err = CreateImage(ioctx, name, testImageSize, options)
	require.NoError(t, err)

	img, err := OpenImage(ioctx, name, NoSnapshot)
	assert.NoError(t, err)

	// write 0 bytes should succeed
	dataOut := make([]byte, 0)
	nOut, err := img.WriteAt(dataOut, 256)
	assert.Equal(t, 0, nOut)
	assert.NoError(t, err)

	// reading 0 bytes should be successful
	dataIn := make([]byte, 0)
	nIn, err := img.ReadAt(dataIn, 256)
	assert.Equal(t, 0, nIn)
	assert.NoError(t, err)

	// write some data at the end of the image
	dataOut = []byte("Hi rbd! Nice to talk through go-ceph :)")

	stats, err := img.Stat()
	require.NoError(t, err)
	offset := int64(stats.Size) - int64(len(dataOut))

	nOut, err = img.WriteAt(dataOut, offset)
	assert.Equal(t, len(dataOut), nOut)
	assert.NoError(t, err)

	dataIn = make([]byte, len(dataOut))
	nIn, err = img.ReadAt(dataIn, offset)
	assert.Equal(t, nIn, len(dataIn))
	assert.Equal(t, dataIn, dataOut)
	assert.NoError(t, err)

	// reading after EOF (needs to be large enough to hit EOF)
	dataIn = make([]byte, len(dataOut)+256)
	nIn, err = img.ReadAt(dataIn, offset)
	assert.Equal(t, nIn, len(dataOut))
	assert.Equal(t, dataIn[0:len(dataOut)], dataOut)
	assert.Equal(t, io.EOF, err)

	err = img.Close()
	assert.NoError(t, err)

	// writing to a read-only image should fail
	img, err = OpenImageReadOnly(ioctx, name, NoSnapshot)
	assert.NoError(t, err)

	_, err = img.WriteAt(dataOut, 256)
	assert.Error(t, err)

	err = img.Close()
	assert.NoError(t, err)

	err = img.Remove()
	assert.NoError(t, err)

	ioctx.Destroy()
	conn.DeletePool(poolname)
	conn.Shutdown()
}

func TestImageCopy(t *testing.T) {
	conn := radosConnect(t)

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	assert.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)

	t.Run("invalidParameters", func(t *testing.T) {
		name := GetUUID()
		options := NewRbdImageOptions()
		defer options.Destroy()
		err = options.SetUint64(ImageOptionOrder, uint64(testImageOrder))
		assert.NoError(t, err)
		err = CreateImage(ioctx, name, testImageSize, options)
		require.NoError(t, err)

		// img not open, should fail
		img := GetImage(ioctx, name)
		err = img.Copy(nil, "")
		assert.Equal(t, err, ErrImageNotOpen)

		img, err := OpenImage(ioctx, name, NoSnapshot)
		require.NoError(t, err)

		// pass invalid parameters
		err = img.Copy(nil, "")
		assert.Error(t, err) // order of errors not enforced

		err = img.Copy(ioctx, "")
		assert.Equal(t, err, ErrNoName)

		err = img.Copy(nil, "duplicate")
		assert.Equal(t, err, ErrNoIOContext)

		err = img.Close()
		assert.NoError(t, err)
		err = RemoveImage(ioctx, name)
		assert.NoError(t, err)
	})

	// try successful copying
	t.Run("successfulCopy", func(t *testing.T) {
		name := GetUUID()
		options := NewRbdImageOptions()
		defer options.Destroy()
		err = options.SetUint64(ImageOptionOrder, uint64(testImageOrder))
		assert.NoError(t, err)
		err = CreateImage(ioctx, name, testImageSize, options)
		require.NoError(t, err)

		img, err := OpenImage(ioctx, name, NoSnapshot)
		require.NoError(t, err)

		name2 := GetUUID()
		err = img.Copy(ioctx, name2)
		require.NoError(t, err)

		img2, err := OpenImage(ioctx, name2, NoSnapshot)
		require.NoError(t, err)

		err = img2.Close()
		assert.NoError(t, err)

		err = img2.Remove()
		assert.NoError(t, err)

		err = img.Close()
		assert.NoError(t, err)
	})

	t.Run("copy2ImageNotOpen", func(t *testing.T) {
		name := GetUUID()
		name2 := GetUUID()
		img := GetImage(ioctx, name)
		img2 := GetImage(ioctx, name2)

		err = img.Copy2(img2)
		assert.Equal(t, err, ErrImageNotOpen)

		options := NewRbdImageOptions()
		defer options.Destroy()
		err = options.SetUint64(ImageOptionOrder, uint64(testImageOrder))
		assert.NoError(t, err)
		err = CreateImage(ioctx, name, testImageSize, options)
		require.NoError(t, err)
		img, err = OpenImage(ioctx, name, NoSnapshot)
		assert.NoError(t, err)

		err = img.Copy2(img2)
		assert.Equal(t, err, ErrImageNotOpen)

		err = img.Close()
		assert.NoError(t, err)
		err = RemoveImage(ioctx, name)
		assert.NoError(t, err)
	})

	t.Run("successfulCopy2", func(t *testing.T) {
		name := GetUUID()
		name2 := GetUUID()

		options := NewRbdImageOptions()
		defer options.Destroy()
		err = options.SetUint64(ImageOptionOrder, uint64(testImageOrder))
		assert.NoError(t, err)
		err = CreateImage(ioctx, name, testImageSize, options)
		require.NoError(t, err)
		img, err := OpenImage(ioctx, name, NoSnapshot)
		assert.NoError(t, err)

		err = CreateImage(ioctx, name2, testImageSize, options)
		require.NoError(t, err)
		img2, err := OpenImage(ioctx, name2, NoSnapshot)
		assert.NoError(t, err)

		err = img.Copy2(img2)
		require.NoError(t, err)

		err = img2.Close()
		assert.NoError(t, err)

		err = img2.Remove()
		assert.NoError(t, err)

		err = img.Close()
		assert.NoError(t, err)

		err = img.Remove()
		assert.NoError(t, err)
	})

	ioctx.Destroy()
	conn.DeletePool(poolname)
	conn.Shutdown()
}

func TestParentInfo(t *testing.T) {
	conn := radosConnect(t)

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	assert.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)

	name := "parent"
	img, err := Create(ioctx, name, testImageSize, testImageOrder, 1)
	assert.NoError(t, err)

	img, err = OpenImage(ioctx, name, NoSnapshot)
	assert.NoError(t, err)

	snapshot, err := img.CreateSnapshot("mysnap")
	assert.NoError(t, err)

	err = snapshot.Protect()
	assert.NoError(t, err)

	// create an image context with the parent+snapshot
	snapImg, err := OpenImage(ioctx, name, "mysnap")
	assert.NoError(t, err)

	// ensure no children prior to clone
	pools, images, err := snapImg.ListChildren()
	assert.NoError(t, err)
	assert.Equal(t, len(pools), 0, "pools equal")
	assert.Equal(t, len(images), 0, "children length equal")

	// invalid order, should fail
	_, err = img.Clone("mysnap", ioctx, "child", 1, -1)
	assert.Error(t, err)

	_, err = img.Clone("mysnap", ioctx, "child", 1, testImageOrder)
	assert.NoError(t, err)

	imgNew, err := OpenImage(ioctx, "child", NoSnapshot)
	assert.NoError(t, err)
	parentPool := make([]byte, 128)
	parentName := make([]byte, 128)
	parentSnapname := make([]byte, 128)

	err = imgNew.GetParentInfo(parentPool, parentName, parentSnapname)
	assert.NoError(t, err)

	n := bytes.Index(parentName, []byte{0})
	pName := string(parentName[:n])

	n = bytes.Index(parentSnapname, []byte{0})
	pSnapname := string(parentSnapname[:n])
	assert.Equal(t, pName, "parent", "they should be equal")
	assert.Equal(t, pSnapname, "mysnap", "they should be equal")

	pools, images, err = snapImg.ListChildren()
	assert.NoError(t, err)
	assert.Equal(t, len(pools), 1, "pools equal")
	assert.Equal(t, len(images), 1, "children length equal")

	err = imgNew.Close()
	assert.NoError(t, err)

	err = imgNew.Remove()
	assert.NoError(t, err)

	err = snapshot.Unprotect()
	assert.NoError(t, err)

	err = snapshot.Remove()
	assert.NoError(t, err)

	err = img.Close()
	assert.NoError(t, err)

	err = snapImg.Close()
	assert.NoError(t, err)

	err = img.Remove()
	assert.NoError(t, err)

	ioctx.Destroy()
	conn.DeletePool(poolname)
	conn.Shutdown()
}

func TestNoIOContext(t *testing.T) {
	image := GetImage(nil, "nonexistent")

	_, err := image.Clone("new snapshot", nil, "clone", 0, 0)
	assert.Equal(t, err, ErrNoIOContext)

	err = image.Remove()
	assert.Equal(t, err, ErrNoIOContext)

	err = image.Trash(15 * time.Second)
	assert.Equal(t, err, ErrNoIOContext)

	err = image.Rename("unknown")
	assert.Equal(t, err, ErrNoIOContext)

	err = image.Open()
	assert.Equal(t, err, ErrNoIOContext)
}

func TestErrorNoName(t *testing.T) {
	image := GetImage(nil, "")

	err := image.Remove()
	assert.Equal(t, err, ErrNoName)

	err = image.Trash(15 * time.Second)
	assert.Equal(t, err, ErrNoName)

	err = image.Rename("unknown")
	assert.Equal(t, err, ErrNoName)

	err = image.Open()
	assert.Equal(t, err, ErrNoName)
}

func TestErrorImageNotOpen(t *testing.T) {
	image := GetImage(nil, "nonexistent")

	err := image.Close()
	assert.Equal(t, err, ErrImageNotOpen)

	err = image.Resize(2 << 22)
	assert.Equal(t, err, ErrImageNotOpen)

	_, err = image.Stat()
	assert.Equal(t, err, ErrImageNotOpen)

	_, err = image.IsOldFormat()
	assert.Equal(t, err, ErrImageNotOpen)

	_, err = image.GetSize()
	assert.Equal(t, err, ErrImageNotOpen)

	_, err = image.GetFeatures()
	assert.Equal(t, err, ErrImageNotOpen)

	_, err = image.GetStripeUnit()
	assert.Equal(t, err, ErrImageNotOpen)

	_, err = image.GetStripeCount()
	assert.Equal(t, err, ErrImageNotOpen)

	_, err = image.GetOverlap()
	assert.Equal(t, err, ErrImageNotOpen)

	err = image.Flatten()
	assert.Equal(t, err, ErrImageNotOpen)

	_, _, err = image.ListChildren()
	assert.Equal(t, err, ErrImageNotOpen)

	_, _, err = image.ListLockers()
	assert.Equal(t, err, ErrImageNotOpen)

	err = image.LockExclusive("a magic cookie")
	assert.Equal(t, err, ErrImageNotOpen)

	err = image.LockShared("a magic cookie", "tasty")
	assert.Equal(t, err, ErrImageNotOpen)

	err = image.BreakLock("a magic cookie", "tasty")
	assert.Equal(t, err, ErrImageNotOpen)

	_, err = image.Read(nil)
	assert.Equal(t, err, ErrImageNotOpen)

	_, err = image.Write(nil)
	assert.Equal(t, err, ErrImageNotOpen)

	_, err = image.ReadAt(nil, 0)
	assert.Equal(t, err, ErrImageNotOpen)

	_, err = image.WriteAt(nil, 0)
	assert.Equal(t, err, ErrImageNotOpen)

	err = image.Flush()
	assert.Equal(t, err, ErrImageNotOpen)
}

func TestNotFound(t *testing.T) {
	conn := radosConnect(t)

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	assert.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)

	name := GetUUID()

	img, err := OpenImage(ioctx, name, NoSnapshot)
	assert.Equal(t, err, ErrNotFound)
	assert.Nil(t, img)

	ioctx.Destroy()
	conn.DeletePool(poolname)
	conn.Shutdown()
}

func TestTrashImage(t *testing.T) {
	conn := radosConnect(t)

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	assert.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)

	name := GetUUID()
	err = quickCreate(ioctx, name, testImageSize, testImageOrder)
	assert.NoError(t, err)

	image := GetImage(ioctx, name)
	err = image.Trash(time.Hour)
	assert.NoError(t, err)

	trashList, err := GetTrashList(ioctx)
	assert.NoError(t, err)
	assert.Equal(t, len(trashList), 1, "trashList length equal")

	err = TrashRestore(ioctx, trashList[0].Id, name+"_restored")
	assert.NoError(t, err)

	image2 := GetImage(ioctx, name+"_restored")
	err = image2.Trash(time.Hour)
	assert.NoError(t, err)

	trashList, err = GetTrashList(ioctx)
	assert.NoError(t, err)
	assert.Equal(t, len(trashList), 1, "trashList length equal")

	err = TrashRemove(ioctx, trashList[0].Id, true)
	assert.NoError(t, err)

	ioctx.Destroy()
	conn.DeletePool(poolname)
	conn.Shutdown()
}

func TestImageMetadata(t *testing.T) {
	metadataKey := "mykey"
	metadataValue := "myvalue"

	conn := radosConnect(t)

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	assert.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)

	name := GetUUID()
	err = quickCreate(ioctx, name, testImageSize, testImageOrder)
	assert.NoError(t, err)

	image := GetImage(ioctx, name)
	// Set a metadata key/value on unopen image
	err = image.SetMetadata(metadataKey, metadataValue)
	assert.Equal(t, err, ErrImageNotOpen)
	// Get the metadata value on unopen image
	value, err := image.GetMetadata(metadataKey)
	assert.Equal(t, err, ErrImageNotOpen)
	assert.Equal(t, "", value)
	// Remove the metadata key on unopen image
	err = image.RemoveMetadata(metadataKey)
	assert.Equal(t, err, ErrImageNotOpen)
	// check key is removed on unopen image
	value, err = image.GetMetadata(metadataKey)
	assert.Equal(t, "", value)
	assert.Equal(t, err, ErrImageNotOpen)

	image, err = OpenImage(ioctx, name, NoSnapshot)
	assert.NoError(t, err)

	// Set a metadata key/value
	err = image.SetMetadata(metadataKey, metadataValue)
	assert.NoError(t, err)
	// Get the metadata value
	value, err = image.GetMetadata(metadataKey)
	assert.NoError(t, err)
	assert.Equal(t, metadataValue, value)
	// Remove the metadata key
	err = image.RemoveMetadata(metadataKey)
	assert.NoError(t, err)
	// check key is removed
	value, err = image.GetMetadata(metadataKey)
	assert.Equal(t, "", value)
	assert.Error(t, err)

	err = image.Close()
	assert.NoError(t, err)
	err = image.Remove()
	assert.NoError(t, err)

	ioctx.Destroy()
	conn.DeletePool(poolname)
	conn.Shutdown()
}

func TestClosedImage(t *testing.T) {
	t.Skipf("many of the following functions cause a panic or hang, skip this test")

	conn := radosConnect(t)

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	assert.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)

	name := GetUUID()
	err = quickCreate(ioctx, name, testImageSize, testImageOrder)
	assert.NoError(t, err)

	image, err := OpenImage(ioctx, name, NoSnapshot)
	assert.NoError(t, err)

	// keep the rbdImage around after close
	rbdImage := image.image

	// close the image
	err = image.Close()
	assert.NoError(t, err)

	// restore the image so image.validate() succeeds
	image.image = rbdImage

	// functions should now fail with an RBDError

	err = image.Close()
	assert.Error(t, err)

	err = image.Resize(1 << 22)
	assert.Error(t, err)

	_, err = image.Stat()
	assert.Error(t, err)

	_, err = image.IsOldFormat()
	assert.Error(t, err)

	_, err = image.GetSize()
	assert.Error(t, err)

	_, err = image.GetFeatures()
	assert.Error(t, err)

	_, err = image.GetStripeUnit()
	assert.Error(t, err)

	_, err = image.GetStripeCount()
	assert.Error(t, err)

	_, err = image.GetOverlap()
	assert.Error(t, err)

	err = image.Flatten()
	assert.Error(t, err)

	_, _, err = image.ListChildren()
	assert.Error(t, err)

	err = image.Flush()
	assert.Error(t, err)

	_, err = image.GetSnapshotNames()
	assert.Error(t, err)

	_, err = image.CreateSnapshot("new_snapshot")
	assert.Error(t, err)

	_, err = image.GetMetadata("metadata-key")
	assert.Error(t, err)

	err = image.SetMetadata("metadata-key", "metadata-value")
	assert.Error(t, err)

	err = image.RemoveMetadata("metadata-key")
	assert.Error(t, err)

	err = image.Remove()
	assert.NoError(t, err)

	ioctx.Destroy()
	conn.DeletePool(poolname)
	conn.Shutdown()
}

func TestOpenImage(t *testing.T) {
	conn := radosConnect(t)

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	assert.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)

	name := GetUUID()

	_, err = OpenImage(ioctx, name, NoSnapshot)
	assert.Error(t, err)

	err = quickCreate(ioctx, name, testImageSize, testImageOrder)
	assert.NoError(t, err)

	oImage, err := OpenImage(ioctx, name, NoSnapshot)
	assert.NoError(t, err)
	assert.Equal(t, name, oImage.name)
	err = oImage.Close()
	assert.NoError(t, err)

	// open read-only
	oImage, err = OpenImageReadOnly(ioctx, name, NoSnapshot)
	assert.NoError(t, err)

	bytesIn := []byte("input data")
	_, err = oImage.Write(bytesIn)
	// writing should fail in read-only mode
	assert.Error(t, err)

	err = oImage.Close()
	assert.NoError(t, err)

	err = oImage.Remove()
	assert.NoError(t, err)

	_, err = OpenImageReadOnly(ioctx, name, NoSnapshot)
	assert.Error(t, err)

	ioctx.Destroy()
	conn.DeletePool(poolname)
	conn.Shutdown()
}

func TestRemoveImage(t *testing.T) {
	conn := radosConnect(t)

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	assert.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)

	// trying to remove a non-existent image is an error
	err = RemoveImage(ioctx, "bananarama")
	require.Error(t, err)

	// create and then remove an image
	name := GetUUID()
	options := NewRbdImageOptions()
	defer options.Destroy()
	err = CreateImage(ioctx, name, testImageSize, options)
	assert.NoError(t, err)

	imageNames, err := GetImageNames(ioctx)
	assert.NoError(t, err)
	assert.Contains(t, imageNames, name)

	err = RemoveImage(ioctx, name)
	assert.NoError(t, err)

	imageNames, err = GetImageNames(ioctx)
	assert.NoError(t, err)
	assert.NotContains(t, imageNames, name)

	ioctx.Destroy()
	conn.DeletePool(poolname)
	conn.Shutdown()
}

func TestCloneImage(t *testing.T) {
	conn := radosConnect(t)

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	assert.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)

	imageName := "parent1"
	snapName := "snap1"
	cloneName := "clone1"
	options := NewRbdImageOptions()
	defer options.Destroy()
	err = options.SetUint64(ImageOptionOrder, uint64(testImageOrder))
	assert.NoError(t, err)
	err = options.SetUint64(ImageOptionFeatures, 1)
	assert.NoError(t, err)
	err = CreateImage(ioctx, imageName, testImageSize, options)
	assert.NoError(t, err)

	image, err := OpenImage(ioctx, imageName, NoSnapshot)
	assert.NoError(t, err)

	snapshot, err := image.CreateSnapshot(snapName)
	assert.NoError(t, err)
	err = snapshot.Protect()
	assert.NoError(t, err)

	t.Run("cloneImage", func(t *testing.T) {
		imageNames, err := GetImageNames(ioctx)
		assert.NoError(t, err)
		assert.Contains(t, imageNames, imageName)
		assert.NotContains(t, imageNames, cloneName)

		options := NewRbdImageOptions()
		defer options.Destroy()
		err = options.SetUint64(ImageOptionFormat, uint64(2))
		assert.NoError(t, err)
		err = CloneImage(ioctx, imageName, snapName, ioctx, cloneName, options)
		assert.NoError(t, err)

		imageNames, err = GetImageNames(ioctx)
		assert.NoError(t, err)
		assert.Contains(t, imageNames, imageName)
		assert.Contains(t, imageNames, cloneName)

		err = RemoveImage(ioctx, cloneName)
		assert.NoError(t, err)
	})

	t.Run("cloneFromImage", func(t *testing.T) {
		imageNames, err := GetImageNames(ioctx)
		assert.NoError(t, err)
		assert.Contains(t, imageNames, imageName)
		assert.NotContains(t, imageNames, cloneName)

		options := NewRbdImageOptions()
		defer options.Destroy()
		err = options.SetUint64(ImageOptionFormat, uint64(2))
		assert.NoError(t, err)
		err = CloneFromImage(image, snapName, ioctx, cloneName, options)
		assert.NoError(t, err)

		imageNames, err = GetImageNames(ioctx)
		assert.NoError(t, err)
		assert.Contains(t, imageNames, imageName)
		assert.Contains(t, imageNames, cloneName)

		err = RemoveImage(ioctx, cloneName)
		assert.NoError(t, err)
	})

	t.Run("cloneFromImageInvalidCtx", func(t *testing.T) {
		options := NewRbdImageOptions()
		defer options.Destroy()
		err = options.SetUint64(ImageOptionFormat, uint64(2))
		assert.NoError(t, err)
		badImage := GetImage(nil, image.name)
		err = CloneFromImage(badImage, snapName, ioctx, cloneName, options)
		assert.Error(t, err)
	})

	t.Run("cloneImageNilOpts", func(t *testing.T) {
		err = CloneImage(ioctx, imageName, snapName, ioctx, cloneName, nil)
		assert.Error(t, err)
	})

	err = snapshot.Unprotect()
	assert.NoError(t, err)

	err = snapshot.Remove()
	assert.NoError(t, err)

	err = image.Close()
	assert.NoError(t, err)

	err = image.Remove()
	assert.NoError(t, err)

	ioctx.Destroy()
	conn.DeletePool(poolname)
	conn.Shutdown()
}

// quickCreate creates an image similar to Create but uses CreateImage.
// If possible, avoid using this function for new code/tests. It mainly exists
// to help with refactoring of existing tests.
func quickCreate(ioctx *rados.IOContext, name string, size uint64, order int) error {
	options := NewRbdImageOptions()
	defer options.Destroy()
	err := options.SetUint64(ImageOptionOrder, uint64(order))
	if err != nil {
		return err
	}
	return CreateImage(ioctx, name, size, options)
}

func TestGetId(t *testing.T) {
	conn := radosConnect(t)

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	assert.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)

	name := GetUUID()
	options := NewRbdImageOptions()
	assert.NoError(t,
		options.SetUint64(ImageOptionOrder, uint64(testImageOrder)))
	err = CreateImage(ioctx, name, testImageSize, options)
	assert.NoError(t, err)

	image, err := OpenImage(ioctx, name, NoSnapshot)
	assert.NoError(t, err)
	id, err := image.GetId()
	assert.NoError(t, err)
	assert.Greater(t, len(id), 8)

	err = image.Close()
	assert.NoError(t, err)

	id, err = image.GetId()
	assert.Error(t, err)
	assert.Equal(t, "", id)

	err = image.Remove()
	assert.NoError(t, err)

	ioctx.Destroy()
	conn.DeletePool(poolname)
	conn.Shutdown()
}

func TestOpenImageById(t *testing.T) {
	conn := radosConnect(t)

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	assert.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)

	name := GetUUID()
	options := NewRbdImageOptions()
	assert.NoError(t,
		options.SetUint64(ImageOptionOrder, uint64(testImageOrder)))
	err = CreateImage(ioctx, name, testImageSize, options)
	assert.NoError(t, err)

	workingImage, err := OpenImage(ioctx, name, NoSnapshot)
	assert.NoError(t, err)
	id, err := workingImage.GetId()
	assert.NoError(t, err)
	err = workingImage.Close()
	assert.NoError(t, err)

	t.Run("ReadWriteBadId", func(t *testing.T) {
		t.Skip("segfaults due to https://tracker.ceph.com/issues/43178")
		// phony id
		img, err := OpenImageById(ioctx, "102f00aaabbbccd", NoSnapshot)
		require.Error(t, err)
		require.Nil(t, img)
	})
	t.Run("ReadOnlyBadId", func(t *testing.T) {
		t.Skip("segfaults due to https://tracker.ceph.com/issues/43178")
		// phony id
		img, err := OpenImageByIdReadOnly(ioctx, "blubb", NoSnapshot)
		require.Error(t, err)
		require.Nil(t, img)
	})
	t.Run("ReadWrite", func(t *testing.T) {
		img, err := OpenImageById(ioctx, id, NoSnapshot)
		require.NoError(t, err)
		require.NotNil(t, img)
		defer func() { assert.NoError(t, img.Close()) }()

		data := []byte("input data")
		_, err = img.Write(data)
		assert.NoError(t, err)
	})
	t.Run("ReadOnly", func(t *testing.T) {
		img, err := OpenImageByIdReadOnly(ioctx, id, NoSnapshot)
		require.NoError(t, err)
		require.NotNil(t, img)
		defer func() { assert.NoError(t, img.Close()) }()

		data := []byte("input data")
		_, err = img.Write(data)
		// writing should fail in read-only mode
		assert.Error(t, err)
	})
	t.Run("Snapshot", func(t *testing.T) {
		img, err := OpenImageById(ioctx, id, NoSnapshot)
		require.NoError(t, err)
		require.NotNil(t, img)
		defer func() { assert.NoError(t, img.Close()) }()

		snapshot, err := img.CreateSnapshot("snerpshort")
		assert.NoError(t, err)
		defer func() { assert.NoError(t, snapshot.Remove()) }()

		snapImage, err := OpenImageById(ioctx, id, "snerpshort")
		require.NoError(t, err)
		require.NotNil(t, snapImage)
		assert.NoError(t, snapImage.Close())
	})
	t.Run("ReadOnlySnapshot", func(t *testing.T) {
		img, err := OpenImageById(ioctx, id, NoSnapshot)
		require.NoError(t, err)
		require.NotNil(t, img)
		defer func() { assert.NoError(t, img.Close()) }()

		snapshot, err := img.CreateSnapshot("snerpshort2")
		assert.NoError(t, err)
		defer func() { assert.NoError(t, snapshot.Remove()) }()

		snapImage, err := OpenImageByIdReadOnly(ioctx, id, "snerpshort2")
		require.NoError(t, err)
		require.NotNil(t, snapImage)
		assert.NoError(t, snapImage.Close())
	})

	err = workingImage.Remove()
	assert.NoError(t, err)

	ioctx.Destroy()
	conn.DeletePool(poolname)
	conn.Shutdown()
}
