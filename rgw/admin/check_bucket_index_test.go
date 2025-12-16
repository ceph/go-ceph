//go:build !octopus

package admin

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func (suite *RadosGWTestSuite) TestCheckBucketIndex() {
	suite.SetupConnection()
	co, err := New(suite.endpoint, suite.accessKey, suite.secretKey, newDebugHTTPClient(http.DefaultClient))
	assert.NoError(suite.T(), err)

	s3, err := newS3Agent(suite.accessKey, suite.secretKey, suite.endpoint, true)
	assert.NoError(suite.T(), err)

	err = s3.createBucket(suite.bucketTestName)
	assert.NoError(suite.T(), err)

	suite.T().Run("check bucket index", func(_ *testing.T) {
		_, err := co.CheckBucketIndex(context.Background(), CheckBucketIndexRequest{Bucket: suite.bucketTestName, CheckObjects: true, Fix: true})
		assert.NoError(suite.T(), err)
	})
}
