package admin

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/ceph/go-ceph/internal/util"
	"github.com/stretchr/testify/assert"
)

func (suite *RadosGWTestSuite) TestBucket() {
	suite.SetupConnection()
	co, err := New(suite.endpoint, suite.accessKey, suite.secretKey, newDebugHTTPClient(http.DefaultClient))
	assert.NoError(suite.T(), err)

	s3Agent, err := newS3Agent(suite.accessKey, suite.secretKey, suite.endpoint, true)
	assert.NoError(suite.T(), err)

	beforeCreate := time.Now()
	err = s3Agent.createBucket(suite.bucketTestName)
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
		bucketInfo, err := co.GetBucketInfo(context.Background(), Bucket{Bucket: suite.bucketTestName})
		assert.NoError(suite.T(), err)

		// check if versioning is disabled
		switch {
		case util.CurrentCephVersion() < util.CephQuincy:
			// No action needed for versions below CephQuincy
		case util.CurrentCephVersion() == util.CephReef:
			assert.False(suite.T(), *bucketInfo.VersioningEnabled)
			assert.False(suite.T(), *bucketInfo.Versioned)
		default:
			assert.Equal(suite.T(), "off", *bucketInfo.Versioning)
		}

		// check if object lock is disabled
		if util.CurrentCephVersion() >= util.CephQuincy {
			assert.False(suite.T(), bucketInfo.ObjectLockEnabled)
		}
	})

	suite.T().Run("enable versioning", func(t *testing.T) {
		if util.CurrentCephVersion() < util.CephQuincy {
			t.Skip("versioning is not reported in bucket stats")
		}

		_, err := s3Agent.Client.PutBucketVersioning(context.Background(), &s3.PutBucketVersioningInput{
			Bucket:                  &suite.bucketTestName,
			VersioningConfiguration: &types.VersioningConfiguration{Status: types.BucketVersioningStatusEnabled},
		})
		assert.NoError(suite.T(), err)

		// check if versioning is enabled
		bucketInfo, err := co.GetBucketInfo(context.Background(), Bucket{Bucket: suite.bucketTestName})
		assert.NoError(suite.T(), err)
		if util.CurrentCephVersion() == util.CephReef {
			assert.True(suite.T(), *bucketInfo.VersioningEnabled)
			assert.True(suite.T(), *bucketInfo.Versioned)
		} else {
			assert.Equal(suite.T(), "enabled", *bucketInfo.Versioning)
		}
	})

	suite.T().Run("enable bucket object lock", func(t *testing.T) {
		if util.CurrentCephVersion() < util.CephQuincy {
			t.Skip("bucket object lock is not reported in bucket stats")
		}

		const bucketName = "bucket-object-lock"

		// create bucket with object lock enabled
		_, err := s3Agent.Client.CreateBucket(context.Background(), &s3.CreateBucketInput{
			Bucket:                     aws.String(bucketName),
			ObjectLockEnabledForBucket: aws.Bool(true),
		})
		assert.NoError(suite.T(), err)

		// check if object lock is enabled
		bucketInfo, err := co.GetBucketInfo(context.Background(), Bucket{Bucket: bucketName})
		assert.NoError(suite.T(), err)
		assert.True(suite.T(), bucketInfo.ObjectLockEnabled)

		// remove bucket
		err = co.RemoveBucket(context.Background(), Bucket{Bucket: bucketName})
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
		if util.CurrentCephVersion() >= util.CephPacific && util.CurrentCephVersion() <= util.CephSquid {
			assert.True(suite.T(), errors.Is(err, ErrNoSuchBucket))
		} else {
			assert.True(suite.T(), errors.Is(err, ErrNoSuchKey))
		}
	})
}
