package cref

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCallbacks(t *testing.T) {
	defer reset()
	assert.Len(t, cmap, 1)

	i1 := Add("foo")
	i2 := Add("bar")
	i3 := Add("baz")
	assert.Len(t, cmap, 4)

	var x interface{}
	x = Lookup(i1)
	assert.NotNil(t, x)
	if s, ok := x.(string); ok {
		assert.EqualValues(t, s, "foo")
	}

	x = Lookup(Ref{5555})
	assert.Nil(t, x)

	x = Lookup(i3)
	assert.NotNil(t, x)
	if s, ok := x.(string); ok {
		assert.EqualValues(t, s, "baz")
	}
	Remove(i3)
	x = Lookup(i3)
	assert.Nil(t, x)

	Remove(i2)
	x = Lookup(i2)
	assert.Nil(t, x)

	Remove(i1)
	assert.Len(t, cmap, 4)
	assert.Len(t, free, 3)
}

func TestCallbacksIndexing(t *testing.T) {
	defer reset()
	assert.Len(t, cmap, 1)

	i1 := Add("foo")
	i2 := Add("bar")
	_ = Add("baz")
	_ = Add("wibble")
	_ = Add("wabble")
	assert.Len(t, cmap, 6)

	// Check that when we remove the first items inserted into the map there are
	// no subsequent issues
	Remove(i1)
	Remove(i2)
	assert.Len(t, free, 2)
	_ = Add("flim")
	ilast := Add("flam")
	assert.Len(t, cmap, 6)
	assert.Len(t, free, 0)

	x := Lookup(ilast)
	assert.NotNil(t, x)
	if s, ok := x.(string); ok {
		assert.EqualValues(t, s, "flam")
	}
}

func TestCallbacksData(t *testing.T) {
	defer reset()
	assert.Len(t, cmap, 1)

	// insert a plain function
	i1 := Add(func(v int) int { return v + 1 })

	// insert a type "containing" a function, note that it doesn't
	// actually have a callable function. Users of the type must
	// check that themselves
	type flup struct {
		Stuff int
		Junk  func(int, int) error
	}
	i2 := Add(flup{
		Stuff: 55,
	})

	// did we get a function back
	x1 := Lookup(i1)
	if assert.NotNil(t, x1) {
		if f, ok := x1.(func(v int) int); ok {
			assert.Equal(t, 2, f(1))
		} else {
			t.Fatalf("conversion failed")
		}
	}

	// did we get our data structure back
	x2 := Lookup(i2)
	if assert.NotNil(t, x2) {
		if d, ok := x2.(flup); ok {
			assert.Equal(t, 55, d.Stuff)
			assert.Nil(t, d.Junk)
		} else {
			t.Fatalf("conversion failed")
		}
	}
}
