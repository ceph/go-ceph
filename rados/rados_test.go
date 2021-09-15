package rados

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tsuite "github.com/stretchr/testify/suite"
)

type RadosTestSuite struct {
	tsuite.Suite
	conn  *Conn
	ioctx *IOContext
	pool  string
	count int
}

func (suite *RadosTestSuite) SetupSuite() {
	conn, err := NewConn()
	require.NoError(suite.T(), err)
	defer conn.Shutdown()

	err = conn.ReadDefaultConfigFile()
	require.NoError(suite.T(), err)

	timeout := time.After(time.Second * 5)
	ch := make(chan error)
	go func(conn *Conn) {
		ch <- conn.Connect()
	}(conn)
	select {
	case err = <-ch:
	case <-timeout:
		err = fmt.Errorf("timed out waiting for connect")
	}

	if assert.NoError(suite.T(), err) {
		pool := uuid.Must(uuid.NewV4()).String()
		if err = conn.MakePool(pool); assert.NoError(suite.T(), err) {
			suite.pool = pool
			return
		}
	}

	suite.T().FailNow()
}

func (suite *RadosTestSuite) SetupTest() {
	suite.conn = nil
	suite.ioctx = nil
	suite.count = 0

	conn, err := NewConn()
	require.NoError(suite.T(), err)
	suite.conn = conn
	err = suite.conn.ReadDefaultConfigFile()
	require.NoError(suite.T(), err)
}

func (suite *RadosTestSuite) SetupConnection() {
	if err := suite.conn.Connect(); assert.NoError(suite.T(), err) {
		ioctx, err := suite.conn.OpenIOContext(suite.pool)
		if assert.NoError(suite.T(), err) {
			suite.ioctx = ioctx
			return
		}
	}
	suite.conn.Shutdown()
	suite.T().FailNow()
}

func (suite *RadosTestSuite) GenObjectName() string {
	name := fmt.Sprintf("%s_%d", suite.T().Name(), suite.count)
	suite.count++
	return name
}

func (suite *RadosTestSuite) RandomBytes(size int) []byte {
	bytes := make([]byte, size)
	n, err := rand.Read(bytes)
	require.Equal(suite.T(), n, size)
	require.NoError(suite.T(), err)
	return bytes
}

func (suite *RadosTestSuite) TearDownTest() {
	if suite.ioctx != nil {
		suite.ioctx.Destroy()
	}
	suite.conn.Shutdown()
}

func (suite *RadosTestSuite) TearDownSuite() {
	conn, err := NewConn()
	require.NoError(suite.T(), err)
	defer conn.Shutdown()

	err = conn.ReadDefaultConfigFile()
	require.NoError(suite.T(), err)

	if err = conn.Connect(); assert.NoError(suite.T(), err) {
		err = conn.DeletePool(suite.pool)
		assert.NoError(suite.T(), err)
	}
}

func TestVersion(t *testing.T) {
	var major, minor, patch = Version()
	assert.False(t, major < 0 || major > 1000, "invalid major")
	assert.False(t, minor < 0 || minor > 1000, "invalid minor")
	assert.False(t, patch < 0 || patch > 1000, "invalid patch")
}

func (suite *RadosTestSuite) TestGetFSID() {
	fsid, err := suite.conn.GetFSID()
	assert.NoError(suite.T(), err)
	assert.NotEqual(suite.T(), fsid, "")
}

func (suite *RadosTestSuite) TestGetSetConfigOption() {
	// rejects invalid options
	err := suite.conn.SetConfigOption("___dne___", "value")
	assert.Error(suite.T(), err, "Invalid option")

	// check error for  get invalid option
	_, err = suite.conn.GetConfigOption("__dne__")
	assert.Error(suite.T(), err)

	// verify SetConfigOption changes a values
	prevVal, err := suite.conn.GetConfigOption("log_file")
	assert.NoError(suite.T(), err, "Invalid option")

	err = suite.conn.SetConfigOption("log_file", "/dev/null")
	assert.NoError(suite.T(), err, "Invalid option")

	currVal, err := suite.conn.GetConfigOption("log_file")
	assert.NoError(suite.T(), err, "Invalid option")

	assert.NotEqual(suite.T(), prevVal, "/dev/null")
	assert.Equal(suite.T(), currVal, "/dev/null")
}

func (suite *RadosTestSuite) TestParseDefaultConfigEnv() {
	prevVal, err := suite.conn.GetConfigOption("log_file")
	assert.NoError(suite.T(), err, "Invalid option")

	err = os.Setenv("CEPH_ARGS", "--log-file /dev/null")
	assert.NoError(suite.T(), err)

	err = suite.conn.ParseDefaultConfigEnv()
	assert.NoError(suite.T(), err)

	currVal, err := suite.conn.GetConfigOption("log_file")
	assert.NoError(suite.T(), err, "Invalid option")

	assert.NotEqual(suite.T(), prevVal, "/dev/null")
	assert.Equal(suite.T(), currVal, "/dev/null")
}

