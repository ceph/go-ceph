package cephfs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSetXattrPath(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)
	fname := "TestGetSetXattrPath.txt"

	f1, err := mount.Open(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	require.NoError(t, err)
	assert.NoError(t, f1.Close())
	defer func() {
		assert.NoError(t, mount.Unlink(fname))
	}()

	for _, s := range xattrSamples {
		t.Run("roundTrip-"+s.name, func(t *testing.T) {
			err := mount.SetXattr(fname, s.name, s.value, XattrDefault)
			assert.NoError(t, err)
			b, err := mount.GetXattr(fname, s.name)
			assert.NoError(t, err)
			assert.EqualValues(t, s.value, b)
		})
	}

	t.Run("missingXattrOnGet", func(t *testing.T) {
		_, err := mount.GetXattr(fname, "user.never-set")
		assert.Error(t, err)
	})

	t.Run("emptyNameGet", func(t *testing.T) {
		_, err := mount.GetXattr(fname, "")
		assert.Error(t, err)
	})

	t.Run("emptyNameSet", func(t *testing.T) {
		err := mount.SetXattr(fname, "", []byte("foo"), XattrDefault)
		assert.Error(t, err)
	})

	t.Run("invalidMount", func(t *testing.T) {
		m := &MountInfo{}
		err := m.SetXattr(fname, xattrSamples[0].name, xattrSamples[0].value, XattrDefault)
		assert.Error(t, err)
		_, err = m.GetXattr(fname, xattrSamples[0].name)
		assert.Error(t, err)
	})
}

func TestListXattrPath(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)
	fname := "TestListXattrPath.txt"

	f1, err := mount.Open(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	require.NoError(t, err)
	assert.NoError(t, f1.Close())
	defer func() {
		assert.NoError(t, mount.Unlink(fname))
	}()

	t.Run("listXattrs1", func(t *testing.T) {
		for _, s := range xattrSamples[:1] {
			err := mount.SetXattr(fname, s.name, s.value, XattrDefault)
			assert.NoError(t, err)
		}
		xl, err := mount.ListXattr(fname)
		assert.NoError(t, err)
		assert.Len(t, xl, 1)
		assert.Contains(t, xl, xattrSamples[0].name)
	})

	t.Run("listXattrs2", func(t *testing.T) {
		for _, s := range xattrSamples {
			err := mount.SetXattr(fname, s.name, s.value, XattrDefault)
			assert.NoError(t, err)
		}
		xl, err := mount.ListXattr(fname)
		assert.NoError(t, err)
		assert.Len(t, xl, 4)
		assert.Contains(t, xl, xattrSamples[0].name)
		assert.Contains(t, xl, xattrSamples[1].name)
		assert.Contains(t, xl, xattrSamples[2].name)
		assert.Contains(t, xl, xattrSamples[3].name)
	})

	t.Run("invalidMount", func(t *testing.T) {
		m := &MountInfo{}
		_, err := m.ListXattr(fname)
		assert.Error(t, err)
	})
}

func TestRemoveXattrPath(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)
	fname := "TestRemoveXattrPath.txt"

	f1, err := mount.Open(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	require.NoError(t, err)
	assert.NoError(t, f1.Close())
	defer func() {
		assert.NoError(t, mount.Unlink(fname))
	}()

	t.Run("removeXattr", func(t *testing.T) {
		s := xattrSamples[0]
		err := mount.SetXattr(fname, s.name, s.value, XattrDefault)
		err = mount.RemoveXattr(fname, s.name)
		assert.NoError(t, err)
	})

	t.Run("removeMissingXattr", func(t *testing.T) {
		s := xattrSamples[1]
		err := mount.RemoveXattr(fname, s.name)
		assert.Error(t, err)
	})

	t.Run("emptyName", func(t *testing.T) {
		err := mount.RemoveXattr(fname, "")
		assert.Error(t, err)
	})

	t.Run("invalidMount", func(t *testing.T) {
		m := &MountInfo{}
		err := m.RemoveXattr(fname, xattrSamples[0].name)
		assert.Error(t, err)
	})
}

