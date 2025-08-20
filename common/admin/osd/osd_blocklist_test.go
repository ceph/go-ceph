//go:build !octopus && ceph_preview

package osd

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func (suite *OSDAdminSuite) TestOSDBlocklist() {
	osda := NewFromConn(suite.vconn.Get(suite.T()))

	res, err := osda.OSDBlocklist()
	assert.NoError(suite.T(), err)
	prev := len(*res)

	suite.T().Run("osd blocklist add address", func(t *testing.T) {
		// empty address
		err := osda.OSDBlocklistAdd(AddressEntry{})
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrEmptyArgument))

		// add invalid ip address
		err = osda.OSDBlocklistAdd(AddressEntry{
			Addr: "192.168.122.257",
		})
		assert.Error(t, err)
		assert.True(t,
			errors.Unwrap(err) == errors.Unwrap(ErrInvalidArgument))

		err = osda.OSDBlocklistAdd(AddressEntry{
			Addr: "192.168.122.2",
		})
		assert.NoError(t, err)

		// add ip address with invalid expire value
		err = osda.OSDBlocklistAdd(AddressEntry{
			Addr:   "192.168.122.3",
			Expire: -1,
		})
		assert.Error(t, err)
		assert.Equal(t, err, ErrInvalidArgument)

		err = osda.OSDBlocklistAdd(AddressEntry{
			Addr:   "192.168.122.3",
			Expire: 22.3,
		})
		assert.NoError(t, err)

		// add invalid network
		err = osda.OSDBlocklistAdd(AddressEntry{
			Addr: "192.168.122.0/40",
		})
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "osd: ret=-22, Invalid argument")

		err = osda.OSDBlocklistAdd(AddressEntry{
			Addr: "192.168.122.0/24",
		})
		assert.NoError(t, err)
	})

	suite.T().Run("osd blocklist remove address", func(t *testing.T) {
		// empty address
		err := osda.OSDBlocklistRemove(AddressEntry{})
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrEmptyArgument))

		err = osda.OSDBlocklistRemove(AddressEntry{
			Addr: "192.168.122.2",
		})
		assert.NoError(t, err)

		err = osda.OSDBlocklistRemove(AddressEntry{
			Addr: "192.168.122.3",
		})
		assert.NoError(t, err)

		// remove non existent ip address
		err = osda.OSDBlocklistRemove(AddressEntry{
			Addr: "192.168.122.4",
		})
		assert.NoError(t, err)

		res, err := osda.OSDBlocklist()
		assert.NoError(t, err)
		assert.Equal(t, prev+1, len(*res))

		err = osda.OSDBlocklistRemove(AddressEntry{
			Addr: "192.168.122.0/24",
		})
		assert.NoError(t, err)

		// remove non existent network
		err = osda.OSDBlocklistRemove(AddressEntry{
			Addr: "192.168.122.0/32",
		})
		assert.NoError(t, err)

		res, err = osda.OSDBlocklist()
		assert.NoError(t, err)
		assert.Equal(t, prev, len(*res))
	})
}
