package rgw

/*
#cgo LDFLAGS: -lrgw
#include <stdlib.h>
#include <sys/stat.h>
#include <rados/librgw.h>
#include <rados/rgw_file.h>

// readdir_callback.go
extern bool common_readdir_cb(const char *name, void *arg, uint64_t offset,
                       struct stat *st, uint32_t mask,
                       uint32_t flags);
*/
import "C"

import (
	"syscall"
	"unsafe"

	gopointer "github.com/mattn/go-pointer"
)

// RGW represents a librgw handle
//
// typedef void* librgw_t;
type RGW struct {
	libRGW *C.librgw_t
}

// Implements:
//   int librgw_create(librgw_t *rgw, int argc, char **argv)
func createRGW(argc C.int, argv **C.char) (*RGW, error) {
	rgw := &RGW{}
	var libRGW C.librgw_t

	ret := C.librgw_create(&libRGW, argc, argv)
	if ret == 0 {
		rgw.libRGW = &libRGW
		return rgw, nil
	}
	return nil, getError(ret)

}

// CreateRGW create a RGW instance
func CreateRGW(argv []string) (*RGW, error) {
	cargv := make([]*C.char, len(argv))
	for i := range argv {
		cargv[i] = C.CString(argv[i])
		defer C.free(unsafe.Pointer(cargv[i]))
	}

	return createRGW(C.int(len(cargv)), &cargv[0])
}

// ShutdownRGW shutdown a RGW instance
//
// Implements:
//   void librgw_shutdown(librgw_t rgw)
func ShutdownRGW(rgw *RGW) {
	C.librgw_shutdown(*rgw.libRGW)
}

// FS represents a rgw fs
//
// FS exports ceph's rgw_fs from include/rados/rgw_file.h
type FS struct {
	rgwFS *C.struct_rgw_fs
}

// MountFlag is used to control behavior of Mount()
type MountFlag uint32

const (
	// MountFlagNone keep the default behavior of Mount()
	MountFlagNone MountFlag = 0
)

// Mount the filesystem
//
// Implements:
//   int rgw_mount(librgw_t rgw, const char *uid, const char *key,
//               const char *secret, rgw_fs **fs, uint32_t flags)
func (fs *FS) Mount(rgw *RGW, uid, key, secret string, flags MountFlag) error {
	cuid := C.CString(uid)
	ckey := C.CString(key)
	csecret := C.CString(secret)

	defer C.free(unsafe.Pointer(cuid))
	defer C.free(unsafe.Pointer(ckey))
	defer C.free(unsafe.Pointer(csecret))
	ret := C.rgw_mount(*rgw.libRGW, cuid, ckey, csecret,
		&fs.rgwFS, C.uint(flags))
	if ret != 0 {
		return getError(ret)
	}

	return nil
}

// UmountFlag is used to control behavior of Umount()
type UmountFlag uint32

const (
	// UmountFlagNone keep the default behavior of Umount()
	UmountFlagNone UmountFlag = 0
)

// Umount the file system.
//
// Implements:
//   int rgw_umount(rgw_fs *fs, uint32_t flags)
func (fs *FS) Umount(flags UmountFlag) error {
	ret := C.rgw_umount(fs.rgwFS, C.uint(flags))
	if ret == 0 {
		return nil
	}
	return getError(ret)
}

// StatVFS instances are returned from the StatFS call. It reports
// file-system wide statistics.
type StatVFS struct {
	// Bsize reports the file system's block size.
	Bsize int64
	// Fragment reports the file system's fragment size.
	Frsize int64
	// Blocks reports the number of blocks in the file system.
	Blocks uint64
	// Bfree reports the number of free blocks.
	Bfree uint64
	// Bavail reports the number of free blocks for unprivileged users.
	Bavail uint64
	// Files reports the number of inodes in the file system.
	Files uint64
	// Ffree reports the number of free indoes.
	Ffree uint64
	// Favail reports the number of free indoes for unprivileged users.
	Favail uint64
	// Fsid reports the file system ID number.
	Fsid [2]int64
	// Flag reports the file system mount flags.
	Flag int64
	// Namemax reports the maximum file name length.
	Namemax int64
}

