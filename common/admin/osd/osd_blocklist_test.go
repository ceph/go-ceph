//go:build ceph_preview

package osd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func (suite *OSDAdminSuite) TestOSDBlocklist() {
	osda := NewFromConn(suite.vconn.Get(suite.T()))

	res, err := osda.OSDBlocklist()
	assert.NoError(suite.T(), err)
	prev := len(res)

	suite.T().Run("osd blocklist add address", func(_ *testing.T) {
		// add invalid ip address
		_, err := osda.OSDBlocklistAdd(AddressEntry{
			addr: "192.168.122.257"},
		)
		assert.Error(suite.T(), err)

		res, err := osda.OSDBlocklistAdd(AddressEntry{
			addr: "192.168.122.2"},
		)
		assert.NoError(suite.T(), err)
		assert.True(suite.T(),
			strings.Contains(res, "blocklisting 192.168.122.2"))

		// add invalid network
		_, err = osda.OSDBlocklistAdd(AddressEntry{
			addr: "192.168.122.0/40"},
		)
		assert.Error(suite.T(), err)

		res, err = osda.OSDBlocklistAdd(AddressEntry{
			addr: "192.168.122.0/24"},
		)
		assert.NoError(suite.T(), err)
		assert.True(suite.T(),
			strings.Contains(res,
				"blocklisting cidr:192.168.122.0:0/24"))
	})

	suite.T().Run("display osd blocklist", func(_ *testing.T) {
		res, err := osda.OSDBlocklist()
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), prev+2, len(res))
	})

	suite.T().Run("osd blocklist remove address", func(_ *testing.T) {
		res, err := osda.OSDBlocklistRemove(AddressEntry{
			addr: "192.168.122.2"},
		)
		assert.NoError(suite.T(), err)
		assert.True(suite.T(),
			strings.Contains(res, "un-blocklisting 192.168.122.2"))

		ret, err := osda.OSDBlocklist()
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), prev+1, len(ret))

		// remove non existent ip address
		_, err = osda.OSDBlocklistRemove(AddressEntry{
			addr: "192.168.122.3"},
		)
		assert.Error(suite.T(), err)

		res, err = osda.OSDBlocklistRemove(AddressEntry{
			addr: "192.168.122.0/24"},
		)
		assert.NoError(suite.T(), err)
		assert.True(suite.T(),
			strings.Contains(res,
				"un-blocklisting cidr:192.168.122.0:0/24"))

		ret, err = osda.OSDBlocklist()
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), prev, len(ret))

		// remove non existent network
		_, err = osda.OSDBlocklistRemove(AddressEntry{
			addr: "192.168.122.0/32"},
		)
		assert.Error(suite.T(), err)
	})
}
