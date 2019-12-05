package cephfs

import (
	"github.com/ceph/go-ceph/logging"
)

var log logging.Logger = logging.NewStubLogger()

// SetLogger sets the logger l as the log destination for the cephfs package.
func SetLogger(l logging.Logger) {
	log = l
}
