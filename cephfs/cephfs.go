package cephfs

/*
#cgo LDFLAGS: -lcephfs
#cgo CPPFLAGS: -D_FILE_OFFSET_BITS=64
#include <stdlib.h>
#include <cephfs/libcephfs.h>
*/
import "C"

import (
	"unsafe"

	"github.com/ceph/go-ceph/internal/retry"
	"github.com/ceph/go-ceph/rados"
)

// MountInfo exports ceph's ceph_mount_info from libcephfs.cc
type MountInfo struct {
	mount *C.struct_ceph_mount_info
}

func createMount(id *C.char) (*MountInfo, error) {
	mount := &MountInfo{}
	ret := C.ceph_create(&mount.mount, id)
	if ret != 0 {
		return nil, getError(ret)
	}
	return mount, nil
}

// CreateMount creates a mount handle for interacting with Ceph.
func CreateMount() (*MountInfo, error) {
	return createMount(nil)
}

// CreateMountWithId creates a mount handle for interacting with Ceph.
// The caller can specify a unique id that will identify this client.
func CreateMountWithId(id string) (*MountInfo, error) {
	cid := C.CString(id)
	defer C.free(unsafe.Pointer(cid))
	return createMount(cid)
}

// CreateFromRados creates a mount handle using an existing rados cluster
// connection.
//
// Implements:
//  int ceph_create_from_rados(struct ceph_mount_info **cmount, rados_t cluster);
func CreateFromRados(conn *rados.Conn) (*MountInfo, error) {
	mount := &MountInfo{}
	ret := C.ceph_create_from_rados(&mount.mount, C.rados_t(conn.Cluster()))
	if ret != 0 {
		return nil, getError(ret)
	}
	return mount, nil
}

// ReadDefaultConfigFile loads the ceph configuration from the specified config file.
func (mount *MountInfo) ReadDefaultConfigFile() error {
	ret := C.ceph_conf_read_file(mount.mount, nil)
	return getError(ret)
}

// SetConfigOption sets the value of the configuration option identified by
// the given name.
//
// Implements:
//  int ceph_conf_set(struct ceph_mount_info *cmount, const char *option, const char *value);
func (mount *MountInfo) SetConfigOption(option, value string) error {
	cOption := C.CString(option)
	defer C.free(unsafe.Pointer(cOption))
	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))
	return getError(C.ceph_conf_set(mount.mount, cOption, cValue))
}

// GetConfigOption returns the value of the Ceph configuration option
// identified by the given name.
//
// Implements:
//  int ceph_conf_get(struct ceph_mount_info *cmount, const char *option, char *buf, size_t len);
func (mount *MountInfo) GetConfigOption(option string) (string, error) {
	cOption := C.CString(option)
	defer C.free(unsafe.Pointer(cOption))

	var (
		err error
		buf []byte
	)
	// range from 4k to 256KiB
	retry.WithSizes(4096, 1<<18, func(size int) retry.Hint {
		buf = make([]byte, size)
		ret := C.ceph_conf_get(
			mount.mount,
			cOption,
			(*C.char)(unsafe.Pointer(&buf[0])),
			C.size_t(len(buf)))
		err = getError(ret)
		return retry.DoubleSize.If(err == errNameTooLong)
	})
	if err != nil {
		return "", err
	}
	value := C.GoString((*C.char)(unsafe.Pointer(&buf[0])))
	return value, nil
}

// Init the file system client without actually mounting the file system.
//
// Implements:
//  int ceph_init(struct ceph_mount_info *cmount);
func (mount *MountInfo) Init() error {
	return getError(C.ceph_init(mount.mount))
}

// Mount the file system, establishing a connection capable of I/O.
//
// Implements:
//  int ceph_mount(struct ceph_mount_info *cmount, const char *root);
func (mount *MountInfo) Mount() error {
	ret := C.ceph_mount(mount.mount, nil)
	return getError(ret)
}

// MountWithRoot mounts the file system using the path provided for the root of
// the mount. This establishes a connection capable of I/O.
//
// Implements:
//  int ceph_mount(struct ceph_mount_info *cmount, const char *root);
func (mount *MountInfo) MountWithRoot(root string) error {
	croot := C.CString(root)
	defer C.free(unsafe.Pointer(croot))
	return getError(C.ceph_mount(mount.mount, croot))
}

// Unmount the file system.
//
// Implements:
//  int ceph_unmount(struct ceph_mount_info *cmount);
func (mount *MountInfo) Unmount() error {
	ret := C.ceph_unmount(mount.mount)
	return getError(ret)
}

// Release destroys the mount handle.
//
// Implements:
//  int ceph_release(struct ceph_mount_info *cmount);
func (mount *MountInfo) Release() error {
	if mount.mount == nil {
		return nil
	}
	ret := C.ceph_release(mount.mount)
	if err := getError(ret); err != nil {
		return err
	}
	mount.mount = nil
	return nil
}

// SyncFs synchronizes all filesystem data to persistent media.
func (mount *MountInfo) SyncFs() error {
	ret := C.ceph_sync_fs(mount.mount)
	return getError(ret)
}

// IsMounted checks mount status.
func (mount *MountInfo) IsMounted() bool {
	ret := C.ceph_is_mounted(mount.mount)
	return ret == 1
}
