package callbacks

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestCallbacks(t *testing.T) {
	t.Cleanup(reset)
	assert.Len(t, cmap, 0)
	assert.Len(t, free, 0)
	assert.Len(t, blocks, 0)

	i1 := Add("foo")
	i2 := Add("bar")
	i3 := Add("baz")
	assert.Len(t, cmap, 3)

	var x interface{}
	x = Lookup(i1)
	assert.NotNil(t, x)
	if s, ok := x.(string); ok {
		assert.EqualValues(t, s, "foo")
	}

	x = Lookup(unsafe.Pointer(&x))
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
	assert.Len(t, cmap, 0)
}

func TestCallbacksIndexing(t *testing.T) {
	t.Cleanup(reset)
	assert.Len(t, cmap, 0)
	assert.Len(t, free, 0)
	assert.Len(t, blocks, 0)

	i1 := Add("foo")
	i2 := Add("bar")
	_ = Add("baz")
	_ = Add("wibble")
	_ = Add("wabble")
	assert.Len(t, cmap, 5)

	// generally we assume that the callback data will be mostly LIFO
	// but can't guarantee it. Thus we check that when we remove the
	// first items inserted into the map there are no subsequent issues
	Remove(i1)
	Remove(i2)
	_ = Add("flim")
	ilast := Add("flam")
	assert.Len(t, cmap, 5)

	x := Lookup(ilast)
	assert.NotNil(t, x)
	if s, ok := x.(string); ok {
		assert.EqualValues(t, s, "flam")
	}
}

func TestCallbacksData(t *testing.T) {
	t.Cleanup(reset)
	assert.Len(t, cmap, 0)
	assert.Len(t, free, 0)
	assert.Len(t, blocks, 0)

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
