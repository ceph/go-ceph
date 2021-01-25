package rbd

/*
#cgo LDFLAGS: -lrbd
#include <stdlib.h>
#include <rbd/librbd.h>
*/
import "C"

import (
	"unsafe"

	"github.com/ceph/go-ceph/rados"
)

// GroupSnapCreate will create a group snapshot.
//
// Implements:
//  int rbd_group_snap_create(rados_ioctx_t group_p,
//                            const char *group_name,
//                            const char *snap_name);
func GroupSnapCreate(ioctx *rados.IOContext, group, snap string) error {
	cGroupName := C.CString(group)
	defer C.free(unsafe.Pointer(cGroupName))
	cSnapName := C.CString(snap)
	defer C.free(unsafe.Pointer(cSnapName))

	ret := C.rbd_group_snap_create(cephIoctx(ioctx), cGroupName, cSnapName)
	return getError(ret)
}

// GroupSnapRemove removes an existing group snapshot.
//
// Implements:
//  int rbd_group_snap_remove(rados_ioctx_t group_p,
//                            const char *group_name,
//                            const char *snap_name);
func GroupSnapRemove(ioctx *rados.IOContext, group, snap string) error {
	cGroupName := C.CString(group)
	defer C.free(unsafe.Pointer(cGroupName))
	cSnapName := C.CString(snap)
	defer C.free(unsafe.Pointer(cSnapName))

	ret := C.rbd_group_snap_remove(cephIoctx(ioctx), cGroupName, cSnapName)
	return getError(ret)
}

// GroupSnapRename will rename an existing group snapshot.
//
// Implements:
//  int rbd_group_snap_rename(rados_ioctx_t group_p,
//                            const char *group_name,
//                            const char *old_snap_name,
//                            const char *new_snap_name);
func GroupSnapRename(ioctx *rados.IOContext, group, src, dest string) error {
	cGroupName := C.CString(group)
	defer C.free(unsafe.Pointer(cGroupName))
	cOldSnapName := C.CString(src)
	defer C.free(unsafe.Pointer(cOldSnapName))
	cNewSnapName := C.CString(dest)
	defer C.free(unsafe.Pointer(cNewSnapName))

	ret := C.rbd_group_snap_rename(
		cephIoctx(ioctx), cGroupName, cOldSnapName, cNewSnapName)
	return getError(ret)
}
