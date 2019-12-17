// +build !luminous,!mimic
//
// Ceph Nautilus introduced rbd_get_parent() and deprecated rbd_get_parent_info().
// Ceph Nautilus introduced rbd_list_children3() and deprecated rbd_list_children().

package rbd

// #cgo LDFLAGS: -lrbd
// #include <rbd/librbd.h>
// #include <errno.h>
import "C"

import (
	"fmt"
	"unsafe"
)

// GetParentInfo looks for the parent of the image and stores the pool, name
// and snapshot-name in the byte-arrays that are passed as arguments.
//
// Implements:
//   int rbd_get_parent(rbd_image_t image,
//                      rbd_linked_image_spec_t *parent_image,
//                      rbd_snap_spec_t *parent_snap)
func (image *Image) GetParentInfo(pool, name, snapname []byte) error {
	if err := image.validate(imageIsOpen); err != nil {
		return err
	}

	parentImage := C.rbd_linked_image_spec_t{}
	parentSnap := C.rbd_snap_spec_t{}
	ret := C.rbd_get_parent(image.image, &parentImage, &parentSnap)
	if ret != 0 {
		return RBDError(ret)
	}

	defer C.rbd_linked_image_spec_cleanup(&parentImage)
	defer C.rbd_snap_spec_cleanup(&parentSnap)

	strlen := int(C.strlen(parentImage.pool_name))
	if len(pool) < strlen {
		return RBDError(C.ERANGE)
	}
	if copy(pool, C.GoString(parentImage.pool_name)) != strlen {
		return RBDError(C.ERANGE)
	}

	strlen = int(C.strlen(parentImage.image_name))
	if len(name) < strlen {
		return RBDError(C.ERANGE)
	}
	if copy(name, C.GoString(parentImage.image_name)) != strlen {
		return RBDError(C.ERANGE)
	}

	strlen = int(C.strlen(parentSnap.name))
	if len(snapname) < strlen {
		return RBDError(C.ERANGE)
	}
	if copy(snapname, C.GoString(parentSnap.name)) != strlen {
		return RBDError(C.ERANGE)
	}

	return nil
}