func (suite *RadosTestSuite) TestParseConfigArgv() {
	prevVal, err := suite.conn.GetConfigOption("log_file")
	assert.NoError(suite.T(), err, "Invalid option")

	argv := []string{"rados.test", "--log_file", "/dev/null"}
	err = suite.conn.ParseConfigArgv(argv)
	assert.NoError(suite.T(), err)

	currVal, err := suite.conn.GetConfigOption("log_file")
	assert.NoError(suite.T(), err, "Invalid option")

	assert.NotEqual(suite.T(), prevVal, "/dev/null")
	assert.Equal(suite.T(), currVal, "/dev/null")

	// ensure that an empty slice triggers an error (not a crash)
	err = suite.conn.ParseConfigArgv([]string{})
	assert.Error(suite.T(), err)

	// ensure we get an error for an invalid conn value
	badConn := &Conn{}
	err = badConn.ParseConfigArgv(
		[]string{"cephfs.test", "--log_file", "/dev/null"})
	assert.Error(suite.T(), err)
}

func (suite *RadosTestSuite) TestParseCmdLineArgs() {
	prevVal, err := suite.conn.GetConfigOption("log_file")
	assert.NoError(suite.T(), err, "Invalid option")

	args := []string{"--log_file", "/dev/null"}
	err = suite.conn.ParseCmdLineArgs(args)
	assert.NoError(suite.T(), err)

	currVal, err := suite.conn.GetConfigOption("log_file")
	assert.NoError(suite.T(), err, "Invalid option")

	assert.NotEqual(suite.T(), prevVal, "/dev/null")
	assert.Equal(suite.T(), currVal, "/dev/null")
}

func (suite *RadosTestSuite) TestReadConfigFile() {
	// check current log_file value
	prevStr, err := suite.conn.GetConfigOption("log_max_new")
	assert.NoError(suite.T(), err)

	prevVal, err := strconv.Atoi(prevStr)
	assert.NoError(suite.T(), err)

	// create conf file that changes log_file conf option
	file, err := ioutil.TempFile("/tmp", "go-rados")
	require.NoError(suite.T(), err)
	defer func() {
		assert.NoError(suite.T(), file.Close())
		assert.NoError(suite.T(), os.Remove(file.Name()))
	}()

	nextVal := prevVal + 1
	conf := fmt.Sprintf("[global]\nlog_max_new = %d\n", nextVal)
	_, err = io.WriteString(file, conf)
	assert.NoError(suite.T(), err)

	// parse the config file
	err = suite.conn.ReadConfigFile(file.Name())
	assert.NoError(suite.T(), err)

	// check current log_file value
	currStr, err := suite.conn.GetConfigOption("log_max_new")
	assert.NoError(suite.T(), err)

	currVal, err := strconv.Atoi(currStr)
	assert.NoError(suite.T(), err)

	assert.NotEqual(suite.T(), prevStr, currStr)
	assert.Equal(suite.T(), currVal, prevVal+1)
}

func (suite *RadosTestSuite) TestGetClusterStats() {
	suite.SetupConnection()

	// grab current stats
	prevStat, err := suite.conn.GetClusterStats()
	fmt.Printf("prevStat: %+v\n", prevStat)
	assert.NoError(suite.T(), err)

	// make some changes to the cluster
	buf := make([]byte, 1<<20)
	for i := 0; i < 10; i++ {
		objname := suite.GenObjectName()
		err = suite.ioctx.Write(objname, buf, 0)
		assert.NoError(suite.T(), err)
	}

	// wait a while for the stats to change
	for i := 0; i < 30; i++ {
		stat, err := suite.conn.GetClusterStats()
		assert.NoError(suite.T(), err)

		// wait for something to change
		if stat == prevStat {
			fmt.Printf("curr_stat: %+v (trying again...)\n", stat)
			time.Sleep(time.Second)
		} else {
			// success
			fmt.Printf("curr_stat: %+v (change detected)\n", stat)
			return
		}
	}

	suite.T().Error("Cluster stats aren't changing")
}

func (suite *RadosTestSuite) TestGetInstanceID() {
	suite.SetupConnection()

	id := suite.conn.GetInstanceID()
	assert.NotEqual(suite.T(), id, 0)
}

func (suite *RadosTestSuite) TestMakeDeletePool() {
	suite.SetupConnection()

	// check that new pool name is unique
	newName := uuid.Must(uuid.NewV4()).String()
	_, err := suite.conn.GetPoolByName(newName)
	if err == nil {
		suite.T().Error("Random pool name exists!")
		return
	}

	// create pool
	err = suite.conn.MakePool(newName)
	assert.NoError(suite.T(), err)

	// verify that the new pool name exists
	pool, err := suite.conn.GetPoolByName(newName)
	assert.NoError(suite.T(), err)

	if pool == 0 {
		suite.T().Error("Cannot find newly created pool")
	}

	// delete the pool
	err = suite.conn.DeletePool(newName)
	assert.NoError(suite.T(), err)

	// verify that it is gone
	pool, _ = suite.conn.GetPoolByName(newName)
	if pool != 0 {
		suite.T().Error("Deleted pool still exists")
	}
}

