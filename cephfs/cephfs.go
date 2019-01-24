package cephfs

/*
#cgo LDFLAGS: -lcephfs
#cgo CPPFLAGS: -D_FILE_OFFSET_BITS=64
#include <stdlib.h>
#include <string.h>
#include <errno.h>
#include <sys/stat.h>
#include <cephfs/libcephfs.h>
*/
import "C"

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"math"
	"strings"
	"syscall"
	"unsafe"
)

type cephError int

func (e cephError) Error() string {
	if e == 0 {
		return fmt.Sprintf("cephfs: no error given")
	}
	err := syscall.Errno(uint(math.Abs(float64(e))))
	return fmt.Sprintf("cephfs: ret=(%d) %v", e, err)
}

const (
	CephStatxMode       = C.CEPH_STATX_MODE
	CephStatxNlink      = C.CEPH_STATX_NLINK
	CephStatxUid        = C.CEPH_STATX_UID
	CephStatxGid        = C.CEPH_STATX_GID
	CephStatxRdev       = C.CEPH_STATX_RDEV
	CephStatxAtime      = C.CEPH_STATX_ATIME
	CephStatxMtime      = C.CEPH_STATX_MTIME
	CephStatxCtime      = C.CEPH_STATX_CTIME
	CephStatxIno        = C.CEPH_STATX_INO
	CephStatxSize       = C.CEPH_STATX_SIZE
	CephStatxBlocks     = C.CEPH_STATX_BLOCKS
	CephStatxBasicStats = C.CEPH_STATX_BASIC_STATS
	CephStatxBtime      = C.CEPH_STATX_BTIME
	CephStatxVersion    = C.CEPH_STATX_VERSION
	CephStatxAllStats   = C.CEPH_STATX_ALL_STATS

	CephStatxMask = (CephStatxUid | CephStatxGid | CephStatxSize | CephStatxBlocks | CephStatxAtime | CephStatxMtime)

	CephSetAttrMode  = C.CEPH_SETATTR_MODE
	CephSetAttrUid   = C.CEPH_SETATTR_UID
	CephSetAttrGid   = C.CEPH_SETATTR_GID
	CephSetAttrMtime = C.CEPH_SETATTR_MTIME
	CephSetAttrAtime = C.CEPH_SETATTR_ATIME
	CephSetAttrSize  = C.CEPH_SETATTR_SIZE
	CephSetAttrCtime = C.CEPH_SETATTR_CTIME
	CephSetAttrBtime = C.CEPH_SETATTR_BTIME
)

type CephStat struct {
	Mode       uint16
	Uid        uint
	Gid        uint
	Size       uint64
	BlkSize    uint
	Blocks     uint64
	AppendTime int64
	ModifyTime int64

	IsFile    bool // S_ISREG
	IsDir     bool // S_ISDIR
	IsSymlink bool // S_ISLNK
}

// MountInfo exports ceph's ceph_mount_info from libcephfs.cc
type MountInfo struct {
	mount *C.struct_ceph_mount_info
}

type DirResult struct {
	dirp *C.struct_ceph_dir_result
}

type Statx struct {
	stx C.struct_ceph_statx
}

// CreateMount creates a mount handle for interacting with Ceph.
func CreateMount() (*MountInfo, error) {
	mount := &MountInfo{}

	ret := C.ceph_create(&mount.mount, nil)
	if ret != 0 {
		log.Errorf("CreateMount: Failed to create mount")
		return nil, cephError(ret)
	}
	return mount, nil
}

// CreateMountWithClient creates a mount handle for interacting with Ceph.
// id is the id of client, this can be an unique id that identifies the client
// and will get appended onto "client."
func CreateMountWithClient(id string) (*MountInfo, error) {
	mount := &MountInfo{}

	var cId *C.char
	if strings.Compare(id, "") == 0 {
		cId = nil
	} else {
		cId = C.CString(id)
		defer C.free(unsafe.Pointer(cId))
	}

	ret := C.ceph_create(&mount.mount, cId)
	if ret != 0 {
		log.Errorf("CreateMount: Failed to create mount")
		return nil, cephError(ret)
	}
	return mount, nil
}

