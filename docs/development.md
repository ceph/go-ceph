# go-ceph - Development Guide

## Preface

This document is aimed at providing a brief introduction to the structure and
development processes used by the go-ceph project. It is aimed at people who
wish to improve go-ceph. We assume familiarity with the Go language and common
tooling as well as some familiarity with C. We hope this document is useful but
can't guarantee it's always up to date. It will never be totally comprehensive.

This document tries to focus on items that may not be obvious by reading the
code itself. One can go very far by simply reading the code and sticking with
what has been observed. However, this doc tries to give a bit more of a
background "philosophy" that hopefully reveals some part of why we do certain
things we do.



## Theme

The primary theme behind go-ceph is one of providing access to Ceph's
functionality via Go as a library of API functions and types. We desire to
expose the full power of the Ceph APIs - this means that we generally aim to
provide a thin layer of code over the APIs provided by Ceph itself.  We try to
do enough to make a user of Go feel like go-ceph is a (mostly) idiomatic Go
library. We also strive to make someone familiar with the Ceph APIs recognize
what C functions are being mapped to Go.

While we may provide some convenience layers we generally plan to provide access
to all of the Ceph APIs on a near 1:1 basis. When we do provide convenience
layers we do not mean to make them the exclusive tool-set provided.

While the focus so far has been accessing APIs in C in Go, the true target of
go-ceph is to express Ceph functionality in Go. As such, not every feature
of go-ceph may be found in the C API. A small but growing set of features
build upon parts of Ceph that make use of the C APIs but do more than just
that. For example, this includes the `cephfs/admin` package.


## Library Structure

Currently, there are three top level sub-packages that reflect three main
functional areas in Ceph. The `rados` package exposes features related to
Ceph's RADOS system and the `librados` C library. The `rbd` package exposes
features related to the RBD subsystem and the `librbd` C library. The `cephfs`
package exposes features related to CephFS and the `libcephfs` C library.

In addition the `internal` directory contains packages that have APIs used
to support the public APIs in rados, rbd, and cephfs but are not exported
publicly themselves. A large proportion of these helper libraries are intended
to ease working with C functions from Go. These are placed under the special
"internal" namespace so that we do not promise outside consumers that these
APIs are part of go-ceph or stable.

Under `cephfs/admin` there is a sub-package aimed at managing aspects of CephFS
such as administering subvolumes, subvolume groups, snapshots, and other facets
of CephFS that can be also be managed using the `ceph` command line tool but are
not directly part of the C API.


## File Structure

When writing new code or updating go-ceph, keep in mind that a single .go file
should be related to a single related "sub-topic" within the scope of the
subsystem. For example, functions related to snapshotting an rbd volume are
in `snapshot.go`. Following Go convention, tests for those functions will be
found in `snashot_test.go`. For historical reasons, there are still a few
"omnibus" .go files in the codebase. Please avoid adding to those files whenever
possible.

