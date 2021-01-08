package cutil

//go:generate genny -in=$GOFILE -out=gen-$GOFILE gen "ElementType=CPtr,CSize"

import (
	"unsafe"

	"github.com/cheekybits/genny/generic"
)

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

// ElementType is the type of the array elements
type ElementType generic.Type

// ElementTypeSize is the size of ElementType
const ElementTypeSize = CSize(unsafe.Sizeof(*(*ElementType)(nil)))

// ElementTypeCSlice is a C allocated slice of pointers.
type ElementTypeCSlice []ElementType

// NewElementTypeCSlice returns a ElementTypeCSlice.
// Similar to CString it must be freed with Free(slice)
func NewElementTypeCSlice(size int) ElementTypeCSlice {
	if size == 0 {
		return nil
	}
	cMem := cMalloc(CSize(size) * ElementTypeSize)
	cSlice := (*[maxIdx]ElementType)(cMem)[:size:size]
	return cSlice
}

// Ptr returns a pointer to ElementTypeCSlice
func (v *ElementTypeCSlice) Ptr() CPtr {
	if len(*v) == 0 {
		return nil
	}
	return CPtr(&(*v)[0])
}

// Free frees a ElementTypeCSlice
func (v *ElementTypeCSlice) Free() {
	cFree(unsafe.Pointer(v.Ptr()))
	*v = nil
}
