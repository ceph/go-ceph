package cutil

// #include <stdlib.h>
import "C"
import (
	"unsafe"
)

const maxIdx = 1<<31 - 1 // 2GB, max int32 value, should be safe

// CPtr is a pointer to C memory.
// Not required, but makes it more obvious, when only pointers to C memory and
// no pointers to Go memory are allowed to be stored.
type CPtr unsafe.Pointer

// CSize XXX
type CSize C.size_t

// Can't use C in generated code
func cMalloc(n CSize) unsafe.Pointer {
	return C.malloc(C.size_t(n))
}

func cFree(p unsafe.Pointer) {
	C.free(p)
}
