package cutil

/*
#include <stdlib.h>
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
	pgs   []*PtrGuard
}

const iovecSize = C.sizeof_struct_iovec

// ByteSlicesToIovec creates an Iovec and links it to Go buffers in data.
func ByteSlicesToIovec(data [][]byte) (v Iovec) {
	v.data = data
	return
}

// Pointer returns a pointer to the iovec
func (v *Iovec) Pointer() unsafe.Pointer {
	v.Free()
	n := len(v.data)
	iovecMem := C.malloc(iovecSize * C.size_t(n))
	v.iovec = (*[maxIdx]C.struct_iovec)(iovecMem)[:n:n]
	for i, b := range v.data {
		pg := NewPtrGuard(CPtr(&v.iovec[i].iov_base), unsafe.Pointer(&b[0]))
		v.pgs = append(v.pgs, pg)
		v.iovec[i].iov_len = C.size_t(len(b))
	}
	return unsafe.Pointer(&v.iovec[0])
}

// Len returns a pointer to the iovec
func (v *Iovec) Len() int {
	return len(v.data)
}

// Free the C memory in the Iovec.
func (v *Iovec) Free() {
	for _, pg := range v.pgs {
		pg.Release()
	}
	if len(v.iovec) != 0 {
		C.free(unsafe.Pointer(&v.iovec[0]))
	}
	v.iovec = nil
}
