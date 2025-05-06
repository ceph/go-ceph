//go:build ceph_preview

package cephfs

/*
#cgo LDFLAGS: -lcephfs
#cgo CPPFLAGS: -D_FILE_OFFSET_BITS=64
#include <stdlib.h>
#include <dirent.h>
#include <cephfs/libcephfs.h>

// ceph_file_blockdiff_init_fn matches the ceph_file_blockdiff_init function signature.
typedef int(*ceph_file_blockdiff_init_fn)(struct ceph_mount_info* cmount,
                                  const char* root_path,
                                  const char* rel_path,
                                  const char* snap1,
                                  const char* snap2,
                                  struct ceph_file_blockdiff_info* out_info);

// ceph_file_blockdiff_init_dlsym take *fn as ceph_file_blockdiff_init and calls the dynamically loaded
// ceph_file_blockdiff_init function passed as 1st argument.
static inline int ceph_file_blockdiff_init_dlsym(void *fn,
                                  struct ceph_mount_info* cmount,
                                  const char* root_path,
                                  const char* rel_path,
                                  const char* snap1,
                                  const char* snap2,
                                  struct ceph_file_blockdiff_info* out_info) {
	// cast function pointer fn to ceph_file_blockdiff_init and call the function
	return ((ceph_file_blockdiff_init_fn) fn)(cmount, root_path, rel_path, snap1, snap2, out_info);
}

// ceph_file_blockdiff_fn matches the ceph_file_blockdiff function signature.
typedef int(*ceph_file_blockdiff_fn)(struct ceph_file_blockdiff_info* info,
                                  struct ceph_file_blockdiff_changedblocks* blocks);

// ceph_file_blockdiff_dlsym take *fn as ceph_file_blockdiff and calls the dynamically loaded
// ceph_file_blockdiff function passed as 1st argument.
static inline int ceph_file_blockdiff_dlsym(void *fn,
                                  struct ceph_file_blockdiff_info* info,
                                  struct ceph_file_blockdiff_changedblocks* blocks) {
	// cast function pointer fn to ceph_file_blockdiff and call the function
	return ((ceph_file_blockdiff_fn) fn)(info, blocks);
}

// ceph_free_file_blockdiff_buffer_fn matches the ceph_free_file_blockdiff_buffer function signature.
typedef void(*ceph_free_file_blockdiff_buffer_fn)(struct ceph_file_blockdiff_changedblocks* blocks);

// ceph_free_file_blockdiff_buffer_dlsym take *fn as ceph_free_file_blockdiff_buffer and calls the dynamically loaded
// ceph_free_file_blockdiff_buffer function passed as 1st argument.
static inline void ceph_free_file_blockdiff_buffer_dlsym(void *fn,
                                  struct ceph_file_blockdiff_changedblocks* blocks) {
	// cast function pointer fn to ceph_free_file_blockdiff_buffer and call the function
	((ceph_free_file_blockdiff_buffer_fn) fn)(blocks);
}

// ceph_file_blockdiff_finish_fn matches the ceph_file_blockdiff_finish function signature.
typedef int(*ceph_file_blockdiff_finish_fn)(struct ceph_file_blockdiff_info* info);

// ceph_file_blockdiff_finish_dlsym take *fn as ceph_file_blockdiff_finish and calls the dynamically loaded
// ceph_file_blockdiff_finish function passed as 1st argument.
static inline int ceph_file_blockdiff_finish_dlsym(void *fn,
                                  struct ceph_file_blockdiff_info* info) {
	// cast function pointer fn to ceph_file_blockdiff_finish and call the function
	return ((ceph_file_blockdiff_finish_fn) fn)(info);
}
*/
import "C"

import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/ceph/go-ceph/internal/dlsym"
)

var (
	cephFileBlockDiffInitOnce       sync.Once
	cephFileBlockDiffOnce           sync.Once
	cephFreeFileBlockDiffBufferOnce sync.Once
	cephFileBlockDiffFinishOnce     sync.Once

	cephFileBlockDiffInitErr       error
	cephFileBlockDiffErr           error
	cephFreeFileBlockDiffBufferErr error
	cephFileBlockDiffFinishErr     error

	cephFileBlockDiffInit       unsafe.Pointer
	cephFileBlockDiff           unsafe.Pointer
	cephFreeFileBlockDiffBuffer unsafe.Pointer
	cephFileBlockDiffFinish     unsafe.Pointer
)

