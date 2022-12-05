package admin

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testAk = "HDNEZQXZAA6NIWOBOL0U"
)

func (suite *RadosGWTestSuite) TestKeys() {
	suite.SetupConnection()
	co, err := New(suite.endpoint, suite.accessKey, suite.secretKey, newDebugHTTPClient(http.DefaultClient))
	assert.NoError(suite.T(), err)

	var keys *[]UserKeySpec

	suite.T().Run("create keys but user ID and SubUser is empty", func(t *testing.T) {
		_, err := co.CreateKey(context.Background(), UserKeySpec{})
		assert.Error(suite.T(), err)
		assert.EqualError(suite.T(), err, errMissingUserID.Error())
	})

	suite.T().Run("create swift keys but SubUser is empty", func(t *testing.T) {
		_, err := co.CreateKey(context.Background(), UserKeySpec{KeyType: "swift"})
		assert.Error(suite.T(), err)
		assert.EqualError(suite.T(), err, errMissingSubuserID.Error())
	})

	suite.T().Run("create some unknown key type", func(t *testing.T) {
		_, err := co.CreateKey(context.Background(), UserKeySpec{KeyType: "some-key-type"})
		assert.Error(suite.T(), err)
		assert.EqualError(suite.T(), err, errUnsupportedKeyType.Error())
	})

	suite.T().Run("create keys without ak or sk provided", func(t *testing.T) {
		keys, err = co.CreateKey(context.Background(), UserKeySpec{UID: "admin"})
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), 2, len(*keys))
	})

	suite.T().Run("create keys with ak provided", func(t *testing.T) {
		keys, err = co.CreateKey(context.Background(), UserKeySpec{UID: "admin", AccessKey: testAk})
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), 3, len(*keys))
	})

	suite.T().Run("remove keys but user ID and SubUser is empty", func(t *testing.T) {
		err := co.RemoveKey(context.Background(), UserKeySpec{})
		assert.Error(suite.T(), err)
		assert.EqualError(suite.T(), err, errMissingUserID.Error())
	})

	suite.T().Run("remove s3 keys but ak is empty", func(t *testing.T) {
		err := co.RemoveKey(context.Background(), UserKeySpec{UID: "admin"})
		assert.Error(suite.T(), err)
		assert.EqualError(suite.T(), err, errMissingUserAccessKey.Error())
	})

	suite.T().Run("remove s3 key", func(t *testing.T) {
		for _, key := range *keys {
			if key.AccessKey != suite.accessKey {
				err := co.RemoveKey(context.Background(), UserKeySpec{UID: "admin", AccessKey: key.AccessKey})
				assert.NoError(suite.T(), err)
			}
		}
	})
}
