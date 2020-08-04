package admin

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var cachedFSAdmin *FSAdmin

func getFSAdmin(t *testing.T) *FSAdmin {
	if cachedFSAdmin != nil {
		return cachedFSAdmin
	}
	var err error
	cachedFSAdmin, err := New()
	require.NoError(t, err)
	require.NotNil(t, cachedFSAdmin)
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
