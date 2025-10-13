//go:build !(octopus || pacific || quincy || reef || squid)

package smb

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tsuite "github.com/stretchr/testify/suite"

	fsadmin "github.com/ceph/go-ceph/cephfs/admin"
	"github.com/ceph/go-ceph/common/admin/manager"
	"github.com/ceph/go-ceph/internal/admintest"
	"github.com/ceph/go-ceph/internal/commands"
)

const modName = "smb"

func TestSMBAdmin(t *testing.T) {
	tsuite.Run(t, new(SMBAdminSuite))
}

// SMBAdminSuite is a suite of tests for the smb admin package.
type SMBAdminSuite struct {
	tsuite.Suite

	fileSystemName string
	vconn          *admintest.Connector
}

func (suite *SMBAdminSuite) SetupSuite() {
	suite.vconn = admintest.NewConnector()
	suite.disableOrch()
	suite.enableSMB()
	suite.waitForSMBResponsive()
	suite.configureSubVolume()
}

func (suite *SMBAdminSuite) TearDownSuite() {
	suite.removeSubVolume()
	suite.disableSMB()
}

func (suite *SMBAdminSuite) enableSMB() {
	t := suite.T()
	t.Logf("enabling smb module")
	mgradmin := manager.NewFromConn(suite.vconn.Get(t))
	err := mgradmin.EnableModule(modName, true)
	if err != nil && strings.Contains(err.Error(), "already enabled") {
		return
	}
	assert.NoError(t, err)
}

