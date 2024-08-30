//go:build ceph_preview && !(nautilus || octopus || pacific || quincy || reef || squid)

package rbd

/*
#cgo LDFLAGS: -lrbd
#include <errno.h>
#include <stdlib.h>
#include <rbd/librbd.h>

*/
import "C"

import (
	"unsafe"

	"github.com/ceph/go-ceph/internal/cutil"
	"github.com/ceph/go-ceph/rados"
)

type imgSnapInfoArray [cutil.MaxIdx]C.rbd_group_image_snap_info_t

// GroupSnapGetInfo returns a slice of RBD image snapshots that are part of a
// group snapshot.
//
// Implements:
//
//	int rbd_group_snap_get_info(rados_ioctx_t group_p,
//	                        const char *group_name,
//	                        const char *snap_name,
//	                        rbd_group_snap_info2_t *snaps);
func GroupSnapGetInfo(ioctx *rados.IOContext, group, snap string) (GroupSnapInfo, error) {
	cGroupName := C.CString(group)
	defer C.free(unsafe.Pointer(cGroupName))
	cSnapName := C.CString(snap)
	defer C.free(unsafe.Pointer(cSnapName))

	cSnapInfo := C.rbd_group_snap_info2_t{}

	ret := C.rbd_group_snap_get_info(
		cephIoctx(ioctx),
		cGroupName,
		cSnapName,
		&cSnapInfo)
	err := getErrorIfNegative(ret)
	if err != nil {
		return GroupSnapInfo{}, err
	}

	snapCount := uint64(cSnapInfo.image_snaps_count)

	snapInfo := GroupSnapInfo{
		ID:        C.GoString(cSnapInfo.id),
		Name:      C.GoString(cSnapInfo.name),
		SnapName:  C.GoString(cSnapInfo.image_snap_name),
		State:     GroupSnapState(cSnapInfo.state),
		Snapshots: make([]GroupSnap, snapCount),
	}

	imgSnaps := (*imgSnapInfoArray)(unsafe.Pointer(cSnapInfo.image_snaps))[0:snapCount]

	for i, imgSnap := range imgSnaps {
		snapInfo.Snapshots[i].Name = C.GoString(imgSnap.image_name)
		snapInfo.Snapshots[i].PoolID = uint64(imgSnap.pool_id)
		snapInfo.Snapshots[i].SnapID = uint64(imgSnap.snap_id)
	}

	// free C memory allocated by C.rbd_group_snap_get_info call
	C.rbd_group_snap_get_info_cleanup(&cSnapInfo)
	return snapInfo, nil
}
