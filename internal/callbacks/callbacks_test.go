package callbacks

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCallbacks(t *testing.T) {
	cbks := New()
	assert.Len(t, cbks.cmap, 0)

	i1 := cbks.Add("foo")
	i2 := cbks.Add("bar")
	i3 := cbks.Add("baz")
	assert.Len(t, cbks.cmap, 3)

	var x interface{}
	x = cbks.Lookup(i1)
	assert.NotNil(t, x)
	if s, ok := x.(string); ok {
		assert.EqualValues(t, s, "foo")
	}

	x = cbks.Lookup(5555)
	assert.Nil(t, x)

	x = cbks.Lookup(i3)
	assert.NotNil(t, x)
	if s, ok := x.(string); ok {
		assert.EqualValues(t, s, "baz")
	}
	cbks.Remove(i3)
	x = cbks.Lookup(i3)
	assert.Nil(t, x)

	cbks.Remove(i2)
	x = cbks.Lookup(i2)
	assert.Nil(t, x)

	cbks.Remove(i1)
	assert.Len(t, cbks.cmap, 0)
}

func TestCallbacksIndexing(t *testing.T) {
	cbks := New()
	assert.Len(t, cbks.cmap, 0)

	i1 := cbks.Add("foo")
	i2 := cbks.Add("bar")
	_ = cbks.Add("baz")
	_ = cbks.Add("wibble")
	_ = cbks.Add("wabble")
	assert.Len(t, cbks.cmap, 5)

	// generally we assume that the callback data will be mostly LIFO
	// but can't guarantee it. Thus we check that when we remove the
	// first items inserted into the map there are no subsequent issues
	cbks.Remove(i1)
	cbks.Remove(i2)
	_ = cbks.Add("flim")
	ilast := cbks.Add("flam")
	assert.Len(t, cbks.cmap, 5)

	x := cbks.Lookup(ilast)
	assert.NotNil(t, x)
	if s, ok := x.(string); ok {
		assert.EqualValues(t, s, "flam")
	}
}

func TestCallbacksData(t *testing.T) {
	cbks := New()
	assert.Len(t, cbks.cmap, 0)

	// insert a plain function
	i1 := cbks.Add(func(v int) int { return v + 1 })

	// insert a type "containing" a function, note that it doesn't
	// actually have a callable function. Users of the type must
	// check that themselves
	type flup struct {
		Stuff int
		Junk  func(int, int) error
	}
	i2 := cbks.Add(flup{
		Stuff: 55,
	})

	// did we get a function back
	x1 := cbks.Lookup(i1)
	if assert.NotNil(t, x1) {
		if f, ok := x1.(func(v int) int); ok {
			assert.Equal(t, 2, f(1))
		} else {
			t.Fatalf("conversion failed")
		}
	}

	// did we get our data structure back
	x2 := cbks.Lookup(i2)
	if assert.NotNil(t, x2) {
		if d, ok := x2.(flup); ok {
			assert.Equal(t, 55, d.Stuff)
			assert.Nil(t, d.Junk)
		} else {
			t.Fatalf("conversion failed")
		}
	}
}

func BenchmarkCallbacks(t *testing.B) {
	cbks := New()
	workers := 1000
	var wg sync.WaitGroup
	f := func() {
		defer wg.Done()
		var x interface{}
		var i1, i2, i3 uintptr
		for i := 0; i < t.N/workers; i++ {
			i1 = cbks.Add("foo")
			i2 = cbks.Add("bar")
			i3 = cbks.Add("baz")
			x = cbks.Lookup(i1)
			assert.NotNil(t, x)
			if s, ok := x.(string); ok {
				assert.Equal(t, "foo", s)
			} else {
				t.Fail()
			}
			x = cbks.Lookup(i3)
			assert.NotNil(t, x)
			if s, ok := x.(string); ok {
				assert.Equal(t, "baz", s)
			} else {
				t.Fail()
			}
			cbks.Remove(i3)
			cbks.Remove(i2)
			cbks.Remove(i1)
		}
	}
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go f()
	}
	wg.Wait()
	assert.Len(t, cbks.cmap, 0)
}
