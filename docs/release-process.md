
# Go-Ceph Release Process

Regular releases are planned starting mid-Februrary 2020. Until the API is
stable we will be issuing v0.y versions.

## Major-Minor release

Tag master branch with a vX.Y.Z version. This should be an annotated tag (it
will have a commit message).

Example:
```shell
git tag -a v0.2.0 -m 'Release v0.2.0'
```

Push the tag to the go-ceph repo (not your own fork).
Example:
```shell
git push --follow-tags
```

Create a release using github:
* https://github.com/ceph/go-ceph/releases/new
* Select the tag you just pushed
* Author release notes, noting:
  * New features
  * Deprecated and Removed items
  * Other items (general improvements, test coverage, etc)
  * Highlight any important items unique to this version


# Notes

As a Go library package that makes use of cgo we will not be producing any
build artifacts at this time. Only source code will be provided. Users of the
library will typically use the go toolchain (go modules, etc), git or tarballs.
Tarballs are automatically provided by github interface when a release is
created. No extra steps are currently required.
