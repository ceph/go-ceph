
# Go-Ceph Release Process

Regular releases are planned starting mid-February 2020. Until the API is
stable we will be issuing v0.y versions.

Today, the release process includes the following stages:

### Pre-release
- [ ] Complete any API stability updates
- [ ] Check milestone for any incomplete issues

### Release Tasks
- [ ] Update the releases table in the README
- [ ] Tag the code
- [ ] Create release notes
- [ ] Finalize the release on GitHub

### Post-release
- [ ] Announce the release
- [ ] Create/verify a milestone for next release
- [ ] Prepare API stability update issue for next release


The sections below go into more detail about these steps in no particular order.

## Tagging a Major-Minor release

First, make sure your local git tree has the latest changes.
Tag master branch with a vX.Y.Z version. This should be an annotated tag (it
will have a commit message).

Example:
```shell
git checkout master
git pull --ff-only
git tag -a v0.2.0 -m 'Release v0.2.0'
```

Push the tag to the go-ceph repo (not your own fork).
Example:
```shell
git push --follow-tags
```


## Create a release using github
* https://github.com/ceph/go-ceph/releases/new
* Select the tag you just pushed
* Add the release notes to the body of the release
* Save the new release

After creating the release the milestone for the release should be closed
(edit milestone -> close).


### Notes

As a Go library package that makes use of cgo we will not be producing any
build artifacts at this time. Only source code will be provided. Users of the
library will typically use the go toolchain (go modules, etc), git or tarballs.
Tarballs are automatically provided by github interface when a release is
created. No extra steps are currently required.


## Scheduling the next release

Future releases are scheduled using the
[milestones](https://github.com/ceph/go-ceph/milestones) feature on github. The
milestone may have features/issues/PRs associated with it, but it does not have
to. The milestone must have a date associated with it.

Currently, go-ceph is released every two months. The project has a tradition of
releasing on a Tuesday, and a Tuesday near the middle of the month. Often, this
is the second Tuesday of the month. However, if the first or second day of the
month is a Tuesday it may be better to choose the 3rd Tuesday of the month.

The title of the milestone should be "Release vX.Y.Z" where X, Y, and Z are the
major, minor, and patch versions. The description is short typically just,
"Regular planned release vX.Y of the go-ceph library." Since go-ceph is using
time based released additional details are largely unnecessary.

So, for example if the release is occurring on Tuesday, Feb. 15, 2022 then the
next release day would be April. The second Tuesday of April is the 12th. So
the due date would be set to 2022-04-12.

Because of the time based process it is fine to create the new milestones before
the current release is done.


## Creating release notes

As of release v0.15.0 the method of writing release notes is largely based on
using the automatically generated list of PRs generated at GitHub's Release UI
and then sorting some of the items into categories. Some of the boilerplate
language in each section can be copied from the previous release notes and then
updated.

Remember, one of the easiest things to do is to look at previous releases and
largely mimic what they do.

#### Introduction
The "Introduction" is a paragraph noting that this is a new version
of go-ceph. It can be copied and the version updated.

#### Highlights
Thank new contributors to the project. This can be derived from the GitHub
notes.  Additional paragraphs can be added to highlight a particularly
important feature or change.

#### Stability Caveat
The "Stability caveat" is a reminder about go-ceph's stability (non)guarantees.
It can be copied from previous releases.

#### New Features
Sort new features by package (cephfs, rados, rbd/admin, etc). Each
PR/change-list is a bullet point under the package. For every PR that adds an
API add a sub-bullet describing new methods and what methods in Ceph it wraps
(if it wraps something).

#### API Stability Updates
Sort changes by package. For each changed API make a bullet point and describe
the state of the API ("x is now stable", etc).

#### Deprecations and Removals
Sort API function changes by package. Note what APIs that are deprecated and/or
what is removed. Add a short paragraph describing any changes to what versions
of Ceph is being deprecated or removed.

#### Internal
The internal changes list is a flat bulleted list of changes (PRs) that do not
add or remove Go-package-visible features. Things like changes to the build
scripts or unit tests, for example.


> NOTE: For context on how previous versions of the release notes were authored
please review older versions of this file from version control history.



## Announcing the release

The release is publicly announced to ceph-devel and ceph users mailing lists.
The body of the email follows the template below.  Change the URL to point to
the new release. Optionally mention some of the packages that have changes:

```
I'm happy to announce another release of the go-ceph API library. This is a
regular release following our every-two-months release cadence.

https://github.com/ceph/go-ceph/releases/tag/v0.13.0

Changes include additions to the rbd and rados packages. More details are
available at the link above.

The library includes bindings that aim to play a similar role to the "pybind"
python bindings in the ceph tree but for the Go language. The library also
includes additional APIs that can be used to administer cephfs, rbd, and rgw
subsystems.
There are already a few consumers of this library in the wild, including the
ceph-csi project.
```

This announcement is also sent to an internal team list inside Red Hat.


## Documenting the Release

The README.md file contains a table of go-ceph releases and what versions of
ceph each release supports. After the release has been created this table
should be updated to reflect the new release and what versions of ceph it
supports.
