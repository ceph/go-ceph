package rados

// #cgo LDFLAGS: -lrados
// #include <errno.h>
// #include <stdlib.h>
// #include <rados/librados.h>
//
// extern void go_notify_handle(void *arg,
//  uint64_t notify_id,
//  uint64_t handle,
//  uint64_t notifier_id,
//  void *data,
//  size_t data_len);
//
// extern void go_notify_error_handle(void *pre,
//  uint64_t cookie,
//  int err);
//
// static inline void notify_handle(void *arg,
//  uint64_t notify_id,
//  uint64_t handle,
//  uint64_t notifier_id,
//  void *data,
//  size_t data_len) {
// 	go_notify_handle(arg, notify_id, handle, notifier_id, data, data_len);
// }
//
// static inline void notify_error_handle(void *pre,
//  uint64_t cookie,
//  int err) {
// 	go_notify_error_handle(pre, cookie, err);
// }
//
// static inline int rados_watch_wrap(rados_ioctx_t io, const char *o, uint64_t *cookie, void *arg) {
//  return rados_watch2(io, o, cookie,
//   notify_handle,
//   notify_error_handle,
//   arg);
// }
import "C"

import (
	"sync"
	"unsafe"
)

//export go_notify_handle
func go_notify_handle(arg unsafe.Pointer, notify_id C.uint64_t,
	handle C.uint64_t, notifier_id C.uint64_t, c_data unsafe.Pointer,
	data_len C.size_t) {
	cookie := uint64(handle)

	mu.Lock()
	if watchers[cookie] == nil {
		// TODO:
		// May loss some notification.
	}
	watcher := watchers[cookie]
	mu.Unlock()

	data := *(*[]byte)(c_data)
	watcher.OnNotify(uint64(notify_id), cookie, uint64(notifier_id), data)
}

//export go_notify_error_handle
func go_notify_error_handle(pre unsafe.Pointer, handle C.uint64_t,
	err C.int) {
	cookie := uint64(handle)

	mu.Lock()
	if watchers[cookie] == nil {
		// TODO:
		// May loss some notification.
	}
	watcher := watchers[cookie]
	mu.Unlock()

	watcher.OnError(cookie, int(err))
}

var (
	mu       sync.Mutex
	watchers = make(map[uint64]Watcher)
)

type Watcher interface {
	OnNotify(notify_id uint64, cookie uint64, notifier_id uint64, data []byte)
	OnError(cookie uint64, err int)
}

func (ioctx *IOContext) Watch(oid string, watcher Watcher) (uint64, error) {
	c_oid := C.CString(oid)
	defer C.free(unsafe.Pointer(c_oid))

	mu.Lock()
	defer mu.Unlock()

	var cookie uint64
	C.rados_watch_wrap(ioctx.ioctx, c_oid,
		(*C.uint64_t)(&cookie),
		nil)
	watchers[cookie] = watcher

	return cookie, nil
}

func (ioctx *IOContext) WatchCheck(cookie uint64) (int, error) {
	ret := C.rados_watch_check(ioctx.ioctx, (C.uint64_t)(cookie))

	if int(ret) <= 0 {
		return 0, GetRadosError(int(ret))
	}
	return int(ret), nil
}

func (ioctx *IOContext) UnWatch(cookie uint64) error {
	mu.Lock()
	defer mu.Unlock()
	delete(watchers, cookie)

	ret := C.rados_unwatch2(ioctx.ioctx, (C.uint64_t)(cookie))

	switch ret {
	case 0:
		return nil
	default:
		return RadosError(int(ret))
	}
}

func (ioctx *IOContext) Notify(oid string, data []byte, timeoutMS uint64) error {
	c_oid := C.CString(oid)
	defer C.free(unsafe.Pointer(c_oid))

	var replyBuffer *C.char
	var replyBufferLen C.size_t
	ret := C.rados_notify2(ioctx.ioctx, c_oid,
		(*C.char)(unsafe.Pointer(&data[0])),
		(C.int)(len(data)),
		(C.uint64_t)(timeoutMS),
		&replyBuffer,
		&replyBufferLen)

	switch ret {
	case 0:
		return nil
	default:
		return RadosError(int(ret))
	}
}

func (ioctx *IOContext) NotifyAck(oid string, notify_id uint64, cookie uint64) error {
	c_oid := C.CString(oid)
	defer C.free(unsafe.Pointer(c_oid))

	data := []byte("done")
	ret := C.rados_notify_ack(ioctx.ioctx, c_oid,
		(C.uint64_t)(notify_id),
		(C.uint64_t)(cookie),
		(*C.char)(unsafe.Pointer(&data[0])),
		(C.int)(len(data)))

	switch ret {
	case 0:
		return nil
	default:
		return RadosError(int(ret))
	}
}

func (ioctx *IOContext) WatchFlush() {
	C.rados_watch_flush(ioctx.cluster)
}