// ReadDefaultConfigFile loads the ceph configuration from the specified config file.
func (mount *MountInfo) ReadDefaultConfigFile() error {
	ret := C.ceph_conf_read_file(mount.mount, nil)
	if ret != 0 {
		log.Errorf("ReadDefaultConfigFile: Failed to read ceph config")
		return cephError(ret)
	}
	return nil
}

// ReadConfigFile loads the ceph configuration from the specified config file.
func (mount *MountInfo) ReadConfigFile(path string) error {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	ret := C.ceph_conf_read_file(mount.mount, cPath)
	if ret != 0 {
		log.Errorf("ReadConfigFile: Failed to read ceph config")
		return cephError(ret)
	}
	return nil
}

// SetConf sets a configuration value from a string
func (mount *MountInfo) SetConf(option, value string) error {
	cOption := C.CString(option)
	defer C.free(unsafe.Pointer(cOption))
	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))

	ret := C.ceph_conf_set(mount.mount, cOption, cValue)
	if ret != 0 {
		log.Errorf("SetConf: Failed to set ceph config")
		return cephError(ret)
	}
	return nil
}

// GetConf gets a configuration value as a string
func (mount *MountInfo) GetConf(option string) (string, error) {
	cOption := C.CString(option)
	defer C.free(unsafe.Pointer(cOption))

	bufLen := 128
	buf := C.malloc(C.sizeof_char * C.size_t(bufLen))
	defer C.free(unsafe.Pointer(buf))

	for {
		ret := C.ceph_conf_get(mount.mount, cOption, (*C.char)(buf), C.size_t(bufLen))
		if ret == -C.ENAMETOOLONG {
			bufLen *= 2
			buf = C.malloc(C.sizeof_char * C.size_t(bufLen))
			continue
		} else if ret < 0 {
			log.Errorf("GetConf: Failed to get ceph config")
			return "", cephError(ret)
		}
		value := C.GoString((*C.char)(buf))
		return value, nil
	}

}

// Mount mounts the mount handle.
func (mount *MountInfo) Mount() error {
	ret := C.ceph_mount(mount.mount, nil)
	if ret != 0 {
		log.Errorf("Mount: Failed to mount")
		return cephError(ret)
	}
	return nil
}

// MountRoot mounts the mount handle using the path for the root of the mount.
func (mount *MountInfo) MountRoot(rootPath string) error {
	cPath := C.CString(rootPath)
	defer C.free(unsafe.Pointer(cPath))

	ret := C.ceph_mount(mount.mount, cPath)
	if ret != 0 {
		log.Errorf("Mount: Failed to mount")
		return cephError(ret)
	}
	return nil
}

// Unmount unmounts the mount handle.
func (mount *MountInfo) Unmount() error {
	ret := C.ceph_unmount(mount.mount)
	if ret != 0 {
		log.Errorf("Unmount: Failed to unmount")
		return cephError(ret)
	}
	return nil
}

// Release destroys the mount handle.
func (mount *MountInfo) Release() error {
	ret := C.ceph_release(mount.mount)
	if ret != 0 {
		log.Errorf("Release: Failed to release mount")
		return cephError(ret)
	}
	return nil
}

// SyncFs synchronizes all filesystem data to persistent media.
func (mount *MountInfo) SyncFs() error {
	ret := C.ceph_sync_fs(mount.mount)
	if ret != 0 {
		log.Errorf("Mount: Failed to sync filesystem")
		return cephError(ret)
	}
	return nil
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
	if ret != 0 {
		log.Errorf("ChangeDir: Failed to change directory")
		return cephError(ret)
	}
	return nil
}

// MakeDir creates a directory.
func (mount *MountInfo) MakeDir(path string, mode uint32) error {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	ret := C.ceph_mkdir(mount.mount, cPath, C.mode_t(mode))
	if ret != 0 {
		log.Errorf("MakeDir: Failed to make directory %s", path)
		return cephError(ret)
	}
	return nil
}

// RemoveDir removes a directory.
func (mount *MountInfo) RemoveDir(path string) error {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	ret := C.ceph_rmdir(mount.mount, cPath)
	if ret != 0 {
		log.Errorf("RemoveDir: Failed to remove directory")
		return cephError(ret)
	}
	return nil
}