// FileHandle represents a file/dir
//
// struct rgw_file_handle
type FileHandle struct {
	handle *C.struct_rgw_file_handle
}

// GetRootFileHandle get the root file handle
func (fs *FS) GetRootFileHandle() *FileHandle {
	return &FileHandle{
		handle: fs.rgwFS.root_fh,
	}
}

// StatFSFlag is used to control behavior of StatFS()
type StatFSFlag uint32

const (
	// StatFSFlagNone keep the default behavior of StatFS()
	StatFSFlagNone StatFSFlag = 0
)

// StatFS returns file system wide statistics.
//
// Implements:
//    int rgw_statfs(rgw_fs *fs, rgw_file_handle *parent_fh,
//                   rgw_statvfs *vfs_st, uint32_t flags)
func (fs *FS) StatFS(pFH *FileHandle, flags StatFSFlag) (*StatVFS, error) {
	var statVFS C.struct_rgw_statvfs

	ret := C.rgw_statfs(fs.rgwFS, pFH.handle, &statVFS, C.uint(flags))
	if ret != 0 {
		return nil, getError(ret)
	}

	stat := &StatVFS{
		Bsize:   int64(statVFS.f_bsize),
		Frsize:  int64(statVFS.f_frsize),
		Blocks:  uint64(statVFS.f_blocks),
		Bfree:   uint64(statVFS.f_bfree),
		Bavail:  uint64(statVFS.f_bavail),
		Files:   uint64(statVFS.f_files),
		Ffree:   uint64(statVFS.f_ffree),
		Favail:  uint64(statVFS.f_favail),
		Fsid:    [2]int64{int64(statVFS.f_fsid[0]), int64(statVFS.f_fsid[1])},
		Flag:    int64(statVFS.f_flag),
		Namemax: int64(statVFS.f_namemax),
	}
	return stat, nil
}

// ReadDirCallback will be applied to each dentry when ReadDir() is called
type ReadDirCallback interface {
	Callback(name string, st *syscall.Stat_t, mask AttrMask, flags uint32, offset uint64) bool
}

//export goCommonReadDirCallback
func goCommonReadDirCallback(name *C.char, arg unsafe.Pointer, offset C.uint64_t,
	stat *C.struct_stat, mask, flags C.uint32_t) bool {

	cb := gopointer.Restore(arg).(ReadDirCallback)

	var st syscall.Stat_t
	if stat != nil {
		st = syscall.Stat_t{
			Dev:     uint64(stat.st_dev),
			Ino:     uint64(stat.st_ino),
			Nlink:   uint64(stat.st_nlink),
			Mode:    uint32(stat.st_mode),
			Uid:     uint32(stat.st_uid),
			Gid:     uint32(stat.st_gid),
			Rdev:    uint64(stat.st_rdev),
			Size:    int64(stat.st_size),
			Blksize: int64(stat.st_blksize),
			Blocks:  int64(stat.st_blocks),
			Atim: syscall.Timespec{
				Sec:  int64(stat.st_atim.tv_sec),
				Nsec: int64(stat.st_atim.tv_nsec),
			},
			Mtim: syscall.Timespec{
				Sec:  int64(stat.st_mtim.tv_sec),
				Nsec: int64(stat.st_mtim.tv_nsec),
			},
			Ctim: syscall.Timespec{
				Sec:  int64(stat.st_ctim.tv_sec),
				Nsec: int64(stat.st_ctim.tv_nsec),
			},
		}
	}
	return cb.Callback(C.GoString(name), &st, AttrMask(mask), uint32(flags), uint64(offset))
}

// ReadDirFlag is used to control behavor of ReadDir()
type ReadDirFlag uint32

const (
	// ReadDirFlagNone keep the default behavior of ReadDir()
	ReadDirFlagNone ReadDirFlag = 0
	// ReadDirFlagDotDot send dot names
	ReadDirFlagDotDot ReadDirFlag = 1
)

