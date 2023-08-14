//go:build !nautilus
// +build !nautilus

package rbd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMirrorPeerSite(t *testing.T) {
	conn := radosConnect(t)
	poolName := GetUUID()
	err := conn.MakePool(poolName)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, conn.DeletePool(poolName))
		conn.Shutdown()
	}()

	ioctx, err := conn.OpenIOContext(poolName)
	assert.NoError(t, err)
	defer func() {
		ioctx.Destroy()
	}()

	err = SetMirrorMode(ioctx, MirrorModePool)
	assert.NoError(t, err)

	t.Run("addRemovePeerSite", func(t *testing.T) {
		uuid, err := AddMirrorPeerSite(ioctx, "site_a", "client_a", MirrorPeerDirectionRxTx)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, RemoveMirrorPeerSite(ioctx, uuid))
		}()
	})

	t.Run("addPeerSiteInvalid", func(t *testing.T) {
		_, err := AddMirrorPeerSite(ioctx, "", "client_b", MirrorPeerDirectionRx)
		assert.Error(t, err)
	})

	t.Run("listPeerSite", func(t *testing.T) {
		uuid, err := AddMirrorPeerSite(ioctx, "site_b", "client_b", MirrorPeerDirectionRxTx)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, RemoveMirrorPeerSite(ioctx, uuid))
		}()

		site, err := ListMirrorPeerSite(ioctx)
		assert.NoError(t, err)
		assert.Len(t, site, 1)
		assert.Equal(t, site[0].UUID, uuid)
		assert.Equal(t, site[0].SiteName, "site_b")
		assert.Equal(t, site[0].ClientName, "client_b")
		assert.Equal(t, site[0].Direction, MirrorPeerDirectionRxTx)
	})

	t.Run("setGetAttributesPeerSite", func(t *testing.T) {
		uuid, err := AddMirrorPeerSite(ioctx, "site_c", "client_c", MirrorPeerDirectionRxTx)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, RemoveMirrorPeerSite(ioctx, uuid))
		}()

		attributes := map[string]string{
			"mon_host": "test_host",
		}
		err = SetAttributesMirrorPeerSite(ioctx, uuid, attributes)
		assert.NoError(t, err)

		attributesList, err := GetAttributesMirrorPeerSite(ioctx, uuid)
		assert.NoError(t, err)
		assert.Equal(t, attributesList, attributes)
	})

	t.Run("setPeerSite", func(t *testing.T) {
		uuid, err := AddMirrorPeerSite(ioctx, "site_d", "client_d", MirrorPeerDirectionRxTx)
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, RemoveMirrorPeerSite(ioctx, uuid))
		}()

		err = SetMirrorPeerSiteClientName(ioctx, uuid, "client_e")
		assert.NoError(t, err)

		err = SetMirrorPeerSiteName(ioctx, uuid, "site_e")
		assert.NoError(t, err)

		err = SetMirrorPeerSiteDirection(ioctx, uuid, MirrorPeerDirectionRx)
		assert.NoError(t, err)

		site, err := ListMirrorPeerSite(ioctx)
		assert.NoError(t, err)
		assert.Len(t, site, 1)
		assert.Equal(t, site[0].UUID, uuid)
		assert.Equal(t, site[0].SiteName, "site_e")
		assert.Equal(t, site[0].ClientName, "client_e")
		assert.Equal(t, site[0].Direction, MirrorPeerDirectionRx)
	})
}
