package rados

import (
	"github.com/stretchr/testify/assert"
)

// timeStamp generates a dummy Timespec value.
func timeStamp() Timespec {
	// Future TODO (maybe?) - vary the value?
	return Timespec{342334800, 0}
}

func (suite *RadosTestSuite) TestWriteOpCreate() {
	suite.SetupConnection()

	op := CreateWriteOp()
	defer op.Release()
	op.Create(CreateIdempotent)
	err := op.Operate(suite.ioctx, "TestWriteOpCreate", OperationNoFlag)
	assert.NoError(suite.T(), err)

	op2 := CreateWriteOp()
	defer op2.Release()
	op.Create(CreateExclusive)
	err = op.Operate(suite.ioctx, "TestWriteOpCreate", OperationNoFlag)
	assert.Error(suite.T(), err)
}

func (suite *RadosTestSuite) TestWriteOpCreateWithTimestamp() {
	suite.SetupConnection()

	oid := "TestWriteOpCreateWithTimestamp"
	gts := timeStamp()
	op := CreateWriteOp()
	defer op.Release()
	op.Create(CreateIdempotent)
	err := op.OperateWithMtime(suite.ioctx, oid, gts, OperationNoFlag)
	assert.NoError(suite.T(), err)

	s, err := suite.ioctx.Stat(oid)
	assert.NoError(suite.T(), err)
	statTime := s.ModTime.Unix()
	assert.Equal(suite.T(), gts.Sec, statTime)
}

func (suite *RadosTestSuite) TestWriteOpSetOmap() {
	suite.SetupConnection()
	ta := assert.New(suite.T())

	oid := "TestWriteOpSetOmap"
	op := CreateWriteOp()
	defer op.Release()
	op.Create(CreateIdempotent)
	op.SetOmap(map[string][]byte{
		"alice":   []byte("car"),
		"boss":    []byte("office"),
		"catbert": []byte("dungeon"),
	})
	err := op.Operate(suite.ioctx, oid, OperationNoFlag)
	ta.NoError(err)

	// the 2nd set of omap values should not be applied because
	// the Create will fail to exclusively make the object
	op2 := CreateWriteOp()
	defer op2.Release()
	op.Create(CreateExclusive)
	op.SetOmap(map[string][]byte{
		"alice":   []byte("home"),
		"boss":    []byte("golf course"),
		"catbert": []byte("dungeon"),
	})
	err = op.Operate(suite.ioctx, oid, OperationNoFlag)
	ta.Error(err)

	fetched, err := suite.ioctx.GetOmapValues(oid, "", "", 10)
	ta.NoError(err)
	ta.Equal("car", string(fetched["alice"]))
	ta.Equal("office", string(fetched["boss"]))
}

func (suite *RadosTestSuite) TestWriteOpRmOmapKeys() {
	suite.SetupConnection()
	ta := assert.New(suite.T())

	oid := "TestWriteOpRmOmapKeys"
	op := CreateWriteOp()
	defer op.Release()
	op.Create(CreateIdempotent)
	op.SetOmap(map[string][]byte{
		"alice":   []byte("car"),
		"boss":    []byte("office"),
		"catbert": []byte("dungeon"),
	})
	err := op.Operate(suite.ioctx, oid, OperationNoFlag)
	ta.NoError(err)

	op2 := CreateWriteOp()
	defer op2.Release()
	op2.Create(CreateIdempotent)
	op2.SetOmap(map[string][]byte{
		"dogbert": []byte("lab"),
	})
	op2.RmOmapKeys([]string{"catbert"})
	err = op2.Operate(suite.ioctx, oid, OperationNoFlag)
	ta.NoError(err)

	fetched, err := suite.ioctx.GetOmapValues(oid, "", "", 10)
	ta.NoError(err)
	ta.Len(fetched, 3)
	ta.Equal("car", string(fetched["alice"]))
	ta.Equal("office", string(fetched["boss"]))
	ta.Equal("lab", string(fetched["dogbert"]))
	ta.NotContains(fetched, "catbert")
}

