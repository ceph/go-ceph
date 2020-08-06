// +build !luminous,!mimic

package admin

import (
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	cachedFSAdmin *FSAdmin

	// set debugTrace to true to use tracing in tests
	debugTrace = false
)

func init() {
	dt := os.Getenv("GO_CEPH_TEST_DEBUG_TRACE")
	if dt == "yes" || dt == "true" {
		debugTrace = true
	}
}

// tracingCommander serves two purposes: first, it allows one to trace the
// input and output json when running the tests. It can help with actually
// debugging the tests. Second, it demonstrates the rationale for using an
// interface in FSAdmin. You can layer any sort of debugging, error injection,
// or whatnot between the FSAdmin layer and the RADOS layer.
type tracingCommander struct {
	conn RadosCommander
}

func tracer(c RadosCommander) RadosCommander {
	return &tracingCommander{c}
}

func (t *tracingCommander) MgrCommand(buf [][]byte) ([]byte, string, error) {
	for i := range buf {
		fmt.Println("IN:", string(buf[i]))
	}
	r, s, err := t.conn.MgrCommand(buf)
	fmt.Println("OUT(result):", string(r))
	if s != "" {
		fmt.Println("OUT(status):", s)
	}
	if err != nil {
		fmt.Println("OUT(error):", err.Error())
	}
	return r, s, err
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
		c = tracer(c)
	}
	cachedFSAdmin = NewFromConn(c)
	// We sleep briefly before returning in order to ensure we have a mgr map
	// before we start executing the tests.
	time.Sleep(50 * time.Millisecond)
	return cachedFSAdmin
}

func TestInvalidFSAdmin(t *testing.T) {
	fsa := &FSAdmin{}
	_, _, err := fsa.rawMgrCommand([]byte("FOOBAR!"))
	assert.Error(t, err)
}

type badMarshalType bool

func (badMarshalType) MarshalJSON() ([]byte, error) {
	return nil, errors.New("Zowie! wow")
}

func TestBadMarshal(t *testing.T) {
	fsa := getFSAdmin(t)

	var bad badMarshalType
	_, _, err := fsa.marshalMgrCommand(bad)
	assert.Error(t, err)
}

func TestParseListNames(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		_, err := parseListNames(nil, "", errors.New("bonk"))
		assert.Error(t, err)
		assert.Equal(t, "bonk", err.Error())
	})
	t.Run("statusSet", func(t *testing.T) {
		_, err := parseListNames(nil, "unexpected!", nil)
		assert.Error(t, err)
	})
	t.Run("badJSON", func(t *testing.T) {
		_, err := parseListNames([]byte("Foo[[["), "", nil)
		assert.Error(t, err)
	})
	t.Run("ok", func(t *testing.T) {
		l, err := parseListNames([]byte(`[{"name":"bob"}]`), "", nil)
		assert.NoError(t, err)
		if assert.Len(t, l, 1) {
			assert.Equal(t, "bob", l[0])
		}
	})
}

func TestCheckEmptyResponseExpected(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		err := checkEmptyResponseExpected(nil, "", errors.New("bonk"))
		assert.Error(t, err)
		assert.Equal(t, "bonk", err.Error())
	})
	t.Run("statusSet", func(t *testing.T) {
		err := checkEmptyResponseExpected(nil, "unexpected!", nil)
		assert.Error(t, err)
	})
	t.Run("someJSON", func(t *testing.T) {
		err := checkEmptyResponseExpected([]byte(`{"trouble": true}`), "", nil)
		assert.Error(t, err)
	})
	t.Run("ok", func(t *testing.T) {
		err := checkEmptyResponseExpected([]byte{}, "", nil)
		assert.NoError(t, err)
	})
}