func (suite *RadosTestSuite) TestGetPoolByName() {
	suite.SetupConnection()

	// get current list of pool
	pools, err := suite.conn.ListPools()
	assert.NoError(suite.T(), err)

	// check that new pool name is unique
	newName := uuid.Must(uuid.NewV4()).String()
	require.NotContains(
		suite.T(), pools, newName, "Random pool name exists!")

	pool, _ := suite.conn.GetPoolByName(newName)
	assert.Equal(suite.T(), int64(0), pool, "Pool does not exist, but was found!")

	// create pool
	err = suite.conn.MakePool(newName)
	assert.NoError(suite.T(), err)

	// verify that the new pool name exists
	pools, err = suite.conn.ListPools()
	assert.NoError(suite.T(), err)
	assert.Contains(
		suite.T(), pools, newName, "Cannot find newly created pool")

	pool, err = suite.conn.GetPoolByName(newName)
	assert.NoError(suite.T(), err)
	assert.NotEqual(suite.T(), int64(0), pool, "Pool not found!")

	// delete the pool
	err = suite.conn.DeletePool(newName)
	assert.NoError(suite.T(), err)

	// verify that it is gone
	pools, err = suite.conn.ListPools()
	assert.NoError(suite.T(), err)
	assert.NotContains(
		suite.T(), pools, newName, "Deleted pool still exists")

	pool, err = suite.conn.GetPoolByName(newName)
	assert.Error(suite.T(), err)
	assert.Equal(
		suite.T(), int64(0), pool, "Pool should have been deleted, but was found!")
}

func (suite *RadosTestSuite) TestGetPoolByID() {
	suite.SetupConnection()

	// check that new pool name is unique
	newName := uuid.Must(uuid.NewV4()).String()
	pool, err := suite.conn.GetPoolByName(newName)
	if pool != 0 || err == nil {
		suite.T().Error("Random pool name exists!")
		return
	}

	// create pool
	err = suite.conn.MakePool(newName)
	assert.NoError(suite.T(), err)

	// verify that the new pool name exists
	id, err := suite.conn.GetPoolByName(newName)
	assert.NoError(suite.T(), err)

	if id == 0 {
		suite.T().Error("Cannot find newly created pool")
	}

	// get the name of the pool
	name, err := suite.conn.GetPoolByID(id)
	assert.NoError(suite.T(), err)

	if name == "" {
		suite.T().Error("Cannot find name of newly created pool")
	}

	// delete the pool
	err = suite.conn.DeletePool(newName)
	assert.NoError(suite.T(), err)

	// verify that it is gone
	name, err = suite.conn.GetPoolByID(id)
	if name != "" || err == nil {
		suite.T().Error("Deleted pool still exists")
	}
}

func (suite *RadosTestSuite) TestGetLargePoolList() {
	suite.SetupConnection()

	// get current list of pool
	pools, err := suite.conn.ListPools()
	assert.NoError(suite.T(), err)

	fill := func(r rune) string {
		b := make([]rune, 512)
		for i := range b {
			b[i] = r
		}
		return string(b)
	}
	// try to ensure we exceed the default 4096 byte initial buffer
	// size and make use of the increased buffer size code path
	names := []string{
		fill('a'),
		fill('b'),
		fill('c'),
		fill('d'),
		fill('e'),
		fill('f'),
		fill('g'),
		fill('h'),
		fill('i'),
		fill('j'),
	}

	defer func(origPools []string) {
		for _, name := range names {
			err := suite.conn.DeletePool(name)
			assert.NoError(suite.T(), err)
		}
		cleanPools, err := suite.conn.ListPools()
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), origPools, cleanPools)
	}(pools)

	for _, name := range names {
		err = suite.conn.MakePool(name)
		require.NoError(suite.T(), err)
	}
	pools, err = suite.conn.ListPools()
	assert.NoError(suite.T(), err)
	for _, name := range names {
		assert.Contains(suite.T(), pools, name)
	}
}

func (suite *RadosTestSuite) TestPingMonitor() {
	suite.SetupConnection()

	// mon id that should work with vstart.sh
	reply, err := suite.conn.PingMonitor("a")
	assert.NoError(suite.T(), err)
	assert.NotEqual(suite.T(), reply, "")

	// invalid mon id
	reply, err = suite.conn.PingMonitor("charlieB")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), reply, "")
}

func (suite *RadosTestSuite) TestWaitForLatestOSDMap() {
	suite.SetupConnection()

	err := suite.conn.WaitForLatestOSDMap()
	assert.NoError(suite.T(), err)
}

func (suite *RadosTestSuite) TestCreate() {
	suite.SetupConnection()

	err := suite.ioctx.Create("unique", CreateExclusive)
	assert.NoError(suite.T(), err)

	err = suite.ioctx.Create("unique", CreateExclusive)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), err, ErrObjectExists)

	err = suite.ioctx.Create("unique", CreateIdempotent)
	assert.NoError(suite.T(), err)
}

