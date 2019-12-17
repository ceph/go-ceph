// +build luminous mimic
// +build !nautilus
//
// Ceph Nautilus introduced rbd_get_parent() and deprecated rbd_get_parent_info().
// Ceph Nautilus introduced rbd_list_children3() and deprecated rbd_list_children().

package rbd

// #cgo LDFLAGS: -lrbd
// #include <rbd/librbd.h>
// #include <errno.h>
import "C"

import (
	"bytes"
	"unsafe"
)

// GetParentInfo looks for the parent of the image and stores the pool, name
// and snapshot-name in the byte-arrays that are passed as arguments.
//
// Implements:
//   int rbd_get_parent_info(rbd_image_t image, char *parent_pool_name,
//                           size_t ppool_namelen, char *parent_name,
//                           size_t pnamelen, char *parent_snap_name,
//                           size_t psnap_namelen)
func (image *Image) GetParentInfo(p_pool, p_name, p_snapname []byte) error {
	if err := image.validate(imageIsOpen); err != nil {
		return err
	}

	ret := C.rbd_get_parent_info(
		image.image,
		(*C.char)(unsafe.Pointer(&p_pool[0])),
		(C.size_t)(len(p_pool)),
		(*C.char)(unsafe.Pointer(&p_name[0])),
		(C.size_t)(len(p_name)),
		(*C.char)(unsafe.Pointer(&p_snapname[0])),
		(C.size_t)(len(p_snapname)))
	if ret == 0 {
		return nil
	} else {
		return RBDError(ret)
	}
}
