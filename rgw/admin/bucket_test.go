package admin

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
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
		if util.CurrentCephVersion() <= util.CephOctopus {
			assert.True(suite.T(), errors.Is(err, ErrNoSuchKey))
		} else {
			assert.True(suite.T(), errors.Is(err, ErrNoSuchBucket))
		}
	})
}

func TestGetBucketInfoMockAPI(t *testing.T) {
	t.Run("test get bucket info", func(t *testing.T) {
		api, err := New("127.0.0.1", "accessKey", "secretKey", returnMockClientGetBucket())
		assert.NoError(t, err)
		b, err := api.GetBucketInfo(context.TODO(), Bucket{Bucket: "my-versioned-bucket"})
		assert.NoError(t, err)
		assert.Equal(t, "my-versioned-bucket", b.Bucket)

		assert.NotNil(t, b.Usage.RgwMain.NumObjects)
		assert.Equal(t, uint64(14), *b.Usage.RgwMain.NumObjects)
		assert.NotNil(t, b.Usage.RgwMultimeta.NumObjects)
		assert.Equal(t, uint64(53), *b.Usage.RgwMultimeta.NumObjects)
		assert.NotNil(t, b.Usage.RgwNone.NumObjects)
		assert.Equal(t, uint64(27), *b.Usage.RgwNone.NumObjects)
	})
}

var fakeBucketResponse = []byte(`{
    "bucket": "my-versioned-bucket",
    "tenant": "",
    "versioning": "suspended",
    "zonegroup": "e0806fd2-fdd2-43e5-b287-5b9234c7ba98",
    "placement_rule": "default-placement",
    "explicit_placement": {
        "data_pool": "",
        "data_extra_pool": "",
        "index_pool": ""
    },
    "id": "72efcb62-08cd-4591-8485-0baa2775c462.790909945.83884",
    "marker": "72efcb62-08cd-4591-8485-0baa2775c462.790909945.83884",
    "index_type": "Normal",
    "index_generation": 0,
    "num_shards": 11,
    "object_lock_enabled": false,
    "mfa_enabled": false,
    "owner": "test-user",
    "ver": "0#10,1#2,2#2,3#2,4#2,5#2,6#2,7#2,8#2,9#2,10#2",
    "master_ver": "0#0,1#0,2#0,3#0,4#0,5#0,6#0,7#0,8#0,9#0,10#0",
    "mtime": "2026-07-03T12:20:07.394869Z",
    "creation_time": "2026-07-03T12:15:57.194448Z",
    "max_marker": "0#,1#,2#,3#,4#,5#,6#,7#,8#,9#,10#",
    "usage": {
        "rgw.none": {
            "size": 0,
            "size_actual": 0,
            "size_utilized": 0,
            "size_kb": 0,
            "size_kb_actual": 0,
            "size_kb_utilized": 0,
            "num_objects": 27
        },
        "rgw.main": {
            "size": 37656,
            "size_actual": 49152,
            "size_utilized": 37656,
            "size_kb": 37,
            "size_kb_actual": 48,
            "size_kb_utilized": 37,
            "num_objects": 14
        },
        "rgw.multimeta": {
            "size": 0,
            "size_actual": 0,
            "size_utilized": 3445,
            "size_kb": 0,
            "size_kb_actual": 0,
            "size_kb_utilized": 4,
            "num_objects": 53
        }
    },
    "bucket_quota": {
        "enabled": false,
        "check_on_raw": false,
        "max_size": -1,
        "max_size_kb": 0,
        "max_objects": -1
    }
}
`)

func returnMockClientGetBucket() *mockClient {
	r := io.NopCloser(bytes.NewReader(fakeBucketResponse))
	return &mockClient{
		mockDo: func(req *http.Request) (*http.Response, error) {
			if req.Method == http.MethodGet && req.URL.Path == "127.0.0.1/admin/bucket" {
				return &http.Response{
					StatusCode: 200,
					Body:       r,
				}, nil
			}
			return nil, fmt.Errorf("unexpected request: %q. method %q. path %q", req.URL.RawQuery, req.Method, req.URL.Path)
		},
	}
}