func (suite *RadosTestSuite) TestReadWrite() {
	suite.SetupConnection()

	bytesIn := []byte("input data")
	err := suite.ioctx.Write("obj", bytesIn, 0)
	assert.NoError(suite.T(), err)

	bytesOut := make([]byte, len(bytesIn))
	numOut, err := suite.ioctx.Read("obj", bytesOut, 0)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), numOut, len(bytesIn))
	assert.Equal(suite.T(), bytesIn, bytesOut)

	bytesIn = []byte("input another data")
	err = suite.ioctx.WriteFull("obj", bytesIn)
	assert.NoError(suite.T(), err)

	bytesOut = make([]byte, len(bytesIn))
	numOut, err = suite.ioctx.Read("obj", bytesOut, 0)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), numOut, len(bytesIn))
	assert.Equal(suite.T(), bytesIn, bytesOut)
}

func (suite *RadosTestSuite) TestAppend() {
	suite.SetupConnection()

	mirror := []byte{}
	oid := suite.GenObjectName()
	for i := 0; i < 3; i++ {
		// append random bytes
		bytes := suite.RandomBytes(33)
		err := suite.ioctx.Append(oid, bytes)
		assert.NoError(suite.T(), err)

		// what the object should contain
		mirror = append(mirror, bytes...)

		// check object contains what we expect
		buf := make([]byte, len(mirror))
		n, err := suite.ioctx.Read(oid, buf, 0)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), n, len(buf))
		assert.Equal(suite.T(), buf, mirror)
	}
}

func (suite *RadosTestSuite) TestReadNotFound() {
	suite.SetupConnection()

	var bytes []byte
	oid := suite.GenObjectName()
	_, err := suite.ioctx.Read(oid, bytes, 0)
	assert.Equal(suite.T(), err, ErrNotFound)
}

func (suite *RadosTestSuite) TestDeleteNotFound() {
	suite.SetupConnection()

	oid := suite.GenObjectName()
	err := suite.ioctx.Delete(oid)
	assert.Equal(suite.T(), err, ErrNotFound)
}

func (suite *RadosTestSuite) TestStatNotFound() {
	suite.SetupConnection()

	oid := suite.GenObjectName()
	_, err := suite.ioctx.Stat(oid)
	assert.Equal(suite.T(), err, ErrNotFound)
}

func (suite *RadosTestSuite) TestObjectStat() {
	suite.SetupConnection()

	oid := suite.GenObjectName()
	bytes := suite.RandomBytes(234)
	err := suite.ioctx.Write(oid, bytes, 0)
	assert.NoError(suite.T(), err)

	stat, err := suite.ioctx.Stat(oid)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), uint64(len(bytes)), stat.Size)
	assert.NotNil(suite.T(), stat.ModTime)
}

func (suite *RadosTestSuite) TestGetPoolStats() {
	suite.SetupConnection()

	// grab current stats
	prevStat, err := suite.ioctx.GetPoolStats()
	fmt.Printf("prevStat: %+v\n", prevStat)
	assert.NoError(suite.T(), err)

	// make some changes to the cluster
	buf := make([]byte, 1<<20)
	for i := 0; i < 10; i++ {
		oid := suite.GenObjectName()
		err = suite.ioctx.Write(oid, buf, 0)
		assert.NoError(suite.T(), err)
	}

	// wait a while for the stats to change
	for i := 0; i < 30; i++ {
		stat, err := suite.ioctx.GetPoolStats()
		assert.NoError(suite.T(), err)

		// wait for something to change
		if stat == prevStat {
			fmt.Printf("curr_stat: %+v (trying again...)\n", stat)
			time.Sleep(time.Second)
		} else {
			// success
			fmt.Printf("curr_stat: %+v (change detected)\n", stat)
			return
		}
	}

	suite.T().Error("Pool stats aren't changing")
}

func (suite *RadosTestSuite) TestGetPoolID() {
	suite.SetupConnection()

	poolIDByName, err := suite.conn.GetPoolByName(suite.pool)
	assert.NoError(suite.T(), err)
	assert.NotEqual(suite.T(), int64(0), poolIDByName)

	poolID := suite.ioctx.GetPoolID()
	assert.Equal(suite.T(), poolIDByName, poolID)
}

func (suite *RadosTestSuite) TestGetPoolName() {
	suite.SetupConnection()

	name, err := suite.ioctx.GetPoolName()
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), name, suite.pool)
}

func (suite *RadosTestSuite) TestObjectListObjects() {
	suite.SetupConnection()

	// objects currently in pool
	prevObjectList := []string{}
	err := suite.ioctx.ListObjects(func(oid string) {
		prevObjectList = append(prevObjectList, oid)
	})
	assert.NoError(suite.T(), err)

	// create some objects
	createdList := []string{}
	for i := 0; i < 10; i++ {
		oid := suite.GenObjectName()
		bytes := []byte("input data")
		err := suite.ioctx.Write(oid, bytes, 0)
		assert.NoError(suite.T(), err)
		createdList = append(createdList, oid)
	}

	// join the lists of objects
	expectedObjectList := prevObjectList
	expectedObjectList = append(expectedObjectList, createdList...)

	// now list the current set of objects in the pool
	currObjectList := []string{}
	err = suite.ioctx.ListObjects(func(oid string) {
		currObjectList = append(currObjectList, oid)
	})
	assert.NoError(suite.T(), err)

	// lists should be equal
	sort.Strings(currObjectList)
	sort.Strings(expectedObjectList)
	assert.Equal(suite.T(), currObjectList, expectedObjectList)
}

