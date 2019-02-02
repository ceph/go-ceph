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
func AioComplete(p unsafe.Pointer, ret int) {
	arg := (*C.callback_args_t)(p)
	*arg.ret = C.int(ret)
	(*sync.Mutex)(arg.lock).Unlock()
}

type future interface {
	Get() (interface{}, error)
}

type aioFuture struct {
	err    error
	n      int
	o      sync.Once
	buf    []byte
	offset uint64
	oid    string
	ioctx  *IOContext
	tp     IOType
}

func (a *aioFuture) Get() (interface{}, error) {
	switch a.tp {
	case IORead:
		return a.readGet()
	case IOWrite:
		return a.writeGet()
	case IOAppend:
		return a.appendGet()
	default:
		panic("unsupported io type")
	}
}

func (a *aioFuture) readGet() (interface{}, error) {
	var err error
	a.o.Do(func() {
		c_oid := C.CString(a.oid)
		defer C.free(unsafe.Pointer(c_oid))
		var comp C.rados_completion_t
		mu := &sync.Mutex{}
		var result C.int
		ret := C.create_aio_read_completion(unsafe.Pointer(mu), &result, &comp)
		if err = GetRadosError(int(ret)); err != nil {
			return
		}
		mu.Lock()
		ret = C.rados_aio_read(a.ioctx.ioctx, c_oid, comp, (*C.char)(unsafe.Pointer(&a.buf[0])), (C.size_t)(len(a.buf)), (C.uint64_t)(a.offset))
		if err = GetRadosError(int(ret)); err != nil {
			return
		}
		mu.Lock()
		if int(result) < 0 {
			a.err = GetRadosError(int(result))
		} else {
			a.n = int(result)
		}
	})
	return a.n, a.err
}

func (a *aioFuture) writeGet() (interface{}, error) {
	var err error
	a.o.Do(func() {
		c_oid := C.CString(a.oid)
		defer C.free(unsafe.Pointer(c_oid))
		var comp C.rados_completion_t
		mu := &sync.Mutex{}
		var result C.int
		ret := C.create_aio_write_completion(unsafe.Pointer(mu), &result, &comp)
		if err = GetRadosError(int(ret)); err != nil {
			return
		}
		mu.Lock()
		ret = C.rados_aio_write(a.ioctx.ioctx, c_oid, comp, (*C.char)(unsafe.Pointer(&a.buf[0])), (C.size_t)(len(a.buf)), (C.uint64_t)(a.offset))
		if err = GetRadosError(int(ret)); err != nil {
			return
		}
		mu.Lock()
		if int(result) < 0 {
			a.err = GetRadosError(int(result))
		} else {
			a.n = int(result)
		}
	})
	return a.n, a.err
}

func (a *aioFuture) appendGet() (interface{}, error) {
	var err error
	a.o.Do(func() {
		c_oid := C.CString(a.oid)
		defer C.free(unsafe.Pointer(c_oid))
		var comp C.rados_completion_t
		mu := &sync.Mutex{}
		var result C.int
		ret := C.create_aio_write_completion(unsafe.Pointer(mu), &result, (*C.rados_completion_t)(&comp))
		if err = GetRadosError(int(ret)); err != nil {
			return
		}
		mu.Lock()
		ret = C.rados_aio_append(a.ioctx.ioctx, c_oid, comp, (*C.char)(unsafe.Pointer(&a.buf[0])), (C.size_t)(len(a.buf)))
		if err = GetRadosError(int(ret)); err != nil {
			return
		}
		mu.Lock()
		if int(result) < 0 {
			a.err = GetRadosError(int(result))
		} else {
			a.n = int(result)
		}
	})
	return a.n, a.err
}
