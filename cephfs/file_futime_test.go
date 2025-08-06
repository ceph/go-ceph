//go:build ceph_preview

package cephfs

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFileFutime(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)

	fname := "futime_file.txt"
	f1, err := mount.Open(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	assert.NoError(t, err)
	assert.NotNil(t, f1)
	assert.NotEqual(t, -1, f1.Fd())
	defer func() {
		assert.NoError(t, f1.Close())
		assert.NoError(t, mount.Unlink(fname))
	}()

	currentTime := Timespec{int64(time.Now().Second()), 0}
	newTime := &Utime{
		AcTime:  currentTime.Sec,
		ModTime: currentTime.Sec,
	}
	err = f1.Futime(newTime)
	assert.NoError(t, err)

	sx, err := mount.Statx(fname, StatxBasicStats, 0)
	assert.NoError(t, err)
	assert.Equal(t, currentTime, sx.Atime)
	assert.Equal(t, currentTime, sx.Mtime)

	// Test invalid file object
	f2 := &File{}
	currentTime = Timespec{int64(time.Now().Second()), 0}
	newTime = &Utime{
		AcTime:  currentTime.Sec,
		ModTime: currentTime.Sec,
	}
	err = f2.Futime(newTime)
	assert.Error(t, err)
}

func TestFileFutimens(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)

	fname := "futimens_file.txt"
	f1, err := mount.Open(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	assert.NoError(t, err)
	assert.NotNil(t, f1)
	assert.NotEqual(t, -1, f1.Fd())
	defer func() {
		assert.NoError(t, f1.Close())
		assert.NoError(t, mount.Unlink(fname))
	}()

	times := []Timespec{
		{int64(time.Now().Second()), 0},
		{int64(time.Now().Second()), 0},
	}
	err = f1.Futimens(times)
	assert.NoError(t, err)

	sx, err := mount.Statx(fname, StatxBasicStats, 0)
	assert.NoError(t, err)
	assert.Equal(t, times[0], sx.Atime)
	assert.Equal(t, times[1], sx.Mtime)

	// Test invalid file object
	f2 := &File{}
	times = []Timespec{
		{int64(time.Now().Second()), 0},
		{int64(time.Now().Second()), 0},
	}
	err = f2.Futimens(times)
	assert.Error(t, err)

	// Test times array length more than 2
	times = []Timespec{
		{int64(time.Now().Second()), 0},
		{int64(time.Now().Second()), 0},
		{int64(time.Now().Second()), 0},
	}
	err = f1.Futimens(times)
	assert.Error(t, err)
}

func TestFileFutimes(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)

	fname := "futimes_file.txt"
	f1, err := mount.Open(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	assert.NoError(t, err)
	assert.NotNil(t, f1)
	assert.NotEqual(t, -1, f1.Fd())
	defer func() {
		assert.NoError(t, f1.Close())
		assert.NoError(t, mount.Unlink(fname))
	}()

	times := []Timespec{
		{int64(time.Now().Second()), 0},
		{int64(time.Now().Second()), 0},
	}
	newTimes := []Timeval{}
	for _, val := range times {
		newTimes = append(newTimes, Timeval{
			Sec:  val.Sec,
			USec: int64(val.Nsec / 1000),
		})
	}
	err = f1.Futimes(newTimes)
	assert.NoError(t, err)

	sx, err := mount.Statx(fname, StatxBasicStats, 0)
	assert.NoError(t, err)
	assert.Equal(t, times[0], sx.Atime)
	assert.Equal(t, times[1], sx.Mtime)

	// Test invalid file object
	f2 := &File{}
	times = []Timespec{
		{int64(time.Now().Second()), 0},
		{int64(time.Now().Second()), 0},
	}
	newTimes = []Timeval{}
	for _, val := range times {
		newTimes = append(newTimes, Timeval{
			Sec:  val.Sec,
			USec: int64(val.Nsec / 1000),
		})
	}
	err = f2.Futimes(newTimes)
	assert.Error(t, err)

	// Test times array length more than 2
	times = []Timespec{
		{int64(time.Now().Second()), 0},
		{int64(time.Now().Second()), 0},
		{int64(time.Now().Second()), 0},
	}
	newTimes = []Timeval{}
	for _, val := range times {
		newTimes = append(newTimes, Timeval{
			Sec:  val.Sec,
			USec: int64(val.Nsec / 1000),
		})
	}
	err = f1.Futimes(newTimes)
	assert.Error(t, err)
}
