// +build !luminous,!mimic

package admin

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResponse(t *testing.T) {
	e1 := errors.New("error one")
	e2 := errors.New("error two")
	r1 := response{
		body: []byte(`{"foo": "bar", "baz": 1}`),
	}
	r2 := response{
		status: "System notice: disabled for maintenance",
		err:    e1,
	}
	r3 := response{
		body:   []byte(`{"oof": "RAB", "baz": 8}`),
		status: "reversed polarity detected",
	}
	r4 := response{
		body:   []byte(`{"whoops": true, "state": "total protonic reversal"}`),
		status: "",
		err:    e2,
	}

	t.Run("ok", func(t *testing.T) {
		assert.True(t, r1.Ok())
		assert.False(t, r2.Ok())
		assert.True(t, r3.Ok())
	})

	t.Run("error", func(t *testing.T) {
		assert.Equal(t,
			"error one: \"System notice: disabled for maintenance\"",
			r2.Error())
		assert.Equal(t,
			e2.Error(),
			r4.Error())
	})

	t.Run("unwrap", func(t *testing.T) {
		assert.Equal(t, e1, r2.Unwrap())
		assert.Equal(t, e2, r4.Unwrap())
	})

	t.Run("status", func(t *testing.T) {
		assert.Equal(t, "", r1.Status())
		assert.Equal(t, "System notice: disabled for maintenance", r2.Status())
		assert.Equal(t, "reversed polarity detected", r3.Status())
	})

	t.Run("end", func(t *testing.T) {
		assert.Nil(t, r1.End())
		assert.NotNil(t, r2.End())
		assert.EqualValues(t, r2, r2.End())
	})

	t.Run("noStatus", func(t *testing.T) {
		assert.EqualValues(t, r1, r1.noStatus())
		assert.EqualValues(t, r2, r2.noStatus())

		x := r3.noStatus()
		assert.EqualValues(t, ErrStatusNotEmpty, x.Unwrap())
		assert.EqualValues(t, r3.Status(), x.Status())
	})

	t.Run("noBody", func(t *testing.T) {
		x := r1.noBody()
		assert.EqualValues(t, ErrBodyNotEmpty, x.Unwrap())
		assert.EqualValues(t, r1.Status(), x.Status())

		assert.EqualValues(t, r2, r2.noBody())

		rtemp := response{}
		assert.EqualValues(t, rtemp, rtemp.noBody())
	})

	t.Run("noData", func(t *testing.T) {
		x := r1.noData()
		assert.EqualValues(t, ErrBodyNotEmpty, x.Unwrap())
		assert.EqualValues(t, r1.Status(), x.Status())

		x = r3.noStatus()
		assert.EqualValues(t, ErrStatusNotEmpty, x.Unwrap())
		assert.EqualValues(t, r3.Status(), x.Status())

		rtemp := response{}
		assert.EqualValues(t, rtemp, rtemp.noData())
	})

	t.Run("filterDeprecated", func(t *testing.T) {
		assert.EqualValues(t, r1, r1.filterDeprecated())
		assert.EqualValues(t, r2, r2.filterDeprecated())

		rtemp := response{
			status: "blorple call is deprecated and will be removed in a future release",
		}
		x := rtemp.filterDeprecated()
		assert.True(t, x.Ok())
		assert.Nil(t, x.End())
		assert.Equal(t, "", x.Status())
	})

	t.Run("unmarshal", func(t *testing.T) {
		var v map[string]interface{}
		assert.EqualValues(t, r1, r1.unmarshal(&v))
		assert.EqualValues(t, "bar", v["foo"])

		assert.EqualValues(t, r2, r2.unmarshal(&v))

		rtemp := response{body: []byte("foo!")}
		x := rtemp.unmarshal(&v)
		assert.False(t, x.Ok())
		assert.Contains(t, x.Error(), "invalid character")
	})

	t.Run("newResponse", func(t *testing.T) {
		rtemp := newResponse(nil, "x", e2)
		assert.False(t, rtemp.Ok())
		assert.Equal(t, "x", rtemp.Status())
	})
}
