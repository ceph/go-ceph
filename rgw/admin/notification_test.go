//go:build ceph_preview

package admin

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func (suite *RadosGWTestSuite) TestNotification() {
	suite.SetupConnection()
	co, err := New(suite.endpoint, suite.accessKey, suite.secretKey, newDebugHTTPClient(http.DefaultClient))
	assert.NoError(suite.T(), err)

	s3Agent, err := newS3Agent(suite.accessKey, suite.secretKey, suite.endpoint, true)
	assert.NoError(suite.T(), err)

	err = s3Agent.createBucket(suite.bucketTestName)
	assert.NoError(suite.T(), err)
	defer func() {
		assert.NoError(suite.T(), co.RemoveBucket(context.Background(), Bucket{Bucket: suite.bucketTestName}))
	}()

	topicName := "test-notification-topic"
	attrs := map[string]string{
		"push-endpoint": "http://localhost:8080",
	}
	topicARN, err := co.CreateTopic(context.Background(), topicName, attrs)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), topicARN)
	defer func() {
		assert.NoError(suite.T(), co.DeleteTopic(context.Background(), topicARN))
	}()

	suite.T().Run("create notification", func(_ *testing.T) {
		config := NotificationConfiguration{
			TopicConfigurations: []TopicConfiguration{
				{
					Id:    "test-notification",
					Topic: topicARN,
					Event: []string{"s3:ObjectCreated:*"},
				},
			},
		}

		err := co.CreateNotification(context.Background(), Bucket{Bucket: suite.bucketTestName}, config)
		assert.NoError(suite.T(), err)
	})

	suite.T().Run("get all notifications", func(_ *testing.T) {
		config, err := co.GetNotifications(context.Background(), Bucket{Bucket: suite.bucketTestName}, "")
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), 1, len(config.TopicConfigurations))
		assert.Equal(suite.T(), "test-notification", config.TopicConfigurations[0].Id)
	})

	suite.T().Run("get specific notification", func(_ *testing.T) {
		config, err := co.GetNotifications(context.Background(), Bucket{Bucket: suite.bucketTestName}, "test-notification")
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), 1, len(config.TopicConfigurations))
		assert.Equal(suite.T(), "test-notification", config.TopicConfigurations[0].Id)
	})

	suite.T().Run("create notification with filter", func(_ *testing.T) {
		config := NotificationConfiguration{
			TopicConfigurations: []TopicConfiguration{
				{
					Id:    "filtered-notification",
					Topic: topicARN,
					Event: []string{"s3:ObjectCreated:*"},
					Filter: &Filter{
						S3Key: &FilterContainer{
							FilterRules: []FilterRule{
								{Name: "prefix", Value: "images/"},
							},
						},
					},
				},
			},
		}

		err := co.CreateNotification(context.Background(), Bucket{Bucket: suite.bucketTestName}, config)
		assert.NoError(suite.T(), err)

		config, err = co.GetNotifications(context.Background(), Bucket{Bucket: suite.bucketTestName}, "")
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), 2, len(config.TopicConfigurations))
	})

	suite.T().Run("delete specific notification", func(_ *testing.T) {
		err := co.DeleteNotification(context.Background(), Bucket{Bucket: suite.bucketTestName}, "test-notification")
		assert.NoError(suite.T(), err)

		config, err := co.GetNotifications(context.Background(), Bucket{Bucket: suite.bucketTestName}, "")
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), 1, len(config.TopicConfigurations))
		assert.Equal(suite.T(), "filtered-notification", config.TopicConfigurations[0].Id)
	})

	suite.T().Run("delete all notifications", func(_ *testing.T) {
		err := co.DeleteNotification(context.Background(), Bucket{Bucket: suite.bucketTestName}, "")
		assert.NoError(suite.T(), err)

		config, err := co.GetNotifications(context.Background(), Bucket{Bucket: suite.bucketTestName}, "")
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), 0, len(config.TopicConfigurations))
	})
}