// cephFileBlockDiffInfo is a struct that holds the block diff stream handle.
// struct ceph_file_blockdiff_info
//
//	{
//	  struct ceph_mount_info* cmount;
//	  struct ceph_file_blockdiff_result* blockp;
//	};
type CephFileBlockDiffInfo struct {
	CMount                  *MountInfo
	CephFileBlockDiffResult *C.struct_ceph_file_blockdiff_result
}

// CBlock is a struct that holds the offset and length of a block.
// struct cblock
//
//	{
//	  uint64_t offset;
//	  uint64_t len;
//	};
type CBlock struct {
	Offset uint64
	Len    uint64
}

// CephFileBlockDiffChangedBlocks is a struct that holds the number of blocks
// and list of CBlocks.
// struct ceph_file_blockdiff_changedblocks
//
//	{
//	  uint64_t num_blocks;
//	  struct cblock *b;
//	};
type CephFileBlockDiffChangedBlocks struct {
	NumBlocks uint64
	CBlocks   []CBlock
}

// CephFileBlockDiffInit initializes the block diff stream to get file block deltas.
// It takes the mount handle, root path, relative path, snapshot names and returns
// a CephFileBlockDiffInfo struct that contains the block diff stream handle.
// It returns an error if the initialization fails.
// Implements:
//
//	    int ceph_file_blockdiff_init(
//		    struct ceph_mount_info* cmount,
//			const char* root_path,
//			const char* rel_path,
//			const char* snap1,
//			const char* snap2,
//			struct ceph_file_blockdiff_info* out_info);
func CephFileBlockDiffInit(mount *MountInfo, rootPath, relPath, snap1, snap2 string) (*CephFileBlockDiffInfo, error) {
	cRootPath := C.CString(rootPath)
	cRelPath := C.CString(relPath)
	cSnap1 := C.CString(snap1)
	cSnap2 := C.CString(snap2)

	// Load the ceph_file_blockdiff_init function from the shared library.
	cephFileBlockDiffInitOnce.Do(func() {
		cephFileBlockDiffInit, cephFileBlockDiffInitErr = dlsym.LookupSymbol("ceph_file_blockdiff_init")
	})
	if cephFileBlockDiffInitErr != nil {
		return nil, fmt.Errorf("%w: %w", ErrNotImplemented, cephFileBlockDiffInitErr)
	}

	rawCephBlockDiffInfo := &C.struct_ceph_file_blockdiff_info{}

	// Call the ceph_file_blockdiff_init function with the provided arguments.
	ret := C.ceph_file_blockdiff_init_dlsym(cephFileBlockDiffInit,
		mount.mount,
		cRootPath,
		cRelPath,
		cSnap1,
		cSnap2,
		rawCephBlockDiffInfo,
	)
	if ret != 0 {
		return nil, getError(ret)
	}

	cephFileBlockDiffInfo := &CephFileBlockDiffInfo{
		CMount:                  &MountInfo{mount: rawCephBlockDiffInfo.cmount},
		CephFileBlockDiffResult: rawCephBlockDiffInfo.blockp,
	}

	return cephFileBlockDiffInfo, nil
}

