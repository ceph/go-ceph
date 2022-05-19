//go:build ceph_preview
// +build ceph_preview

// Package log allows to enable go-ceph logging and integrate it with the
// logging of the go-ceph consuming code.
package log

import (
	intLog "github.com/ceph/go-ceph/internal/log"
)

// SetWarnf sets the log.Printf compatible receiver for warning logs.
func SetWarnf(f func(format string, v ...interface{})) {
	intLog.Warnf = f
}

// SetDebugf sets the log.Printf compatible receiver for debug logs.
func SetDebugf(f func(format string, v ...interface{})) {
	intLog.Debugf = f
}
