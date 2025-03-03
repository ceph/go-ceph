//go:build ceph_preview

package rbd

// #cgo LDFLAGS: -lrbd
// #include <rbd/librbd.h>
// #include <errno.h>
import "C"
import (
	"unsafe"

	"github.com/ceph/go-ceph/internal/retry"
)

// ListChildrenAttributes returns an array of struct with the names and ids of the
// images and pools and the trash of the images that are children of the given image.
//
// Implements:
//
//	int rbd_list_children3(rbd_image_t image, rbd_linked_image_spec_t *images,
//	                       size_t *max_images);
func (image *Image) ListChildrenAttributes() (imgSpec []ImageSpec, err error) {
	if err := image.validate(imageIsOpen); err != nil {
		return nil, err
	}
	var (
		csize    C.size_t
		children []C.rbd_linked_image_spec_t
	)
	retry.WithSizes(16, 4096, func(size int) retry.Hint {
		csize = C.size_t(size)
		children = make([]C.rbd_linked_image_spec_t, csize)
		ret := C.rbd_list_children3(
			image.image,
			(*C.rbd_linked_image_spec_t)(unsafe.Pointer(&children[0])),
			&csize)
		err = getErrorIfNegative(ret)
		return retry.Size(int(csize)).If(err == errRange)
	})
	if err != nil {
		return nil, err
	}
	defer C.rbd_linked_image_spec_list_cleanup((*C.rbd_linked_image_spec_t)(unsafe.Pointer(&children[0])), csize)

	imgSpec = make([]ImageSpec, csize)
	for i, child := range children[:csize] {
		imgSpec[i] = ImageSpec{
			ImageName:     C.GoString(child.image_name),
			ImageID:       C.GoString(child.image_id),
			PoolName:      C.GoString(child.pool_name),
			PoolNamespace: C.GoString(child.pool_namespace),
			PoolID:        uint64(child.pool_id),
			Trash:         bool(child.trash),
		}
	}
	return imgSpec, nil
}
