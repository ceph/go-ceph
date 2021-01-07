package rgw

import (
	"fmt"
	"math/rand"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	rgw *RGW

	cephConf   = os.Getenv("CEPH_CONF")
	clientName = os.Getenv("RGW_TEST_CLIENT_NAME")
	s3User     = os.Getenv("RGW_TEST_S3_USER")
	ak         = os.Getenv("RGW_TEST_S3_AK")
	sk         = os.Getenv("RGW_TEST_S3_SK")
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestCreateRGW(t *testing.T) {
	confStr := "-c " + cephConf
	nameStr := "--name " + clientName
	rgwLocal, err := CreateRGW([]string{confStr, nameStr})
	assert.NoError(t, err)
	assert.NotNil(t, rgwLocal)

	rgw = rgwLocal
}

func TestMountUmount(t *testing.T) {
	fs := FS{}
	err := fs.Mount(rgw, s3User, ak, sk, 0)
	assert.NoError(t, err)

	err = fs.Umount(0)
	assert.NoError(t, err)
}

func TestVersion(_ *testing.T) {
	Version()
}

func TestStatFS(t *testing.T) {
	fs := FS{}
	err := fs.Mount(rgw, s3User, ak, sk, 0)
	assert.NoError(t, err)

	statVFS, err := fs.StatFS(fs.GetRootFileHandle(), 0)
	assert.NotNil(t, statVFS)
	assert.NoError(t, err)

	err = fs.Umount(0)
	assert.NoError(t, err)
}

type ReadDirCallbackDump struct {
	items int
}

func (cb *ReadDirCallbackDump) Callback(_ string, _ *syscall.Stat_t,
	_ AttrMask, _ uint32, _ uint64) bool {
	cb.items++
	return true
}

func TestMkdir(t *testing.T) {
	fs := FS{}
	err := fs.Mount(rgw, s3User, ak, sk, 0)
	assert.NoError(t, err)

	bName := fmt.Sprintf("mybucket-%v", rand.Int63())
	fhB, stB, err := fs.Mkdir(fs.GetRootFileHandle(), bName, 0, 0)
	assert.NotNil(t, fhB)
	assert.NotNil(t, stB)
	assert.NoError(t, err)

	err = fs.Umount(0)
	assert.NoError(t, err)
}

func TestReadDir(t *testing.T) {
	fs := FS{}
	err := fs.Mount(rgw, s3User, ak, sk, 0)
	assert.NoError(t, err)

	cb := &ReadDirCallbackDump{}
	_, _, err = fs.ReadDir(fs.GetRootFileHandle(), cb, 0, 0)
	assert.NoError(t, err)
	assert.Equal(t, true, (cb.items >= 1))
	err = fs.Umount(0)
	assert.NoError(t, err)
}

func TestCreateUnlink(t *testing.T) {
	fs := FS{}
	err := fs.Mount(rgw, s3User, ak, sk, 0)
	assert.NoError(t, err)

	bName := fmt.Sprintf("mybucket-%v", rand.Int63())
	fhB, stB, err := fs.Mkdir(fs.GetRootFileHandle(), bName, 0, 0)
	assert.NotNil(t, fhB)
	assert.NotNil(t, stB)
	assert.NoError(t, err)

	objName := fmt.Sprintf("obj-%v", rand.Int63())
	fhObj, _, err := fs.Create(fhB, objName, 0, 0, 0)
	assert.NotNil(t, fhObj)
	assert.NoError(t, err)

	err = fs.Unlink(fhB, objName, 0)
	assert.NoError(t, err)

	err = fs.Umount(0)
	assert.NoError(t, err)
}

func TestOpenClose(t *testing.T) {
	fs := FS{}
	err := fs.Mount(rgw, s3User, ak, sk, 0)
	assert.NoError(t, err)

	bName := fmt.Sprintf("mybucket-%v", rand.Int63())
	fhB, stB, err := fs.Mkdir(fs.GetRootFileHandle(), bName, 0, 0)
	assert.NotNil(t, fhB)
	assert.NotNil(t, stB)
	assert.NoError(t, err)

	objName := fmt.Sprintf("obj-%v", rand.Int63())
	fhObj, _, err := fs.Create(fhB, objName, 0, 0, 0)
	assert.NotNil(t, fhObj)
	assert.NoError(t, err)

	err = fs.Open(fhObj, 0, 0)
	assert.NoError(t, err)

	err = fs.Close(fhObj, 0)
	assert.NoError(t, err)

	err = fs.Umount(0)
	assert.NoError(t, err)
}

func TestSetAttrGetAttr(t *testing.T) {
	fs := FS{}
	err := fs.Mount(rgw, s3User, ak, sk, 0)
	assert.NoError(t, err)

	bName := fmt.Sprintf("mybucket-%v", rand.Int63())
	fhB, stB, err := fs.Mkdir(fs.GetRootFileHandle(), bName, 0, 0)
	assert.NotNil(t, fhB)
	assert.NotNil(t, stB)
	assert.NoError(t, err)

	objName := fmt.Sprintf("obj-%v", rand.Int63())
	fhObj, _, err := fs.Create(fhB, objName, 0, 0, 0)
	assert.NotNil(t, fhObj)
	assert.NoError(t, err)

	st := &syscall.Stat_t{
		Gid: 123,
		Uid: 234,
	}
	err = fs.SetAttr(fhObj, st, AttrGid|AttrUid, 0)
	assert.NoError(t, err)

	st2, err := fs.GetAttr(fhObj, 0)
	assert.NoError(t, err)
	assert.EqualValues(t, st2.Gid, st.Gid)
	assert.EqualValues(t, st2.Uid, st.Uid)

	err = fs.Umount(0)
	assert.NoError(t, err)
}

func TestReadWrite(t *testing.T) {
	fs := FS{}
	err := fs.Mount(rgw, s3User, ak, sk, 0)
	assert.NoError(t, err)

	bName := fmt.Sprintf("mybucket-%v", rand.Int63())
	fhB, stB, err := fs.Mkdir(fs.GetRootFileHandle(), bName, 0, 0)
	assert.NotNil(t, fhB)
	assert.NotNil(t, stB)
	assert.NoError(t, err)

	objName := fmt.Sprintf("obj-%v", rand.Int63())
	fhObj, _, err := fs.Create(fhB, objName, 0, 0, 0)
	assert.NotNil(t, fhObj)
	assert.NoError(t, err)

	err = fs.Open(fhObj, 0, 0)
	assert.NoError(t, err)

	buffer := []byte{'h', 'e', 'l', 'l', 'o'}
	written, err := fs.Write(fhObj, buffer, 0, uint64(len(buffer)), 0)
	assert.EqualValues(t, written, uint64(len(buffer)))

	err = fs.Close(fhObj, 0)
	assert.NoError(t, err)

	buffer2 := make([]byte, len(buffer))
	bytes, err := fs.Read(fhObj, 0, uint64(len(buffer)), buffer2, 0)
	assert.NotNil(t, bytes)
	assert.NoError(t, err)
	assert.Equal(t, string(buffer), string(buffer2))

	err = fs.Umount(0)
	assert.NoError(t, err)
}

func TestRename(t *testing.T) {
	fs := FS{}
	err := fs.Mount(rgw, s3User, ak, sk, 0)
	assert.NoError(t, err)

	bName := fmt.Sprintf("mybucket-%v", rand.Int63())
	fhB, stB, err := fs.Mkdir(fs.GetRootFileHandle(), bName, 0, 0)
	assert.NotNil(t, fhB)
	assert.NotNil(t, stB)
	assert.NoError(t, err)

	dirName := fmt.Sprintf("dir-%v", rand.Int63())
	fhDir, stDir, err := fs.Mkdir(fhB, dirName, 0, 0)
	assert.NotNil(t, fhDir)
	assert.NotNil(t, stDir)
	assert.NoError(t, err)

	objName := fmt.Sprintf("obj-%v", rand.Int63())
	fhObj, _, err := fs.Create(fhB, objName, 0, 0, 0)
	assert.NotNil(t, fhObj)
	assert.NoError(t, err)

	objName2 := fmt.Sprintf("obj2-%v", rand.Int63())
	err = fs.Rename(fhB, objName, fhDir, objName2, 0)
	assert.NoError(t, err)

	fhObj2, stObj2, err := fs.Lookup(fhDir, objName2, 0, 0)
	assert.NotNil(t, fhObj2)
	assert.NotNil(t, stObj2)
	assert.NoError(t, err)

	err = fs.Umount(0)
	assert.NoError(t, err)
}

func TestShutdownRGW(_ *testing.T) {
	ShutdownRGW(rgw)
}
