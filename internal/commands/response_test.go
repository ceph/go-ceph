package commands

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResponse(t *testing.T) {
	e1 := errors.New("error one")
	e2 := errors.New("error two")
	r1 := Response{
		body: []byte(`{"foo": "bar", "baz": 1}`),
	}
	r2 := Response{
		status: "System notice: disabled for maintenance",
		err:    e1,
	}
	r3 := Response{
		body:   []byte(`{"oof": "RAB", "baz": 8}`),
		status: "reversed polarity detected",
	}
	r4 := Response{
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
		assert.EqualValues(t, r1, r1.NoStatus())
		assert.EqualValues(t, r2, r2.NoStatus())

		x := r3.NoStatus()
		assert.EqualValues(t, ErrStatusNotEmpty, x.Unwrap())
		assert.EqualValues(t, r3.Status(), x.Status())
	})

	t.Run("noBody", func(t *testing.T) {
		x := r1.NoBody()
		assert.EqualValues(t, ErrBodyNotEmpty, x.Unwrap())
		assert.EqualValues(t, r1.Status(), x.Status())

		assert.EqualValues(t, r2, r2.NoBody())

		rtemp := Response{}
		assert.EqualValues(t, rtemp, rtemp.NoBody())
	})

	t.Run("noData", func(t *testing.T) {
		x := r1.NoData()
		assert.EqualValues(t, ErrBodyNotEmpty, x.Unwrap())
		assert.EqualValues(t, r1.Status(), x.Status())

		x = r3.NoStatus()
		assert.EqualValues(t, ErrStatusNotEmpty, x.Unwrap())
		assert.EqualValues(t, r3.Status(), x.Status())

		rtemp := Response{}
		assert.EqualValues(t, rtemp, rtemp.NoData())
	})

	t.Run("filterDeprecated", func(t *testing.T) {
		assert.EqualValues(t, r1, r1.FilterDeprecated())
		assert.EqualValues(t, r2, r2.FilterDeprecated())

		rtemp := Response{
			status: "blorple call is deprecated and will be removed in a future release",
		}
		x := rtemp.FilterDeprecated()
		assert.True(t, x.Ok())
		assert.Nil(t, x.End())
		assert.Equal(t, "", x.Status())
	})

	t.Run("unmarshal", func(t *testing.T) {
		var v map[string]interface{}
		assert.EqualValues(t, r1, r1.Unmarshal(&v))
		assert.EqualValues(t, "bar", v["foo"])

		assert.EqualValues(t, r2, r2.Unmarshal(&v))

		rtemp := Response{body: []byte("foo!")}
		x := rtemp.Unmarshal(&v)
		assert.False(t, x.Ok())
		assert.Contains(t, x.Error(), "invalid character")
	})

	t.Run("newResponse", func(t *testing.T) {
		rtemp := NewResponse(nil, "x", e2)
		assert.False(t, rtemp.Ok())
		assert.Equal(t, "x", rtemp.Status())
	})

	t.Run("notImplemented", func(t *testing.T) {
		rtemp := Response{
			status: "No handler found for this function",
			err:    myCephError(-22),
		}
		if assert.False(t, rtemp.Ok()) {
			err := rtemp.End()
			assert.Error(t, err)
			var n NotImplementedError
			assert.True(t, errors.As(err, &n))
			assert.Contains(t, err.Error(), "not implemented")
		}
	})

	t.Run("filterBodyPrefix", func(t *testing.T) {
		rtemp := Response{
			body: []byte("No way, no how"),
		}
		if assert.True(t, rtemp.Ok()) {
			r2 := rtemp.FilterBodyPrefix("No way")
			assert.True(t, r2.Ok())
			assert.Equal(t, []byte(""), r2.body)
			err := r2.NoBody().End()
			assert.NoError(t, err)
		}
	})
	t.Run("filterBodyPrefixEmpty", func(t *testing.T) {
		rtemp := Response{
			body: []byte(""),
		}
		if assert.True(t, rtemp.Ok()) {
			r2 := rtemp.FilterBodyPrefix("No way")
			assert.True(t, r2.Ok())
			assert.Equal(t, []byte(""), r2.body)
			err := r2.NoBody().End()
			assert.NoError(t, err)
		}
	})
	t.Run("filterBodyPrefixNoMatch", func(t *testing.T) {
		rtemp := Response{
			body: []byte("No way, no how"),
		}
		if assert.True(t, rtemp.Ok()) {
			r2 := rtemp.FilterBodyPrefix("No foolin")
			assert.True(t, r2.Ok())
			assert.Equal(t, []byte("No way, no how"), r2.body)
			err := r2.NoBody().End()
			assert.Error(t, err)
		}
	})
}

type myCephError int

func (myCephError) Error() string {
	return "oops"
}

func (e myCephError) ErrorCode() int {
	return int(e)
}