// ReadDir read directory content
//
// Implements:
//   int rgw_readdir(struct rgw_fs *rgw_fs,
//                 struct rgw_file_handle *parent_fh, uint64_t *offset,
//                 rgw_readdir_cb rcb, void *cb_arg, bool *eof,
//                 uint32_t flags)
func (fs *FS) ReadDir(parentHdl *FileHandle, cb ReadDirCallback, offset uint64, flags ReadDirFlag) (uint64, bool, error) {
	coffset := C.uint64_t(offset)
	var eof C.bool = false

	cbArg := gopointer.Save(cb)
	defer gopointer.Unref(cbArg)

	ret := C.rgw_readdir(fs.rgwFS, parentHdl.handle, &coffset, C.rgw_readdir_cb(C.common_readdir_cb),
		unsafe.Pointer(cbArg), &eof, C.uint(flags))
	if ret != 0 {
		return 0, false, getError(ret)
	}

	next := uint64(coffset)
	return next, bool(eof), nil
}

// Version get rgwfile version
//
// Implements:
//   void rgwfile_version(int *major, int *minor, int *extra)
func Version() (int, int, int) {
	var major, minor, extra C.int
	C.rgwfile_version(&major, &minor, &extra)
	return int(major), int(minor), int(extra)
}

// AttrMask specifies which part(s) of attrs is/are concerned
type AttrMask uint32

const (
	// AttrMode mode is concerned
	AttrMode AttrMask = 1
	// AttrUid uid is concerned
	AttrUid AttrMask = 2
	// AttrGid gid is concerned
	AttrGid AttrMask = 4
	// AttrMtime mtime is concerned
	AttrMtime AttrMask = 8
	// AttrAtime atime is concerned
	AttrAtime AttrMask = 16
	// AttrSize size is concerned
	AttrSize AttrMask = 32
	// AttrCtime ctime is concerned
	AttrCtime AttrMask = 64
)

// LookupFlag is used to control behavor of Lookup()
type LookupFlag uint32

const (
	// LookupFlagNone keep the default behavior of Lookup()
	LookupFlagNone LookupFlag = 0
	// LookupFlagCreate create if not exist
	LookupFlagCreate LookupFlag = 1
	// LookupFlagRCB readdir callback hint
	LookupFlagRCB LookupFlag = 2
	// LookupFlagDir FileHandle type is dir
	LookupFlagDir LookupFlag = 4
	// LookupFlagFile FileHandle type is file
	LookupFlagFile LookupFlag = 8
	// LookupTypeFlags FileHandle type is file or dir
	LookupTypeFlags LookupFlag = LookupFlagDir | LookupFlagFile
)

// Lookup object by name (POSIX style)
//
// Implements:
//    int rgw_lookup(rgw_fs *fs,
//                   rgw_file_handle *parent_fh, const char *path,
//                   rgw_file_handle **fh, stat* st, uint32_t st_mask,
//                   uint32_t flags)
func (fs *FS) Lookup(parentHdl *FileHandle, path string, stMask AttrMask,
	flags LookupFlag) (*FileHandle, *syscall.Stat_t, error) {
	fh := &FileHandle{}
	var stat C.struct_stat

	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	ret := C.rgw_lookup(fs.rgwFS, parentHdl.handle, cPath, &fh.handle, &stat,
		C.uint32_t(stMask), C.uint32_t(flags))
	if ret == 0 {
		st := syscall.Stat_t{
			Dev:     uint64(stat.st_dev),
			Ino:     uint64(stat.st_ino),
			Nlink:   uint64(stat.st_nlink),
			Mode:    uint32(stat.st_mode),
			Uid:     uint32(stat.st_uid),
			Gid:     uint32(stat.st_gid),
			Rdev:    uint64(stat.st_rdev),
			Size:    int64(stat.st_size),
			Blksize: int64(stat.st_blksize),
			Blocks:  int64(stat.st_blocks),
			Atim: syscall.Timespec{
				Sec:  int64(stat.st_atim.tv_sec),
				Nsec: int64(stat.st_atim.tv_nsec),
			},
			Mtim: syscall.Timespec{
				Sec:  int64(stat.st_mtim.tv_sec),
				Nsec: int64(stat.st_mtim.tv_nsec),
			},
			Ctim: syscall.Timespec{
				Sec:  int64(stat.st_ctim.tv_sec),
				Nsec: int64(stat.st_ctim.tv_nsec),
			},
		}
		return fh, &st, nil
	}
	return nil, nil, getError(ret)
}

