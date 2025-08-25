//go:build ceph_preview

package osd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	tsuite "github.com/stretchr/testify/suite"

	"github.com/ceph/go-ceph/internal/admintest"
	"github.com/ceph/go-ceph/internal/commands"
)

func TestOSDAdmin(t *testing.T) {
	tsuite.Run(t, new(OSDAdminSuite))
}

// OSDAdminSuite is a suite of tests for the osd admin package.
type OSDAdminSuite struct {
	tsuite.Suite

	vconn *admintest.Connector
}

func (suite *OSDAdminSuite) SetupSuite() {
	suite.vconn = admintest.NewConnector()
}

func (suite *OSDAdminSuite) TestOSDList() {
	cmd := map[string]string{"prefix": "osd ls", "format": "json"}

	buf := commands.MarshalMonCommand(suite.vconn.Get(suite.T()), cmd)
	assert.NoError(suite.T(), buf.End())

	resp := make([]int, 0)
	assert.NoError(suite.T(), buf.Unmarshal(&resp).End())
	assert.Equal(suite.T(), 1, len(resp))
	assert.EqualValues(suite.T(), 0, resp[0])
}
