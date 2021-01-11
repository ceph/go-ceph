package cutil

/*
#include <stdlib.h>
typedef void* voidptr;
*/
import "C"

import (
	"unsafe"
)

// PtrSize is the size of a pointer
const PtrSize = C.sizeof_voidptr

// SizeTSize is the size of C.size_t
const SizeTSize = C.sizeof_size_t

// SizeT wraps size_t from C.
type SizeT C.size_t

// This section contains a bunch of types that are basically just
// unsafe.Pointer but have specific types to help "self document" what the
// underlying pointer is really meant to represent.

// CPtr is an unsafe.Pointer to C allocated memory
type CPtr unsafe.Pointer

// CharPtrPtr is an unsafe pointer wrapping C's `char**`.
type CharPtrPtr unsafe.Pointer

// CharPtr is an unsafe pointer wrapping C's `char*`.
type CharPtr unsafe.Pointer

// SizeTPtr is an unsafe pointer wrapping C's `size_t*`.
type SizeTPtr unsafe.Pointer

// FreeFunc is a wrapper around calls to, or act like, C's free function.
type FreeFunc func(unsafe.Pointer)

// Malloc is C.malloc
func Malloc(s SizeT) CPtr { return CPtr(C.malloc(C.size_t(s))) }

// Free is C.free
func Free(p CPtr) { C.free(unsafe.Pointer(p)) }
