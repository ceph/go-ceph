//go:build ceph_preview && ceph_main

package nfs

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ceph/go-ceph/internal/commands"
)

func (suite *NFSAdminSuite) TestExportSecType() {
	require := suite.Require()
	ra := radosConnector.Get(suite.T())
	nfsa := NewFromConn(ra)

	_, err := nfsa.CreateCephFSExport(CephFSExportSpec{
		FileSystemName: suite.fileSystemName,
		ClusterID:      suite.clusterID,
		PseudoPath:     "/12",
		Path:           "/sept",
		SecType:        []SecType{Krb5pSec, Krb5iSec, SysSec},
	})
	require.NoError(err)

	defer func() {
		err = nfsa.RemoveExport(suite.clusterID, "/12")
		require.NoError(err)
	}()

	e1, err := nfsa.ExportInfo(suite.clusterID, "/12")
	require.NoError(err)
	require.Equal(e1.PseudoPath, "/12")
	require.Equal(e1.SecType, []SecType{Krb5pSec, Krb5iSec, SysSec})
}

// # ceph nfs export info --cluster-id foobar --pseudo-path /cheese
const exportInfo2 = `
{
  "export_id": 2,
  "path": "/",
  "cluster_id": "goceph",
  "pseudo": "/bar",
  "access_type": "RW",
  "squash": "none",
  "security_label": true,
  "protocols": [
    4
  ],
  "transports": [
    "TCP"
  ],
  "fsal": {
    "name": "CEPH",
    "user_id": "nfs.goceph.2",
    "fs_name": "cephfs"
  },
  "clients": [],
  "sectype": [
    "krb5p",
    "sys"
  ]
}
`

func TestParseExportInfoWithSecType(t *testing.T) {
	t.Run("exportInfo1", func(t *testing.T) {
		r := commands.NewResponse([]byte(exportInfo1), "", nil)
		e, err := parseExportInfo(r)
		assert.NoError(t, err)
		assert.Equal(t, e.Path, "/")
		assert.Equal(t, e.PseudoPath, "/cheese")
		assert.Len(t, e.SecType, 0)
	})
	t.Run("exportInfo2", func(t *testing.T) {
		r := commands.NewResponse([]byte(exportInfo2), "", nil)
		e, err := parseExportInfo(r)
		assert.NoError(t, err)
		assert.Equal(t, e.Path, "/")
		assert.Equal(t, e.PseudoPath, "/bar")
		if assert.Len(t, e.SecType, 2) {
			assert.Equal(t, e.SecType[0], Krb5pSec)
			assert.Equal(t, e.SecType[1], SysSec)
		}
	})
}
