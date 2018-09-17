#!/bin/bash

set -e

mkdir /tmp/ceph
/micro-osd.sh /tmp/ceph
export CEPH_CONF=/tmp/ceph/ceph.conf

export PATH=/usr/lib/go-1.10/bin:$PATH

go get -t -v ./...
diff -u <(echo -n) <(gofmt -d -s .)
#go vet ./...
#go list ./...
go test -v $(go list ./... | grep -v cephfs)
