package admin

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func (suite *RadosGWTestSuite) TestCaps() {
	suite.SetupConnection()
	co, err := New(suite.endpoint, suite.accessKey, suite.secretKey, newDebugHTTPClient(http.DefaultClient))

	assert.NoError(suite.T(), err)
	suite.T().Run("create test user", func(t *testing.T) {
		user, err := co.CreateUser(context.Background(), User{ID: "test", DisplayName: "test-user", Email: "test@example.com"})
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), "test", user.ID)
		assert.Zero(suite.T(), len(user.Caps))
	})

	suite.T().Run("add caps to the user but user ID is empty", func(t *testing.T) {
		_, err := co.AddUserCap(context.Background(), "", "users=read")
		assert.Error(suite.T(), err)
		assert.EqualError(suite.T(), err, errMissingUserID.Error())
	})

	suite.T().Run("add caps to the user but no cap is specified", func(t *testing.T) {
		_, err := co.AddUserCap(context.Background(), "test", "")
		assert.Error(suite.T(), err)
		assert.EqualError(suite.T(), err, errMissingUserCap.Error())

	})

	suite.T().Run("add caps to the user, returns success", func(t *testing.T) {
		usercap, err := co.AddUserCap(context.Background(), "test", "users=read")
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), "users", usercap[0].Type)
		assert.Equal(suite.T(), "read", usercap[0].Perm)

	})

	suite.T().Run("remove caps from the user but user ID is empty", func(t *testing.T) {
		_, err := co.RemoveUserCap(context.Background(), "", "users=read")
		assert.Error(suite.T(), err)
		assert.EqualError(suite.T(), err, errMissingUserID.Error())
	})

	suite.T().Run("remove caps from the user but no cap is specified", func(t *testing.T) {
		_, err := co.RemoveUserCap(context.Background(), "test", "")
		assert.Error(suite.T(), err)
		assert.EqualError(suite.T(), err, errMissingUserCap.Error())

	})

	suite.T().Run("remove caps from the user returns success", func(t *testing.T) {
		usercap, err := co.RemoveUserCap(context.Background(), "test", "users=read")
		assert.NoError(suite.T(), err)
		assert.Zero(suite.T(), len(usercap))
	})

	suite.T().Run("delete test user", func(t *testing.T) {
		err := co.RemoveUser(context.Background(), User{ID: "test"})
		assert.NoError(suite.T(), err)
	})
}
