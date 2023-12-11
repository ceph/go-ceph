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
* rgw/admin - interact with [radosgw admin ops API](https://docs.ceph.com/en/latest/radosgw/adminops)

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

On MacOS you can use brew to install the libraries:
```sh
brew tap mulbc/ceph-client
brew install ceph-client
```

NOTE: CentOS users may want to use a
[CentOS Storage SIG](https://wiki.centos.org/SpecialInterestGroup/Storage/Ceph)
repository to enable packages for a supported ceph version.
Example: `dnf -y install centos-release-ceph-pacific`.
(CentOS 7 users should use "yum" rather than "dnf")


To quickly test if one can build with go-ceph on your system, run:
```sh
go get github.com/ceph/go-ceph
```

Once compiled, code using go-ceph is expected to dynamically link to the Ceph
libraries. These libraries must be available on the system where the go based
binaries will be run. Our use of cgo and ceph libraries does not allow for
fully static binaries.

go-ceph tries to support different Ceph versions. However some functions might
only be available in recent versions, and others may be deprecated. In order to
work with non-current versions of Ceph, it is required to pass build-tags to
the `go` command line. A tag with the named Ceph release will enable/disable
certain features of the go-ceph packages, and prevent warnings or compile
problems. For example, to ensure you select the library features that match
the "pacific" release, use:
```sh
go build -tags pacific ....
go test -tags pacific ....
```

### Supported Ceph Versions

| go-ceph version | Supported Ceph Versions | Deprecated Ceph Versions |
| --------------- | ------------------------| -------------------------|
| v0.25.0         | pacific, quincy, reef   | nautilus, octopus        |
| v0.24.0         | pacific, quincy, reef   | nautilus, octopus        |
| v0.23.0         | pacific, quincy, reef   | nautilus, octopus        |
| v0.22.0         | pacific, quincy         | nautilus, octopus        |
| v0.21.0         | pacific, quincy         | nautilus, octopus        |
| v0.20.0         | pacific, quincy         | nautilus, octopus        |
| v0.19.0         | pacific, quincy         | nautilus, octopus        |
| v0.18.0         | octopus, pacific, quincy | nautilus                |
| v0.17.0         | octopus, pacific, quincy | nautilus                |

The tags affect what is supported at compile time. What version of the Ceph
cluster the client libraries support, and vice versa, is determined entirely
by what version of the Ceph C libraries go-ceph is compiled with.

To see what older versions of go-ceph supported refer to the [older
releases](./docs/older-releases.md) file in the documentation.


## Documentation

Detailed API documentation is available at
<https://pkg.go.dev/github.com/ceph/go-ceph>.

Some [API Hints and How-Tos](./docs/hints.md) are also available to quickly
introduce how some of API calls work together.


## Development

```
docker run --rm -it --net=host \
  --security-opt apparmor:unconfined \
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

### Getting in Touch

Want to get in touch with the go-ceph team? We're available through a few
different channels:
* Have a question, comment, or feedback:
  [Use the Discussions Board](https://github.com/ceph/go-ceph/discussions)
* Report an issue or request a feature:
  [Issues Tracker](https://github.com/ceph/go-ceph/issues)
* We participate in the Ceph
  [user's mailing list](https://lists.ceph.io/hyperkitty/list/ceph-users@ceph.io/)
  and [dev list](https://lists.ceph.io/hyperkitty/list/dev@ceph.io/)
  and we also announce our releases on those lists
* You can sometimes find us in the
  [#ceph-devel IRC channel](https://ceph.io/irc/) - hours may vary
