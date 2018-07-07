#!/bin/bash

set -e

mkdir /tmp/ceph
/micro-osd.sh /tmp/ceph
export CEPH_CONF=/tmp/ceph/ceph.conf

go vet ./...
go get -t -v ./...
go list ./...
go test -v $(go list ./... | grep -v cephfs)
