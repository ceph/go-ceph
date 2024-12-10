package striper

import (
	ts "github.com/ceph/go-ceph/internal/timespec"
)

// Timespec behaves similarly to C's struct timespec.
type Timespec ts.Timespec

// StatInfo contains values returned by a Striper's Stat call.
type StatInfo struct {
	Size    uint64
	ModTime Timespec
}
