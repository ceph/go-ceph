//go:build !nautilus && ceph_preview
// +build !nautilus,ceph_preview

package cephfs

/*
#cgo LDFLAGS: -lcephfs
#cgo CPPFLAGS: -D_FILE_OFFSET_BITS=64
#include <stdlib.h>
#include <cephfs/libcephfs.h>
*/
import "C"

import (
	"unsafe"
)

// Mknod creates a regular, block or character special file.
//
// Implements:
//
//	int ceph_mknod(struct ceph_mount_info *cmount, const char *path, mode_t mode,
//				   dev_t rdev);
func (mount *MountInfo) Mknod(path string, mode uint16, dev uint16) error {
	if err := mount.validate(); err != nil {
		return err
	}

	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	ret := C.ceph_mknod(mount.mount, cPath, C.mode_t(mode), C.dev_t(dev))
	return getError(ret)
}

// Utime struct is the equivalent of C.struct_utimbuf
type Utime struct {
	// AcTime  represents the file's access time in seconds since the Unix epoch.
	AcTime int64
	// ModTime represents the file's modification time in seconds since the Unix epoch.
	ModTime int64
}

// Futime changes file/directory last access and modification times.
//
// Implements:
//
//	int ceph_futime(struct ceph_mount_info *cmount, int fd, struct utimbuf *buf);
func (mount *MountInfo) Futime(fd int, times *Utime) error {
	if err := mount.validate(); err != nil {
		return err
	}

	cFd := C.int(fd)
	uTimeBuf := &C.struct_utimbuf{
		actime:  C.time_t(times.AcTime),
		modtime: C.time_t(times.ModTime),
	}

	ret := C.ceph_futime(mount.mount, cFd, uTimeBuf)
	return getError(ret)
}