// CreateFlag is used to control behavor of Create()
type CreateFlag uint32

const (
	// CreateFlagNone keep the default behavior of Create()
	CreateFlagNone CreateFlag = 0
)

// Create file
//
// Implements:
//    int rgw_create(rgw_fs *fs, rgw_file_handle *parent_fh,
//                   const char *name, stat *st, uint32_t mask,
//                   rgw_file_handle **fh, uint32_t posix_flags,
//                   uint32_t flags)
func (fs *FS) Create(parentHdl *FileHandle, name string, mask AttrMask,
	posixFlags uint32, flags CreateFlag) (
	*FileHandle, *syscall.Stat_t, error) {

	fh := &FileHandle{}
	var stat C.struct_stat

	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	ret := C.rgw_create(fs.rgwFS, parentHdl.handle, cName, &stat, C.uint32_t(mask), &fh.handle,
		C.uint32_t(posixFlags), C.uint32_t(flags))
	if ret == 0 {
		st := syscall.Stat_t{
			Dev:     uint64(stat.st_dev),
			Ino:     uint64(stat.st_ino),
			Nlink:   uint64(stat.st_nlink),
			Mode:    uint32(stat.st_mode),
			Uid:     uint32(stat.st_uid),
			Gid:     uint32(stat.st_gid),
			Rdev:    uint64(stat.st_rdev),
			Size:    int64(stat.st_size),
			Blksize: int64(stat.st_blksize),
			Blocks:  int64(stat.st_blocks),
			Atim: syscall.Timespec{
				Sec:  int64(stat.st_atim.tv_sec),
				Nsec: int64(stat.st_atim.tv_nsec),
			},
			Mtim: syscall.Timespec{
				Sec:  int64(stat.st_mtim.tv_sec),
				Nsec: int64(stat.st_mtim.tv_nsec),
			},
			Ctim: syscall.Timespec{
				Sec:  int64(stat.st_ctim.tv_sec),
				Nsec: int64(stat.st_ctim.tv_nsec),
			},
		}
		return fh, &st, nil
	}
	return nil, nil, getError(ret)
}

// MkdirFlag is used to control behavor of Mkdir()
type MkdirFlag uint32

const (
	// MkdirFlagNone keep the default behavior of Mkdir()
	MkdirFlagNone MkdirFlag = 0
)

// Mkdir creates a new directory
//
// Implements:
//    int rgw_mkdir(rgw_fs *fs,
//                  rgw_file_handle *parent_fh,
//                  const char *name, stat *st, uint32_t mask,
//                  rgw_file_handle **fh, uint32_t flags)
//
func (fs *FS) Mkdir(parentHdl *FileHandle, name string, mask AttrMask, flags MkdirFlag) (
	*FileHandle, *syscall.Stat_t, error) {

	fh := &FileHandle{}
	var stat C.struct_stat

	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	ret := C.rgw_mkdir(fs.rgwFS, parentHdl.handle, cName, &stat, C.uint32_t(mask), &fh.handle,
		C.uint32_t(flags))
	if ret == 0 {
		st := syscall.Stat_t{
			Dev:     uint64(stat.st_dev),
			Ino:     uint64(stat.st_ino),
			Nlink:   uint64(stat.st_nlink),
			Mode:    uint32(stat.st_mode),
			Uid:     uint32(stat.st_uid),
			Gid:     uint32(stat.st_gid),
			Rdev:    uint64(stat.st_rdev),
			Size:    int64(stat.st_size),
			Blksize: int64(stat.st_blksize),
			Blocks:  int64(stat.st_blocks),
			Atim: syscall.Timespec{
				Sec:  int64(stat.st_atim.tv_sec),
				Nsec: int64(stat.st_atim.tv_nsec),
			},
			Mtim: syscall.Timespec{
				Sec:  int64(stat.st_mtim.tv_sec),
				Nsec: int64(stat.st_mtim.tv_nsec),
			},
			Ctim: syscall.Timespec{
				Sec:  int64(stat.st_ctim.tv_sec),
				Nsec: int64(stat.st_ctim.tv_nsec),
			},
		}
		return fh, &st, nil
	}
	return nil, nil, getError(ret)
}

