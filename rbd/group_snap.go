package rbd

/*
#cgo LDFLAGS: -lrbd
#include <stdlib.h>
#include <rbd/librbd.h>
*/
import "C"

import (
	"unsafe"

	"github.com/ceph/go-ceph/internal/retry"
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

// GroupSnapState represents the state of a group snapshot in GroupSnapInfo.
type GroupSnapState int

const (
	// GroupSnapStateIncomplete is equivalent to RBD_GROUP_SNAP_STATE_INCOMPLETE.
	GroupSnapStateIncomplete = GroupSnapState(C.RBD_GROUP_SNAP_STATE_INCOMPLETE)
	// GroupSnapStateComplete is equivalent to RBD_GROUP_SNAP_STATE_COMPLETE.
	GroupSnapStateComplete = GroupSnapState(C.RBD_GROUP_SNAP_STATE_COMPLETE)
)

// GroupSnapInfo values are returned by GroupSnapList, representing the
// snapshots that are part of an rbd group.
type GroupSnapInfo struct {
	Name  string
	State GroupSnapState
}

// GroupSnapList returns a slice of snapshots in a group.
//
// Implements:
//  int rbd_group_snap_list(rados_ioctx_t group_p,
//                          const char *group_name,
//                          rbd_group_snap_info_t *snaps,
//                          size_t group_snap_info_size,
//                          size_t *num_entries);
func GroupSnapList(ioctx *rados.IOContext, group string) ([]GroupSnapInfo, error) {
	cGroupName := C.CString(group)
	defer C.free(unsafe.Pointer(cGroupName))

	var (
		cSnaps []C.rbd_group_snap_info_t
		cSize  C.size_t
		err    error
	)
	retry.WithSizes(1024, 262144, func(size int) retry.Hint {
		cSize = C.size_t(size)
		cSnaps = make([]C.rbd_group_snap_info_t, cSize)
		ret := C.rbd_group_snap_list(
			cephIoctx(ioctx),
			cGroupName,
			(*C.rbd_group_snap_info_t)(unsafe.Pointer(&cSnaps[0])),
			C.sizeof_rbd_group_snap_info_t,
			&cSize)
		err = getErrorIfNegative(ret)
		return retry.Size(int(cSize)).If(err == errRange)
	})

	if err != nil {
		return nil, err
	}

	snaps := make([]GroupSnapInfo, cSize)
	for i := range snaps {
		snaps[i].Name = C.GoString(cSnaps[i].name)
		snaps[i].State = GroupSnapState(cSnaps[i].state)
	}

	// free C memory allocated by C.rbd_group_snap_list call
	ret := C.rbd_group_snap_list_cleanup(
		(*C.rbd_group_snap_info_t)(unsafe.Pointer(&cSnaps[0])),
		C.sizeof_rbd_group_snap_info_t,
		cSize)
	return snaps, getError(ret)
}

// GroupSnapRollback will roll back the images in the group to that of the
// given snapshot.
//
// Implements:
//  int rbd_group_snap_rollback(rados_ioctx_t group_p,
//                              const char *group_name,
//                              const char *snap_name);
func GroupSnapRollback(ioctx *rados.IOContext, group, snap string) error {
	cGroupName := C.CString(group)
	defer C.free(unsafe.Pointer(cGroupName))
	cSnapName := C.CString(snap)
	defer C.free(unsafe.Pointer(cSnapName))

	ret := C.rbd_group_snap_rollback(cephIoctx(ioctx), cGroupName, cSnapName)
	return getError(ret)
}
