package cephfs

/*
#cgo LDFLAGS: -lcephfs
#cgo CPPFLAGS: -D_FILE_OFFSET_BITS=64
#include <stdlib.h>
#include <cephfs/libcephfs.h>
*/
import "C"

import (
	"fmt"
	"unsafe"

	"github.com/ceph/go-ceph/errutil"
	"github.com/ceph/go-ceph/rados"
)

type CephFSError int

func (e CephFSError) Error() string {
	errno, s := errutil.FormatErrno(int(e))
	if s == "" {
		return fmt.Sprintf("cephfs: ret=%d", errno)
	}
	return fmt.Sprintf("cephfs: ret=%d, %s", errno, s)
}

func getError(e C.int) error {
	if e == 0 {
		return nil
	}
	return CephFSError(e)
}

// MountInfo exports ceph's ceph_mount_info from libcephfs.cc
type MountInfo struct {
	mount *C.struct_ceph_mount_info
}

// CreateMount creates a mount handle for interacting with Ceph.
func CreateMount() (*MountInfo, error) {
	mount := &MountInfo{}
	ret := C.ceph_create(&mount.mount, nil)
	if ret != 0 {
		return nil, getError(ret)
	}
	return mount, nil
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

// Mount mounts the mount handle.
func (mount *MountInfo) Mount() error {
	ret := C.ceph_mount(mount.mount, nil)
	return getError(ret)
}

// Unmount unmounts the mount handle.
func (mount *MountInfo) Unmount() error {
	ret := C.ceph_unmount(mount.mount)
	return getError(ret)
}

// Release destroys the mount handle.
func (mount *MountInfo) Release() error {
	ret := C.ceph_release(mount.mount)
	return getError(ret)
}

// SyncFs synchronizes all filesystem data to persistent media.
func (mount *MountInfo) SyncFs() error {
	ret := C.ceph_sync_fs(mount.mount)
	return getError(ret)
}

// CurrentDir gets the current working directory.
func (mount *MountInfo) CurrentDir() string {
	cDir := C.ceph_getcwd(mount.mount)
	return C.GoString(cDir)
}

// ChangeDir changes the current working directory.
func (mount *MountInfo) ChangeDir(path string) error {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	ret := C.ceph_chdir(mount.mount, cPath)
	return getError(ret)
}

// MakeDir creates a directory.
func (mount *MountInfo) MakeDir(path string, mode uint32) error {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	ret := C.ceph_mkdir(mount.mount, cPath, C.mode_t(mode))
	return getError(ret)
}

// RemoveDir removes a directory.
func (mount *MountInfo) RemoveDir(path string) error {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	ret := C.ceph_rmdir(mount.mount, cPath)
	return getError(ret)
}

// Chmod changes the mode bits (permissions) of a file/directory.
func (mount *MountInfo) Chmod(path string, mode uint32) error {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	ret := C.ceph_chmod(mount.mount, cPath, C.mode_t(mode))
	return getError(ret)
}

// Chown changes the ownership of a file/directory.
func (mount *MountInfo) Chown(path string, user uint32, group uint32) error {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	ret := C.ceph_chown(mount.mount, cPath, C.int(user), C.int(group))
	return getError(ret)
}

// IsMounted checks mount status.
func (mount *MountInfo) IsMounted() bool {
	ret := C.ceph_is_mounted(mount.mount)
	return ret == 1
}
