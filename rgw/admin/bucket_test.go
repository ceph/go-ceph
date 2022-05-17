package admin

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/ceph/go-ceph/internal/util"
	"github.com/stretchr/testify/assert"
)

func (suite *RadosGWTestSuite) TestBucket() {
	suite.SetupConnection()
	co, err := New(suite.endpoint, suite.accessKey, suite.secretKey, newDebugHTTPClient(http.DefaultClient))
	assert.NoError(suite.T(), err)

	s3, err := newS3Agent(suite.accessKey, suite.secretKey, suite.endpoint, true)
	assert.NoError(suite.T(), err)

	err = s3.createBucket(suite.bucketTestName)
	assert.NoError(suite.T(), err)

	suite.T().Run("list buckets", func(t *testing.T) {
		buckets, err := co.ListBuckets(context.Background())
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), 1, len(buckets))
	})

	suite.T().Run("info non-existing bucket", func(t *testing.T) {
		_, err := co.GetBucketInfo(context.Background(), Bucket{Bucket: "foo"})
		assert.Error(suite.T(), err)
		assert.True(suite.T(), errors.Is(err, ErrNoSuchBucket), err)
	})

	suite.T().Run("info existing bucket", func(t *testing.T) {
		_, err := co.GetBucketInfo(context.Background(), Bucket{Bucket: suite.bucketTestName})
		assert.NoError(suite.T(), err)
	})

	suite.T().Run("remove bucket", func(t *testing.T) {
		err := co.RemoveBucket(context.Background(), Bucket{Bucket: suite.bucketTestName})
		assert.NoError(suite.T(), err)
	})

	suite.T().Run("list bucket is now zero", func(t *testing.T) {
		buckets, err := co.ListBuckets(context.Background())
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), 0, len(buckets))
	})

	suite.T().Run("remove non-existing bucket", func(t *testing.T) {
		err := co.RemoveBucket(context.Background(), Bucket{Bucket: "foo"})
		assert.Error(suite.T(), err)
		if util.CurrentCephVersion() <= util.CephOctopus {
			assert.True(suite.T(), errors.Is(err, ErrNoSuchKey))
		} else {
			assert.True(suite.T(), errors.Is(err, ErrNoSuchBucket))
		}
	})
}
