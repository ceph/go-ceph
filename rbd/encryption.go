// +build !octopus,!nautilus

package rbd

// #cgo LDFLAGS: -lrbd
// /* force XSI-complaint strerror_r() */
// #define _POSIX_C_SOURCE 200112L
// #undef _GNU_SOURCE
// #include <errno.h>
// #include <stdlib.h>
// #include <rados/librados.h>
// #include <rbd/librbd.h>
import "C"

import (
	"unsafe"
)

// cData contains the data needed by the encryption functions
type cData struct {
	format   C.rbd_encryption_format_t
	opts     C.rbd_encryption_options_t
	optsSize C.size_t
	free     func()
}

// EncryptionAlgorithm is the encryption algorithm
type EncryptionAlgorithm C.rbd_encryption_algorithm_t

// Possible values for EncryptionAlgorithm:
// EncryptionAlgorithmAES128: AES 128bits
// EncryptionAlgorithmAES256: AES 256bits
const (
	EncryptionAlgorithmAES128 = EncryptionAlgorithm(C.RBD_ENCRYPTION_ALGORITHM_AES128)
	EncryptionAlgorithmAES256 = EncryptionAlgorithm(C.RBD_ENCRYPTION_ALGORITHM_AES256)
)

// EncryptionOptionsLUKS1 and EncryptionOptionsLUKS2 are identical
// structures at the moment, just as they are in the librbd api.
// The purpose behind creating different identical structures, is to facilitate
// future modifications of one of the formats, while maintaining backwards
// compatibility with the other.

// EncryptionOptionsLUKS1 options required for LUKS v1
type EncryptionOptionsLUKS1 struct {
	Alg        EncryptionAlgorithm
	Passphrase []byte
}

// EncryptionOptionsLUKS2 options required for LUKS v2
type EncryptionOptionsLUKS2 struct {
	Alg        EncryptionAlgorithm
	Passphrase []byte
}

// EncryptionOptions interface is used to encapsulate the different encryption
// formats options and enable converting them from go to C structures.
type EncryptionOptions interface {
	allocateEncryptionOptions() cData
}

func (opts EncryptionOptionsLUKS1) allocateEncryptionOptions() cData {
	var cOpts C.rbd_encryption_luks1_format_options_t
	var retData cData
	cOpts.alg = C.rbd_encryption_algorithm_t(opts.Alg)
	//CBytes allocates memory which we'll free by calling cOptsFree()
	cOpts.passphrase = (*C.char)(C.CBytes(opts.Passphrase))
	cOpts.passphrase_size = C.size_t(len(opts.Passphrase))
	retData.opts = C.rbd_encryption_options_t(&cOpts)
	retData.optsSize = C.size_t(C.sizeof_rbd_encryption_luks1_format_options_t)
	retData.free = func() { C.free(unsafe.Pointer(cOpts.passphrase)) }
	retData.format = C.RBD_ENCRYPTION_FORMAT_LUKS1
	return retData
}

func (opts EncryptionOptionsLUKS2) allocateEncryptionOptions() cData {
	var cOpts C.rbd_encryption_luks2_format_options_t
	var retData cData
	cOpts.alg = C.rbd_encryption_algorithm_t(opts.Alg)
	//CBytes allocates memory which we'll free by calling cOptsFree()
	cOpts.passphrase = (*C.char)(C.CBytes(opts.Passphrase))
	cOpts.passphrase_size = C.size_t(len(opts.Passphrase))
	retData.opts = C.rbd_encryption_options_t(&cOpts)
	retData.optsSize = C.size_t(C.sizeof_rbd_encryption_luks2_format_options_t)
	retData.free = func() { C.free(unsafe.Pointer(cOpts.passphrase)) }
	retData.format = C.RBD_ENCRYPTION_FORMAT_LUKS2
	return retData
}

// EncryptionFormat creates an encryption format header
//
// Implements:
//  int rbd_encryption_format(rbd_image_t image,
//                            rbd_encryption_format_t format,
//                            rbd_encryption_options_t opts,
//                            size_t opts_size);
//
// To issue an IO against the image, you need to mount the image
// with libvirt/qemu using the LUKS format, or make a call to
// rbd_encryption_load().
func (image *Image) EncryptionFormat(opts EncryptionOptions) error {
	if image.image == nil {
		return ErrImageNotOpen
	}

	encryptionOpts := opts.allocateEncryptionOptions()
	defer encryptionOpts.free()

	ret := C.rbd_encryption_format(
		image.image,
		encryptionOpts.format,
		encryptionOpts.opts,
		encryptionOpts.optsSize)

	return getError(ret)
}
