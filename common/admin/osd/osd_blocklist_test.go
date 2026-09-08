//go:build !octopus

package osd

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ceph/go-ceph/internal/commands"
)

func TestParseBlocklist(t *testing.T) {
	const (
		ipUntil      = "2026-09-06T12:34:56.000000+0000"
		rangeUntil   = "2026-09-07T12:34:56.000000+0000"
		ipEntry      = `{"addr":"192.0.2.1:0/0","until":"` + ipUntil + `"}`
		rangeEntry   = `{"range":"198.51.100.0/24","until":"` + rangeUntil + `"}`
		legacyFormat = `[` + ipEntry + `][` + rangeEntry + `]`
		newFormat    = `{"blocklist":[` + ipEntry + `],"range_blocklist":[` + rangeEntry + `]}`
	)

	ipTime, err := time.Parse(layout, ipUntil)
	assert.NoError(t, err)
	rangeTime, err := time.Parse(layout, rangeUntil)
	assert.NoError(t, err)

	want := []Blocklist{
		{Addr: "192.0.2.1:0/0", Until: ipTime},
		{Addr: "198.51.100.0/24", Until: rangeTime},
	}

	got, err := parseBlocklist(commands.NewResponse([]byte(legacyFormat), "", nil))
	assert.NoError(t, err)
	assert.Equal(t, &want, got)

	got, err = parseBlocklist(commands.NewResponse([]byte(newFormat), "", nil))
	assert.NoError(t, err)
	assert.Equal(t, &want, got)
}

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

		t1 := time.Now().UTC().Truncate(time.Microsecond)

		err = osda.OSDBlocklistAdd(AddressEntry{
			Addr:   "192.168.122.3",
			Expire: 22.3,
		})
		assert.NoError(t, err)

		res, err := osda.OSDBlocklist()
		assert.NoError(t, err)

		for _, entry := range *res {
			if strings.Contains(entry.Addr, "192.168.122.3") {
				t2 := entry.Until.UTC()
				assert.InDelta(t,
					22.3, t2.Sub(t1).Seconds(), 0.05)
				break
			}
		}

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
