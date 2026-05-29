//go:build !(pacific || quincy || reef || squid) && ceph_preview

package smb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var genericResults = `
{
  "results": [
    {
      "resource": {
        "resource_type": "ceph.smb.fish",
        "external_id": "tuna",
        "info": {
          "count": "123456"
        }
      },
      "state": "created",
      "success": true
    },
    {
      "resource": {
        "resource_type": "ceph.smb.share",
        "cluster_id": "c1",
        "share_id": "s1",
        "intent": "present",
        "name": "s1",
        "readonly": false,
        "browseable": true,
        "cephfs": {
          "volume": "cephfs",
          "path": "/"
        }
      },
      "state": "created",
      "success": true
    }
  ],
  "success": true
}
`

func TestApplyWithGenericResults(t *testing.T) {
	sa := NewFromConn(&phoneyConn{
		mgrCommandWithInputBuffer: func(_ [][]byte, _ []byte) ([]byte, string, error) {
			return []byte(genericResults), "", nil
		},
	})
	gr := &GenericResource{
		Values: map[string]any{
			"resource_type": "ceph.smb.fish",
			"intent":        "present",
			"external_id":   "tuna",
			"info": map[string]any{
				"count": "123456",
			},
		},
	}
	share := NewShare("c1", "s1").SetCephFS("cephfs", "", "", "/")
	opts := &ApplyOptions{}
	opts.SetGeneric(true)
	rgroup, err := sa.Apply([]Resource{gr, share}, opts)
	assert.NoError(t, err)

	assert.True(t, rgroup.Ok())
	assert.Len(t, rgroup.Results, 2)

	gr1, ok := rgroup.Results[0].Resource().(*GenericResource)
	assert.True(t, ok)
	assert.Equal(t, gr1.Type(), ResourceType("ceph.smb.fish"))
	assert.Equal(t, gr1.Values["external_id"].(string), "tuna")
	assert.Equal(t, rgroup.Results[0].State(), "created")

	gr2, ok := rgroup.Results[1].Resource().(*GenericResource)
	assert.True(t, ok)
	assert.Equal(t, gr2.Type(), ShareType)
}
