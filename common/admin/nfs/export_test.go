//go:build !(nautilus || octopus) && ceph_preview && ceph_ci_untested
// +build !nautilus,!octopus,ceph_preview,ceph_ci_untested

package nfs

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	tsuite "github.com/stretchr/testify/suite"

	"github.com/ceph/go-ceph/internal/admintest"
	"github.com/ceph/go-ceph/internal/commands"
)

var radosConnector = admintest.NewConnector()

func TestNFSAdmin(t *testing.T) {
	tsuite.Run(t, new(NFSAdminSuite))
}

// NFSAdminSuite is a suite of tests for the nfs admin package.
// A suite is used because creating/managing NFS has certain expectations for
// the cluster that we may need to mock, especially when running in the
// standard go-ceph test container. Using a suite allows us to have suite
// setup/validate the environment and the tests can largely be ignorant of the
// environment.
type NFSAdminSuite struct {
	tsuite.Suite

	fileSystemName string
	clusterID      string
}

func (suite *NFSAdminSuite) SetupSuite() {
	suite.fileSystemName = "cephfs"
	suite.clusterID = "goceph"
}

func (suite *NFSAdminSuite) TestCreateDeleteCephFSExport() {
	require := suite.Require()
	ra := radosConnector.Get(suite.T())
	nfsa := NewFromConn(ra)

	res, err := nfsa.CreateCephFSExport(CephFSExportSpec{
		FileSystemName: suite.fileSystemName,
		ClusterID:      suite.clusterID,
		PseudoPath:     "/cheese",
	})
	require.NoError(err)
	require.Equal("/cheese", res.Bind)

	err = nfsa.RemoveExport(suite.clusterID, "/cheese")
	require.NoError(err)
}

func (suite *NFSAdminSuite) TestListDetailedExports() {
	require := suite.Require()
	ra := radosConnector.Get(suite.T())
	nfsa := NewFromConn(ra)

	_, err := nfsa.CreateCephFSExport(CephFSExportSpec{
		FileSystemName: suite.fileSystemName,
		ClusterID:      suite.clusterID,
		PseudoPath:     "/01",
		Path:           "/january",
	})
	require.NoError(err)

	defer func() {
		err = nfsa.RemoveExport(suite.clusterID, "/01")
		require.NoError(err)
	}()

	_, err = nfsa.CreateCephFSExport(CephFSExportSpec{
		FileSystemName: suite.fileSystemName,
		ClusterID:      suite.clusterID,
		PseudoPath:     "/02",
		Path:           "/february",
	})
	require.NoError(err)

	defer func() {
		err = nfsa.RemoveExport(suite.clusterID, "/02")
		require.NoError(err)
	}()

	l, err := nfsa.ListDetailedExports(suite.clusterID)
	require.NoError(err)
	require.Len(l, 2)
	var e1, e2 ExportInfo
	for _, e := range l {
		if e.PseudoPath == "/01" {
			e1 = e
		}
		if e.PseudoPath == "/02" {
			e2 = e
		}
	}
	require.Equal(e1.PseudoPath, "/01")
	require.Equal(e2.PseudoPath, "/02")
}

func (suite *NFSAdminSuite) TestExportInfo() {
	require := suite.Require()
	ra := radosConnector.Get(suite.T())
	nfsa := NewFromConn(ra)

	_, err := nfsa.CreateCephFSExport(CephFSExportSpec{
		FileSystemName: suite.fileSystemName,
		ClusterID:      suite.clusterID,
		PseudoPath:     "/03",
		Path:           "/march",
	})
	require.NoError(err)

	defer func() {
		err = nfsa.RemoveExport(suite.clusterID, "/03")
		require.NoError(err)
	}()

	e1, err := nfsa.ExportInfo(suite.clusterID, "/03")
	require.NoError(err)
	require.Equal(e1.PseudoPath, "/03")

	_, err = nfsa.ExportInfo(suite.clusterID, "/88")
	require.Error(err)
}