// WriteFlag is used to control behavor of Write()
type WriteFlag uint32

const (
	// WriteFlagNone keep the default behavior of Write()
	WriteFlagNone WriteFlag = 0
)

// Write data to file
//
// Implements:
//    int rgw_write(rgw_fs *fs,
//                  rgw_file_handle *fh, uint64_t offset,
//                  size_t length, size_t *bytes_written, void *buffer,
//                  uint32_t flags)
//
func (fs *FS) Write(fh *FileHandle, buffer []byte, offset uint64, length uint64,
	flags uint32) (bytesWritten uint, err error) {
	var written C.size_t

	ret := C.rgw_write(fs.rgwFS, fh.handle, C.uint64_t(offset),
		C.size_t(length), &written, unsafe.Pointer(&buffer[0]),
		C.uint32_t(flags))
	if ret == 0 {
		return uint(written), nil
	}
	return uint(written), getError(ret)

}

// ReadFlag is used to control behavior of Read()
type ReadFlag uint32

const (
	// ReadFlagNone keep the default behavior of Read()
	ReadFlagNone ReadFlag = 0
)

// Read data from file
//
// Implements:
//    int rgw_read(rgw_fs *fs,
//                 rgw_file_handle *fh, uint64_t offset,
//                 size_t length, size_t *bytes_read, void *buffer,
//                 uint32_t flags)
func (fs *FS) Read(fh *FileHandle, offset, length uint64, buffer []byte,
	flags ReadFlag) (bytes_read uint64, err error) {
	var cbytes_read C.size_t
	bufptr := unsafe.Pointer(&buffer[0])
	ret := C.rgw_read(fs.rgwFS, fh.handle, C.uint64_t(offset),
		C.size_t(length), &cbytes_read, bufptr,
		C.uint32_t(flags))
	if ret == 0 {
		return uint64(cbytes_read), nil
	}
	return 0, getError(ret)
}

// OpenFlag is used to control behavior of Open()
type OpenFlag uint32

const (
	// OpenFlagNone keep the default behavior of Open()
	OpenFlagNone OpenFlag = 0
	// OpenFlagCreate create if not exist
	OpenFlagCreate OpenFlag = 1
	// OpenFlagV3 ops have v3 semantics
	OpenFlagV3 OpenFlag = 2
	// OpenFlagStateless alias of OpenFlagV3
	OpenFlagStateless OpenFlag = 2
)

// Open file
//
// Implements:
//   int rgw_open(struct rgw_fs *rgw_fs,
//             struct rgw_file_handle *fh, uint32_t posix_flags, uint32_t flags)
func (fs *FS) Open(fh *FileHandle, posixFlags uint32, flags OpenFlag) error {
	ret := C.rgw_open(fs.rgwFS, fh.handle, C.uint32_t(posixFlags),
		C.uint32_t(flags))
	if ret == 0 {
		return nil
	}
	return getError(ret)
}

// CloseFlag is used to control behavior of Close()
type CloseFlag uint32

const (
	// CloseFlagNone keep the default behavior of Close()
	CloseFlagNone CloseFlag = 0
	// CloseFlagRele also decreases reference count of related FileHandle
	CloseFlagRele CloseFlag = 1
)

