//go:build !(pacific || quincy) && ceph_preview

package admin

import (
	"context"
	"net/http"
	"testing"

	"github.com/ceph/go-ceph/internal/util"
	"github.com/stretchr/testify/assert"
)

func (suite *RadosGWTestSuite) TestBucketRateLimit() {
	suite.SetupConnection()
	co, err := New(suite.endpoint, suite.accessKey, suite.secretKey, newDebugHTTPClient(http.DefaultClient))
	assert.NoError(suite.T(), err)

	const (
		userName   = "test-user-bucket-ratelimits"
		bucketName = "test-bucket-ratelimits"
	)

	suite.T().Run("create test user for bucket ratelimits", func(_ *testing.T) {
		user, err := co.CreateUser(context.Background(), User{ID: userName, DisplayName: userName})
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), userName, user.ID)
		assert.Zero(suite.T(), len(user.Caps))
	})

	suite.T().Run("create test bucket for bucket ratelimits", func(t *testing.T) {
		s3, err := newS3Agent(suite.accessKey, suite.secretKey, suite.endpoint, true)
		assert.NoError(t, err)

		err = s3.createBucket(bucketName)
		assert.NoError(t, err)
	})

	suite.T().Run("set bucket rate-limit but no bucket is specified", func(_ *testing.T) {
		err := co.SetIndividualBucketRateLimit(context.Background(), RateLimitSpec{})
		assert.Error(suite.T(), err)
		assert.EqualError(suite.T(), err, errMissingBucket.Error())
	})

	wantRateLimits := RateLimitSpec{
		Bucket:        bucketName,
		Enabled:       &[]bool{true}[0],
		MaxReadOps:    &[]int64{1}[0],
		MaxWriteOps:   &[]int64{1}[0],
		MaxReadBytes:  &[]int64{1}[0],
		MaxWriteBytes: &[]int64{1}[0],
		MaxListOps:    &[]int64{1}[0],
		MaxDeleteOps:  &[]int64{1}[0],
	}

	suite.T().Run("set bucket rate-limit", func(_ *testing.T) {
		err := co.SetIndividualBucketRateLimit(context.Background(), wantRateLimits)
		assert.NoError(suite.T(), err)
	})

	suite.T().Run("get bucket rate-limit", func(_ *testing.T) {
		gotRateLimits, err := co.GetIndividualBucketRateLimit(context.Background(), bucketName)
		assert.NoError(suite.T(), err)

		if util.CurrentCephVersion() <= util.CephTentacle {
			wantRateLimits.MaxListOps = nil
			wantRateLimits.MaxDeleteOps = nil
		}

		assert.EqualValues(suite.T(), wantRateLimits, gotRateLimits)
	})

	suite.T().Run("remove bucket for bucket ratelimits", func(_ *testing.T) {
		err := co.RemoveBucket(context.Background(), Bucket{Bucket: bucketName})
		assert.NoError(suite.T(), err)
	})

	suite.T().Run("delete test user for bucket ratelimits", func(_ *testing.T) {
		err := co.RemoveUser(context.Background(), User{ID: userName})
		assert.NoError(suite.T(), err)
	})
}