func (suite *RadosTestSuite) TestWriteOpCleanOmap() {
	suite.SetupConnection()
	ta := assert.New(suite.T())

	oid := "TestWriteOpCleanOmap"
	op := CreateWriteOp()
	defer op.Release()
	op.Create(CreateIdempotent)
	op.SetOmap(map[string][]byte{
		"alice":   []byte("car"),
		"boss":    []byte("office"),
		"catbert": []byte("dungeon"),
	})
	err := op.Operate(suite.ioctx, oid, OperationNoFlag)
	ta.NoError(err)

	// this test simulates wanting to start a fresh new set of
	// omap keys, atomically clearing and setting a new key & value.
	op2 := CreateWriteOp()
	defer op2.Release()
	op2.Create(CreateIdempotent)
	op2.CleanOmap()
	op2.SetOmap(map[string][]byte{
		"dogbert": []byte("lab"),
	})
	op2.RmOmapKeys([]string{"catbert"})
	err = op2.Operate(suite.ioctx, oid, OperationNoFlag)
	ta.NoError(err)

	fetched, err := suite.ioctx.GetOmapValues(oid, "", "", 10)
	ta.NoError(err)
	ta.Len(fetched, 1)
	ta.Equal("lab", string(fetched["dogbert"]))
}

func (suite *RadosTestSuite) TestWriteOpAssertExists() {
	suite.SetupConnection()
	ta := assert.New(suite.T())

	oid := "TestWriteOpRmOmapKeys"
	op := CreateWriteOp()
	defer op.Release()
	op.Create(CreateIdempotent)
	err := op.Operate(suite.ioctx, oid, OperationNoFlag)
	ta.NoError(err)

	op2 := CreateWriteOp()
	defer op2.Release()
	op2.AssertExists()
	op2.CleanOmap()
	err = op2.Operate(suite.ioctx, oid, OperationNoFlag)
	ta.NoError(err)

	op3 := CreateWriteOp()
	defer op3.Release()
	op3.AssertExists()
	op3.CleanOmap()
	err = op3.Operate(suite.ioctx, oid+"dne", OperationNoFlag)
	ta.Error(err)
}

func (suite *RadosTestSuite) TestWriteOpWrite() {
	suite.SetupConnection()
	ta := assert.New(suite.T())

	oid := "TestWriteOpWrite"
	op := CreateWriteOp()
	defer op.Release()
	op.Create(CreateIdempotent)
	op.Write([]byte("go-go-gadget"), 0)
	op.Write([]byte("ceph project!"), 6)
	err := op.Operate(suite.ioctx, oid, OperationNoFlag)
	ta.NoError(err)

	d := make([]byte, 32)
	l, err := suite.ioctx.Read(oid, d, 0)
	ta.NoError(err)
	ta.Equal("go-go-ceph project!", string(d[:l]))
}

func (suite *RadosTestSuite) TestWriteOpWriteFull() {
	suite.SetupConnection()
	ta := assert.New(suite.T())

	oid := "TestWriteOpWriteFull"
	op := CreateWriteOp()
	defer op.Release()
	op.Create(CreateIdempotent)
	op.WriteFull([]byte("one, two"))
	op.WriteFull([]byte("buckle my shoe"))
	err := op.Operate(suite.ioctx, oid, OperationNoFlag)
	ta.NoError(err)

	d := make([]byte, 32)
	l, err := suite.ioctx.Read(oid, d, 0)
	ta.NoError(err)
	ta.Equal("buckle my shoe", string(d[:l]))
}

func (suite *RadosTestSuite) TestWriteOpWriteSame() {
	suite.SetupConnection()
	ta := assert.New(suite.T())

	oid := "TestWriteOpWriteSame"
	op := CreateWriteOp()
	defer op.Release()
	op.Create(CreateIdempotent)
	op.WriteSame([]byte("repetition "), 44, 0)
	op.Write([]byte("is fun"), 44)
	err := op.Operate(suite.ioctx, oid, OperationNoFlag)
	ta.NoError(err)

	d := make([]byte, 64)
	l, err := suite.ioctx.Read(oid, d, 0)
	ta.NoError(err)
	ta.Equal("repetition repetition repetition repetition is fun", string(d[:l]))
}
