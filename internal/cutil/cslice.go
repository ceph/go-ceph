package cutil

// #include <stdlib.h>
import "C"
import "unsafe"

// The following code needs some explanation:
// This creates slices on top of the C memory buffers allocated before in
// order to safely and comfortably use them as arrays. First the void pointer
// is cast to a pointer to an array of the type that will be stored in the
// array. Because the size of an array is a constant, but the real array size
// is dynamic, we just use the biggest possible index value maxIdx, to make
// sure it's always big enough. (Nothing is allocated by casting, so the size
// can be arbitrarily big.) So, if the array should store items of myType, the
// cast would be (*[maxIdx]myItem)(myCMemPtr).
// From that array pointer a slice is created with the [start:end:capacity]
// syntax. The capacity must be set explicitly here, because by default it
// would be set to the size of the original array, which is maxIdx, which
// doesn't reflect reality in this case. This results in definitions like:
// cSlice := (*[maxIdx]myItem)(myCMemPtr)[:numOfItems:numOfItems]

const maxIdx = 1<<31 - 1 // 2GB, max int32 value, should be safe
const ptrSize = C.size_t(unsafe.Sizeof(unsafe.Pointer(nil)))

// CPtr is a pointer to C memory.
// Not required, but makes it more obvious, when only pointers to C memory and
// no pointers to Go memory are allowed to be stored.
type CPtr unsafe.Pointer

// CSize is a C.size_t
// (required, because C.size_t is different for every package)
type CSize C.size_t

////////// CPtr //////////

// CPtrSlice is a C allocated slice of pointers.
type CPtrSlice []CPtr

// NewCPtrSlice returns a CPtrSlice.
// Similar to CString it must be freed with Free(slice)
func NewCPtrSlice(size int) CPtrSlice {
	if size == 0 {
		return nil
	}
	cMem := C.malloc(C.size_t(size) * ptrSize)
	cSlice := (*[maxIdx]CPtr)(cMem)[:size:size]
	return cSlice
}

// Ptr returns a pointer to CPtrSlice
func (v *CPtrSlice) Ptr() CPtr {
	if len(*v) == 0 {
		return nil
	}
	return CPtr(&(*v)[0])
}

// Free frees a CPtrSlice
func (v *CPtrSlice) Free() {
	C.free(unsafe.Pointer(v.Ptr()))
	*v = nil
}

////////// CSize //////////

// CSizeSlice is a C allocated slice of C.size_t.
type CSizeSlice []CSize

// NewCSizeSlice returns a CSizeSlice.
// Similar to CString it must be freed with Free(slice)
func NewCSizeSlice(size int) CSizeSlice {
	if size == 0 {
		return nil
	}
	cMem := C.malloc(C.size_t(size) * ptrSize)
	cSlice := (*[maxIdx]CSize)(cMem)[:size:size]
	return cSlice
}

// Ptr returns a pointer to CSizeSlice
func (v *CSizeSlice) Ptr() CPtr {
	if len(*v) == 0 {
		return nil
	}
	return CPtr(&(*v)[0])
}

// Free frees a CSizeSlice
func (v *CSizeSlice) Free() {
	C.free(unsafe.Pointer(v.Ptr()))
	*v = nil
}
