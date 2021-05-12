package admin

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func (suite *RadosGWTestSuite) TestBucket() {
	suite.SetupConnection()
	co, err := New(suite.endpoint, suite.accessKey, suite.secretKey, nil)
	co.Debug = true
	assert.NoError(suite.T(), err)

	suite.T().Run("list buckets", func(t *testing.T) {
		buckets, err := co.ListBuckets(context.Background())
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), 0, len(buckets))
	})

	suite.T().Run("info non-existing bucket", func(t *testing.T) {
		_, err := co.GetBucketInfo(context.Background(), Bucket{Bucket: "foo"})
		assert.Error(suite.T(), err)
		assert.True(suite.T(), errors.Is(err, ErrNoSuchBucket), err)
	})

	suite.T().Run("remove non-existing bucket", func(t *testing.T) {
		err := co.RemoveBucket(context.Background(), Bucket{Bucket: "foo"})
		assert.Error(suite.T(), err)
		// TODO: report to rgw team, this should return NoSuchBucket?
		assert.True(suite.T(), errors.Is(err, ErrNoSuchKey))
	})
}
