// +build !luminous
//
// Ceph Mimic introduced rbd_snap_get_namespace_type().

package rbd

// #cgo LDFLAGS: -lrbd
// #include <rbd/librbd.h>
import "C"

// SnapNamespaceType indicates the namespace to which the snapshot belongs to.
type SnapNamespaceType C.rbd_snap_namespace_type_t

const (
	// SnapNamespaceTypeUser indicates that the snapshot belongs to user namespace.
	SnapNamespaceTypeUser = SnapNamespaceType(C.RBD_SNAP_NAMESPACE_TYPE_USER)

	// SnapNamespaceTypeGroup indicates that the snapshot belongs to group namespace.
	// Such snapshots will have associated group information.
	SnapNamespaceTypeGroup = SnapNamespaceType(C.RBD_SNAP_NAMESPACE_TYPE_GROUP)

	// SnapNamespaceTypeTrash indicates that the snapshot belongs to trash namespace.
	SnapNamespaceTypeTrash = SnapNamespaceType(C.RBD_SNAP_NAMESPACE_TYPE_TRASH)
)

// GetSnapNamespaceType gets the type of namespace to which the snapshot belongs to,
// returns error on failure.
//
// Implements:
//  int rbd_snap_get_namespace_type(rbd_image_t image, uint64_t snap_id, rbd_snap_namespace_type_t *namespace_type)
func (image *Image) GetSnapNamespaceType(snapID uint64) (SnapNamespaceType, error) {
	var nsType SnapNamespaceType

	if err := image.validate(imageIsOpen); err != nil {
		return nsType, err
	}

	ret := C.rbd_snap_get_namespace_type(image.image,
		C.uint64_t(snapID),
		(*C.rbd_snap_namespace_type_t)(&nsType))
	return nsType, getError(ret)
}
