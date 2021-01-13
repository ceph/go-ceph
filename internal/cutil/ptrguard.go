package cutil

/*
#include <stdlib.h>

extern void release_wait(void*);
extern void stored_signal(void*);

static inline void storeAndWait(void** c_ptr, void* go_ptr, void* v) {
	*c_ptr = go_ptr;
	stored_signal(v);
	release_wait(v);
	*c_ptr = NULL;
	stored_signal(v);
}
*/
import "C"

import (
	"sync"
	"unsafe"
)

type semaphore struct {
	sync.Mutex
}

func (v *semaphore) init() {
	v.wait()
}

func (v *semaphore) wait() {
	v.Lock()
}

func (v *semaphore) signal() {
	v.Unlock()
}

// PtrGuard respresents a guarded Go pointer (pointing to memory allocated by Go
// runtime) stored in C memory (allocated by C)
type PtrGuard struct {
	stored, release semaphore
	released        bool
}

// WARNING: using binary semaphores (mutexes) for signalling like this is quite
// a delicate task in order to avoid deadlocks or panics. Whenever changing the
// code logic, please review at least three times that there is no unexpected
// state possible. Usually the natural choice would be to use channels instead,
// but these can not easily passed to C code because of the pointer-to-pointer
// cgo rule, and would require the use of a Go object registry.

// NewPtrGuard writes the goPtr (pointing to Go memory) into C memory at the
// position cPtr, and returns a PtrGuard object.
func NewPtrGuard(cPtr CPtr, goPtr unsafe.Pointer) *PtrGuard {
	var v PtrGuard
	v.release.init()
	v.stored.init()
	go C.storeAndWait((*unsafe.Pointer)(cPtr), goPtr, unsafe.Pointer(&v))
	v.stored.wait()
	return &v
}

// Release removes the guarded Go pointer from the C memory by overwriting it
// with NULL.
func (v *PtrGuard) Release() {
	if !v.released {
		v.released = true
		v.release.signal() // send release signal
		v.stored.wait()    // wait for stored signal
	}
}

//export release_wait
func release_wait(p unsafe.Pointer) {
	v := (*PtrGuard)(p)
	v.release.wait()
}

//export stored_signal
func stored_signal(p unsafe.Pointer) {
	v := (*PtrGuard)(p)
	v.stored.signal()
}
