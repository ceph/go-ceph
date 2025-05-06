package cephfs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var xattrSamples = []struct {
	name  string
	value []byte
}{
	{
		name:  "user.xPhrase",
		value: []byte("june and july"),
	},
	{
		name:  "user.xHasNulls",
		value: []byte("\x00got\x00null?\x00"),
	},
	{
		name:  "user.x2kZeros",
		value: make([]byte, 2048),
	},
	// Older versions of ceph had a bug where using an empty value caused the
	// xattr to be unset. This has been fixed for nautilus and octopus.
	{
		name:  "user.xEmpty",
		value: []byte(""),
	},
}

func TestGetSetXattr(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)
	fname := "TestGetSetXattr.txt"

	f, err := mount.Open(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, f.Close())
		assert.NoError(t, mount.Unlink(fname))
	}()

	for _, s := range xattrSamples {
		t.Run("roundTrip-"+s.name, func(t *testing.T) {
			err := f.SetXattr(s.name, s.value, XattrDefault)
			assert.NoError(t, err)
			b, err := f.GetXattr(s.name)
			assert.NoError(t, err)
			assert.EqualValues(t, s.value, b)
		})
	}

	t.Run("missingXattrOnGet", func(t *testing.T) {
		_, err := f.GetXattr("user.never-set")
		assert.Error(t, err)
	})

	t.Run("emptyNameGet", func(t *testing.T) {
		_, err := f.GetXattr("")
		assert.Error(t, err)
	})

	t.Run("emptyNameSet", func(t *testing.T) {
		err := f.SetXattr("", []byte("foo"), XattrDefault)
		assert.Error(t, err)
	})

	t.Run("invalidFile", func(t *testing.T) {
		f1 := &File{}
		err := f1.SetXattr(xattrSamples[0].name, xattrSamples[0].value, XattrDefault)
		assert.Error(t, err)
		_, err = f1.GetXattr(xattrSamples[0].name)
		assert.Error(t, err)
	})
}

func TestListXattr(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)
	fname := "TestListXattr.txt"

	f, err := mount.Open(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, f.Close())
		assert.NoError(t, mount.Unlink(fname))
	}()

	t.Run("listXattrs1", func(t *testing.T) {
		for _, s := range xattrSamples[:1] {
			err := f.SetXattr(s.name, s.value, XattrDefault)
			assert.NoError(t, err)
		}
		xl, err := f.ListXattr()
		assert.NoError(t, err)
		assert.Len(t, xl, 1)
		assert.Contains(t, xl, xattrSamples[0].name)
	})

	t.Run("listXattrs2", func(t *testing.T) {
		for _, s := range xattrSamples {
			err := f.SetXattr(s.name, s.value, XattrDefault)
			assert.NoError(t, err)
		}
		xl, err := f.ListXattr()
		assert.NoError(t, err)
		assert.Len(t, xl, 4)
		assert.Contains(t, xl, xattrSamples[0].name)
		assert.Contains(t, xl, xattrSamples[1].name)
		assert.Contains(t, xl, xattrSamples[2].name)
		assert.Contains(t, xl, xattrSamples[3].name)
	})

	t.Run("invalidFile", func(t *testing.T) {
		f1 := &File{}
		_, err := f1.ListXattr()
		assert.Error(t, err)
	})
}

func TestRemoveXattr(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)
	fname := "TestRemoveXattr.txt"

	f, err := mount.Open(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, f.Close())
		assert.NoError(t, mount.Unlink(fname))
	}()

	t.Run("removeXattr", func(t *testing.T) {
		s := xattrSamples[0]
		err := f.SetXattr(s.name, s.value, XattrDefault)
		err = f.RemoveXattr(s.name)
		assert.NoError(t, err)
	})

	t.Run("removeMissingXattr", func(t *testing.T) {
		s := xattrSamples[1]
		err := f.RemoveXattr(s.name)
		assert.Error(t, err)
	})

	t.Run("emptyName", func(t *testing.T) {
		err := f.RemoveXattr("")
		assert.Error(t, err)
	})

	t.Run("invalidFile", func(t *testing.T) {
		f1 := &File{}
		err := f1.RemoveXattr(xattrSamples[0].name)
		assert.Error(t, err)
	})
}
