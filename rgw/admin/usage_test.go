package admin

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func (suite *RadosGWTestSuite) TestUsage() {
	suite.SetupConnection()
	co, err := New(suite.endpoint, suite.accessKey, suite.secretKey, newDebugHTTPClient(http.DefaultClient))
	assert.NoError(suite.T(), err)

	suite.T().Run("get usage", func(t *testing.T) {
		pTrue := true
		usage, err := co.GetUsage(context.Background(), Usage{ShowSummary: &pTrue})
		assert.NoError(suite.T(), err)
		assert.NotEmpty(t, usage)
	})

	suite.T().Run("trim usage", func(_ *testing.T) {
		pFalse := false
		_, err := co.GetUsage(context.Background(), Usage{RemoveAll: &pFalse})
		assert.NoError(suite.T(), err)
	})
}