// Chmod changes the mode bits (permissions) of a file/directory.
func (mount *MountInfo) Chmod(path string, mode uint32) error {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	ret := C.ceph_chmod(mount.mount, cPath, C.mode_t(mode))
	if ret != 0 {
		log.Errorf("Chmod: Failed to chmod :%s", path)
		return cephError(ret)
	}
	return nil
}

// Fchmod changes the mode bits (permissions) of an open file
func (mount *MountInfo) Fchmod(fd int, mode uint32) error {
	ret := C.ceph_fchmod(mount.mount, C.int(fd), C.mode_t(mode))
	if ret != 0 {
		log.Errorf("Fchmod: Failed to fchmod")
		return cephError(ret)
	}
	return nil
}

// Chown changes the ownership of a file/directory.
func (mount *MountInfo) Chown(path string, user uint32, group uint32) error {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	ret := C.ceph_chown(mount.mount, cPath, C.int(user), C.int(group))
	if ret != 0 {
		log.Errorf("Chown: Failed to chown :%s", path)
		return cephError(ret)
	}
	return nil
}

// IsMounted checks mount status.
func (mount *MountInfo) IsMounted() bool {
	ret := C.ceph_is_mounted(mount.mount)
	return ret == 1
}

// Create and/or open a file
func (mount *MountInfo) Open(path string, flags int, mode uint32) (int, error) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	ret := C.ceph_open(mount.mount, cPath, C.int(flags), C.mode_t(mode))
	if ret < 0 {
		log.Errorf("Open: Failed to open: %s", path)
		return int(ret), cephError(ret)
	}
	return int(ret), nil
}

// Close the open file
func (mount *MountInfo) Close(fd int) error {
	ret := C.ceph_close(mount.mount, C.int(fd))
	if ret < 0 {
		log.Errorf("Close: Failed to close")
		return cephError(ret)
	}
	return nil
}

// Lseek seek to a position in a file
func (mount *MountInfo) Lseek(fd int, offset int64, whence int) error {
	ret := C.ceph_lseek(mount.mount, C.int(fd), C.int64_t(offset), C.int(whence))
	if ret < 0 {
		log.Errorf("Lseek: Failed to lseek")
		return cephError(ret)
	}
	return nil
}

// Read read data from the file
func (mount *MountInfo) Read(fd int, size uint64, offset uint64) ([]byte, error) {
	buf := C.malloc(C.sizeof_char * C.uint64_t(size))
	defer C.free(unsafe.Pointer(buf))

	ret := C.ceph_read(mount.mount, C.int(fd), (*C.char)(unsafe.Pointer(buf)), C.int64_t(size), C.int64_t(offset))
	if ret < 0 {
		log.Errorf("Read: Failed to read")
		return nil, cephError(ret)
	}

	b := C.GoBytes(buf, ret)
	return b, nil
}

// Write write data to a file
func (mount *MountInfo) Write(fd int, data []byte, size uint64, offset uint64) (int, error) {
	buf := C.CBytes(data)
	defer C.free(unsafe.Pointer(buf))

	ret := C.ceph_write(mount.mount, C.int(fd), (*C.char)(buf), C.int64_t(size), C.int64_t(offset))
	if ret < 0 {
		log.Errorf("Write: Failed to write")
		return int(ret), cephError(ret)
	}
	return int(ret), nil
}

