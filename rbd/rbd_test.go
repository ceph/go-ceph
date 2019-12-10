package rbd

import (
	"bytes"
	"encoding/json"
	"sort"
	"testing"
	"time"

	"github.com/ceph/go-ceph/rados"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func GetUUID() string {
	return uuid.Must(uuid.NewV4()).String()
}

func TestRBDError(t *testing.T) {
	err := GetError(0)
	assert.NoError(t, err)

	err = GetError(-39) // NOTEMPTY (image still has a snapshot)
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "rbd: ret=-39, Directory not empty")

	err = GetError(345) // no such errno
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "rbd: ret=345")
}

func TestVersion(t *testing.T) {
	var major, minor, patch = Version()
	assert.False(t, major < 0 || major > 1000, "invalid major")
	assert.False(t, minor < 0 || minor > 1000, "invalid minor")
	assert.False(t, patch < 0 || patch > 1000, "invalid patch")
}

func TestImageCreate(t *testing.T) {
	conn, _ := rados.NewConn()
	conn.ReadDefaultConfigFile()
	conn.Connect()

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	assert.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)

	name := GetUUID()
	image, err := Create(ioctx, name, 1<<22, 22)
	assert.NoError(t, err)
	err = image.Remove()
	assert.NoError(t, err)

	name = GetUUID()
	image, err = Create(ioctx, name, 1<<22, 22,
		RbdFeatureLayering|RbdFeatureStripingV2)
	assert.NoError(t, err)
	err = image.Remove()
	assert.NoError(t, err)

	name = GetUUID()
	image, err = Create(ioctx, name, 1<<22, 22,
		RbdFeatureLayering|RbdFeatureStripingV2, 4096, 2)
	assert.NoError(t, err)
	err = image.Remove()
	assert.NoError(t, err)

	// invalid order
	name = GetUUID()
	_, err = Create(ioctx, name, 1<<22, -1)
	assert.Error(t, err)

	// too many arguments
	_, err = Create(ioctx, name, 1<<22, 22,
		RbdFeatureLayering|RbdFeatureStripingV2, 4096, 2, 123)
	assert.Error(t, err)

	ioctx.Destroy()
	conn.DeletePool(poolname)
	conn.Shutdown()
}

func TestImageCreate2(t *testing.T) {
	conn, _ := rados.NewConn()
	conn.ReadDefaultConfigFile()
	conn.Connect()

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	assert.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolname)
	assert.NoError(t, err)

	name := GetUUID()
	image, err := Create2(ioctx, name, 1<<22,
		RbdFeatureLayering|RbdFeatureStripingV2, 22)
	assert.NoError(t, err)
	err = image.Remove()
	assert.NoError(t, err)

	ioctx.Destroy()
	conn.DeletePool(poolname)
	conn.Shutdown()
}

func TestImageCreate3(t *testing.T) {
	conn, _ := rados.NewConn()
	conn.ReadDefaultConfigFile()
	conn.Connect()

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	assert.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolname)
	assert.NoError(t, err)

	name := GetUUID()
	image, err := Create3(ioctx, name, 1<<22,
		RbdFeatureLayering|RbdFeatureStripingV2, 22, 4096, 2)
	assert.NoError(t, err)
	err = image.Remove()
	assert.NoError(t, err)

	ioctx.Destroy()
	conn.DeletePool(poolname)
	conn.Shutdown()
}

func TestCreateImageWithOptions(t *testing.T) {
	conn, _ := rados.NewConn()
	conn.ReadDefaultConfigFile()
	conn.Connect()

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	assert.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolname)
	assert.NoError(t, err)

	// nil options, causes a panic if not handled correctly
	name := GetUUID()
	image, err := Create4(ioctx, name, 1<<22, nil)
	assert.Error(t, err)

	options := NewRbdImageOptions()

	// empty/default options
	name = GetUUID()
	image, err = Create4(ioctx, name, 1<<22, options)
	assert.NoError(t, err)
	err = image.Remove()
	assert.NoError(t, err)

	// create image with RbdImageOptionOrder
	err = options.SetUint64(RbdImageOptionOrder, 22)
	assert.NoError(t, err)
	name = GetUUID()
	image, err = Create4(ioctx, name, 1<<22, options)
	assert.NoError(t, err)
	err = image.Remove()
	assert.NoError(t, err)
	options.Clear()

	// create image with a different data pool
	datapool := GetUUID()
	err = conn.MakePool(datapool)
	assert.NoError(t, err)
	err = options.SetString(RbdImageOptionDataPool, datapool)
	assert.NoError(t, err)
	name = GetUUID()
	image, err = Create4(ioctx, name, 1<<22, options)
	assert.NoError(t, err)
	err = image.Remove()
	assert.NoError(t, err)
	conn.DeletePool(datapool)

	// cleanup
	options.Destroy()
	ioctx.Destroy()
	conn.DeletePool(poolname)
	conn.Shutdown()
}

