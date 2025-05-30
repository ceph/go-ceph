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

	"github.com/ceph/go-ceph/internal/util"
)

var (
	serVersion string
)

func init() {
	serVersion = util.CurrentCephVersionString()
}

// Chmod changes the mode bits (permissions) of a file/directory.
func (mount *MountInfo) Chmod(path string, mode uint32) error {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	ret := C.ceph_chmod(mount.mount, cPath, C.mode_t(mode))
	return getError(ret)
}

// Chown changes the ownership of a file/directory.
func (mount *MountInfo) Chown(path string, user uint32, group uint32) error {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	var ret C.int

	if util.CurrentCephVersion() > util.CephTentacle {
		ret = C.ceph_chown(mount.mount, cPath, C.uid_t(user), C.gid_t(group))
	} else {
		ret = C.ceph_chown(mount.mount, cPath, C.int(user), C.int(group))
	}
	return getError(ret)
}

// Lchown changes the ownership of a file/directory/etc without following symbolic links
func (mount *MountInfo) Lchown(path string, user uint32, group uint32) error {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	if util.CurrentCephVersion() > util.CephTentacle {
		ret = C.ceph_lchown(mount.mount, cPath, C.uid_t(user), C.gid_t(group))
	} else {
		ret = C.ceph_lchown(mount.mount, cPath, C.int(user), C.int(group))
	}

	return getError(ret)
}
