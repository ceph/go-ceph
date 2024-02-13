package admin

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func (suite *RadosGWTestSuite) TestUserBucket() {
	suite.SetupConnection()
	co, err := New(suite.endpoint, suite.accessKey, suite.secretKey, newDebugHTTPClient(http.DefaultClient))
	assert.NoError(suite.T(), err)

	s3, err := newS3Agent(suite.accessKey, suite.secretKey, suite.endpoint, true)
	assert.NoError(suite.T(), err)

	err = s3.createBucket(suite.bucketTestName)
	assert.NoError(suite.T(), err)

	suite.T().Run("list empty user's buckets", func(_ *testing.T) {
		_, err := co.ListUsersBuckets(context.Background(), "")
		assert.Error(suite.T(), err)
		assert.True(suite.T(), errors.Is(err, errMissingUserID), err)
	})

	suite.T().Run("list user's buckets", func(_ *testing.T) {
		buckets, err := co.ListUsersBuckets(context.Background(), "admin")
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), 1, len(buckets))
	})

	suite.T().Run("list unknown user's buckets", func(_ *testing.T) {
		buckets, err := co.ListUsersBuckets(context.Background(), "foo")
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), 0, len(buckets))
	})

	suite.T().Run("list empty user's buckets with stat", func(_ *testing.T) {
		_, err := co.ListUsersBucketsWithStat(context.Background(), "")
		assert.Error(suite.T(), err)
		assert.True(suite.T(), errors.Is(err, errMissingUserID), err)
	})

	suite.T().Run("list user's buckets with stat", func(_ *testing.T) {
		buckets, err := co.ListUsersBucketsWithStat(context.Background(), "admin")
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), 1, len(buckets))
		b := buckets[0]
		assert.NotNil(suite.T(), b)
		assert.Equal(suite.T(), suite.bucketTestName, b.Bucket)
		assert.Equal(suite.T(), "admin", b.Owner)
		assert.NotNil(suite.T(), b.BucketQuota.MaxSize)
	})

	suite.T().Run("list unknown user's buckets with stat", func(_ *testing.T) {
		buckets, err := co.ListUsersBucketsWithStat(context.Background(), "foo")
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), 0, len(buckets))
	})
}
