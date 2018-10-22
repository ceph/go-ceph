package api

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsInBlackList(t *testing.T) {
	var obl OsdBlacklistLs
	body := `{"status": "listed 1 entries", "output": [{"addr": "1.2.3.4:0/0", "until": "2018-04-05 10:36:29.496087"},{"addr": "10.160.10.102:0/1708448121", "until": "2018-01-03 19:44:08.080229"}]}`
	err := json.Unmarshal([]byte(body), &obl)
	if err != nil {
		t.Fatal(err)
	}

	assert.True(t, obl.IsInBlacklist("1.2.3.4"))
	assert.False(t, obl.IsInBlacklist("10.160.10.102"))
}

func TestIsInBlackListAndContained(t *testing.T) {
	var obl OsdBlacklistLs
	body := `{"status": "listed 1 entries", "output": [{"addr": "1.2.3.40:0/0", "until": "2018-04-05 10:36:29.496087"},{"addr": "10.160.10.102:0/1708448121", "until": "2018-01-03 19:44:08.080229"}]}`
	err := json.Unmarshal([]byte(body), &obl)
	if err != nil {
		t.Fatal(err)
	}

	assert.False(t, obl.IsInBlacklist("1.2.3.4"))
}
