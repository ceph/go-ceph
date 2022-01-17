package admintest

import (
	"errors"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	ccom "github.com/ceph/go-ceph/common/commands"
	"github.com/ceph/go-ceph/internal/commands"
	"github.com/ceph/go-ceph/rados"
)

var errTimedOut = errors.New("timed out waiting for connect")

// DebugTraceEnabled returns true if the environment variable
// GO_CEPH_TEST_DEBUG_TRACE indicates that the JSON requests and responses
// should be traced (logged).
func DebugTraceEnabled() bool {
	dt := os.Getenv("GO_CEPH_TEST_DEBUG_TRACE")
	ok, err := strconv.ParseBool(dt)
	return ok && err == nil
}

// NewConn return a new rados connection based on the default config.
func NewConn(t *testing.T) *rados.Conn {
	return NewConnFromConfig(t, "")
}

// NewConnFromConfig returns a new rados connection based on the specified
// config file. If configFile is the empty string, the default config is
// used.
func NewConnFromConfig(t *testing.T, configFile string) *rados.Conn {
	conn, err := rados.NewConn()
	require.NoError(t, err)
	require.NotNil(t, conn)
	if configFile == "" {
		err = conn.ReadDefaultConfigFile()
		require.NoError(t, err)
	} else {
		err = conn.ReadConfigFile(configFile)
		require.NoError(t, err)
	}

	timeout := time.After(time.Second * 5)
	ch := make(chan error)
	go func(conn *rados.Conn) {
		ch <- conn.Connect()
	}(conn)
	select {
	case err = <-ch:
	case <-timeout:
		err = errTimedOut
	}
	require.NoError(t, err)

	return conn
}

// WrapConn wraps the given conn in as a "tracer" if so enabled in the
// environment.
func WrapConn(conn *rados.Conn) ccom.RadosCommander {
	var c ccom.RadosCommander = conn
	if DebugTraceEnabled() {
		c = commands.NewTraceCommander(conn)
	}
	return c
}

// Connector is a caching and convenience type for rados connections.
// Typically it is instantiated as a global in a go _test.go file.
type Connector struct {
	conn *rados.Conn
}

// NewConnector returns a new Connector.
func NewConnector() *Connector {
	return &Connector{}
}

func (c *Connector) realConn(t *testing.T) *rados.Conn {
	if c.conn == nil {
		c.conn = NewConn(t)
	}
	return c.conn
}

// GetConn returns the underlying rados connection, either cached or
// newly created.
func (c *Connector) GetConn(t *testing.T) *rados.Conn {
	conn := c.realConn(t)
	// We sleep briefly before returning in order to ensure we have a mgr map
	// before we start executing the tests.
	time.Sleep(50 * time.Millisecond)
	return conn
}

// Get returns a RadosCommander that may or may not be wrapped.
// Typically this is the one you want.
func (c *Connector) Get(t *testing.T) ccom.RadosCommander {
	conn := WrapConn(c.realConn(t))
	// We sleep briefly before returning in order to ensure we have a mgr map
	// before we start executing the tests.
	time.Sleep(50 * time.Millisecond)
	return conn
}