const resultExport1 = `{
    "bind": "/cheese",
    "fs": "cephfs",
    "path": "/",
    "cluster": "foobar",
    "mode": "RW"
}
`

func TestParseExportResult(t *testing.T) {
	t.Run("resultExport", func(t *testing.T) {
		r := commands.NewResponse([]byte(resultExport1), "", nil)
		e, err := parseExportResult(r)
		assert.NoError(t, err)
		assert.Equal(t, e.Bind, "/cheese")
	})
	t.Run("errorSet", func(t *testing.T) {
		r := commands.NewResponse([]byte(""), "", errors.New("beep"))
		_, err := parseExportResult(r)
		assert.Error(t, err)
	})
	t.Run("statusSet", func(t *testing.T) {
		r := commands.NewResponse([]byte(""), "boo", nil)
		_, err := parseExportResult(r)
		assert.Error(t, err)
	})
}

// # ceph nfs export ls --cluster-id foobar --detailed
const exportList1 = `[
  {
    "export_id": 1,
    "path": "/",
    "cluster_id": "foobar",
    "pseudo": "/cheese",
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
      "user_id": "nfs.foobar.1",
      "fs_name": "cephfs"
    },
    "clients": []
  }
]`

func TestParseExportsList(t *testing.T) {
	t.Run("exportList1", func(t *testing.T) {
		r := commands.NewResponse([]byte(exportList1), "", nil)
		l, err := parseExportsList(r)
		assert.NoError(t, err)
		if assert.Len(t, l, 1) {
			e := l[0]
			assert.Equal(t, e.Path, "/")
			assert.Equal(t, e.PseudoPath, "/cheese")
			if assert.Len(t, e.Protocols, 1) {
				assert.Equal(t, e.Protocols[0], 4)
			}
			if assert.Len(t, e.Transports, 1) {
				assert.Equal(t, e.Transports[0], "TCP")
			}
			assert.Equal(t, e.FSAL.Name, "CEPH")
		}
	})
	t.Run("errorSet", func(t *testing.T) {
		r := commands.NewResponse([]byte(""), "", errors.New("beep"))
		_, err := parseExportsList(r)
		assert.Error(t, err)
	})
	t.Run("statusSet", func(t *testing.T) {
		r := commands.NewResponse([]byte(""), "boo", nil)
		_, err := parseExportsList(r)
		assert.Error(t, err)
	})
}

// # ceph nfs export info --cluster-id foobar --pseudo-path /cheese
const exportInfo1 = `{
  "export_id": 1,
  "path": "/",
  "cluster_id": "foobar",
  "pseudo": "/cheese",
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
    "user_id": "nfs.foobar.1",
    "fs_name": "cephfs"
  },
  "clients": []
}
`

func TestParseExportInfo(t *testing.T) {
	t.Run("exportInfo1", func(t *testing.T) {
		r := commands.NewResponse([]byte(exportInfo1), "", nil)
		e, err := parseExportInfo(r)
		assert.NoError(t, err)
		assert.Equal(t, e.Path, "/")
		assert.Equal(t, e.PseudoPath, "/cheese")
		if assert.Len(t, e.Protocols, 1) {
			assert.Equal(t, e.Protocols[0], 4)
		}
		if assert.Len(t, e.Transports, 1) {
			assert.Equal(t, e.Transports[0], "TCP")
		}
		assert.Equal(t, e.FSAL.Name, "CEPH")
	})
	t.Run("errorSet", func(t *testing.T) {
		r := commands.NewResponse([]byte(""), "", errors.New("beep"))
		_, err := parseExportInfo(r)
		assert.Error(t, err)
	})
	t.Run("statusSet", func(t *testing.T) {
		r := commands.NewResponse([]byte(""), "boo", nil)
		_, err := parseExportInfo(r)
		assert.Error(t, err)
	})
}
