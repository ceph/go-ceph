package cutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var tbl = []struct {
	val  []byte
	res1 []string
	res2 []string
}{
	// simple inputs
	{
		val:  []byte("foo\x00bar\x00baz\x00"),
		res1: []string{"foo", "bar", "baz"},
		res2: []string{"foo", "bar", "baz"},
	},
	// no trailing null bytes
	{
		val:  []byte("meow mix"),
		res1: []string{"meow mix"},
		res2: []string{"meow mix"},
	},
	// one item
	{
		val:  []byte("fancy feast\x00"),
		res1: []string{"fancy feast"},
		res2: []string{"fancy feast"},
	},
	// nuttin dare
	{
		val:  []byte(""),
		res1: []string{},
		res2: []string{},
	},
	// almost nuttin
	{
		val:  []byte("\x00"),
		res1: []string{},
		res2: []string{},
	},
	// how multiple adjacent nulls are handled
	{
		val:  []byte("kibbles\x00\x00and\x00bits"),
		res1: []string{"kibbles", "and", "bits"},
		res2: []string{"kibbles", "", "and", "bits"},
	},
	{
		val:  []byte("dinki\x00\x00\x00di\x00\x00"),
		res1: []string{"dinki", "di"},
		res2: []string{"dinki", "", "", "di", ""},
	},
	// starting with a null
	{
		val:  []byte("\x00caesar\x00"),
		res1: []string{"caesar"},
		res2: []string{"", "caesar"},
	},
}

func TestSplitBufStrings(t *testing.T) {
	for _, x := range tbl {
		assert.Equal(t, x.res1, SplitSparseBuffer(x.val))
	}
	for _, x := range tbl {
		assert.Equal(t, x.res2, SplitBuffer(x.val))
	}
}
