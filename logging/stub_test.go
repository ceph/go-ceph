package logging

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestInterface verifies that stub logger is something and that something
// meets the Logger interface. StubLogger doesn't actually do anything
// (intentionally) so there's not a lot more to assert.
func TestInterface(t *testing.T) {
	s := NewStubLogger()
	assert.NotNil(t, s)

	var l Logger = s
	l.Errorf("foo %v", 77)
	l.Errorf("bar %v, %s", 101, "blat")
}