func TestGetImageNames(t *testing.T) {
	conn, _ := rados.NewConn()
	conn.ReadDefaultConfigFile()
	conn.Connect()

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	assert.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)

	createdList := []string{}
	for i := 0; i < 10; i++ {
		name := GetUUID()
		_, err := Create(ioctx, name, 1<<22, 22)
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

func TestImageReadOnly(t *testing.T) {
	conn, _ := rados.NewConn()
	conn.ReadDefaultConfigFile()
	conn.Connect()

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	assert.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)

	name := GetUUID()
	image, err := Create(ioctx, name, 1<<22, 22)
	assert.NoError(t, err)

	err = image.Open(true)
	assert.NoError(t, err)

	bytes_in := []byte("input data")
	_, err = image.Write(bytes_in)
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
	conn, _ := rados.NewConn()
	conn.ReadDefaultConfigFile()
	conn.Connect()

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	assert.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)

	name := GetUUID()
	reqSize := uint64(1024 * 1024 * 4) // 4MB
	image, err := Create(ioctx, name, reqSize, 22)
	assert.NoError(t, err)

	err = image.Resize(reqSize * 2)
	assert.Equal(t, err, ErrImageNotOpen)

	err = image.Open()
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

	err = image.Remove()
	assert.NoError(t, err)

	ioctx.Destroy()
	conn.DeletePool(poolname)
	conn.Shutdown()
}

func TestImageProperties(t *testing.T) {
	conn, _ := rados.NewConn()
	conn.ReadDefaultConfigFile()
	conn.Connect()

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	require.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)

	name := GetUUID()
	reqSize := uint64(1024 * 1024 * 4) // 4MB
	img, err := Create3(ioctx, name, reqSize,
		RbdFeatureLayering|RbdFeatureStripingV2, 22, 4096, 2)
	require.NoError(t, err)

	err = img.Open()
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
	assert.Equal(t, features&(RbdFeatureLayering|RbdFeatureStripingV2),
		RbdFeatureLayering|RbdFeatureStripingV2)

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
	conn, _ := rados.NewConn()
	conn.ReadDefaultConfigFile()
	conn.Connect()

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	assert.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)

	name := GetUUID()
	img, err := Create(ioctx, name, 1<<22, 22)
	assert.NoError(t, err)

	err = img.Rename(name)
	assert.Error(t, err)

	err = img.Rename(GetUUID())
	assert.NoError(t, err)

	img.Remove()

	ioctx.Destroy()
	conn.DeletePool(poolname)
	conn.Shutdown()
}

func TestIOReaderWriter(t *testing.T) {
	conn, _ := rados.NewConn()
	conn.ReadDefaultConfigFile()
	conn.Connect()

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	assert.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)

	name := GetUUID()
	img, err := Create(ioctx, name, 1<<22, 22)
	assert.NoError(t, err)

	err = img.Open()
	assert.NoError(t, err)

	stats, err := img.Stat()
	assert.NoError(t, err)

	encoder := json.NewEncoder(img)
	encoder.Encode(stats)

	err = img.Flush()
	assert.NoError(t, err)

	_, err = img.Seek(0, 0)
	assert.NoError(t, err)

	var stats2 *ImageInfo
	decoder := json.NewDecoder(img)
	decoder.Decode(&stats2)

	assert.Equal(t, &stats, &stats2)

	_, err = img.Seek(0, 0)
	bytes_in := []byte("input data")
	_, err = img.Write(bytes_in)
	assert.NoError(t, err)

	_, err = img.Seek(0, 0)
	assert.NoError(t, err)

	bytes_out := make([]byte, len(bytes_in))
	n_out, err := img.Read(bytes_out)

	assert.Equal(t, n_out, len(bytes_in))
	assert.Equal(t, bytes_in, bytes_out)

	err = img.Close()
	assert.NoError(t, err)

	img.Remove()
	assert.NoError(t, err)

	ioctx.Destroy()
	conn.DeletePool(poolname)
	conn.Shutdown()
}