// CephFileBlockDiff retrieves the next set of file block diffs.
// It takes a CephFileBlockDiffInfo struct and returns a CephFileBlockDiffChangedBlocks
// struct that contains the number of blocks and list of CBlocks.
// It returns an error if the retrieval fails.
//
// Implements:
//
// int ceph_file_blockdiff(struct ceph_file_blockdiff_info* info,
//
//	struct ceph_file_blockdiff_changedblocks* blocks);
func CephFileBlockDiff(info *CephFileBlockDiffInfo) (*CephFileBlockDiffChangedBlocks, error) {
	// Load the ceph_file_blockdiff function from the shared library.
	cephFileBlockDiffOnce.Do(func() {
		cephFileBlockDiff, cephFileBlockDiffErr = dlsym.LookupSymbol("ceph_file_blockdiff")
	})
	if cephFileBlockDiffErr != nil {
		return nil, fmt.Errorf("%w: %w", ErrNotImplemented, cephFileBlockDiffErr)
	}

	rawCephFileBlockDiffInfo := &C.struct_ceph_file_blockdiff_info{
		cmount: info.CMount.mount,
		blockp: info.CephFileBlockDiffResult,
	}
	// rawCBlocks := make([]C.struct_cblock{}, 0)
	rawCephBlockDiffChangedBlocks := &C.struct_ceph_file_blockdiff_changedblocks{}

	// Call the ceph_file_blockdiff function with the provided arguments.
	ret := C.ceph_file_blockdiff_dlsym(cephFileBlockDiff,
		rawCephFileBlockDiffInfo,
		rawCephBlockDiffChangedBlocks,
	)
	if ret != 0 {
		return nil, getError(ret)
	}

	// Convert the C struct to Go struct.
	cBlocks := make([]CBlock, rawCephBlockDiffChangedBlocks.num_blocks)
	if rawCephBlockDiffChangedBlocks.num_blocks == 0 {
		return &CephFileBlockDiffChangedBlocks{
			NumBlocks: 0,
			CBlocks:   cBlocks,
		}, nil
	}

	currentCBlock := (*C.struct_cblock)(unsafe.Pointer(rawCephBlockDiffChangedBlocks.b))
	for i := uint64(0); i < uint64(rawCephBlockDiffChangedBlocks.num_blocks); i++ {
		cBlocks[i] = CBlock{
			Offset: uint64(currentCBlock.offset),
			Len:    uint64(currentCBlock.len),
		}
		currentCBlock = (*C.struct_cblock)(unsafe.Pointer(uintptr(unsafe.Pointer(currentCBlock)) + unsafe.Sizeof(C.struct_cblock{})))
	}

	// Free the C struct.
	err := CephFreeFileBlockDiffBuffer(rawCephBlockDiffChangedBlocks)
	if err != nil {
		return nil, fmt.Errorf("failed to free block diff buffer: %w", err)
	}

	return &CephFileBlockDiffChangedBlocks{
		NumBlocks: uint64(rawCephBlockDiffChangedBlocks.num_blocks),
		CBlocks:   cBlocks,
	}, nil
}

// CephFreeFileBlockDiffBuffer frees the block diff buffer.
// It takes a CephFileBlockDiffChangedBlocks struct and returns an error
// if the freeing fails.
//
// Implements:
//
// void ceph_free_file_blockdiff_buffer(struct ceph_file_blockdiff_changedblocks* blocks);
func CephFreeFileBlockDiffBuffer(rawCephFileBlockDiffChangedBlocks *C.struct_ceph_file_blockdiff_changedblocks) error {
	// Load the ceph_free_file_blockdiff_buffer function from the shared library.
	cephFreeFileBlockDiffBufferOnce.Do(func() {
		cephFreeFileBlockDiffBuffer, cephFreeFileBlockDiffBufferErr = dlsym.LookupSymbol("ceph_free_file_blockdiff_buffer")
	})
	if cephFreeFileBlockDiffBufferErr != nil {
		return fmt.Errorf("%w: %w", ErrNotImplemented, cephFreeFileBlockDiffBufferErr)
	}

	// Call the ceph_free_file_blockdiff_buffer function with the provided arguments.
	C.ceph_free_file_blockdiff_buffer_dlsym(
		cephFreeFileBlockDiffBuffer,
		rawCephFileBlockDiffChangedBlocks)

	return nil
}

// CephFileBlockDiffFinish closes the block diff stream.
// It takes a CephFileBlockDiffInfo struct and returns an error
// if the closing fails.
//
// Implements:
//
// int ceph_file_blockdiff_finish(struct ceph_file_blockdiff_info* info);
func CephFileBlockDiffFinish(info *CephFileBlockDiffInfo) error {
	// Load the ceph_file_blockdiff_finish function from the shared library.
	cephFileBlockDiffFinishOnce.Do(func() {
		cephFileBlockDiffFinish, cephFileBlockDiffFinishErr = dlsym.LookupSymbol("ceph_file_blockdiff_finish")
	})
	if cephFileBlockDiffFinishErr != nil {
		return fmt.Errorf("%w: %w", ErrNotImplemented, cephFileBlockDiffFinishErr)
	}

	rawCephFileBlockDiffInfo := &C.struct_ceph_file_blockdiff_info{
		cmount: info.CMount.mount,
		blockp: info.CephFileBlockDiffResult,
	}

	// Call the ceph_file_blockdiff_finish function with the provided arguments.
	ret := C.ceph_file_blockdiff_finish_dlsym(
		cephFileBlockDiffFinish,
		rawCephFileBlockDiffInfo,
	)
	if ret != 0 {
		return getError(ret)
	}

	return nil
}