func (suite *SMBAdminSuite) waitForSMBOLD() {
	t := suite.T()
	t.Logf("waiting for smb module")
	time.Sleep(100 * time.Millisecond)
	mgradmin := manager.NewFromConn(suite.vconn.Get(t))
	for i := 0; i < 30; i++ {
		modinfo, err := mgradmin.ListModules()
		require.NoError(t, err)
		for _, emod := range modinfo.EnabledModules {
			if emod == modName {
				// give additional time for mgr to restart(?)
				time.Sleep(200 * time.Millisecond)
				return
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for smb module")
}

func (suite *SMBAdminSuite) waitForSMBResponsive() {
	// wait until smb module is responsive
	sa := NewFromConn(suite.vconn.Get(suite.T()))
	for i := 0; i < 30; i++ {
		_, err := sa.Show(nil, nil)
		if err == nil {
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
	suite.T().Fatalf("show command never succeeded - module not ready?")
}

func (suite *SMBAdminSuite) disableSMB() {
	suite.T().Logf("disabling smb module")
	mgradmin := manager.NewFromConn(suite.vconn.Get(suite.T()))
	err := mgradmin.DisableModule("smb")
	assert.NoError(suite.T(), err)
	time.Sleep(500 * time.Millisecond)
}

func (suite *SMBAdminSuite) disableOrch() {
	suite.T().Logf("disabling smb orch")
	time.Sleep(10 * time.Millisecond)
	// ceph config set mgr mgr/smb/update_orchestration false
	m := map[string]any{
		"prefix": "config set",
		"who":    "mgr",
		"name":   "mgr/smb/update_orchestration",
		"value":  "false",
		"force":  true, // needed to run this before enabling module
	}
	a := suite.vconn.Get(suite.T())
	err := commands.MarshalMonCommand(a, m).NoData().End()
	assert.NoError(suite.T(), err)
}

func (suite *SMBAdminSuite) configureSubVolume() {
	// set up a subvolume for use by smb commands
	fsa := fsadmin.NewFromConn(suite.vconn.Get(suite.T()))
	err := fsa.CreateSubVolumeGroup("cephfs", "smb", nil)
	assert.NoError(suite.T(), err)
	err = fsa.CreateSubVolume("cephfs", "smb", "v1", nil)
	assert.NoError(suite.T(), err)
}

func (suite *SMBAdminSuite) removeSubVolume() {
	fsa := fsadmin.NewFromConn(suite.vconn.Get(suite.T()))
	err := fsa.RemoveSubVolume("cephfs", "smb", "v1")
	assert.NoError(suite.T(), err)
	err = fsa.RemoveSubVolumeGroup("cephfs", "smb")
	assert.NoError(suite.T(), err)
}

func (suite *SMBAdminSuite) SetupTest() {
	suite.waitForSMBResponsive()
}

func (suite *SMBAdminSuite) TearDownTest() {
	sa := NewFromConn(suite.vconn.Get(suite.T()))
	r, err := sa.Show(nil, nil)
	suite.Assert().NoError(err)

	for i := range r {
		if sio, ok := r[i].(interface{ SetIntent(Intent) }); ok {
			sio.SetIntent(Removed)
			continue
		}
		// TODO: maybe go back and make a SetIntent receiver for the
		// existing core resource types
		switch res := r[i].(type) {
		case *Cluster:
			res.IntentValue = Removed
		case *Share:
			res.IntentValue = Removed
		case *JoinAuth:
			res.IntentValue = Removed
		case *UsersAndGroups:
			res.IntentValue = Removed
		}
	}
	if len(r) > 0 {
		_, err = sa.Apply(r, nil)
		suite.Assert().NoError(err)
	}
}

func (suite *SMBAdminSuite) TestShowEmpty() {
	t := suite.T()
	sa := NewFromConn(suite.vconn.Get(t))
	r, err := sa.Show(nil, nil)
	assert.NoError(t, err)
	assert.Len(t, r, 0)
}

func (suite *SMBAdminSuite) TestCreateCluster() {
	cluster := NewUserCluster("clu1")
	ug := NewLinkedUsersAndGroups(cluster).SetValues(
		[]UserInfo{
			{"alice", "W0nder14nd"},
			{"billy", "p14n0m4N"},
		},
		[]GroupInfo{{"clients"}},
	)

	t := suite.T()
	sa := NewFromConn(suite.vconn.Get(t))
	rg, err := sa.Apply([]Resource{cluster, ug}, nil)
	assert.NoError(t, err)
	assert.NotNil(t, rg)
}

func (suite *SMBAdminSuite) TestRemoveCluster() {
	t := suite.T()
	sa := NewFromConn(suite.vconn.Get(t))
	r := []Resource{
		NewActiveDirectoryCluster("cc1", "zoid"),
	}
	rg, err := sa.Apply(r, nil)
	assert.NoError(t, err)
	assert.True(t, rg.Ok())

	err = sa.RemoveCluster("cc1")
	assert.NoError(t, err)
}

func (suite *SMBAdminSuite) TestRemoveShare() {
	t := suite.T()
	sa := NewFromConn(suite.vconn.Get(t))
	err := sa.RemoveShare("foo", "bar")
	assert.NoError(t, err)

	r := []Resource{
		NewActiveDirectoryCluster("cc1", "zoid"),
		NewShare("cc1", "zap").SetCephFS("cephfs", "smb", "v1", "/"),
	}
	rg, err := sa.Apply(r, nil)
	assert.NoError(t, err)
	assert.True(t, rg.Ok())

	err = sa.RemoveShare("cc1", "zap")
	assert.NoError(t, err)
}

func (suite *SMBAdminSuite) TestRemoveJoinAuth() {
	t := suite.T()
	sa := NewFromConn(suite.vconn.Get(t))
	err := sa.RemoveJoinAuth("foo")
	assert.NoError(t, err)

	r := []Resource{
		NewJoinAuth("jauth1").SetAuth("admin", "foobar"),
	}
	rg, err := sa.Apply(r, nil)
	assert.NoError(t, err)
	assert.True(t, rg.Ok())

	err = sa.RemoveJoinAuth("jauth1")
	assert.NoError(t, err)
}

func (suite *SMBAdminSuite) TestRemoveUsersAndGroups() {
	t := suite.T()
	sa := NewFromConn(suite.vconn.Get(t))
	err := sa.RemoveUsersAndGroups("foo")
	assert.NoError(t, err)

	r := []Resource{
		NewUsersAndGroups("ug1").SetValues(
			[]UserInfo{{"foo", "s3cr3t"}},
			[]GroupInfo{{"bloop"}},
		),
	}
	rg, err := sa.Apply(r, nil)
	assert.NoError(t, err)
	assert.True(t, rg.Ok())

	err = sa.RemoveUsersAndGroups("ug1")
	assert.NoError(t, err)
}

func (suite *SMBAdminSuite) TestShowSelect() {
	t := suite.T()
	sa := NewFromConn(suite.vconn.Get(t))
	cluster := NewUserCluster("cc1")
	ug := NewLinkedUsersAndGroups(cluster).SetValues(
		[]UserInfo{
			{"alice", "W0nder14nd"},
			{"billy", "p14n0m4N"},
		},
		[]GroupInfo{{"clients"}},
	)
	share := NewShare("cc1", "ss1").SetCephFS("cephfs", "smb", "v1", "/")
	rgroup, err := sa.Apply([]Resource{cluster, ug, share}, nil)
	assert.NoError(t, err)
	assert.True(t, rgroup.Ok())

	t.Run("byIdentity", func(t *testing.T) {
		matching, err := sa.Show([]ResourceRef{cluster.Identity()}, nil)
		assert.NoError(t, err)
		assert.Len(t, matching, 1)
		assert.Equal(t, matching[0].Type(), ClusterType)
	})

	t.Run("byResouceID", func(t *testing.T) {
		ref := ResourceID{ResourceType: ClusterType, ID: "cc1"}
		matching, err := sa.Show([]ResourceRef{ref}, nil)
		assert.NoError(t, err)
		assert.Len(t, matching, 1)
		assert.Equal(t, matching[0].Type(), ClusterType)
	})

	t.Run("byType", func(t *testing.T) {
		matching, err := sa.Show([]ResourceRef{ClusterType}, nil)
		assert.NoError(t, err)
		assert.Len(t, matching, 1)
		assert.Equal(t, matching[0].Type(), ClusterType)
	})
}

func (suite *SMBAdminSuite) TestApplyErrorInvalid() {
	t := suite.T()
	sa := NewFromConn(suite.vconn.Get(t))
	badShare := NewShare("abc", "xyz")
	results, err := sa.Apply([]Resource{badShare}, nil)
	assert.Error(t, err)
	assert.False(t, results.Success)
}

func (suite *SMBAdminSuite) TestApplyFilterBase64() {
	t := suite.T()
	sa := NewFromConn(suite.vconn.Get(t))
	ja1 := NewJoinAuth("jjb64").SetAuth("joiner", "bXlzZWNyZXRQYXNzd29yZA==")
	results, err := sa.Apply(
		[]Resource{ja1},
		&ApplyOptions{PasswordFilter: PasswordFilterBase64},
	)
	assert.NoError(t, err)
	assert.True(t, results.Ok())
	if assert.Len(t, results.Results, 1) {
		assert.Equal(
			t,
			results.Results[0].Resource().(*JoinAuth).Auth.Password,
			"bXlzZWNyZXRQYXNzd29yZA==",
		)
	}
}

func (suite *SMBAdminSuite) TestApplyFilterMixed() {
	t := suite.T()
	sa := NewFromConn(suite.vconn.Get(t))
	ja1 := NewJoinAuth("jjb64").SetAuth("joiner", "bXlzZWNyZXRQYXNzd29yZA==")
	results, err := sa.Apply(
		[]Resource{ja1},
		&ApplyOptions{
			PasswordFilter:    PasswordFilterBase64,
			PasswordFilterOut: PasswordFilterNone,
		},
	)
	assert.NoError(t, err)
	assert.True(t, results.Ok())
	if assert.Len(t, results.Results, 1) {
		assert.Equal(
			t,
			results.Results[0].Resource().(*JoinAuth).Auth.Password,
			"mysecretPassword",
		)
	}
}

func (suite *SMBAdminSuite) TestApplyFilterShowFilter() {
	t := suite.T()
	sa := NewFromConn(suite.vconn.Get(t))
	ja1 := NewJoinAuth("jjb64").SetAuth("joiner", "bXlzZWNyZXRQYXNzd29yZA==")
	results, err := sa.Apply(
		[]Resource{ja1},
		&ApplyOptions{
			PasswordFilter:    PasswordFilterBase64,
			PasswordFilterOut: PasswordFilterHidden,
		},
	)
	assert.NoError(t, err)
	assert.True(t, results.Ok())

	t.Run("showHidden", func(t *testing.T) {
		res, err := sa.Show(
			[]ResourceRef{ja1.Identity()},
			&ShowOptions{PasswordFilter: PasswordFilterHidden},
		)
		assert.NoError(t, err)
		if assert.Len(t, res, 1) {
			assert.Equal(t, res[0].(*JoinAuth).Auth.Password, "****************")
		}
	})
	t.Run("showBase64", func(t *testing.T) {
		res, err := sa.Show(
			[]ResourceRef{ja1.Identity()},
			&ShowOptions{PasswordFilter: PasswordFilterBase64},
		)
		assert.NoError(t, err)
		if assert.Len(t, res, 1) {
			assert.Equal(t, res[0].(*JoinAuth).Auth.Password, "bXlzZWNyZXRQYXNzd29yZA==")
		}
	})
	t.Run("showFilterNone", func(t *testing.T) {
		res, err := sa.Show(
			[]ResourceRef{ja1.Identity()},
			&ShowOptions{PasswordFilter: PasswordFilterNone},
		)
		assert.NoError(t, err)
		if assert.Len(t, res, 1) {
			assert.Equal(t, res[0].(*JoinAuth).Auth.Password, "mysecretPassword")
		}
	})
}
