// +build !luminous,!mimic

package admin

import (
	"errors"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ceph/go-ceph/internal/commands"
)

var (
	cachedFSAdmin *FSAdmin

	// set debugTrace to true to use tracing in tests
	debugTrace = false

	// some tests are sensitive to the server version
	serverIsNautilus = false
	serverIsOctopus  = false
)

func init() {
	dt := os.Getenv("GO_CEPH_TEST_DEBUG_TRACE")
	if ok, err := strconv.ParseBool(dt); ok && err == nil {
		debugTrace = true
	}
	switch os.Getenv("CEPH_VERSION") {
	case "nautilus":
		serverIsNautilus = true
	case "octopus":
		serverIsOctopus = true
	}
}

func TestServerSentinel(t *testing.T) {
	// there probably *is* a better way to do this but I'm doing what's easy
	// and expedient at the moment. That's tying the tests to the environment
	// var to tell us what version of the *server* we are testing against. The
	// build tags control what version of the *client libs* we use.  These
	// happen to be the same for our CI tests today, but its a lousy way to
	// organize things IMO.
	// This check is intended to fail the test suite if you don't tell it a
	// server version it expects and force us to update the tests if a new
	// version of ceph is added.
	if !serverIsNautilus && !serverIsOctopus {
		t.Fatalf("server must be nautilus or octopus (do the tests need updating?)")
	}
}

func getFSAdmin(t *testing.T) *FSAdmin {
	if cachedFSAdmin != nil {
		return cachedFSAdmin
	}
	fsa, err := New()
	require.NoError(t, err)
	require.NotNil(t, fsa)
	// We steal the connection set up by the New() method and wrap it in an
	// optional tracer.
	c := fsa.conn
	if debugTrace {
		c = commands.NewTraceCommander(c)
	}
	cachedFSAdmin = NewFromConn(c)
	// We sleep briefly before returning in order to ensure we have a mgr map
	// before we start executing the tests.
	time.Sleep(50 * time.Millisecond)
	return cachedFSAdmin
}

func TestInvalidFSAdmin(t *testing.T) {
	fsa := &FSAdmin{}
	res := fsa.rawMgrCommand([]byte("FOOBAR!"))
	assert.Error(t, res.Unwrap())
}

type badMarshalType bool

func (badMarshalType) MarshalJSON() ([]byte, error) {
	return nil, errors.New("Zowie! wow")
}

func TestBadMarshal(t *testing.T) {
	fsa := getFSAdmin(t)

	var bad badMarshalType
	res := fsa.marshalMgrCommand(bad)
	assert.Error(t, res.Unwrap())
}

func TestParseListNames(t *testing.T) {
	R := newResponse
	t.Run("error", func(t *testing.T) {
		_, err := parseListNames(R(nil, "", errors.New("bonk")))
		assert.Error(t, err)
		assert.Equal(t, "bonk", err.Error())
	})
	t.Run("statusSet", func(t *testing.T) {
		_, err := parseListNames(R(nil, "unexpected!", nil))
		assert.Error(t, err)
	})
	t.Run("badJSON", func(t *testing.T) {
		_, err := parseListNames(R([]byte("Foo[[["), "", nil))
		assert.Error(t, err)
	})
	t.Run("ok", func(t *testing.T) {
		l, err := parseListNames(R([]byte(`[{"name":"bob"}]`), "", nil))
		assert.NoError(t, err)
		if assert.Len(t, l, 1) {
			assert.Equal(t, "bob", l[0])
		}
	})
}

func TestCheckEmptyResponseExpected(t *testing.T) {
	R := newResponse
	t.Run("error", func(t *testing.T) {
		err := R(nil, "", errors.New("bonk")).NoData().End()
		assert.Error(t, err)
		assert.Equal(t, "bonk", err.Error())
	})
	t.Run("statusSet", func(t *testing.T) {
		err := R(nil, "unexpected!", nil).NoData().End()
		assert.Error(t, err)
	})
	t.Run("someJSON", func(t *testing.T) {
		err := R([]byte(`{"trouble": true}`), "", nil).NoData().End()
		assert.Error(t, err)
	})
	t.Run("ok", func(t *testing.T) {
		err := R([]byte{}, "", nil).NoData().End()
		assert.NoError(t, err)
	})
}
