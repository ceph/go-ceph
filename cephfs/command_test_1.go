package cephfs

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMdsCommand(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)

	cmd := []byte(`{"prefix": "client ls"}`)
	buf, info, err := mount.MdsCommand(
		testMdsName,
		[][]byte{cmd})
	assert.NoError(t, err)
	assert.NotEqual(t, "", string(buf))
	assert.Equal(t, "", string(info))
	assert.Contains(t, string(buf), "ceph_version")
	// response should also be valid json
	var j []interface{}
	err = json.Unmarshal(buf, &j)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(j), 1)
}

func TestMdsCommandError(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)

	cmd := []byte("iAMinValId~~~")
	buf, info, err := mount.MdsCommand(
		testMdsName,
		[][]byte{cmd})
	assert.Error(t, err)
	assert.Equal(t, "", string(buf))
	assert.NotEqual(t, "", string(info))
	assert.Contains(t, string(info), "unparseable JSON")
}