// Close file
//
// Implements:
//    int rgw_close(rgw_fs *fs, rgw_file_handle *fh,
//                  uint32_t flags)
func (fs *FS) Close(fh *FileHandle, flags CloseFlag) error {
	ret := C.rgw_close(fs.rgwFS, fh.handle, C.uint32_t(flags))
	if ret == 0 {
		return nil
	}
	return getError(ret)
}

// FsyncFlag is used to control behavior of Fsync()
type FsyncFlag uint32

const (
	// FsyncFlagNone keep the default behavior of Fsync()
	FsyncFlagNone FsyncFlag = 0
)

// Fsync sync written data.
// NOTE: Actually, do nothing
//
// Implements:
//    int rgw_fsync(rgw_fs *fs, rgw_file_handle *fh,
//                  uint32_t flags)
func (fs *FS) Fsync(fh *FileHandle, flags FsyncFlag) error {
	ret := C.rgw_fsync(fs.rgwFS, fh.handle, C.uint32_t(flags))
	if ret == 0 {
		return nil
	}
	return getError(ret)
}

// CommitFlag is used to control behavior of Commit()
type CommitFlag uint32

const (
	// CommitFlagNone keep the default behavior of Commit()
	CommitFlagNone CommitFlag = 0
)

// Commit nfs commit operation
//
// Implements:
//   int rgw_commit(struct rgw_fs *rgw_fs, struct rgw_file_handle *fh,
//               uint64_t offset, uint64_t length, uint32_t flags)
func (fs *FS) Commit(fh *FileHandle, offset, length uint64, flags CommitFlag) error {
	ret := C.rgw_commit(fs.rgwFS, fh.handle, C.uint64_t(offset),
		C.uint64_t(length), C.uint32_t(flags))
	if ret == 0 {
		return nil
	}
	return getError(ret)
}

// TruncFlag is used to control behavior of Truncate()
type TruncFlag uint32

const (
	// TruncFlagNone keep the default behavior of Truncate()
	TruncFlagNone TruncFlag = 0
)

// Truncate file.
// Actually, do nothing
//
// Implements:
//   int rgw_truncate(rgw_fs *fs, rgw_file_handle *fh, uint64_t size, uint32_t flags)
func (fs *FS) Truncate(fh *FileHandle, size uint64, flags TruncFlag) error {
	ret := C.rgw_truncate(fs.rgwFS, fh.handle, C.uint64_t(size), C.uint32_t(flags))
	if ret == 0 {
		return nil
	}
	return getError(ret)
}

// UnlinkFlag is used to control behavior of Unlink()
type UnlinkFlag uint32

const (
	// UnlinkFlagNone keep the default behavior of Unlink()
	UnlinkFlagNone UnlinkFlag = 0
)

// Unlink remove file or directory
//
// Implements:
//    int rgw_unlink(rgw_fs *fs,
//                   rgw_file_handle *parent_fh, const char* path,
//                   uint32_t flags)
func (fs *FS) Unlink(parentHdl *FileHandle, path string, flags UnlinkFlag) error {
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))

	ret := C.rgw_unlink(fs.rgwFS, parentHdl.handle, cpath, C.uint32_t(flags))
	if ret == 0 {
		return nil
	}
	return getError(ret)
}

// RenameFlag is used to control behavior of Rename()
type RenameFlag uint32

const (
	// RenameFlagNone keep the default behavior of Rename()
	RenameFlagNone RenameFlag = 0
)

// Rename object
//
// Implements:
//    int rgw_rename(rgw_fs *fs,
//                   rgw_file_handle *olddir, const char* old_name,
//                   rgw_file_handle *newdir, const char* new_name,
//                   uint32_t flags)
func (fs *FS) Rename(oldDirHdl *FileHandle, oldName string,
	newDirHdl *FileHandle, newName string, flags RenameFlag) error {
	cOldName := C.CString(oldName)
	defer C.free(unsafe.Pointer(cOldName))
	cNewName := C.CString(newName)
	defer C.free(unsafe.Pointer(cNewName))

	ret := C.rgw_rename(fs.rgwFS, oldDirHdl.handle, cOldName,
		newDirHdl.handle, cNewName, C.uint32_t(flags))
	if ret == 0 {
		return nil
	}
	return getError(ret)
}

