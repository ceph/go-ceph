//go:build !(octopus || pacific || quincy || reef || squid) && ceph_preview

package nvmegw

import (
	"testing"

	tsuite "github.com/stretchr/testify/suite"

	"github.com/ceph/go-ceph/internal/admintest"
)

var (
	radosConnector = admintest.NewConnector()
)

func TestNVMeGWAdmin(t *testing.T) {
	tsuite.Run(t, new(NVMeGWAdminSuite))
}

// NVMeGWAdminSuite is a suite of tests for the nfs admin package.
// A suite is used because creating/managing NFS has certain expectations for
// the cluster that we may need to mock, especially when running in the
// standard go-ceph test container. Using a suite allows us to have suite
// setup/validate the environment and the tests can largely be ignorant of the
// environment.
type NVMeGWAdminSuite struct {
	tsuite.Suite

	poolname string
}

func (suite *NVMeGWAdminSuite) SetupSuite() {
	require := suite.Require()

	suite.poolname = "nvme-create-delete"

	conn := radosConnector.GetConn(suite.T())
	err := conn.MakePool(suite.poolname)
	require.NoError(err)
}

func (suite *NVMeGWAdminSuite) TearDownSuite() {
	conn := radosConnector.GetConn(suite.T())
	conn.DeletePool(suite.poolname)
}

func (suite *NVMeGWAdminSuite) TestCreateDeleteGateway() {
	require := suite.Require()

	ra := radosConnector.Get(suite.T())
	nvmea := NewFromConn(ra)

	gw := "ceph-vm"
	anaGroup := "nqn.2025-10.io.ceph:create-delete"

	err := nvmea.CreateGateway(gw, suite.poolname, anaGroup)
	require.NoError(err)

	err = nvmea.DeleteGateway(gw, suite.poolname, anaGroup)
	require.NoError(err)
}

func (suite *NVMeGWAdminSuite) TestShowGateways() {
	require := suite.Require()

	ra := radosConnector.Get(suite.T())
	nvmea := NewFromConn(ra)

	gw1 := "ceph-vm-01"
	gw2 := "ceph-vm-02"
	anaGroup := "nqn.2025-10.io.ceph:show-gateways"

	err := nvmea.CreateGateway(gw1, suite.poolname, anaGroup)
	require.NoError(err)

	defer func() {
		err = nvmea.DeleteGateway(gw1, suite.poolname, anaGroup)
		require.NoError(err)
	}()

	err = nvmea.CreateGateway(gw2, suite.poolname, anaGroup)
	require.NoError(err)

	defer func() {
		err = nvmea.DeleteGateway(gw2, suite.poolname, anaGroup)
		require.NoError(err)
	}()

	l, err := nvmea.ShowGateways(suite.poolname, anaGroup)
	require.NoError(err)
	require.Len(l.Gateways, 2)
}
