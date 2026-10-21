//go:build !(pacific || quincy || reef)

package admin

import (
	"context"
	"net/http"
	"testing"

	"github.com/ceph/go-ceph/internal/util"
	"github.com/stretchr/testify/assert"
)

func (suite *RadosGWTestSuite) TestAccount() {
	if util.CurrentCephVersion() < util.CephTentacle {
		suite.T().Skip("account get/delete REST APIs have broken capability checks on Squid")
	}

	suite.SetupConnection()
	co, err := New(suite.endpoint, suite.accessKey, suite.secretKey, newDebugHTTPClient(http.DefaultClient))
	assert.NoError(suite.T(), err)

	suite.T().Run("successfully create account", func(_ *testing.T) {
		account, err := co.CreateAccount(context.Background(), Account{ID: "RGW12345678901234567", Name: "test-account"})
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), "RGW12345678901234567", account.ID)
		assert.Equal(suite.T(), "test-account", account.Name)
	})

	suite.T().Run("try to create account that already exists", func(_ *testing.T) {
		_, err := co.CreateAccount(context.Background(), Account{ID: "RGW12345678901234567", Name: "test-account"})
		assert.ErrorIs(suite.T(), err, ErrAccountAlreadyExists)
	})

	suite.T().Run("fail to get account since no ID provided", func(_ *testing.T) {
		_, err := co.GetAccount(context.Background(), "")
		assert.ErrorIs(suite.T(), err, ErrInvalidArgument)
	})

	suite.T().Run("successfully get account", func(t *testing.T) {
		account, err := co.GetAccount(context.Background(), "RGW12345678901234567")
		assert.NoError(t, err)
		assert.Equal(t, "RGW12345678901234567", account.ID)
		assert.Equal(t, "test-account", account.Name)
	})

	suite.T().Run("fail to modify account since no ID provided", func(_ *testing.T) {
		_, err := co.ModifyAccount(context.Background(), Account{Name: "modified-account"})
		assert.ErrorIs(suite.T(), err, ErrInvalidArgument)
	})

	suite.T().Run("successfully modify account", func(_ *testing.T) {
		account, err := co.ModifyAccount(context.Background(), Account{ID: "RGW12345678901234567", Name: "modified-account"})
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), "RGW12345678901234567", account.ID)
		assert.Equal(suite.T(), "modified-account", account.Name)
	})

	suite.T().Run("fail to delete account since no ID provided", func(_ *testing.T) {
		err := co.DeleteAccount(context.Background(), "")
		assert.ErrorIs(suite.T(), err, ErrInvalidArgument)
	})

	suite.T().Run("successfully delete account", func(t *testing.T) {
		err := co.DeleteAccount(context.Background(), "RGW12345678901234567")
		assert.NoError(t, err)
	})
}
