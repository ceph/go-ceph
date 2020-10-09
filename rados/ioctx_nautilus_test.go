// +build !luminous,!mimic

package rados

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func (suite *RadosTestSuite) TestSetGetNamespace() {
	suite.SetupConnection()

	suite.T().Run("validNS", func(t *testing.T) {
		suite.ioctx.SetNamespace("space1")
		ns, err := suite.ioctx.GetNamespace()
		assert.NoError(t, err)
		assert.Equal(t, "space1", ns)
	})

	suite.T().Run("allNamespaces", func(t *testing.T) {
		suite.ioctx.SetNamespace(AllNamespaces)
		ns, err := suite.ioctx.GetNamespace()
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), AllNamespaces, ns)
	})

	suite.T().Run("invalidIoctx", func(t *testing.T) {
		i := &IOContext{}
		ns, err := i.GetNamespace()
		assert.Error(suite.T(), err)
		assert.Equal(suite.T(), "", ns)
	})
}
