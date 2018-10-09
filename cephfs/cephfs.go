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
	log "github.com/sirupsen/logrus"
	"math"
	"syscall"
	"unsafe"
)

type CephError int

func (e CephError) Error() string {
	if e == 0 {
		return fmt.Sprintf("")
	} else {
		err := syscall.Errno(uint(math.Abs(float64(e))))
		return fmt.Sprintf("cephfs: ret=(%d) %v", e, err)
	}
}

type MountInfo struct {
	mount *C.struct_ceph_mount_info
}

func CreateMount() (*MountInfo, error) {
	mount := &MountInfo{}
	ret := C.ceph_create(&mount.mount, nil)
	if ret == 0 {
		return mount, nil
	} else {
		log.Errorf("CreateMount: Failed to create mount")
		return nil, CephError(ret)
	}
}

func (mount *MountInfo) RemoveDir(path string) error {
	c_path := C.CString(path)
	defer C.free(unsafe.Pointer(c_path))

	ret := C.ceph_rmdir(mount.mount, c_path)
	if ret == 0 {
		return nil
	} else {
		log.Errorf("RemoveDir: Failed to remove directory")
		return CephError(ret)
	}
}

func (mount *MountInfo) Unmount() error {
	ret := C.ceph_unmount(mount.mount)
	if ret == 0 {
		return nil
	} else {
		log.Errorf("Unmount: Failed to unmount")
		return CephError(ret)
	}
}

func (mount *MountInfo) Release() error {
	ret := C.ceph_release(mount.mount)
	if ret == 0 {
		return nil
	} else {
		log.Errorf("Release: Failed to release mount")
		return CephError(ret)
	}
}

func (mount *MountInfo) ReadDefaultConfigFile() error {
	ret := C.ceph_conf_read_file(mount.mount, nil)
	if ret == 0 {
		return nil
	} else {
		log.Errorf("ReadDefaultConfigFile: Failed to read ceph config")
		return CephError(ret)
	}
}

func (mount *MountInfo) Mount() error {
	ret := C.ceph_mount(mount.mount, nil)
	if ret == 0 {
		return nil
	} else {
		log.Errorf("Mount: Failed to mount")
		return CephError(ret)
	}
}

func (mount *MountInfo) SyncFs() error {
	ret := C.ceph_sync_fs(mount.mount)
	if ret == 0 {
		return nil
	} else {
		log.Errorf("Mount: Failed to sync filesystem")
		return CephError(ret)
	}
}

func (mount *MountInfo) CurrentDir() string {
	c_dir := C.ceph_getcwd(mount.mount)
	return C.GoString(c_dir)
}

func (mount *MountInfo) ChangeDir(path string) error {
	c_path := C.CString(path)
	defer C.free(unsafe.Pointer(c_path))

	ret := C.ceph_chdir(mount.mount, c_path)
	if ret == 0 {
		return nil
	} else {
		log.Errorf("ChangeDir: Failed to change directory")
		return CephError(ret)
	}
}

func (mount *MountInfo) MakeDir(path string, mode uint32) error {
	c_path := C.CString(path)
	defer C.free(unsafe.Pointer(c_path))

	ret := C.ceph_mkdir(mount.mount, c_path, C.mode_t(mode))
	if ret == 0 {
		return nil
	} else {
		log.Errorf("MakeDir: Failed to make directory %s", path)
		return CephError(ret)
	}
}

func (mount *MountInfo) Chmod(path string, mode uint32) error {
	c_path := C.CString(path)
	defer C.free(unsafe.Pointer(c_path))

	ret := C.ceph_chmod(mount.mount, c_path, C.mode_t(mode))
	if ret == 0 {
		return nil
	} else {
		log.Errorf("Chmod: Failed to chmod :%s", path)
		return CephError(ret)
	}
}

func (mount *MountInfo) Chown(path string, user uint32, group uint32) error {
	c_path := C.CString(path)
	defer C.free(unsafe.Pointer(c_path))

	ret := C.ceph_chown(mount.mount, c_path, C.int(user), C.int(group))
	if ret == 0 {
		return nil
	} else {
		log.Errorf("Chown: Failed to chown :%s", path)
		return CephError(ret)
	}
}

/*
 * Helper functions
 */

func (mount *MountInfo) IsMounted() bool {
	ret := C.ceph_is_mounted(mount.mount)
	return ret == 0
}

func (mount *MountInfo) GetMount() *C.struct_ceph_mount_info {
	return mount.mount
}
