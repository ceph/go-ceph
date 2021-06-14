/*
Package ceph is the root of a set of packages that wrap the Ceph APIs.

Generally, this package only exists to host subsystem specific packages
and contains no functions.

The "rados" sub-package wraps APIs that handle RADOS specific functions.

The "rbd" sub-package wraps APIs that handle RBD specific functions.

The "cephfs" sub-package wraps APIs that handle CephFS specific functions.

The "common" sub-package contains sub-packages related to implementing
common interfaces and utilities shared across the above.

Consult the documentation for each package for additional details.
*/
package ceph
