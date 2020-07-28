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
type Callbacks struct {
	mutex sync.RWMutex
	cmap  []interface{}
	free  []uintptr
}

var nullPtr unsafe.Pointer

func getPtr(i uintptr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(nullPtr) + i)
}

func (cb *Callbacks) validIndex(i uintptr) bool {
	return i > 0 && i < uintptr(len(cb.cmap))
}

func (cb *Callbacks) nextIndex() uintptr {
	var i uintptr
	if len(cb.free) > 0 {
		n := len(cb.free) - 1
		i = cb.free[n]
		cb.free = cb.free[:n]
	} else {
		cb.cmap = append(cb.cmap, nil)
		i = uintptr(len(cb.cmap) - 1)
	}
	return i
}

// New returns a new callbacks tracker.
func New() *Callbacks {
	return &Callbacks{cmap: []interface{}{nil}}
}

// Add a callback/object to the tracker and return a new fake pointer
// for the object.
func (cb *Callbacks) Add(v interface{}) unsafe.Pointer {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	i := cb.nextIndex()
	cb.cmap[i] = v
	return getPtr(i)
}

// Remove a callback/object given it's fake pointer.
func (cb *Callbacks) Remove(p unsafe.Pointer) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	i := uintptr(p)
	if cb.validIndex(i) {
		cb.cmap[i] = nil
		cb.free = append(cb.free, i)
	}
}

// Lookup returns a mapped callback/object given an fake pointer.
func (cb *Callbacks) Lookup(p unsafe.Pointer) interface{} {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	i := uintptr(p)
	if cb.validIndex(i) {
		return cb.cmap[i]
	}
	return nil
}
