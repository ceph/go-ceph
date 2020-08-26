# go-ceph - Go bindings for Ceph APIs

[![Godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/ceph/go-ceph) [![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://raw.githubusercontent.com/ceph/go-ceph/master/LICENSE)

## Introduction

The go-ceph project is a collection of API bindings that support the use of
native Ceph APIs, which are C language functions, in Go. These bindings make
use of Go's cgo feature.
There are three main Go sub-packages that make up go-ceph:
* rados - exports functionality from Ceph's librados
* rbd - exports functionality from Ceph's librbd
* cephfs - exports functionality from Ceph's libcephfs

We aim to provide comprehensive support for the Ceph APIs over time. This
includes both I/O related functions and management functions.  If your project
makes use of Ceph command line tools and is written in Go, you may be able to
switch away from shelling out to the CLI and to these native function calls.

## Installation

The code in go-ceph is purely a library module. Typically, one will import
go-ceph in another Go based project. When building the code the native RADOS,
RBD, & CephFS library and development headers are expected to be installed.

On debian based systems (apt) these may be:
```sh
libcephfs-dev librbd-dev librados-dev
```

On rpm based systems (dnf, yum, etc) these may be:
```sh
libcephfs-devel librbd-devel librados-devel
```

To quickly test if one can build with go-ceph on your system, run:
```sh
go get github.com/ceph/go-ceph
```

Once compiled, code using go-ceph is expected to dynamically link to the Ceph
libraries. These libraries must be available on the system where the go based
binaries will be run. The use of cgo does not allow for fully static binaries.

go-ceph tries to support different Ceph versions. However some functions might
only be available in recent versions, and others may be deprecated. In order to
work with non-current versions of Ceph, it is required to pass build-tags to
the `go` command line. A tag with the named Ceph release will enable/disable
certain features of the go-ceph packages, and prevent warnings or compile
problems. For example, to ensure you select the library features that match
the "nautilus" release, use:
```sh
go build -tags nautilus ....
go test -tags nautilus ....
```

### Supported Ceph Versions

| go-ceph version | Supported Ceph Versions | Deprecated Ceph Versions |
| --------------- | ------------------------| -------------------------|
| v0.5.0          | nautilus, octopus       | luminous, mimic          |
| v0.4.0          | luminous, mimic, nautilus, octopus | |
| v0.3.0          | luminous, mimic, nautilus, octopus | |
| v0.2.0          | luminous, mimic, nautilus          | |
| (pre release)   | luminous, mimic  (see note)        | |

These tags affect what is supported at compile time. What version of the Ceph
cluster the client libraries support, and vice versa, is determined entirely
by what version of the Ceph C libraries go-ceph is compiled with.

NOTE: Prior to 2020 the project did not make versioned releases. The ability to
compile with a particular Ceph version before go-ceph v0.2.0 is not guaranteed.


## Documentation

Detailed API documentation is available at
<https://pkg.go.dev/github.com/ceph/go-ceph>.

Some [API Hints and How-Tos](./docs/hints.md) are also available to quickly
introduce how some of API calls work together.


# Development

```
docker run --rm -it --net=host \
  --device /dev/fuse --cap-add SYS_ADMIN --security-opt apparmor:unconfined \
  -v ${PWD}:/go/src/github.com/ceph/go-ceph:z \
  -v /home/nwatkins/src/ceph/build:/home/nwatkins/src/ceph/build:z \
  -e CEPH_CONF=/home/nwatkins/src/ceph/build/ceph.conf \
  ceph-golang
```

Run against a `vstart.sh` cluster without installing Ceph:

```
export CGO_CPPFLAGS="-I/ceph/src/include"
export CGO_LDFLAGS="-L/ceph/build/lib"
go build
```

## Contributing

Contributions are welcome & greatly appreciated, every little bit helps. Make code changes via Github pull requests:

- Fork the repo and create a topic branch for every feature/fix. Avoid
  making changes directly on master branch.
- All incoming features should be accompanied with tests.
- Make sure that you run `go fmt` before submitting a change
  set. Alternatively the Makefile has a flag for this, so you can call
  `make fmt` as well.
- The integration tests can be run in a docker container, for this run:

```
make test-docker
```

### Interactive "Office Hours"

The maintenance team plans to be available regularly for questions, comments,
pings, etc for about an hour twice a week. The current schedule is:

* 2:00pm EDT (currently 18:00 UTC) Mondays
* 9:00am EDT (currently 13:00 UTC) Thursdays

We will use the [#ceph-devel IRC channel](https://ceph.io/irc/)