func (suite *RadosTestSuite) TestObjectIterator() {
	suite.SetupConnection()

	// current objs in default namespace
	prevObjectList := []string{}
	iter, err := suite.ioctx.Iter()
	assert.NoError(suite.T(), err)
	for iter.Next() {
		prevObjectList = append(prevObjectList, iter.Value())
	}
	iter.Close()
	assert.NoError(suite.T(), iter.Err())

	// create an object in a different namespace to verify that
	// iteration within a namespace does not return it
	suite.ioctx.SetNamespace("ns1")
	bytesIn := []byte("input data")
	err = suite.ioctx.Write(suite.GenObjectName(), bytesIn, 0)
	assert.NoError(suite.T(), err)

	// create some objects in default namespace
	suite.ioctx.SetNamespace("")
	createdList := []string{}
	for i := 0; i < 10; i++ {
		oid := suite.GenObjectName()
		bytesIn := []byte("input data")
		err = suite.ioctx.Write(oid, bytesIn, 0)
		assert.NoError(suite.T(), err)
		createdList = append(createdList, oid)
	}

	// prev list plus new oids
	expectedObjectList := prevObjectList
	expectedObjectList = append(expectedObjectList, createdList...)

	currObjectList := []string{}
	iter, err = suite.ioctx.Iter()
	assert.NoError(suite.T(), err)
	for iter.Next() {
		currObjectList = append(currObjectList, iter.Value())
	}
	iter.Close()
	assert.NoError(suite.T(), iter.Err())

	// curr list doesn't include the obj in ns1
	sort.Strings(expectedObjectList)
	sort.Strings(currObjectList)
	assert.Equal(suite.T(), currObjectList, expectedObjectList)
}

func (suite *RadosTestSuite) TestObjectIteratorAcrossNamespaces() {
	suite.SetupConnection()

	const perNamespace = 100

	// tests use a shared pool so namespaces need to be unique across tests.
	// below ns1=nsX and ns2=nsY. ns1 is used elsewhere.
	objectListNS1 := []string{}
	objectListNS2 := []string{}

	// populate list of current objects
	suite.ioctx.SetNamespace(AllNamespaces)
	existingList := []string{}
	iter, err := suite.ioctx.Iter()
	assert.NoError(suite.T(), err)
	for iter.Next() {
		existingList = append(existingList, iter.Value())
	}
	iter.Close()
	assert.NoError(suite.T(), iter.Err())

	// create some new objects in namespace: nsX
	createdList := []string{}
	suite.ioctx.SetNamespace("nsX")
	for i := 0; i < 10; i++ {
		oid := suite.GenObjectName()
		bytesIn := []byte("input data")
		err = suite.ioctx.Write(oid, bytesIn, 0)
		assert.NoError(suite.T(), err)
		createdList = append(createdList, oid)
	}
	assert.True(suite.T(), len(createdList) == 10)

	// create some new objects in namespace: nsY
	suite.ioctx.SetNamespace("nsY")
	for i := 0; i < 10; i++ {
		oid := suite.GenObjectName()
		bytesIn := []byte("input data")
		err = suite.ioctx.Write(oid, bytesIn, 0)
		assert.NoError(suite.T(), err)
		createdList = append(createdList, oid)
	}
	assert.True(suite.T(), len(createdList) == 20)

	suite.ioctx.SetNamespace(AllNamespaces)
	iter, err = suite.ioctx.Iter()
	assert.NoError(suite.T(), err)
	rogueList := []string{}
	for iter.Next() {
		if iter.Namespace() == "nsX" {
			objectListNS1 = append(objectListNS1, iter.Value())
		} else if iter.Namespace() == "nsY" {
			objectListNS2 = append(objectListNS2, iter.Value())
		} else {
			rogueList = append(rogueList, iter.Value())
		}
	}
	iter.Close()
	assert.NoError(suite.T(), iter.Err())

	assert.Equal(suite.T(), len(existingList), len(rogueList))
	assert.Equal(suite.T(), len(objectListNS1), 10)
	assert.Equal(suite.T(), len(objectListNS2), 10)

	objectList := []string{}
	objectList = append(objectList, objectListNS1...)
	objectList = append(objectList, objectListNS2...)
	sort.Strings(objectList)
	sort.Strings(createdList)

	assert.Equal(suite.T(), objectList, createdList)

	sort.Strings(rogueList)
	sort.Strings(existingList)
	assert.Equal(suite.T(), rogueList, existingList)
}

func (suite *RadosTestSuite) TestNewConnWithUser() {
	_, err := NewConnWithUser("admin")
	assert.Equal(suite.T(), err, nil)
}

func (suite *RadosTestSuite) TestNewConnWithClusterAndUser() {
	_, err := NewConnWithClusterAndUser("ceph", "client.admin")
	assert.Equal(suite.T(), err, nil)
}

