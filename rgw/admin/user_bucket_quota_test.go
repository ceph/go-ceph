//go:build ceph_preview

package admin

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func (suite *RadosGWTestSuite) TestUserBucketQuota() {
	suite.SetupConnection()
	co, err := New(suite.endpoint, suite.accessKey, suite.secretKey, newDebugHTTPClient(http.DefaultClient))
	assert.NoError(suite.T(), err)

	usercaps := "users=read"
	_, err = co.CreateUser(context.Background(), User{ID: "leseb", DisplayName: "This is leseb", Email: "leseb@example.com", UserCaps: usercaps})
	assert.NoError(suite.T(), err)
	defer func() {
		err = co.RemoveUser(context.Background(), User{ID: "leseb"})
		assert.NoError(suite.T(), err)
	}()

	suite.T().Run("set bucket quota without uid", func(t *testing.T) {
		err := co.SetBucketQuota(context.Background(), QuotaSpec{})
		assert.Error(suite.T(), err)
		assert.EqualError(suite.T(), err, errMissingUserID.Error())
	})

	suite.T().Run("set bucket quota", func(t *testing.T) {
		maxObjects := int64(101)
		err := co.SetBucketQuota(context.Background(), QuotaSpec{UID: "leseb", MaxObjects: &maxObjects})
		assert.NoError(suite.T(), err)
	})

	suite.T().Run("get bucket quota without uid", func(t *testing.T) {
		_, err := co.GetBucketQuota(context.Background(), QuotaSpec{})
		assert.Error(suite.T(), err)
		assert.EqualError(suite.T(), err, errMissingUserID.Error())
	})

	suite.T().Run("get bucket quota", func(t *testing.T) {
		q, err := co.GetBucketQuota(context.Background(), QuotaSpec{UID: "leseb"})
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), int64(101), *q.MaxObjects)
		assert.Equal(suite.T(), false, *q.Enabled)
	})
}
