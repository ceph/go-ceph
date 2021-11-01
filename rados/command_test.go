package rados

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	osdNumber = 0
	monName   = "a"
)

func (suite *RadosTestSuite) TestMonCommand() {
	suite.SetupConnection()

	command, err := json.Marshal(
		map[string]string{"prefix": "df", "format": "json"})
	assert.NoError(suite.T(), err)

	buf, info, err := suite.conn.MonCommand(command)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), info, "")

	var message map[string]interface{}
	err = json.Unmarshal(buf, &message)
	assert.NoError(suite.T(), err)
}

// NB: ceph octopus appears to be stricter about the formatting of the keyring
// and now rejects whitespace that older versions did not have a problem with.
const clientKeyFormat = `
[%s]
key = AQD4PGNXBZJNHhAA582iUgxe9DsN+MqFN4Z6Jw==
`

func (suite *RadosTestSuite) TestMonCommandWithInputBuffer() {
	suite.SetupConnection()

	entity := fmt.Sprintf("client.testMonCmdUser%d", time.Now().UnixNano())

	// first add the new test user, specifying its key in the input buffer
	command, err := json.Marshal(map[string]interface{}{
		"prefix": "auth add",
		"format": "json",
		"entity": entity,
	})
	assert.NoError(suite.T(), err)

	clientKey := fmt.Sprintf(clientKeyFormat, entity)

	inbuf := []byte(clientKey)

	buf, info, err := suite.conn.MonCommandWithInputBuffer(command, inbuf)
	assert.NoError(suite.T(), err)
	expectedInfo := fmt.Sprintf("added key for %s", entity)
	assert.Equal(suite.T(), expectedInfo, info)
	assert.Equal(suite.T(), "", string(buf[:]))

	// get the key and verify that it's what we previously set
	command, err = json.Marshal(map[string]interface{}{
		"prefix": "auth get-key",
		"format": "json",
		"entity": entity,
	})
	assert.NoError(suite.T(), err)

	buf, info, err = suite.conn.MonCommand(command)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "", info)
	assert.Equal(suite.T(),
		`{"key":"AQD4PGNXBZJNHhAA582iUgxe9DsN+MqFN4Z6Jw=="}`,
		string(buf[:]))
}

func (suite *RadosTestSuite) TestPGCommand() {
	suite.SetupConnection()

	pgid := "1.2"

	command, err := json.Marshal(
		map[string]string{"prefix": "query", "pgid": pgid, "format": "json"})
	assert.NoError(suite.T(), err)

	buf, info, err := suite.conn.PGCommand([]byte(pgid), [][]byte{[]byte(command)})
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), info, "")

	var message map[string]interface{}
	err = json.Unmarshal(buf, &message)
	assert.NoError(suite.T(), err)
}

func (suite *RadosTestSuite) TestMgrCommandDescriptions() {
	suite.SetupConnection()

	command, err := json.Marshal(
		map[string]string{"prefix": "get_command_descriptions", "format": "json"})
	assert.NoError(suite.T(), err)

	buf, info, err := suite.conn.MgrCommand([][]byte{command})
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), info, "")

	var message map[string]interface{}
	err = json.Unmarshal(buf, &message)
	assert.NoError(suite.T(), err)
}

func (suite *RadosTestSuite) TestMgrCommand() {
	suite.SetupConnection()

	command, err := json.Marshal(
		map[string]string{"prefix": "balancer status", "format": "json"})
	assert.NoError(suite.T(), err)

	buf, info, err := suite.conn.MgrCommand([][]byte{command})
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), info, "")

	var message map[string]interface{}
	err = json.Unmarshal(buf, &message)
	assert.NoError(suite.T(), err)
}

func (suite *RadosTestSuite) TestMgrCommandMalformedCommand() {
	suite.SetupConnection()

	command := []byte("JUNK!")
	buf, info, err := suite.conn.MgrCommand([][]byte{command})
	assert.Error(suite.T(), err)
	assert.NotEqual(suite.T(), info, "")
	assert.Len(suite.T(), buf, 0)
}

func (suite *RadosTestSuite) TestOsdCommandDescriptions() {
	suite.SetupConnection()

	command, err := json.Marshal(
		map[string]string{"prefix": "get_command_descriptions", "format": "json"})
	assert.NoError(suite.T(), err)

	buf, info, err := suite.conn.OsdCommand(osdNumber, [][]byte{command})
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), info, "")

	var message map[string]interface{}
	err = json.Unmarshal(buf, &message)
	assert.NoError(suite.T(), err)
}

func (suite *RadosTestSuite) TestOsdCommand() {
	suite.SetupConnection()

	command, err := json.Marshal(
		map[string]string{"prefix": "version", "format": "json"})
	assert.NoError(suite.T(), err)

	buf, info, err := suite.conn.OsdCommand(osdNumber, [][]byte{command})
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), info, "")

	var message map[string]interface{}
	err = json.Unmarshal(buf, &message)
	assert.NoError(suite.T(), err)
}

func (suite *RadosTestSuite) TestOsdCommandMalformedCommand() {
	suite.SetupConnection()

	command := []byte("JUNK!")
	buf, info, err := suite.conn.OsdCommand(osdNumber, [][]byte{command})
	assert.Error(suite.T(), err)
	assert.NotEqual(suite.T(), info, "")
	assert.Len(suite.T(), buf, 0)
}

func (suite *RadosTestSuite) TestMonCommandTargetDescriptions() {
	suite.SetupConnection()

	command, err := json.Marshal(
		map[string]string{"prefix": "get_command_descriptions", "format": "json"})
	assert.NoError(suite.T(), err)

	buf, info, err := suite.conn.MonCommandTarget(monName, [][]byte{command})
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), info, "")

	var message map[string]interface{}
	err = json.Unmarshal(buf, &message)
	assert.NoError(suite.T(), err)
}

func (suite *RadosTestSuite) TestMonCommandTarget() {
	suite.SetupConnection()

	command, err := json.Marshal(
		map[string]string{"prefix": "mon_status"})
	assert.NoError(suite.T(), err)

	buf, info, err := suite.conn.MonCommandTarget(monName, [][]byte{command})
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), info, "")

	var message map[string]interface{}
	err = json.Unmarshal(buf, &message)
	assert.NoError(suite.T(), err)
}

func (suite *RadosTestSuite) TestMonCommandTargetMalformedCommand() {
	suite.SetupConnection()

	command := []byte("JUNK!")
	buf, info, err := suite.conn.MonCommandTarget(monName, [][]byte{command})
	assert.Error(suite.T(), err)
	assert.NotEqual(suite.T(), info, "")
	assert.Len(suite.T(), buf, 0)
}

// Does not work on ceph luminous, but we do not support ceph luminous.
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