// GetAttrFlag is used to control behavior of GetAttr()
type GetAttrFlag uint32

const (
	// GetAttrFlagNone keep the default behavior of GetAttr()
	GetAttrFlagNone GetAttrFlag = 0
)

// GetAttr gets unix attributes for object
//
// Implements:
//    int rgw_getattr(rgw_fs *fs,
//                    rgw_file_handle *fh, stat *st,
//                    uint32_t flags)
func (fs *FS) GetAttr(fh *FileHandle, flags GetAttrFlag) (*syscall.Stat_t, error) {
	var stat C.struct_stat
	ret := C.rgw_getattr(fs.rgwFS, fh.handle, &stat, C.uint32_t(flags))
	if ret == 0 {
		st := syscall.Stat_t{
			Dev:     uint64(stat.st_dev),
			Ino:     uint64(stat.st_ino),
			Nlink:   uint64(stat.st_nlink),
			Mode:    uint32(stat.st_mode),
			Uid:     uint32(stat.st_uid),
			Gid:     uint32(stat.st_gid),
			Rdev:    uint64(stat.st_rdev),
			Size:    int64(stat.st_size),
			Blksize: int64(stat.st_blksize),
			Blocks:  int64(stat.st_blocks),
			Atim: syscall.Timespec{
				Sec:  int64(stat.st_atim.tv_sec),
				Nsec: int64(stat.st_atim.tv_nsec),
			},
			Mtim: syscall.Timespec{
				Sec:  int64(stat.st_mtim.tv_sec),
				Nsec: int64(stat.st_mtim.tv_nsec),
			},
			Ctim: syscall.Timespec{
				Sec:  int64(stat.st_ctim.tv_sec),
				Nsec: int64(stat.st_ctim.tv_nsec),
			},
		}
		return &st, nil
	}
	return nil, getError(ret)
}

// SetAttrFlag is used to control behavior of SetAttr()
type SetAttrFlag uint32

const (
	// SetAttrFlagNone keep the default behavior of SetAttr()
	SetAttrFlagNone SetAttrFlag = 0
)

// SetAttr sets unix attributes for object
//
// Implements:
//    int rgw_setattr(rgw_fs *fs, rgw_file_handle *fh, stat *st,
//                    uint32_t mask, uint32_t flags)
func (fs *FS) SetAttr(fh *FileHandle, stat *syscall.Stat_t, mask AttrMask, flags SetAttrFlag) error {
	st := C.struct_stat{
		st_dev:     C.uint64_t(stat.Dev),
		st_ino:     C.uint64_t(stat.Ino),
		st_nlink:   C.uint64_t(stat.Nlink),
		st_mode:    C.uint32_t(stat.Mode),
		st_uid:     C.uint32_t(stat.Uid),
		st_gid:     C.uint32_t(stat.Gid),
		st_rdev:    C.uint64_t(stat.Rdev),
		st_size:    C.int64_t(stat.Size),
		st_blksize: C.int64_t(stat.Blksize),
		st_blocks:  C.int64_t(stat.Blocks),
		st_atim: C.struct_timespec{
			tv_sec:  C.long(stat.Atim.Sec),
			tv_nsec: C.long(stat.Atim.Nsec),
		},
		st_mtim: C.struct_timespec{
			tv_sec:  C.long(stat.Mtim.Sec),
			tv_nsec: C.long(stat.Mtim.Nsec),
		},
		st_ctim: C.struct_timespec{
			tv_sec:  C.long(stat.Ctim.Sec),
			tv_nsec: C.long(stat.Ctim.Nsec),
		},
	}

	ret := C.rgw_setattr(fs.rgwFS, fh.handle, &st, C.uint32_t(mask),
		C.uint32_t(flags))
	if ret == 0 {
		return nil
	}
	return getError(ret)
}
