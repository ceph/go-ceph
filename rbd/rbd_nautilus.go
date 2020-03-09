// +build !luminous,!mimic
//
// Ceph Nautilus is the first release that includes rbd_list2().

package rbd

// #cgo LDFLAGS: -lrbd
// #include <rados/librados.h>
// #include <rbd/librbd.h>
// #include <errno.h>
import "C"

import (
	"unsafe"

	"github.com/ceph/go-ceph/internal/retry"
	"github.com/ceph/go-ceph/rados"
)

// GetImageNames returns the list of current RBD images.
func GetImageNames(ioctx *rados.IOContext) ([]string, error) {
	var (
		err    error
		images []C.rbd_image_spec_t
		size   C.size_t
	)
	for sizer := retry.NewSizerEV(32, 4096, errRange); sizer.Continue(); {
		size = C.size_t(sizer.Size())
		images = make([]C.rbd_image_spec_t, size)
		ret := C.rbd_list2(
			cephIoctx(ioctx),
			(*C.rbd_image_spec_t)(unsafe.Pointer(&images[0])),
			&size)
		err = sizer.UpdateWants(getErrorIfNegative(ret), int(size))
	}
	if err != nil {
		return nil, err
	}
	defer C.rbd_image_spec_list_cleanup((*C.rbd_image_spec_t)(unsafe.Pointer(&images[0])), size)

	names := make([]string, size)
	for i, image := range images[:size] {
		names[i] = C.GoString(image.name)
	}
	return names, nil
}