func (suite *RadosTestSuite) TestReadWriteXattr() {
	suite.SetupConnection()

	oid := suite.GenObjectName()
	val := []byte("value")
	err := suite.ioctx.SetXattr(oid, "key", val)
	assert.NoError(suite.T(), err)

	out := make([]byte, len(val))
	n, err := suite.ioctx.GetXattr(oid, "key", out)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), n, len(out))

	assert.Equal(suite.T(), out, val)
}

func (suite *RadosTestSuite) TestListXattrs() {
	suite.SetupConnection()

	oid := suite.GenObjectName()
	xattrs := make(map[string][]byte)
	val := []byte("value")
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("key_%d", i)
		err := suite.ioctx.SetXattr(oid, name, val)
		assert.NoError(suite.T(), err)
		xattrs[name] = val
	}

	out, err := suite.ioctx.ListXattrs(oid)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), xattrs, out)
}

func (suite *RadosTestSuite) TestRmXattr() {
	suite.SetupConnection()

	oid := suite.GenObjectName()

	// 2 xattrs
	xattrs := make(map[string][]byte)
	xattrs["key1"] = []byte("val")
	xattrs["key2"] = []byte("val")
	assert.Equal(suite.T(), len(xattrs), 2)

	// add them to the object
	for key, value := range xattrs {
		err := suite.ioctx.SetXattr(oid, key, value)
		assert.NoError(suite.T(), err)
	}
	out, err := suite.ioctx.ListXattrs(oid)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), len(out), 2)
	assert.Equal(suite.T(), out, xattrs)

	// remove key1
	err = suite.ioctx.RmXattr(oid, "key1")
	assert.NoError(suite.T(), err)
	delete(xattrs, "key1")

	// verify key1 is gone
	out, err = suite.ioctx.ListXattrs(oid)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), len(out), 1)
	assert.Equal(suite.T(), out, xattrs)

	// remove key2
	err = suite.ioctx.RmXattr(oid, "key2")
	assert.NoError(suite.T(), err)
	delete(xattrs, "key2")

	// verify key2 is gone
	out, err = suite.ioctx.ListXattrs(oid)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), len(out), 0)
	assert.Equal(suite.T(), out, xattrs)
}

func (suite *RadosTestSuite) TestReadWriteOmap() {
	suite.SetupConnection()

	// set some key/value pairs on an object
	orig := map[string][]byte{
		"key1":          []byte("value1"),
		"key2":          []byte("value2"),
		"prefixed-key3": []byte("value3"),
		"empty":         []byte(""),
	}

	oid := suite.GenObjectName()
	err := suite.ioctx.SetOmap(oid, orig)
	assert.NoError(suite.T(), err)

	// verify that they can all be read back
	remaining := map[string][]byte{}
	for k, v := range orig {
		remaining[k] = v
	}

	err = suite.ioctx.ListOmapValues(oid, "", "", 4, func(key string, value []byte) {
		assert.Equal(suite.T(), remaining[key], value)
		delete(remaining, key)
	})

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), len(remaining), 0)

	// Get (with a fixed number of keys)
	fetched, err := suite.ioctx.GetOmapValues(oid, "", "", 4)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), orig, fetched)

	// Get All (with an iterator size bigger than the map size)
	fetched, err = suite.ioctx.GetAllOmapValues(oid, "", "", 100)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), orig, fetched)

	// Get All (with an iterator size smaller than the map size)
	fetched, err = suite.ioctx.GetAllOmapValues(oid, "", "", 1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), orig, fetched)

	// Remove
	err = suite.ioctx.RmOmapKeys(oid, []string{"key1", "prefixed-key3"})
	assert.NoError(suite.T(), err)

	fetched, err = suite.ioctx.GetOmapValues(oid, "", "", 4)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), map[string][]byte{
		"key2":  []byte("value2"),
		"empty": []byte(""),
	}, fetched)

	// Clear
	err = suite.ioctx.CleanOmap(oid)
	assert.NoError(suite.T(), err)

	fetched, err = suite.ioctx.GetOmapValues(oid, "", "", 4)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), map[string][]byte{}, fetched)
}

func (suite *RadosTestSuite) TestReadFilterOmap() {
	suite.SetupConnection()

	orig := map[string][]byte{
		"key1":          []byte("value1"),
		"prefixed-key3": []byte("value3"),
		"key2":          []byte("value2"),
	}

	oid := suite.GenObjectName()
	err := suite.ioctx.SetOmap(oid, orig)
	assert.NoError(suite.T(), err)

	// filter by prefix
	fetched, err := suite.ioctx.GetOmapValues(oid, "", "prefixed", 4)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), map[string][]byte{
		"prefixed-key3": []byte("value3"),
	}, fetched)

	// "start_after" a key
	fetched, err = suite.ioctx.GetOmapValues(oid, "key1", "", 4)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), map[string][]byte{
		"prefixed-key3": []byte("value3"),
		"key2":          []byte("value2"),
	}, fetched)

	// maxReturn
	fetched, err = suite.ioctx.GetOmapValues(oid, "", "key", 1)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), map[string][]byte{
		"key1": []byte("value1"),
	}, fetched)
}

