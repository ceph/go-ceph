//go:build !(pacific || quincy || reef) && ceph_preview

package admin

import (
	"context"
	"net/http"
	"testing"

	"github.com/ceph/go-ceph/internal/util"
	"github.com/stretchr/testify/assert"
)

func (suite *RadosGWTestSuite) TestAccountQuota() {
	suite.SetupConnection()
	co, err := New(suite.endpoint, suite.accessKey, suite.secretKey, newDebugHTTPClient(http.DefaultClient))
	assert.NoError(suite.T(), err)

	suite.T().Run("fail to set quota since no ID provided", func(t *testing.T) {
		err := co.SetAccountQuota(context.Background(), AccountQuotaSpec{QuotaType: AccountQuotaTypeAccount})
		assert.ErrorIs(t, err, ErrInvalidArgument)
	})

	suite.T().Run("fail to set quota since quota-type is invalid", func(t *testing.T) {
		err := co.SetAccountQuota(context.Background(), AccountQuotaSpec{ID: "RGW98765432109876543", QuotaType: "invalid"})
		assert.ErrorIs(t, err, ErrInvalidArgument)
	})

	suite.T().Run("successfully set and read account quotas", func(t *testing.T) {
		if util.CurrentCephVersion() < util.CephTentacle {
			t.Skipf("account quota admin ops are not yet supported on %s", util.CurrentCephVersionString())
		}

		account, err := co.CreateAccount(context.Background(), Account{Name: "quota-test-account"})
		assert.NoError(t, err)
		id := account.ID
		defer func() {
			assert.NoError(t, co.DeleteAccount(context.Background(), id))
		}()

		enabled := true
		var maxSize int64 = 1073741824
		var maxObjects int64 = 1000
		err = co.SetAccountQuota(context.Background(), AccountQuotaSpec{
			ID:         id,
			QuotaType:  AccountQuotaTypeAccount,
			Enabled:    &enabled,
			MaxSize:    &maxSize,
			MaxObjects: &maxObjects,
		})
		assert.NoError(t, err)

		var bucketMaxSize int64 = 524288000
		var bucketMaxObjects int64 = 500
		err = co.SetAccountQuota(context.Background(), AccountQuotaSpec{
			ID:         id,
			QuotaType:  AccountQuotaTypeBucket,
			Enabled:    &enabled,
			MaxSize:    &bucketMaxSize,
			MaxObjects: &bucketMaxObjects,
		})
		assert.NoError(t, err)

		got, err := co.GetAccount(context.Background(), id)
		assert.NoError(t, err)
		assert.Equal(t, &enabled, got.Quota.Enabled)
		assert.Equal(t, &maxSize, got.Quota.MaxSize)
		assert.Equal(t, &maxObjects, got.Quota.MaxObjects)
		assert.Equal(t, &enabled, got.BucketQuota.Enabled)
		assert.Equal(t, &bucketMaxSize, got.BucketQuota.MaxSize)
		assert.Equal(t, &bucketMaxObjects, got.BucketQuota.MaxObjects)
	})
}
