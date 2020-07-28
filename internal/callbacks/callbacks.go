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
// by a pointer to C code and so instead integer indexes into an internal
// map are used.
// Typically the item being added will either be a callback function or
// a data structure containing a callback function. It is up to the caller
// to control and validate what "callbacks" get used.
type Callbacks struct {
	mutex sync.RWMutex
	cmap  map[unsafe.Pointer]interface{}
	free  []unsafe.Pointer
	last  unsafe.Pointer
}

func (cb *Callbacks) nextPtr() unsafe.Pointer {
	var p unsafe.Pointer
	if len(cb.free) > 0 {
		n := len(cb.free) - 1
		p = cb.free[n]
		cb.free = cb.free[:n]
	} else {
		cb.last = unsafe.Pointer(uintptr(cb.last) + 1)
		p = cb.last
	}
	return p
}

// New returns a new callbacks tracker.
func New() *Callbacks {
	return &Callbacks{cmap: make(map[unsafe.Pointer]interface{})}
}

// Add a callback/object to the tracker and return a new fake pointer
// for the object.
func (cb *Callbacks) Add(v interface{}) unsafe.Pointer {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	p := cb.nextPtr()
	cb.cmap[p] = v
	return p
}

// Remove a callback/object given it's fake pointer.
func (cb *Callbacks) Remove(p unsafe.Pointer) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	delete(cb.cmap, p)
	cb.free = append(cb.free, p)
}

// Lookup returns a mapped callback/object given an index.
func (cb *Callbacks) Lookup(p unsafe.Pointer) interface{} {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.cmap[p]
}
