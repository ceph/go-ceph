//go:build ceph_preview

package admin

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func (suite *RadosGWTestSuite) TestTopic() {
	suite.SetupConnection()
	co, err := New(suite.endpoint, suite.accessKey, suite.secretKey, newDebugHTTPClient(http.DefaultClient))
	assert.NoError(suite.T(), err)

	var topicARN string

	suite.T().Run("create topic", func(_ *testing.T) {
		var err error
		topicARN, err = co.CreateTopic(context.Background(), "test-topic", map[string]string{
			"push-endpoint": "http://localhost:8080",
			"persistent":    "true",
		})
		assert.NoError(suite.T(), err)
		assert.NotEmpty(suite.T(), topicARN)
	})

	suite.T().Run("get topic attributes", func(_ *testing.T) {
		assert.NotEmpty(suite.T(), topicARN)
		topic, err := co.GetTopicAttributes(context.Background(), topicARN)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), topicARN, topic.TopicArn)
		assert.Equal(suite.T(), "test-topic", topic.Name)
		assert.Equal(suite.T(), "http://localhost:8080", topic.EndPoint.EndpointAddress)
		assert.True(suite.T(), bool(topic.EndPoint.Persistent))
	})

	suite.T().Run("list topics", func(_ *testing.T) {
		topics, err := co.ListTopics(context.Background())
		assert.NoError(suite.T(), err)
		assert.GreaterOrEqual(suite.T(), len(topics), 1)
	})

	suite.T().Run("delete topic", func(_ *testing.T) {
		assert.NotEmpty(suite.T(), topicARN)
		err := co.DeleteTopic(context.Background(), topicARN)
		assert.NoError(suite.T(), err)
	})
}
