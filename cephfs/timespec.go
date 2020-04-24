package cephfs

/*
#include <time.h>
*/
import "C"

import (
	"golang.org/x/sys/unix"
)

// Timespec behaves similarly to C's struct timespec.
// Timespec is used to retain fidelity to the C based file systems
// apis that could be lossy with the use of Go time types.
type Timespec unix.Timespec

func cStructToTimespec(t C.struct_timespec) Timespec {
	return Timespec{
		Sec:  int64(t.tv_sec),
		Nsec: int64(t.tv_nsec),
	}
}
