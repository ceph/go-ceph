//go:build !luminous
// +build !luminous

package rados

import (
	"encoding/json"

	"github.com/stretchr/testify/assert"
)

// A real test using input buffer is hard to find for mgr.
// The simplest does not work on luminous, so we simply don't
// provide the test for luminous.

func (suite *RadosTestSuite) TestMgrCommandWithInputBuffer() {
	suite.SetupConnection()

	command, err := json.Marshal(
		map[string]string{"prefix": "crash post", "format": "json"})
	assert.NoError(suite.T(), err)

	buf, info, err := suite.conn.MgrCommandWithInputBuffer(
		[][]byte{command}, []byte(`{"crash_id": "foobar", "timestamp": "2020-04-10 15:08:34.659679Z"}`))
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), info, "")
	assert.Len(suite.T(), buf, 0)

	command, err = json.Marshal(
		map[string]string{"prefix": "crash rm", "id": "foobar", "format": "json"})
	assert.NoError(suite.T(), err)

	buf, info, err = suite.conn.MgrCommandWithInputBuffer(
		[][]byte{command}, nil)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), info, "")
	assert.Len(suite.T(), buf, 0)
}
