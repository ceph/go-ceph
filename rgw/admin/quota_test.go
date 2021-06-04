package admin

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func (suite *RadosGWTestSuite) TestQuota() {
	suite.SetupConnection()
	co, err := New(suite.endpoint, suite.accessKey, suite.secretKey, nil)
	co.Debug = true
	assert.NoError(suite.T(), err)

	suite.T().Run("set quota to user but user ID is empty", func(t *testing.T) {
		err := co.SetUserQuota(context.Background(), QuotaSpec{})
		assert.Error(suite.T(), err)
		assert.EqualError(suite.T(), err, errMissingUserID.Error())
	})

	suite.T().Run("get user quota but no user is specified", func(t *testing.T) {
		_, err := co.GetUserQuota(context.Background(), QuotaSpec{})
		assert.Error(suite.T(), err)
		assert.EqualError(suite.T(), err, errMissingUserID.Error())

	})
}
