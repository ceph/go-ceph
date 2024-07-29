//go:build ceph_preview

package striper

// #cgo LDFLAGS: -lrados -lradosstriper
// #include <errno.h>
// #include <stdlib.h>
// #include <radosstriper/libradosstriper.h>
import "C"

import (
	"github.com/ceph/go-ceph/rados"
)

// Striper helps manage the reading, writing, and management of RADOS
// striped objects.
type Striper struct {
	striper C.rados_striper_t

	// Hold a reference back to the ioctx that the striper depends on so
	// that Go doesn't garbage collect it prematurely.
	ioctx *rados.IOContext
}

// Layout contains a group of values used to define the size parameters of
// striped objects. Note that these parameters only effect new striped objects.
// Existing striped objects retain the parameters they were created with.
type Layout struct {
	StripeUnit  uint
	StripeCount uint
	ObjectSize  uint
}

// New returns a rados Striper object created from a rados IOContext.
func New(ioctx *rados.IOContext) (*Striper, error) {
	var s C.rados_striper_t
	ret := C.rados_striper_create(cephIoctx(ioctx), &s)
	if err := getError(ret); err != nil {
		return nil, err
	}
	return &Striper{s, ioctx}, nil
}

// NewWithLayout returns a rados Striper object created from a rados IOContext
// and striper layout parameters. These parameters will be used when new
// objects are created.
func NewWithLayout(ioctx *rados.IOContext, layout Layout) (*Striper, error) {
	striper, err := New(ioctx)
	if err != nil {
		return nil, err
	}
	if err := striper.SetObjectLayoutStripeUnit(layout.StripeUnit); err != nil {
		return nil, err
	}
	if err := striper.SetObjectLayoutStripeCount(layout.StripeCount); err != nil {
		return nil, err
	}
	if err := striper.SetObjectLayoutObjectSize(layout.ObjectSize); err != nil {
		return nil, err
	}
	return striper, nil
}

// Destroy the radosstriper object at the Ceph API level.
func (s *Striper) Destroy() {
	C.rados_striper_destroy(s.striper)
}

// SetObjectLayoutStripeUnit sets the stripe unit value used to layout
// new objects.
//
// Implements:
//
//	int rados_striper_set_object_layout_stripe_unit(rados_striper_t striper,
//	                                                unsigned int stripe_unit);
func (s *Striper) SetObjectLayoutStripeUnit(count uint) error {
	ret := C.rados_striper_set_object_layout_stripe_unit(
		s.striper,
		C.uint(count),
	)
	return getError(ret)
}

// SetObjectLayoutStripeCount sets the stripe count value used to layout
// new objects.
//
// Implements:
//
//	int rados_striper_set_object_layout_stripe_count(rados_striper_t striper,
//	                                                 unsigned int stripe_count);
func (s *Striper) SetObjectLayoutStripeCount(count uint) error {
	ret := C.rados_striper_set_object_layout_stripe_count(
		s.striper,
		C.uint(count),
	)
	return getError(ret)
}

// SetObjectLayoutObjectSize sets the object size value used to layout
// new objects.
//
// Implements:
//
//	int rados_striper_set_object_layout_object_size(rados_striper_t striper,
//	                                                unsigned int object_size);
func (s *Striper) SetObjectLayoutObjectSize(count uint) error {
	ret := C.rados_striper_set_object_layout_object_size(
		s.striper,
		C.uint(count),
	)
	return getError(ret)
}

// cephIoctx returns a ceph rados_ioctx_t given a go-ceph rados IOContext.
func cephIoctx(radosIoctx *rados.IOContext) C.rados_ioctx_t {
	p := radosIoctx.Pointer()
	if p == nil {
		panic("invalid IOContext pointer")
	}
	return C.rados_ioctx_t(p)
}
