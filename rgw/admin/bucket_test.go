package admin

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/ceph/go-ceph/internal/util"
	"github.com/stretchr/testify/assert"
)

func (suite *RadosGWTestSuite) TestBucket() {
	suite.SetupConnection()
	co, err := New(suite.endpoint, suite.accessKey, suite.secretKey, newDebugHTTPClient(http.DefaultClient))
	assert.NoError(suite.T(), err)

	s3, err := newS3Agent(suite.accessKey, suite.secretKey, suite.endpoint, true)
	assert.NoError(suite.T(), err)

	beforeCreate := time.Now()
	err = s3.createBucket(suite.bucketTestName)
	assert.NoError(suite.T(), err)

	suite.T().Run("list buckets", func(_ *testing.T) {
		buckets, err := co.ListBuckets(context.Background())
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), 1, len(buckets))
	})

	suite.T().Run("info non-existing bucket", func(_ *testing.T) {
		_, err := co.GetBucketInfo(context.Background(), Bucket{Bucket: "foo"})
		assert.Error(suite.T(), err)
		assert.True(suite.T(), errors.Is(err, ErrNoSuchBucket), err)
	})

	suite.T().Run("info existing bucket", func(_ *testing.T) {
		_, err := co.GetBucketInfo(context.Background(), Bucket{Bucket: suite.bucketTestName})
		assert.NoError(suite.T(), err)
	})

	suite.T().Run("existing bucket has valid creation date", func(_ *testing.T) {
		b, err := co.GetBucketInfo(context.Background(), Bucket{Bucket: suite.bucketTestName})
		assert.NoError(suite.T(), err)
		assert.NotNil(suite.T(), b.CreationTime)
		assert.WithinDuration(suite.T(), beforeCreate, *b.CreationTime, time.Minute)
	})

	suite.T().Run("get policy non-existing bucket", func(_ *testing.T) {
		_, err := co.GetBucketPolicy(context.Background(), Bucket{Bucket: "foo"})
		assert.Error(suite.T(), err)
		assert.True(suite.T(), errors.Is(err, ErrNoSuchKey), err)
	})

	suite.T().Run("get policy existing bucket", func(_ *testing.T) {
		_, err := co.GetBucketPolicy(context.Background(), Bucket{Bucket: suite.bucketTestName})
		assert.NoError(suite.T(), err)
	})

	suite.T().Run("remove bucket", func(_ *testing.T) {
		err := co.RemoveBucket(context.Background(), Bucket{Bucket: suite.bucketTestName})
		assert.NoError(suite.T(), err)
	})

	suite.T().Run("list bucket is now zero", func(_ *testing.T) {
		buckets, err := co.ListBuckets(context.Background())
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), 0, len(buckets))
	})

	suite.T().Run("remove non-existing bucket", func(_ *testing.T) {
		err := co.RemoveBucket(context.Background(), Bucket{Bucket: "foo"})
		assert.Error(suite.T(), err)
		if util.CurrentCephVersion() <= util.CephOctopus {
			assert.True(suite.T(), errors.Is(err, ErrNoSuchKey))
		} else {
			assert.True(suite.T(), errors.Is(err, ErrNoSuchBucket))
		}
	})
}