// ListDir list the contents of a directory
func (mount *MountInfo) ListDir(path string) ([]string, error) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	dir := &DirResult{}

	ret := C.ceph_opendir(mount.mount, cPath, (**C.struct_ceph_dir_result)(&dir.dirp))
	if ret != 0 {
		log.Errorf("ListDir: Failed to open dir: %s", path)
		return nil, cephError(ret)
	}

	var dirs []string

	bufLen := 256
	buf := C.malloc(C.sizeof_char * C.size_t(bufLen))
	defer C.free(unsafe.Pointer(buf))

	for {
		ret = C.ceph_getdnames(mount.mount, dir.dirp, (*C.char)(buf), C.int(bufLen))
		if ret == -C.ERANGE {
			C.free(unsafe.Pointer(buf))
			bufLen *= 2
			buf = C.malloc(C.sizeof_char * C.size_t(bufLen))
			continue
		}
		if ret <= 0 {
			break
		}

		bufpos := 0
		ent := ""
		for bufpos < int(ret) {
			ptr := unsafe.Pointer((uintptr(buf) + uintptr(bufpos)))
			ent = C.GoString((*C.char)(ptr))
			if strings.Compare(ent, ".") != 0 && strings.Compare(ent, "..") != 0 {
				dirs = append(dirs, ent)
			}
			bufpos += len(ent) + 1
		}
	}

	ret = C.ceph_closedir(mount.mount, dir.dirp)
	if ret != 0 {
		log.Errorf("ListDir: Failed to close dir: %s", path)
		return nil, cephError(ret)
	}

	return dirs, nil
}

func (mount *MountInfo) fillCephStat(stx Statx) CephStat {
	var stat CephStat
	stat.Mode = uint16(stx.stx.stx_mode)
	stat.Uid = uint(stx.stx.stx_uid)
	stat.Gid = uint(stx.stx.stx_gid)
	stat.Size = uint64(stx.stx.stx_size)
	stat.BlkSize = uint(stx.stx.stx_blksize)
	stat.Blocks = uint64(stx.stx.stx_blocks)

	time := stx.stx.stx_atime.tv_sec
	time *= 1000
	time += stx.stx.stx_atime.tv_nsec / 1000000
	stat.AppendTime = int64(time)

	time = stx.stx.stx_mtime.tv_sec
	time *= 1000
	time += stx.stx.stx_mtime.tv_nsec / 1000000
	stat.ModifyTime = int64(time)

	if (stx.stx.stx_mode & C.S_IFMT) == C.S_IFREG {
		stat.IsFile = true
	} else {
		stat.IsFile = false
	}

	if (stx.stx.stx_mode & C.S_IFMT) == C.S_IFDIR {
		stat.IsDir = true
	} else {
		stat.IsDir = false
	}

	if (stx.stx.stx_mode & C.S_IFMT) == C.S_IFLNK {
		stat.IsSymlink = true
	} else {
		stat.IsSymlink = false
	}

	return stat
}

// Stat get file status from a given path
func (mount *MountInfo) Stat(path string) (CephStat, error) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	stx := Statx{}
	ret := C.ceph_statx(mount.mount, cPath, &stx.stx, CephStatxMask, 0)
	// ret := C.ceph_statx(mount.mount, cPath, &stx.stx, C.CEPH_STATX_UID, 0)
	if ret != 0 {
		log.Errorf("Stat: Failed to get file status: %s", path)
		return CephStat{}, cephError(ret)
	}

	stat := mount.fillCephStat(stx)
	return stat, nil
}

// FStat get file status from a file descriptor
func (mount *MountInfo) FStat(fd int) (CephStat, error) {
	stx := Statx{}
	ret := C.ceph_fstatx(mount.mount, C.int(fd), &stx.stx, CephStatxMask, 0)
	if ret != 0 {
		log.Errorf("Stat: Failed to get file status")
		return CephStat{}, cephError(ret)
	}

	stat := mount.fillCephStat(stx)
	return stat, nil
}

// LStat get file status, without following symlinks
func (mount *MountInfo) LStat(path string) (CephStat, error) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	stx := Statx{}
	ret := C.ceph_statx(mount.mount, cPath, &stx.stx, CephStatxMask, C.AT_SYMLINK_NOFOLLOW)
	if ret != 0 {
		log.Errorf("Stat: Failed to get file status: %s", path)
		return CephStat{}, cephError(ret)
	}

	stat := mount.fillCephStat(stx)
	return stat, nil
}

