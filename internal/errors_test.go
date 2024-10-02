package internal

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ceph/go-ceph/cephfs"
	"github.com/ceph/go-ceph/rados"
	"github.com/ceph/go-ceph/rbd"
)

func TestErrorCompare(t *testing.T) {
	t.Run("ErrNotConnected", func(t *testing.T) {
		err := rados.ErrNotConnected
		assert.True(t, errors.Is(err, cephfs.ErrNotConnected))
		assert.False(t, errors.Is(err, cephfs.ErrNotExist))
	})

	t.Run("ErrNotExist", func(t *testing.T) {
		err := rbd.ErrNotExist
		assert.True(t, errors.Is(err, rbd.ErrNotFound))
		assert.True(t, errors.Is(err, cephfs.ErrNotExist))
		assert.False(t, errors.Is(err, cephfs.ErrNotConnected))
	})

	t.Run("ErrNotFound", func(t *testing.T) {
		err := rbd.ErrNotFound
		assert.True(t, errors.Is(rbd.ErrNotFound, rados.ErrNotFound))
		assert.False(t, errors.Is(err, rados.ErrNotConnected))
	})
}
