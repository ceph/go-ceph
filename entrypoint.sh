#!/bin/bash

set -e

rm -rf /tmp/ceph
mkdir /tmp/ceph
/micro-osd.sh /tmp/ceph
export CEPH_CONF=/tmp/ceph/ceph.conf

export PATH=/usr/lib/go-1.10/bin:$PATH

go get -t -v ./...
diff -u <(echo -n) <(gofmt -d -s .)
#go vet ./...
#go list ./...
P=github.com/ceph/go-ceph
GOCACHE=off go test -v -covermode=count -coverprofile=cover.out -coverpkg=$P/cephfs,$P/rados,$P/rbd ./...
mkdir -p /results/coverage
go tool cover -html=cover.out -o /results/coverage/go-ceph.html