func TestCreateSnapshot(t *testing.T) {
	conn, _ := rados.NewConn()
	conn.ReadDefaultConfigFile()
	conn.Connect()

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	assert.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)

	name := GetUUID()
	img, err := Create(ioctx, name, 1<<22, 22)
	assert.NoError(t, err)

	err = img.Open()
	assert.NoError(t, err)

	snapshot, err := img.CreateSnapshot("mysnap")
	assert.NoError(t, err)

	err = img.Close()
	err = img.Open("mysnap")
	assert.NoError(t, err)

	snapshot.Remove()
	assert.NoError(t, err)

	err = img.Close()
	assert.NoError(t, err)

	img.Remove()
	assert.NoError(t, err)

	ioctx.Destroy()
	conn.DeletePool(poolname)
	conn.Shutdown()
}

func TestParentInfo(t *testing.T) {
	conn, _ := rados.NewConn()
	conn.ReadDefaultConfigFile()
	conn.Connect()

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	assert.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)

	name := "parent"
	img, err := Create(ioctx, name, 1<<22, 22, 1)
	assert.NoError(t, err)

	err = img.Open()
	assert.NoError(t, err)

	snapshot, err := img.CreateSnapshot("mysnap")
	assert.NoError(t, err)

	err = snapshot.Protect()
	assert.NoError(t, err)

	// create an image context with the parent+snapshot
	snapImg := GetImage(ioctx, "parent")
	err = snapImg.Open("mysnap")
	assert.NoError(t, err)

	// ensure no children prior to clone
	pools, images, err := snapImg.ListChildren()
	assert.NoError(t, err)
	assert.Equal(t, len(pools), 0, "pools equal")
	assert.Equal(t, len(images), 0, "children length equal")

	imgNew, err := img.Clone("mysnap", ioctx, "child", 1, 22)
	assert.NoError(t, err)

	err = imgNew.Open()
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
	conn, _ := rados.NewConn()
	conn.ReadDefaultConfigFile()
	conn.Connect()

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	assert.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)

	name := GetUUID()

	img := GetImage(ioctx, name)
	err = img.Open()
	assert.Equal(t, err, ErrNotFound)

	err = img.Remove()
	assert.Equal(t, err, ErrNotFound)

	ioctx.Destroy()
	conn.DeletePool(poolname)
	conn.Shutdown()
}

func TestErrorSnapshotNoName(t *testing.T) {
	conn, _ := rados.NewConn()
	conn.ReadDefaultConfigFile()
	conn.Connect()

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	assert.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)

	name := GetUUID()
	img, err := Create(ioctx, name, 1<<22, 22)
	assert.NoError(t, err)

	err = img.Open()
	assert.NoError(t, err)

	// this actually works for some reason?!
	snapshot, err := img.CreateSnapshot("")
	assert.NoError(t, err)

	err = img.Close()
	assert.NoError(t, err)

	err = snapshot.Remove()
	assert.Equal(t, err, ErrSnapshotNoName)

	err = snapshot.Rollback()
	assert.Equal(t, err, ErrSnapshotNoName)

	err = snapshot.Protect()
	assert.Equal(t, err, ErrSnapshotNoName)

	err = snapshot.Unprotect()
	assert.Equal(t, err, ErrSnapshotNoName)

	_, err = snapshot.IsProtected()
	assert.Equal(t, err, ErrSnapshotNoName)

	err = snapshot.Set()
	assert.Equal(t, err, ErrSnapshotNoName)

	// image can not be removed as the snapshot still exists
	// err = img.Remove()
	// assert.NoError(t, err)

	ioctx.Destroy()
	conn.DeletePool(poolname)
	conn.Shutdown()
}

func TestTrashImage(t *testing.T) {
	conn, _ := rados.NewConn()
	conn.ReadDefaultConfigFile()
	conn.Connect()

	poolname := GetUUID()
	err := conn.MakePool(poolname)
	assert.NoError(t, err)

	ioctx, err := conn.OpenIOContext(poolname)
	require.NoError(t, err)

	name := GetUUID()
	image, err := Create(ioctx, name, 1<<22, 22)
	assert.NoError(t, err)

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
