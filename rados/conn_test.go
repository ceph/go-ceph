package rados

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShutdownBeforeConnect(t *testing.T) {
	conn, err := NewConn()
	require.NoError(t, err)
	require.NotNil(t, conn.cluster)
	require.False(t, conn.connected)

	conn.Shutdown()

	assert.Nil(t, conn.cluster)
	assert.False(t, conn.connected)
}

func TestShutdownIsIdempotent(t *testing.T) {
	conn, err := NewConn()
	require.NoError(t, err)

	conn.Shutdown()
	conn.Shutdown()

	assert.Nil(t, conn.cluster)
	assert.False(t, conn.connected)
}
