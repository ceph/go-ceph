package admin

import (
	"context"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func (suite *RadosGWTestSuite) TestListBucketsWithStat() {
	suite.SetupConnection()
	co, err := New(suite.endpoint, suite.accessKey, suite.secretKey, newDebugHTTPClient(http.DefaultClient))
	assert.NoError(suite.T(), err)

	s3, err := newS3Agent(suite.accessKey, suite.secretKey, suite.endpoint, true)
	assert.NoError(suite.T(), err)

	err = s3.createBucket(suite.bucketTestName)
	assert.NoError(suite.T(), err)

	suite.T().Run("list buckets with stat", func(t *testing.T) {
		buckets, err := co.ListBucketsWithStat(context.Background())
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), 1, len(buckets))
		b := buckets[0]
		assert.NotNil(suite.T(), b)
		assert.Equal(suite.T(), suite.bucketTestName, b.Bucket)
		assert.Equal(suite.T(), "admin", b.Owner)
		assert.NotNil(suite.T(), b.BucketQuota.MaxSize)
	})

}
