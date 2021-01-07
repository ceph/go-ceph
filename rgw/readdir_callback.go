package rgw

/*
#cgo LDFLAGS: -lrgw
#include <stdlib.h>
#include <sys/stat.h>
#include <rados/librgw.h>
#include <rados/rgw_file.h>

extern bool goCommonReadDirCallback(char *name, void* arg, uint64_t offset,
                                    struct stat *st, uint32_t mask, uint32_t flags);
bool common_readdir_cb(const char *name, void *arg, uint64_t offset,
                       struct stat *st, uint32_t mask,
                       uint32_t flags) {
  return goCommonReadDirCallback((char *)name, arg, offset, st, mask, flags);
}
*/
import "C"
