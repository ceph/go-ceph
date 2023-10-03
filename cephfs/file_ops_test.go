//go:build !nautilus
// +build !nautilus

package cephfs

import (
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMknod(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)

	file1 := "/file1"
	mode1 := uint16(syscall.S_IFIFO | syscall.S_IRUSR | syscall.S_IWUSR)
	err := mount.Mknod(file1, mode1, 0)
	assert.NoError(t, err)

	file2 := "/file2"
	mode2 := uint16(syscall.S_IFCHR)
	err = mount.Mknod(file2, mode2, 89)
	assert.NoError(t, err)

	file3 := "/file3"
	mode3 := uint16(syscall.S_IFBLK)
	err = mount.Mknod(file3, mode3, 129)
	assert.NoError(t, err)

	defer func() {
		assert.NoError(t, mount.Unlink(file1))
		assert.NoError(t, mount.Unlink(file2))
		assert.NoError(t, mount.Unlink(file3))
	}()

	sx, err := mount.Statx(file1, StatxBasicStats, 0)
	assert.Equal(t, mode1, sx.Mode&mode1)

	sx, err = mount.Statx(file2, StatxBasicStats, 0)
	assert.Equal(t, mode2, sx.Mode&mode2)
	assert.Equal(t, uint64(89), sx.Rdev)

	sx, err = mount.Statx(file3, StatxBasicStats, 0)
	assert.Equal(t, mode3, sx.Mode&mode3)
	assert.Equal(t, uint64(129), sx.Rdev)

	// Test invalid mount value
	mount1 := &MountInfo{}
	file4 := "/file4"
	err = mount1.Mknod(file4, uint16(syscall.S_IFCHR), 64)
	assert.Error(t, err)
}

func TestFutime(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)

	fname := "futime_file.txt"
	f1, err := mount.Open(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	assert.NoError(t, err)
	assert.NotNil(t, f1)
	defer func() {
		assert.NoError(t, f1.Close())
		assert.NoError(t, mount.Unlink(fname))
	}()

	currentTime := Timespec{int64(time.Now().Second()), 0}
	newTime := &Utime{
		AcTime:  currentTime.Sec,
		ModTime: currentTime.Sec,
	}
	err = mount.Futime(int(f1.fd), newTime)
	assert.NoError(t, err)

	sx, err := mount.Statx(fname, StatxBasicStats, 0)
	assert.NoError(t, err)
	assert.Equal(t, currentTime, sx.Atime)
	assert.Equal(t, currentTime, sx.Mtime)

	// Test invalid mount value
	mount1 := &MountInfo{}
	currentTime = Timespec{int64(time.Now().Second()), 0}
	newTime = &Utime{
		AcTime:  currentTime.Sec,
		ModTime: currentTime.Sec,
	}
	err = mount1.Futime(int(f1.fd), newTime)
	assert.Error(t, err)
}

func TestFutimens(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)

	fname := "futimens_file.txt"
	f1, err := mount.Open(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	assert.NoError(t, err)
	assert.NotNil(t, f1)
	defer func() {
		assert.NoError(t, f1.Close())
		assert.NoError(t, mount.Unlink(fname))
	}()

	times := []Timespec{
		{int64(time.Now().Second()), 0},
		{int64(time.Now().Second()), 0},
	}
	err = mount.Futimens(int(f1.fd), times)
	assert.NoError(t, err)

	sx, err := mount.Statx(fname, StatxBasicStats, 0)
	assert.NoError(t, err)
	assert.Equal(t, times[0], sx.Atime)
	assert.Equal(t, times[1], sx.Mtime)

	// Test invalid mount value
	mount1 := &MountInfo{}
	times = []Timespec{
		{int64(time.Now().Second()), 0},
		{int64(time.Now().Second()), 0},
	}
	err = mount1.Futimens(int(f1.fd), times)
	assert.Error(t, err)

	// Test times array length more than 2
	times = []Timespec{
		{int64(time.Now().Second()), 0},
		{int64(time.Now().Second()), 0},
		{int64(time.Now().Second()), 0},
	}
	err = mount.Futimens(int(f1.fd), times)
	assert.Error(t, err)
}

func TestFutimes(t *testing.T) {
	mount := fsConnect(t)
	defer fsDisconnect(t, mount)

	fname := "futimes_file.txt"
	f1, err := mount.Open(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	assert.NoError(t, err)
	assert.NotNil(t, f1)
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
	err = mount.Futimes(int(f1.fd), newTimes)
	assert.NoError(t, err)

	sx, err := mount.Statx(fname, StatxBasicStats, 0)
	assert.NoError(t, err)
	assert.Equal(t, times[0], sx.Atime)
	assert.Equal(t, times[1], sx.Mtime)

	// Test invalid mount value
	mount1 := &MountInfo{}
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
	err = mount1.Futimes(int(f1.fd), newTimes)
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
	err = mount.Futimes(int(f1.fd), newTimes)
	assert.Error(t, err)
}
