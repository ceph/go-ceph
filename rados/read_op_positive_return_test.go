package rados

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// luaReturn42 is input for the lua class's eval_json method: a script that
// appends "lua" to the output and returns 42 without changing the object.
const luaReturn42 = `{"handler": "f", "script": "` +
	`function f(input, output) output:append('lua'); return 42; end ` +
	`objclass.register(f)"}`

// The hello class's write_return_data method returns 42.
// With RETURNVEC, both the operation and later omap steps can return 42.
// These positive results must be treated as success.
func (suite *RadosTestSuite) TestReadOpPositiveReturn() {
	suite.SetupConnection()

	suite.T().Run("getOmapValuesByKeys", func(t *testing.T) {
		ta := assert.New(t)
		oid := suite.GenObjectName()
		require.NoError(t, suite.ioctx.SetOmap(
			oid, map[string][]byte{"key": []byte("value")}))

		rop := CreateReadOp()
		defer rop.Release()
		es := rop.Exec("hello", "write_return_data", nil)
		kvStep := rop.GetOmapValuesByKeys([]string{"key"})
		ta.NoError(rop.Operate(suite.ioctx, oid, OperationReturnVec))

		out, err := es.Bytes()
		ta.NoError(err)
		ta.Equal([]byte("you might see this"), out)
		kv, err := kvStep.Next()
		ta.NoError(err)
		ta.Equal(&OmapKeyValue{Key: "key", Value: []byte("value")}, kv)
	})

	// without omap the OSD skips the omap iteration and leaves the
	// previous result in place
	suite.T().Run("getOmapValues", func(t *testing.T) {
		ta := assert.New(t)
		oid := suite.GenObjectName()

		rop := CreateReadOp()
		defer rop.Release()
		es := rop.Exec("hello", "write_return_data", nil)
		gos := rop.GetOmapValues("", "", 10)
		ta.NoError(rop.Operate(suite.ioctx, oid, OperationReturnVec))

		out, err := es.Bytes()
		ta.NoError(err)
		ta.Equal([]byte("you might see this"), out)
		kv, err := gos.Next()
		ta.NoError(err)
		ta.Nil(kv)
	})

	// The lua class's methods are writes, but a script that changes nothing
	// leaves the transaction empty, so its 42 reaches the client without
	// RETURNVEC.
	suite.T().Run("exec", func(t *testing.T) {
		ta := assert.New(t)
		oid := suite.GenObjectName()

		rop := CreateReadOp()
		defer rop.Release()
		es := rop.Exec("lua", "eval_json", []byte(luaReturn42))
		ta.NoError(rop.Operate(suite.ioctx, oid, OperationNoFlag))

		out, err := es.Bytes()
		ta.NoError(err)
		ta.Equal([]byte("lua"), out)
	})
}
