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
