package dlsym

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLookupSymbol(t *testing.T) {
	t.Run("ValidSymbol", func(t *testing.T) {
		sym, err := LookupSymbol("dlsym")
		assert.NotNil(t, sym)
		assert.NoError(t, err)
	})

	t.Run("InvalidSymbol", func(t *testing.T) {
		sym, err := LookupSymbol("go_ceph_dlsym")
		assert.Nil(t, sym)
		assert.True(t, errors.Is(err, ErrUndefinedSymbol))
	})
}