func (suite *RadosTestSuite) TestSetNamespace() {
	suite.SetupConnection()

	// create oid
	oid := suite.GenObjectName()
	bytesIn := []byte("input data")
	err := suite.ioctx.Write(oid, bytesIn, 0)
	assert.NoError(suite.T(), err)

	stat, err := suite.ioctx.Stat(oid)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), uint64(len(bytesIn)), stat.Size)
	assert.NotNil(suite.T(), stat.ModTime)

	// oid isn't seen in space1 ns
	suite.ioctx.SetNamespace("space1")
	stat, err = suite.ioctx.Stat(oid)
	assert.Equal(suite.T(), err, ErrNotFound)

	// create oid2 in space1 ns
	oid2 := suite.GenObjectName()
	bytesIn = []byte("input data")
	err = suite.ioctx.Write(oid2, bytesIn, 0)
	assert.NoError(suite.T(), err)

	suite.ioctx.SetNamespace("")
	stat, err = suite.ioctx.Stat(oid2)
	assert.Equal(suite.T(), err, ErrNotFound)

	stat, err = suite.ioctx.Stat(oid)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), uint64(len(bytesIn)), stat.Size)
	assert.NotNil(suite.T(), stat.ModTime)
}

func (suite *RadosTestSuite) TestListAcrossNamespaces() {
	suite.SetupConnection()

	// count objects in pool
	origObjects := 0
	err := suite.ioctx.ListObjects(func(string) {
		origObjects++
	})
	assert.NoError(suite.T(), err)

	// create oid
	oid := suite.GenObjectName()
	bytesIn := []byte("input data")
	err = suite.ioctx.Write(oid, bytesIn, 0)
	assert.NoError(suite.T(), err)

	// create oid2 in space1 ns
	suite.ioctx.SetNamespace("space1")
	oid2 := suite.GenObjectName()
	bytesIn = []byte("input data")
	err = suite.ioctx.Write(oid2, bytesIn, 0)
	assert.NoError(suite.T(), err)

	// count objects in space1 ns
	nsFoundObjects := 0
	err = suite.ioctx.ListObjects(func(string) {
		nsFoundObjects++
	})
	assert.NoError(suite.T(), err)
	assert.EqualValues(suite.T(), 1, nsFoundObjects)

	// count objects in pool
	suite.ioctx.SetNamespace(AllNamespaces)
	allFoundObjects := 0
	err = suite.ioctx.ListObjects(func(oid string) {
		allFoundObjects++
	})
	assert.NoError(suite.T(), err)
	assert.EqualValues(suite.T(), (origObjects + 2), allFoundObjects)
}

func (suite *RadosTestSuite) TestLocking() {
	suite.SetupConnection()

	oid := suite.GenObjectName()

	// lock ex
	res, err := suite.ioctx.LockExclusive(oid, "myLock", "myCookie", "this is a test lock", 0, nil)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, res)

	// verify lock ex
	info, err := suite.ioctx.ListLockers(oid, "myLock")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, len(info.Clients))
	assert.Equal(suite.T(), true, info.Exclusive)

	// fail to lock ex again
	res, err = suite.ioctx.LockExclusive(oid, "myLock", "myCookie", "this is a description", 0, nil)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), -17, res)

	// fail to lock sh
	res, err = suite.ioctx.LockShared(oid, "myLock", "myCookie", "", "a description", 0, nil)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), -17, res)

	// unlock
	res, err = suite.ioctx.Unlock(oid, "myLock", "myCookie")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, res)

	// verify unlock
	info, err = suite.ioctx.ListLockers(oid, "myLock")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, len(info.Clients))

	// lock sh
	res, err = suite.ioctx.LockShared(oid, "myLock", "myCookie", "", "a description", 0, nil)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, res)

	// verify lock sh
	info, err = suite.ioctx.ListLockers(oid, "myLock")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, len(info.Clients))
	assert.Equal(suite.T(), false, info.Exclusive)

	// fail to lock sh again
	res, err = suite.ioctx.LockExclusive(oid, "myLock", "myCookie", "a description", 0, nil)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), -17, res)

	// fail to lock ex
	res, err = suite.ioctx.LockExclusive(oid, "myLock", "myCookie", "this is a test lock", 0, nil)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), res, -17)

	// break the lock
	res, err = suite.ioctx.BreakLock(oid, "myLock", info.Clients[0], "myCookie")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, res)

	// verify lock broken
	info, err = suite.ioctx.ListLockers(oid, "myLock")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, len(info.Clients))

	// lock sh with duration
	res, err = suite.ioctx.LockShared(oid, "myLock", "myCookie", "", "a description", time.Millisecond, nil)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, res)

	// verify lock sh expired
	time.Sleep(time.Second)
	info, err = suite.ioctx.ListLockers(oid, "myLock")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, len(info.Clients))

	// lock sh with duration
	res, err = suite.ioctx.LockExclusive(oid, "myLock", "myCookie", "a description", time.Millisecond, nil)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, res)

	// verify lock sh expired
	time.Sleep(time.Second)
	info, err = suite.ioctx.ListLockers(oid, "myLock")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, len(info.Clients))
}

func (suite *RadosTestSuite) TestOmapOnNonexistentObjectError() {
	suite.SetupConnection()
	oid := suite.GenObjectName()
	_, err := suite.ioctx.GetAllOmapValues(oid, "", "", 100)
	assert.Equal(suite.T(), err, ErrNotFound)
}

