//go:build !octopus && !pacific && !quincy && ceph_preview

package rbd

// #cgo LDFLAGS: -lrbd
// /* force XSI-complaint strerror_r() */
// #define _POSIX_C_SOURCE 200112L
// #undef _GNU_SOURCE
// #include <rbd/librbd.h>
import "C"

import (
	"unsafe"
)

// toEncryptionSpec returns a rbd_encryption_spec_t converted from the
// cEncryptionData type.
func (edata cEncryptionData) toEncryptionSpec() C.rbd_encryption_spec_t {
	var cSpec C.rbd_encryption_spec_t
	cSpec.format = edata.format
	cSpec.opts = edata.opts
	cSpec.opts_size = edata.optsSize
	return cSpec
}

// EncryptionLoad2 enables IO on an open encrypted image. The difference
// between EncryptionLoad and EncryptionLoad2 is that EncryptionLoad2 can open
// ancestor images with a different encryption options than the current image.
// The first EncryptionOptions in the slice is applied to the current image,
// the second to the first ancestor, the third to the second ancestor and so
// on.  If the length of the slice is smaller than the number of ancestors the
// final item in the slice will be applied to all remaining ancestors, or if
// the ancestor does not match the encryption format the ancestor will be
// interpreted as plain-text.
//
// Implements:
//
//	int rbd_encryption_load2(rbd_image_t image,
//	                         const rbd_encryption_spec_t *specs,
//	                         size_t spec_count);
func (image *Image) EncryptionLoad2(opts []EncryptionOptions) error {
	if image.image == nil {
		return ErrImageNotOpen
	}

	length := len(opts)
	eos := make([]cEncryptionData, length)
	cspecs := (*C.rbd_encryption_spec_t)(C.malloc(
		C.size_t(C.sizeof_rbd_encryption_spec_t * length)))
	specs := unsafe.Slice(cspecs, length)

	for idx, option := range opts {
		eos[idx] = option.allocateEncryptionOptions()
		specs[idx] = eos[idx].toEncryptionSpec()
	}
	defer func() {
		for _, eopt := range eos {
			eopt.free()
		}
	}()

	ret := C.rbd_encryption_load2(
		image.image,
		cspecs,
		C.size_t(length))
	return getError(ret)
}
