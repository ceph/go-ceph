//go:build !nautilus
// +build !nautilus

package admin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLevelSpec(t *testing.T) {
	ls := NewLevelSpec("bob", "", "")
	assert.Equal(t, "bob/", ls.spec)

	ls = NewLevelSpec("bob", "", "foo")
	assert.Equal(t, "bob/foo", ls.spec)

	ls = NewLevelSpec("bob", "ns", "foo")
	assert.Equal(t, "bob/ns/foo", ls.spec)

	ls = NewLevelSpec("bob", "ns", "")
	assert.Equal(t, "bob/ns/", ls.spec)
}

func TestRawLevelSpec(t *testing.T) {
	rls := NewRawLevelSpec("foo/bar")
	assert.Equal(t, "foo/bar", rls.spec)

	// NewRawLevelSpec takes whatever junk it's given and does not validate
	rls = NewRawLevelSpec("totally! invalid! haha. ha...")
	assert.Equal(t, "totally! invalid! haha. ha...", rls.spec)
}
