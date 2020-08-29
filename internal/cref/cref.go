package cref

import "C"

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
	mutex sync.RWMutex
	cmap  = map[unsafe.Pointer]interface{}{}
	free  []unsafe.Pointer
)

func reset() {
	for p := range cmap {
		Remove(p)
	}
}

func allocMem() {
	size := 1024
	mem := C.malloc(1024)
	if mem == nil {
		panic("can't allocate memory for C pointer")
	}
	for i := size - 1; i >= 0; i-- {
		p := unsafe.Pointer(uintptr(mem) + uintptr(i))
		free = append(free, p)
	}
}

func getPtr() unsafe.Pointer {
	if len(free) == 0 {
		allocMem()
	}
	n := len(free) - 1
	p := free[n]
	free = free[:n]
	return p
}

// Add a go object to the registry and return a C compatible pointer for the
// object.
func Add(v interface{}) unsafe.Pointer {
	mutex.Lock()
	defer mutex.Unlock()
	p := getPtr()
	cmap[p] = v
	return p
}

// Remove a object given it's reference.
func Remove(p unsafe.Pointer) {
	mutex.Lock()
	defer mutex.Unlock()
	if _, ok := cmap[p]; ok {
		cmap[p] = nil
		delete(cmap, p)
		free = append(free, p)
	}
}

// Lookup returns a mapped object given an reference.
func Lookup(p unsafe.Pointer) interface{} {
	mutex.RLock()
	defer mutex.RUnlock()
	return cmap[p]
}
