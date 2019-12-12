package rados

// #cgo LDFLAGS: -lrados
// #include <stdlib.h>
// #include <rados/librados.h>
// extern void commit_callback(rados_completion_t comp, void *arg);
// extern int create_aio_read_completion(void* p, int* ret, rados_completion_t* comp);
// extern int create_aio_write_completion(void* p, int* ret, rados_completion_t* comp);
// typedef struct callback_args_t {
//     void* lock;
//     int*  ret;
// } callback_args_t;
import "C"

import (
	"sync"
	"unsafe"
)

type IOType int

const (
	IORead IOType = iota
	IOWrite
	IOAppend
)

//AioComplete is a wrapper of the callback
//export AioComplete
func AioComplete(p unsafe.Pointer, ret int32) {
	arg := (*C.callback_args_t)(p)
	*arg.ret = C.int(ret)
	(*sync.Mutex)(arg.lock).Unlock()
}

type future interface {
	// Get should block until the work is completed.
	Get() (interface{}, error)
}

type aioFuture struct {
	err    error
	n      *int32
	o      sync.Once
	buf    []byte
	offset uint64
	oid    string
	ioctx  *IOContext
	tp     IOType
	mu     *sync.Mutex
}

func (a *aioFuture) read() {
	c_oid := C.CString(a.oid)
	defer C.free(unsafe.Pointer(c_oid))
	var comp C.rados_completion_t
	ret := C.create_aio_read_completion(unsafe.Pointer(a.mu), (*C.int)(unsafe.Pointer(a.n)), &comp)
	if err := GetRadosError(int(ret)); err != nil {
		a.err = err
		return
	}
	a.mu.Lock()
	ret = C.rados_aio_read(a.ioctx.ioctx, c_oid, comp, (*C.char)(unsafe.Pointer(&a.buf[0])), (C.size_t)(len(a.buf)), (C.uint64_t)(a.offset))
	if err := GetRadosError(int(ret)); err != nil {
		a.err = err
		return
	}
}

// Get will be blocked until the aio callback is called.
func (a *aioFuture) Get() (interface{}, error) {
	a.o.Do(func() {
		if a.err != nil {
			return
		}
		a.mu.Lock()
		if (*a.n) < 0 {
			a.err = GetRadosError(int(*a.n))
		}
	})
	return int(*a.n), a.err
}

func (a *aioFuture) write() {
	c_oid := C.CString(a.oid)
	defer C.free(unsafe.Pointer(c_oid))
	var comp C.rados_completion_t
	ret := C.create_aio_write_completion(unsafe.Pointer(a.mu), (*C.int)(unsafe.Pointer(a.n)), &comp)
	if err := GetRadosError(int(ret)); err != nil {
		a.err = err
		return
	}
	a.mu.Lock()
	ret = C.rados_aio_write(a.ioctx.ioctx, c_oid, comp, (*C.char)(unsafe.Pointer(&a.buf[0])), (C.size_t)(len(a.buf)), (C.uint64_t)(a.offset))
	if err := GetRadosError(int(ret)); err != nil {
		a.err = err
		return
	}

}

func (a *aioFuture) append() {
	c_oid := C.CString(a.oid)
	defer C.free(unsafe.Pointer(c_oid))
	var comp C.rados_completion_t
	ret := C.create_aio_write_completion(unsafe.Pointer(a.mu), (*C.int)(unsafe.Pointer(a.n)), (*C.rados_completion_t)(&comp))
	if err := GetRadosError(int(ret)); err != nil {
		a.err = err
		return
	}
	a.mu.Lock()
	ret = C.rados_aio_append(a.ioctx.ioctx, c_oid, comp, (*C.char)(unsafe.Pointer(&a.buf[0])), (C.size_t)(len(a.buf)))
	if err := GetRadosError(int(ret)); err != nil {
		a.err = err
		return
	}

}
