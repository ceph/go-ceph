package admin

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func (suite *RadosGWTestSuite) TestLink() {
	suite.SetupConnection()
	co, err := New(suite.endpoint, suite.accessKey, suite.secretKey, newDebugHTTPClient(http.DefaultClient))
	assert.NoError(suite.T(), err)

	suite.T().Run("create test user1", func(_ *testing.T) {
		user, err := co.CreateUser(context.Background(), User{ID: "test-user1", DisplayName: "test-user1", Email: "test1@example.com"})
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), "test-user1", user.ID)
		assert.Zero(suite.T(), len(user.Caps))
	})

	suite.T().Run("create test bucket", func(t *testing.T) {
		s3, err := newS3Agent(suite.accessKey, suite.secretKey, suite.endpoint, true)
		assert.NoError(t, err)

		err = s3.createBucket(suite.bucketTestName)
		assert.NoError(t, err)
	})

	suite.T().Run("create test user2", func(_ *testing.T) {
		user, err := co.CreateUser(context.Background(), User{ID: "test-user2", DisplayName: "test-user2", Email: "test2@example.com"})
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), "test-user2", user.ID)
		assert.Zero(suite.T(), len(user.Caps))
	})

	suite.T().Run("link test-user2", func(t *testing.T) {
		bucket, err := co.GetBucketInfo(context.Background(), Bucket{Bucket: suite.bucketTestName})
		assert.NoError(t, err)

		err = co.LinkBucket(context.Background(), BucketLinkInput{
			Bucket:   suite.bucketTestName,
			BucketID: bucket.ID,
			UID:      "test-user2",
		})
		assert.NoError(t, err)

		bucket, err = co.GetBucketInfo(context.Background(), Bucket{Bucket: suite.bucketTestName})
		assert.NoError(t, err)
		assert.Equal(t, bucket.Owner, "test-user2")
	})

	suite.T().Run("unlink test-user2", func(t *testing.T) {
		bucket, err := co.GetBucketInfo(context.Background(), Bucket{Bucket: suite.bucketTestName})
		assert.NoError(t, err)

		err = co.UnlinkBucket(context.Background(), BucketLinkInput{
			Bucket: suite.bucketTestName,
			UID:    bucket.Owner,
		})
		assert.NoError(t, err)
	})

	suite.T().Run("remove bucket", func(_ *testing.T) {
		err := co.RemoveBucket(context.Background(), Bucket{Bucket: suite.bucketTestName})
		assert.NoError(suite.T(), err)
	})

	suite.T().Run("delete test user1", func(_ *testing.T) {
		err := co.RemoveUser(context.Background(), User{ID: "test-user1"})
		assert.NoError(suite.T(), err)
	})

	suite.T().Run("delete test user2", func(_ *testing.T) {
		err := co.RemoveUser(context.Background(), User{ID: "test-user2"})
		assert.NoError(suite.T(), err)
	})

}

func (suite *RadosGWTestSuite) TestLinkRename() {
	suite.SetupConnection()
	co, err := New(suite.endpoint, suite.accessKey, suite.secretKey, newDebugHTTPClient(http.DefaultClient))
	assert.NoError(suite.T(), err)

	const (
		userName          = "test-user-bucket-rename"
		initialBucketName = "initial-name"
		renamedBucketName = "renamed-name"
	)

	suite.T().Run("create test user for bucket rename", func(_ *testing.T) {
		user, err := co.CreateUser(context.Background(), User{ID: userName, DisplayName: userName, Email: "test-user-bucket-rename@example.com"})
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), userName, user.ID)
		assert.Zero(suite.T(), len(user.Caps))
	})

	suite.T().Run("create test bucket for rename", func(t *testing.T) {
		s3, err := newS3Agent(suite.accessKey, suite.secretKey, suite.endpoint, true)
		assert.NoError(t, err)

		err = s3.createBucket(initialBucketName)
		assert.NoError(t, err)
	})

	suite.T().Run("rename bucket", func(t *testing.T) {
		err = co.LinkBucket(context.Background(), BucketLinkInput{
			UID:           userName,
			Bucket:        initialBucketName,
			NewBucketName: renamedBucketName,
		})
		assert.NoError(suite.T(), err)

		_, err = co.GetBucketInfo(context.Background(), Bucket{Bucket: initialBucketName})
		assert.ErrorIs(t, err, ErrNoSuchBucket)

		_, err = co.GetBucketInfo(context.Background(), Bucket{Bucket: renamedBucketName})
		assert.NoError(t, err)
	})

	suite.T().Run("remove bucket", func(_ *testing.T) {
		err := co.RemoveBucket(context.Background(), Bucket{Bucket: renamedBucketName})
		assert.NoError(suite.T(), err)
	})

	suite.T().Run("delete test user for bucket rename", func(_ *testing.T) {
		err := co.RemoveUser(context.Background(), User{ID: userName})
		assert.NoError(suite.T(), err)
	})
}
