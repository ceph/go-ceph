package callbacks

import (
	"sync"
	"unsafe"
)

// The logic of this file is largely adapted from:
// https://github.com/golang/go/wiki/cgo#function-variables
//
// Also helpful:
// https://eli.thegreenplace.net/2019/passing-callbacks-and-pointers-to-cgo/

// Callbacks provides a tracker for data that is to be passed between Go
// and C callback functions. The Go callback/object may not be passed
// by a pointer to C code and so instead fake pointers into an internal
// map are used.
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

func getPtr(i uintptr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(nullPtr) + i)
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

// Add a callback/object to the tracker and return a new fake pointer
// for the object.
func Add(v interface{}) unsafe.Pointer {
	mutex.Lock()
	defer mutex.Unlock()
	i := nextIdx()
	cmap[i] = v
	return getPtr(i)
}

// Remove a callback/object given it's fake pointer.
func Remove(p unsafe.Pointer) {
	mutex.Lock()
	defer mutex.Unlock()
	i := uintptr(p)
	if validIdx(i) {
		cmap[i] = nil
		free = append(free, i)
	}
}

// Lookup returns a mapped callback/object given an fake pointer.
func Lookup(p unsafe.Pointer) interface{} {
	mutex.RLock()
	defer mutex.RUnlock()
	i := uintptr(p)
	if validIdx(i) {
		return cmap[i]
	}
	return nil
}
