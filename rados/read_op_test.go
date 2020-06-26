package rados

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func (suite *RadosTestSuite) TestReadOpAssertExists() {
	suite.SetupConnection()
	oid := "TestReadOpAssertExists"

	wrop := CreateWriteOp()
	defer wrop.Release()
	wrop.Create(CreateIdempotent)
	err := wrop.Operate(suite.ioctx, oid, OperationNoFlag)
	assert.NoError(suite.T(), err)

	op := CreateReadOp()
	defer op.Release()
	op.AssertExists()
	err = op.Operate(suite.ioctx, oid, OperationNoFlag)
	assert.NoError(suite.T(), err)

	op2 := CreateReadOp()
	defer op2.Release()
	op2.AssertExists()
	err = op2.Operate(suite.ioctx, oid+"junk", OperationNoFlag)
	assert.Error(suite.T(), err)
}

func omeMap(ome *OmapGetElement) map[string][]byte {
	r := make(map[string][]byte)
	for {
		kv, err := ome.Next()
		if err != nil {
			panic(err)
		}
		if kv == nil {
			break
		}
		r[kv.Key] = kv.Value
	}
	return r
}

func (suite *RadosTestSuite) TestReadOpGetOmapValues() {
	suite.SetupConnection()
	ta := assert.New(suite.T())
	oid := "TestReadOpGetOmapValues"

	wrop := CreateWriteOp()
	defer wrop.Release()
	wrop.Create(CreateIdempotent)
	wrop.SetOmap(map[string][]byte{
		"tos.captain":       []byte("Kirk"),
		"tos.first-officer": []byte("Spock"),
		"tos.doctor":        []byte("McCoy"),
		"tng.captain":       []byte("Picard"),
		"tng.first-officer": []byte("Riker"),
		"tng.doctor":        []byte("Crusher"),
		"random.value":      []byte("foobar"),
		"no.value":          []byte(""),
	})
	err := wrop.Operate(suite.ioctx, oid, OperationNoFlag)
	ta.NoError(err)

	suite.T().Run("simple", func(t *testing.T) {
		ta := assert.New(t)
		op := CreateReadOp()
		defer op.Release()
		op.AssertExists()
		ome := op.GetOmapValues("", "", 16)
		err = op.Operate(suite.ioctx, oid, OperationNoFlag)
		ta.NoError(err)

		omap := omeMap(ome)
		ta.Len(omap, 8)
		ta.Contains(omap, "tos.captain")
		ta.Contains(omap, "tng.captain")
		ta.False(ome.More())
	})

	suite.T().Run("twoIterations", func(t *testing.T) {
		// test two iterations over different subsets of omap keys
		ta := assert.New(t)
		op := CreateReadOp()
		defer op.Release()
		op.AssertExists()
		ome1 := op.GetOmapValues("", "tos", 16)
		ome2 := op.GetOmapValues("", "tng", 16)
		err = op.Operate(suite.ioctx, oid, OperationNoFlag)
		ta.NoError(err)

		omap1 := omeMap(ome1)
		ta.Len(omap1, 3)
		ta.Contains(omap1, "tos.captain")
		ta.Contains(omap1, "tos.first-officer")
		ta.Contains(omap1, "tos.doctor")
		omap2 := omeMap(ome2)
		ta.Len(omap2, 3)
		ta.Contains(omap2, "tng.captain")
		ta.Contains(omap2, "tng.first-officer")
		ta.Contains(omap2, "tng.doctor")
	})

	suite.T().Run("checkForMore", func(t *testing.T) {
		// test two iterations over different subsets of omap keys
		ta := assert.New(t)
		op := CreateReadOp()
		defer op.Release()
		op.AssertExists()
		ome := op.GetOmapValues("", "", 6)
		err = op.Operate(suite.ioctx, oid, OperationNoFlag)
		ta.NoError(err)

		omap1 := omeMap(ome)
		ta.Len(omap1, 6)
		ta.True(ome.More())
	})

	suite.T().Run("iterateTooEarly", func(t *testing.T) {
		// test two iterations over different subsets of omap keys
		ta := assert.New(t)
		op := CreateReadOp()
		defer op.Release()
		op.AssertExists()
		ome := op.GetOmapValues("", "", 6)
		_, err := ome.Next()
		ta.Error(err)
		ta.Equal(ErrOperationIncomplete, err)
	})
}
