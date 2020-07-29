package cref

import (
	"sync"
	"unsafe"
)

// The logic of this file is largely adapted from:
// https://github.com/golang/go/wiki/cgo#function-variables
//
// Also helpful:
// https://eli.thegreenplace.net/2019/passing-callbacks-and-pointers-to-cgo/

// Cref provides a registry for data that is to be passed between Go
// and C callback functions. The Go callback/object may not be passed
// by a pointer to C code and so instead references into an internal
// map are used. The references can produce types that can be safely passed
// to and received from C functions, that is void pointers and ints.
// Typically the item being added will either be a callback function or
// a data structure containing a callback function. It is up to the caller
// to control and validate what "callbacks" get used.

var (
	mutex   sync.RWMutex
	cmap    = []interface{}{nil}
	free    []uintptr
	nullPtr unsafe.Pointer
)

func reset() {
	cmap = cmap[:1]
	free = nil
}

func idx2Ptr(idx uintptr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(nullPtr) + idx)
}

func validIdx(i uintptr) bool {
	return i > 0 && i < uintptr(len(cmap))
}

func nextIdx() uintptr {
	var i uintptr
	if len(free) > 0 {
		n := len(free) - 1
		i = free[n]
		free = free[:n]
	} else {
		cmap = append(cmap, nil)
		i = uintptr(len(cmap) - 1)
	}
	return i
}

// Ref is a reference to a stored object
type Ref struct {
	idx uintptr
}

// Ptr creates a raw pointer from the reference
func (r Ref) Ptr() unsafe.Pointer {
	return idx2Ptr(r.idx)
}

// Uintptr creates a uintptr value from the reference
func (r Ref) Uintptr() uintptr {
	return r.idx
}

// Uintptr restores a reference from an uintptr value
func Uintptr(idx uintptr) Ref {
	if validIdx(idx) {
		return Ref{idx}
	}
	return Ref{0}
}

// Ptr restores a reference from a raw pointer
func Ptr(p unsafe.Pointer) Ref {
	return Uintptr(uintptr(p))
}

// Add a go object to the registry and return a new reference
// for the object.
func Add(v interface{}) Ref {
	mutex.Lock()
	defer mutex.Unlock()
	i := nextIdx()
	cmap[i] = v
	return Ref{i}
}

// Remove a object given it's reference.
func Remove(r Ref) {
	mutex.Lock()
	defer mutex.Unlock()
	i := r.idx
	if validIdx(i) {
		cmap[i] = nil
		free = append(free, i)
	}
}

// Lookup returns a mapped object given an reference.
func Lookup(r Ref) interface{} {
	mutex.RLock()
	defer mutex.RUnlock()
	i := r.idx
	if validIdx(i) {
		return cmap[i]
	}
	return nil
}