func TestGetSetXattrLinkPath(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)
	fname := "TestGetSetXattrLinkPath.txt"
	lname := "TestGetSetXattrLinkPath.lnk"

	f1, err := mount.Open(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	require.NoError(t, err)
	assert.NoError(t, f1.Close())
	err = mount.Symlink(fname, lname)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, mount.Unlink(fname))
		assert.NoError(t, mount.Unlink(lname))
	}()

	for _, s := range xattrSamples {
		t.Run("roundTrip-"+s.name, func(t *testing.T) {
			err := mount.LsetXattr(lname, s.name, s.value, XattrDefault)
			assert.NoError(t, err)
			b, err := mount.LgetXattr(lname, s.name)
			assert.NoError(t, err)
			assert.EqualValues(t, s.value, b)
		})
	}

	t.Run("linkVsFile", func(t *testing.T) {
		s := xattrSamples[0]
		err := mount.LsetXattr(lname, s.name, s.value, XattrDefault)
		assert.NoError(t, err)

		// not on the file
		err = mount.LremoveXattr(fname, s.name)
		assert.Error(t, err)
		// on the link
		err = mount.LremoveXattr(lname, s.name)
		assert.NoError(t, err)
	})

	t.Run("missingXattrOnGet", func(t *testing.T) {
		_, err := mount.LgetXattr(lname, "user.never-set")
		assert.Error(t, err)
	})

	t.Run("emptyNameGet", func(t *testing.T) {
		_, err := mount.LgetXattr(lname, "")
		assert.Error(t, err)
	})

	t.Run("emptyNameSet", func(t *testing.T) {
		err := mount.LsetXattr(lname, "", []byte("foo"), XattrDefault)
		assert.Error(t, err)
	})

	t.Run("invalidMount", func(t *testing.T) {
		m := &MountInfo{}
		err := m.LsetXattr(lname, xattrSamples[0].name, xattrSamples[0].value, XattrDefault)
		assert.Error(t, err)
		_, err = m.LgetXattr(lname, xattrSamples[0].name)
		assert.Error(t, err)
	})
}

func TestListXattrLinkPath(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)
	fname := "TestListXattrLinkPath.txt"
	lname := "TestListXattrLinkPath.lnk"

	f1, err := mount.Open(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	require.NoError(t, err)
	assert.NoError(t, f1.Close())
	err = mount.Symlink(fname, lname)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, mount.Unlink(fname))
		assert.NoError(t, mount.Unlink(lname))
	}()

	t.Run("listXattrs1", func(t *testing.T) {
		for _, s := range xattrSamples[:1] {
			err := mount.LsetXattr(lname, s.name, s.value, XattrDefault)
			assert.NoError(t, err)
		}

		// not on the file
		xl, err := mount.LlistXattr(fname)
		assert.NoError(t, err)
		assert.Len(t, xl, 0)
		// on the link
		xl, err = mount.LlistXattr(lname)
		assert.NoError(t, err)
		assert.Len(t, xl, 1)
		assert.Contains(t, xl, xattrSamples[0].name)
	})

	t.Run("listXattrs2", func(t *testing.T) {
		for _, s := range xattrSamples {
			err := mount.LsetXattr(lname, s.name, s.value, XattrDefault)
			assert.NoError(t, err)
		}
		xl, err := mount.LlistXattr(lname)
		assert.NoError(t, err)
		assert.Len(t, xl, 4)
		assert.Contains(t, xl, xattrSamples[0].name)
		assert.Contains(t, xl, xattrSamples[1].name)
		assert.Contains(t, xl, xattrSamples[2].name)
		assert.Contains(t, xl, xattrSamples[3].name)
	})

	t.Run("invalidMount", func(t *testing.T) {
		m := &MountInfo{}
		_, err := m.LlistXattr(lname)
		assert.Error(t, err)
	})
}

func TestRemoveXattrLinkPath(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)
	fname := "TestRemoveXattrLinkPath.txt"
	lname := "TestRemoveXattrLinkPath.lnk"

	f1, err := mount.Open(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	require.NoError(t, err)
	assert.NoError(t, f1.Close())
	err = mount.Symlink(fname, lname)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, mount.Unlink(fname))
		assert.NoError(t, mount.Unlink(lname))
	}()

	t.Run("removeXattr", func(t *testing.T) {
		s := xattrSamples[0]
		err := mount.LsetXattr(lname, s.name, s.value, XattrDefault)
		assert.NoError(t, err)

		// not on the file
		err = mount.LremoveXattr(fname, s.name)
		assert.Error(t, err)
		// on the link
		err = mount.LremoveXattr(lname, s.name)
		assert.NoError(t, err)
	})

	t.Run("removeMissingXattr", func(t *testing.T) {
		s := xattrSamples[1]
		err := mount.LremoveXattr(lname, s.name)
		assert.Error(t, err)
	})

	t.Run("emptyName", func(t *testing.T) {
		err := mount.LremoveXattr(lname, "")
		assert.Error(t, err)
	})

	t.Run("invalidMount", func(t *testing.T) {
		m := &MountInfo{}
		err := m.LremoveXattr(lname, xattrSamples[0].name)
		assert.Error(t, err)
	})
}