The go-ceph project uses "build tags" to support multiple versions of Ceph with
a single version of go-ceph. [Build
tags](https://www.digitalocean.com/community/tutorials/customizing-go-binaries-with-build-tags)
are a feature of Go that we use to conditionally build some files based on Ceph
release versions. Typically, we use the release code name of Ceph (nautilus,
octopus, etc) to choose what APIs in Ceph we expect to be available. Because
some APIs for a topical area may vary across ceph releases we some time name
files like `snapshot_nautilus.go` to support compiling some APIs only
conditionally for given versions of Ceph. Depending on the feature, an API
function in go-ceph may be disabled entirely or if the APIs in Ceph are largely
compatible one go-ceph function may be written in terms of different C APIs
functions (that produce the same effect).


## API Naming

Generally, the go-ceph project aims to name functions and types similarly to
the names used in Ceph libraries and documentation. We also follow the standard
Go naming conventions. This leads us to converting some names from
`underscore_style` to `CamelCase` style. However, we try to retain the same
terms used by the Ceph functions. Occasionally, we will tweak the word order
to fit the object-method approach. For example, if a function in Ceph is called
`pantry_cheese_get` and we've created a type `Pantry` to encapsulate functions
related to the pantry topic, we might add a function `func (p *Pantry)
GetCheese(...)` rather than keeping the word order of the original. When in
doubt, do what seems reasonable and ask for additional feedback during code
review.


## Documentation Conventions

The project intends to rigorously document the public facing APIs of go-ceph.
This starts by adding godoc comments to exported functions and types. This
is currently enforced by tools in our CI. Additionally, we've established a
"home grown" convention to help map between Ceph APIs and go-ceph ones.
For functions that have an equivalent C API, add a block at the bottom of
the doc comment that starts with "Implements:" and is followed by
the C function definition, indented, like so:

```
// GroupImageAdd will add the specified image to the named group.
// An io context must be supplied for both the group and image.
//
// Implements:
//  int rbd_group_image_add(rados_ioctx_t group_p,
//                          const char *group_name,
//                          rados_ioctx_t image_p,
//                          const char *image_name);
```

The typical Go-style doc comment goes first, followed by the "Implements" line,
indicating that what follows is what C function being implemented by the
go-ceph function, and then the C function. This is indented so that the
godoc system treats it as a quoted block.

These lines help readers who are familiar with the C API and may even aid
search engines. In addition, we have some simplistic tooling that uses these
comments to help us determine how much of the Ceph APIs we're covering.

For the `cephfs/admin` package, and any similar cases where there's CLI support
for something but no Ceph API, we replace "Implements" by "Similar To" and
record a simplified version of the command it most closely matches. Example:

```
// ListSubVolumes returns a list of subvolumes belonging to the volume and
// optional subvolume group.
//
// Similar To:
//  ceph fs subvolume ls <volume> --group-name=<group>
```

Recently, go-ceph has adopted an [API Stability Policy](./api-stability.md) to
help users of our library know what APIs are deprecated and what APIs are
available for preview.  In short, APIs that are deprecated must contain a line
starting with "Deprecated:" and APIs that are preview must be in files that
contain a build constraint for the `ceph_preview` tag.

Deprecated function Example:

```
// AllocateBlocks pre-allocates the specified number of memory blocks
// for caching.
//
// Deprecated: this API is no longer supported.
//
// Implements.
//  int allocate_blocks(...)
```

Preview function example:

```
//go:build ceph_preview

...

// Energize the particle buffers with anti-nutrinos. This can be used to
// warm up the Heisenberg compensator.
//
// Implements:
//  int energize(...)
```

### API Status

In order to better track the status of our deprecated and preview APIs we have
an [API Status document](./api-status.md). This document is generated from a
JSON file in our `docs/` directory. When a new API is being added, one or more
additional patches need to be provided to update the API status doc and JSON
file. If you have no unusual requirements, you can run `make api-update` and
commit the changes that have been made to the `docs/` directory.

This command will automatically update the `api-status.*` files, indicating
that the API is added in the next expected release and will become stable
two release after that. If you need to, you can customize this behavior
by editing the JSON file by hand, or running `./contib/apiage.py` with
different options, followed by running `make api-doc` to update the generated
markdown file.


## Testing

The go-ceph project makes heavy use of unit and functional tests to ensure it
matches the behaviors of Ceph as intended. We are not strict about the
distinction between the types of tests and generally treat the majority of
tests as functional. Unless you're running a manually specified subset of
tests we require a running Ceph cluster to ensure the APIs we've implemented
are correct in terms of the Ceph features we need.

As of this writing the test automation is preformed using the github actions
system. The YAML files under `.github/workflows` define the jobs that get
executed automatically.


### Running Tests

For both running tests locally or in our CI jobs we build and run the tests in
(OCI/docker) containers. This also includes containers to run a self-contained
"micro" Ceph cluster. The container images can be build by running `make
ci-image` an optional `CEPH_VERSION` variable can be provided which will be
used to select the base image for the container. Currently it can be supplied
as either "nautilus" or "octopus".

The entire suite of tests can be run via the makefile rule `test-container`.
For example: `make test-container`. The behavior of the test container is
controlled by the script `entrypoint.sh`. This script takes a number of command
line options, and can be used to restrict what tests will be run. For example,
the command line option `--test-pkg=rados`. Will only test the `rados`
subpackage. The `--help` option can be provided to view the options the script
supports.


### Adding New Tests


We stress the importance of all new features and code changes having
corresponding tests. We try not to obsess over having 100% line coverage, but
do want to see everything that can be tested have a test case. By default, our
test container enables coverage reports so it is fairly easy to see what parts
of the library have coverage or not. The CI also captures the generated
coverage HTML reports and makes them available to download.

The go-ceph project makes use of the [testify
library](https://github.com/stretchr/testify). Depending on the circumstances,
tests use a mix of checks (assert) and requirements (require). If a test
function must not proceed past a certain point, use require. Otherwise, we
default to assert calls.


### Misc

Code quality and formatting checks can be run via `make check`. These checks
require `gofmt` as well as [revive](https://github.com/mgechev/revive).

A custom tool called `implements` is available under `contrib/implements`. This
tool is designed to help compare what is available in ceph vs. go-ceph. It
checks both the "Implements" sections in the comments as well as the code
itself.


## Contribution Guidelines

The go-ceph project makes use of the pull-request workflow provided by github.
Work should be submitted as a series of patches organized on a git branch and
then submitted together as a PR. As a general rule, small PRs - both in terms
of number of patches and lines of code - are processed, reviewed and merged
faster than larger ones. However, if you have to err on one side or the other,
a larger number of small patches is preferable to a small number of large
patches.

Each patch should have a well formed commit message. The go-ceph project
prefers the topic-subject-body style, followed by a Signed-off-by line. Example:

```
[topic]: [short description]

[Longer description - multiple lines or paragraphs as needed]

Signed-off-by: [Your Name] <[your email]>
```

The commit message should help others understand the where, what, and why of
the patch. A topic is a functional area within the go-ceph project. For
example, `rados` or `cephfs`. When in doubt, let the directory name of files
changed help be your guide. So if you worked on files in the "cephfs/admin"
directory, a topic of "cephfs admin" would be appropriate.

Every new patch should be complete enough that it does not rely on any code
changes that follow it. In other words, add API A before B if B relies on A.
Add tests for a feature either in the patch that adds the feature or following
it. No (submitted) patch should cause a test failure. Keeping patches clean
this way enables the use of tools like `git bisect` to find real issues in the
future.

As noted previously, changes should be accompanied by documentation and tests
appropriate.


## Closing Remarks

When in doubt, feel free to reach out to the project via the github discussions
feature, IRC chat, etc. This document is part of the go-ceph project and so
feedback and contributions to this doc are very welcome.
