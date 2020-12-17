package cutil

/*
extern void waitForReleaseSignal(void*);
extern void sendStoredSignal(void*);

static inline void storeUntilRelease(void** c_ptr, void* go_ptr, void* v) {
	*c_ptr = go_ptr;         // store Go pointer in C memory at c_ptr
	sendStoredSignal(v);     // send "stored" signal to main thread -->(1)
  waitForReleaseSignal(v); // wait for "release" signal from main thread when
                           // Release() has been called. <--(2)
	*c_ptr = NULL;           // reset C memory to NULL
	sendStoredSignal(v);     // send second "stored" signal to main thread -->(3)
}
*/
import "C"

import (
	"sync"
	"unsafe"
)

// PtrGuard respresents a guarded Go pointer (pointing to memory allocated by Go
// runtime) stored in C memory (allocated by C)
type PtrGuard struct {
	// These mutexes will be used as binary semaphores for signalling events from
	// one thread to another, which - in contrast to other languages like C++ - is
	// possible in Go, that is a Mutex can be locked in one thread and unlocked in
	// another.
	stored, release sync.Mutex
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
	// Since the mutexes are used for signalling, they have to be initialized to
	// locked state, so that following lock attempts will block.
	v.release.Lock()
	v.stored.Lock()
	// Start a background go routine that lives until Release is called. This
	// calls a C function that stores goPtr into C memory at position cPtr and
	// then waits until it reveices the "release" signal, after which it nulls out
	// the C memory at cPtr and then exits.
	go C.storeUntilRelease((*unsafe.Pointer)(cPtr), goPtr, unsafe.Pointer(&v))
	// Wait for the "stored" signal from the go routine when the Go pointer has
	// been stored to the C memory. <--(1)
	v.stored.Lock()
	return &v
}

// Release removes the guarded Go pointer from the C memory by overwriting it
// with NULL.
func (v *PtrGuard) Release() {
	if !v.released {
		v.released = true
		v.release.Unlock() // Send the "release" signal to the go routine. -->(2)
		// Wait for the second "stored" signal when the C memory has been nulled
		// out. <--(3)
		v.stored.Lock()
	}
}

//export waitForReleaseSignal
func waitForReleaseSignal(p unsafe.Pointer) {
	v := (*PtrGuard)(p)
	v.release.Lock()
}

//export sendStoredSignal
func sendStoredSignal(p unsafe.Pointer) {
	v := (*PtrGuard)(p)
	v.stored.Unlock()
}