func (suite *RadosTestSuite) TestOpenIOContextInvalidPool() {
	ioctx, err := suite.conn.OpenIOContext("cmartel")
	require.Error(suite.T(), err)
	require.Nil(suite.T(), ioctx)
}

func (suite *RadosTestSuite) TestGetLastVersion() {
	suite.SetupConnection()
	// NOTE: Reusing the ioctx set up by SetupConnection seems to make this
	// test flakey. This is possibly be due to the value being "global" to the
	// ioctx. Thus we create a new ioctx for the subtests.

	suite.T().Run("write", func(t *testing.T) {
		ioctx, err := suite.conn.OpenIOContext(suite.pool)
		require.NoError(suite.T(), err)
		oid := suite.GenObjectName()
		defer suite.ioctx.Delete(oid)

		v1, _ := ioctx.GetLastVersion()

		err = ioctx.Write(oid, []byte("something to write"), 0)
		assert.NoError(t, err)

		v2, _ := ioctx.GetLastVersion()
		assert.NotEqual(t, v1, v2)

		v3, _ := ioctx.GetLastVersion()
		assert.Equal(t, v2, v3)

		err = ioctx.Write(oid, []byte("something completely different"), 0)
		assert.NoError(t, err)

		v4, _ := ioctx.GetLastVersion()
		assert.NotEqual(t, v1, v4)
		assert.NotEqual(t, v2, v4)
		assert.NotEqual(t, v3, v4)
	})

	suite.T().Run("writeAndRead", func(t *testing.T) {
		ioctx, err := suite.conn.OpenIOContext(suite.pool)
		require.NoError(suite.T(), err)
		oid := suite.GenObjectName()
		defer ioctx.Delete(oid)

		v1, _ := ioctx.GetLastVersion()

		err = ioctx.Write(oid, []byte("presto"), 0)
		assert.NoError(t, err)

		v2, _ := ioctx.GetLastVersion()
		assert.NotEqual(t, v1, v2)

		bytes := make([]byte, 1024)
		_, err = ioctx.Read(oid, bytes, 0)
		assert.NoError(t, err)

		v3, _ := ioctx.GetLastVersion()
		assert.Equal(t, v2, v3)

		err = ioctx.Write(oid, []byte("abracadabra"), 0)
		assert.NoError(t, err)

		v4, _ := ioctx.GetLastVersion()
		assert.NotEqual(t, v1, v4)
		assert.NotEqual(t, v2, v4)
		assert.NotEqual(t, v3, v4)

		_, err = ioctx.Read(oid, bytes, 0)
		assert.NoError(t, err)

		v5, _ := ioctx.GetLastVersion()
		assert.Equal(t, v4, v5)
	})

	suite.T().Run("writeAndReadMultiple", func(t *testing.T) {
		ioctx, err := suite.conn.OpenIOContext(suite.pool)
		require.NoError(suite.T(), err)

		oids := make([]string, 5)
		vers := make([]uint64, 5)
		for i := 0; i < 5; i++ {
			oid := suite.GenObjectName()
			defer ioctx.Delete(oid)
			err = ioctx.Write(oid, []byte(oid), 0)
			assert.NoError(t, err)

			oids[i] = oid
			vers[i], _ = ioctx.GetLastVersion()
		}

		var v uint64
		bytes := make([]byte, 1024)

		_, err = ioctx.Read(oids[0], bytes, 0)
		assert.NoError(t, err)
		v, _ = ioctx.GetLastVersion()
		assert.Equal(t, vers[0], v)

		_, err = ioctx.Read(oids[4], bytes, 0)
		assert.NoError(t, err)
		v, _ = ioctx.GetLastVersion()
		assert.Equal(t, vers[4], v)

		_, err = ioctx.Read(oids[1], bytes, 0)
		assert.NoError(t, err)
		v, _ = ioctx.GetLastVersion()
		assert.Equal(t, vers[1], v)
	})

	suite.T().Run("invalidIOContext", func(t *testing.T) {
		ioctx := &IOContext{}
		_, err := ioctx.GetLastVersion()
		assert.Error(t, err)
	})
}

func (suite *RadosTestSuite) TestSetGetNamespace() {
	suite.SetupConnection()
	suite.T().Run("validNS", func(t *testing.T) {
		suite.ioctx.SetNamespace("space1")
		ns, err := suite.ioctx.GetNamespace()
		assert.NoError(t, err)
		assert.Equal(t, "space1", ns)
	})

	suite.T().Run("allNamespaces", func(t *testing.T) {
		suite.ioctx.SetNamespace(AllNamespaces)
		ns, err := suite.ioctx.GetNamespace()
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), AllNamespaces, ns)
	})

	suite.T().Run("invalidIoctx", func(t *testing.T) {
		i := &IOContext{}
		ns, err := i.GetNamespace()
		assert.Error(suite.T(), err)
		assert.Equal(suite.T(), "", ns)
	})
}

func TestRadosTestSuite(t *testing.T) {
	tsuite.Run(t, new(RadosTestSuite))
}