// SetAttr set file attributes
func (mount *MountInfo) SetAttr(path string, stat CephStat, mask int) error {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	stx := Statx{}
	stx.stx.stx_mode = C.uint16_t(stat.Mode)
	stx.stx.stx_uid = C.uint32_t(stat.Uid)
	stx.stx.stx_gid = C.uint32_t(stat.Gid)
	stx.stx.stx_blksize = C.uint32_t(stat.BlkSize)
	stx.stx.stx_size = C.uint64_t(stat.Size)
	stx.stx.stx_blocks = C.uint64_t(stat.Blocks)
	stx.stx.stx_atime.tv_sec = C.int64_t(stat.AppendTime / 1000)
	stx.stx.stx_atime.tv_nsec = C.int64_t((stat.AppendTime % 1000) * 1000000)
	stx.stx.stx_mtime.tv_sec = C.int64_t(stat.ModifyTime / 1000)
	stx.stx.stx_mtime.tv_nsec = C.int64_t((stat.ModifyTime % 1000) * 1000000)

	ret := C.ceph_setattrx(mount.mount, cPath, &stx.stx, C.int(mask), 0)
	if ret != 0 {
		log.Errorf("SetAttr: Failed to set file attributes: %s", path)
		return cephError(ret)
	}
	return nil
}

// Truncate truncate a file to a specified length
func (mount *MountInfo) Truncate(path string, size int64) error {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	ret := C.ceph_truncate(mount.mount, cPath, C.int64_t(size))
	if ret != 0 {
		log.Errorf("Truncate: Failed to truncate file: %s", path)
		return cephError(ret)
	}
	return nil
}

// FTruncate truncate a file to a specified length
func (mount *MountInfo) FTruncate(fd int, size int64) error {
	ret := C.ceph_ftruncate(mount.mount, C.int(fd), C.int64_t(size))
	if ret != 0 {
		log.Errorf("FTruncate: Failed to truncate file")
		return cephError(ret)
	}
	return nil
}

// Link create a hard link to an existing file
func (mount *MountInfo) Link(OldPath, NewPath string) error {
	cOldPath, cNewPath := C.CString(OldPath), C.CString(NewPath)
	defer C.free(unsafe.Pointer(cOldPath))
	defer C.free(unsafe.Pointer(cNewPath))

	ret := C.ceph_link(mount.mount, cOldPath, cNewPath)
	if ret != 0 {
		log.Errorf("Link: Failed to link oldPath: %s, newPath: %s", OldPath, NewPath)
		return cephError(ret)
	}
	return nil
}

// Symlink create a symbolic link
func (mount *MountInfo) Symlink(OldPath, NewPath string) error {
	cOldPath, cNewPath := C.CString(OldPath), C.CString(NewPath)
	defer C.free(unsafe.Pointer(cOldPath))
	defer C.free(unsafe.Pointer(cNewPath))

	ret := C.ceph_symlink(mount.mount, cOldPath, cNewPath)
	if ret != 0 {
		log.Errorf("Symlink: Failed to link oldPath: %s, newPath: %s", OldPath, NewPath)
		return cephError(ret)
	}
	return nil
}

// ReadLink read the value of a symbolic link.
func (mount *MountInfo) ReadLink(path string) (string, error) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	stx := Statx{}
	ret := C.ceph_statx(mount.mount, cPath, &stx.stx, CephStatxMask, C.AT_SYMLINK_NOFOLLOW)
	if ret != 0 {
		log.Errorf("Stat: Failed to get file status: %s", path)
		return "", cephError(ret)
	}

	bufLen := stx.stx.stx_size + 1
	buf := C.malloc(C.sizeof_char * C.size_t(bufLen))
	defer C.free(unsafe.Pointer(buf))

	ret = C.ceph_readlink(mount.mount, cPath, (*C.char)(buf), C.int64_t(bufLen))
	if ret != 0 {
		log.Errorf("ReadLink: Failed to read link: %s", path)
		return "", cephError(ret)
	}

	ptr := unsafe.Pointer((uintptr(buf) + uintptr(ret)))
	buf = C.memset(ptr, 0, 1)
	linkName := C.GoString((*C.char)(buf))
	return linkName, nil
}

// Unlink delete a name from the file system
func (mount *MountInfo) Unlink(path string) error {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	ret := C.ceph_unlink(mount.mount, cPath)
	if ret != 0 {
		log.Errorf("Unlink: Failed to unlink, path: %s", path)
		return cephError(ret)
	}
	return nil
}
