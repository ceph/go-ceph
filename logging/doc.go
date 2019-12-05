/*
Package logging provides hooks to support logging within the go-ceph library
packages without depending on any particular logging library.

The package provides a Logger interface that is fairly minimal with functions
for logging error, info, and debug messages. Users of this library can provide
any compatible type or write an adapter type for the logger in use as needed.

Each package that supports logging will provide a SetLogger function to replace
the current logger with the specified one.

The logrus package is already compatible.
Example:
  package main

  import (
    "github.com/ceph/go-ceph/cephfs"
    "github.com/sirupsen/logrus"
  )

  func main() {
    // set up cephfs to use logrus logging
    cephfs.SetLogger(logrus.New())
  }

*/
package logging
