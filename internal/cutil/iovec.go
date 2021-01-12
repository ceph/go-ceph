package cutil

/*
#include <stdlib.h>
#include <string.h>
#include <sys/uio.h>
*/
import "C"
import (
	"unsafe"
)

// Iovec is a slice of iovec structs. Might have allocated C memory, so it must
// be freed with the Free() method.
type Iovec struct {
	iovec []C.struct_iovec
	data  [][]byte
}

// ByteSlicesToIovec creates an Iovec and links it to Go buffers in data.
func ByteSlicesToIovec(data [][]byte) (v Iovec) {
	v.iovec = make([]C.struct_iovec, len(data))
	v.data = data
	for i, b := range v.data {
		v.iovec[i].iov_base = C.CBytes(b)
		v.iovec[i].iov_len = C.size_t(len(b))
	}
	return
}

// SyncToData writes the iovec buffers into the linked Go buffers
func (v Iovec) SyncToData() {
	for i, b := range v.iovec {
		if len(v.data[i]) != int(b.iov_len) {
			panic("buffers have been modified")
		}
		C.memcpy(unsafe.Pointer(&v.data[i][0]), b.iov_base, b.iov_len)
	}
}

// Pointer returns a pointer to the iovec
func (v Iovec) Pointer() unsafe.Pointer {
	if len(v.iovec) == 0 {
		return nil
	}
	return unsafe.Pointer(&v.iovec[0])
}

// Len returns a pointer to the iovec
func (v Iovec) Len() int {
	return len(v.iovec)
}

// Free the C memory in the Iovec.
func (v Iovec) Free() {
	for i := range v.iovec {
		C.free(v.iovec[i].iov_base)
		v.iovec[i].iov_base = nil
		v.iovec[i].iov_len = 0
	}
}
