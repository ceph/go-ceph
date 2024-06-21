//go:build !(nautilus || octopus || pacific || quincy || reef) && ceph_preview

package rbd

// #cgo LDFLAGS: -lrbd
// #include <errno.h>
// #include <stdlib.h>
// #include <rados/librados.h>
// #include <rbd/librbd.h>
import "C"

import (
	"unsafe"

	"github.com/ceph/go-ceph/rados"
)

// CloneImageByID creates a clone of the image from a snapshot with the given
// ID in the provided io-context with the given name and image options.
//
// Implements:
//
//	int rbd_clone4(rados_ioctx_t p_ioctx, const char *p_name,
//	               uint64_t p_snap_id, rados_ioctx_t c_ioctx,
//	               const char *c_name, rbd_image_options_t c_opts);
func CloneImageByID(ioctx *rados.IOContext, parentName string, snapID uint64,
	destctx *rados.IOContext, name string, rio *ImageOptions) error {

	if rio == nil {
		return rbdError(C.EINVAL)
	}

	cParentName := C.CString(parentName)
	defer C.free(unsafe.Pointer(cParentName))
	cCloneName := C.CString(name)
	defer C.free(unsafe.Pointer(cCloneName))

	ret := C.rbd_clone4(
		cephIoctx(ioctx),
		cParentName,
		C.uint64_t(snapID),
		cephIoctx(destctx),
		cCloneName,
		C.rbd_image_options_t(rio.options))
	return getError(ret)
}
