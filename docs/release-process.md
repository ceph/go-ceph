
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

The go-ceph release notes have the following structure. Some of the boilerplate
language in each section can be copied from the previous release notes and then
updated:
* Introduction
* Highlights
* Stability caveat
* New features
* Now stable
* Deprecations and Removals
* Other

The "Introduction" is a paragraph noting that this is a new version
of go-ceph. It can be copied and the version updated.

The "Highlights" are one or two paragraphs that highlight notable new feature(s)
in the library. This includes a sentence or two describing the feature with a
large impact to users of the library or one consumers were waiting for.
If the contribution comes from someone who is not a project maintainer the
paragraph should include a "shout out" to the contributor by name (usually
taken from the contributor's signed-off-by line). We prefer to highlight changes
from outside contributors whenever possible.

The "Stability caveat" is a reminder about go-ceph's stability (non)guarantees.
It can be copied from previous releases.

The "New Features" section is a bullet list of of the sub packages that make up
go-ceph and then sub-lists for each new API call. If the feature is a C binding
the line should be in the form `Add [GoApiName] implementing [c_api_name]`.
Other calls and types can be listed like `Add [GoApiName] function` or `Add
[GoTypeName] method [GoFuncName]`. If a method name is unambiguous feel free to
skip documenting the type it recives, but if the name is not unique it may be
best to include both the type and method like `Add [GoTypeName] method
[GoApiName] implementing [c_api_name]`.

Example:
```
# New Features

* In the rados package:
  * Add RambleOn implementing rados_ramble_on

* In the cephfs admin package:
  * Add SuperSecretSnapshot function
```

The "Now Stable" section consists of bulleted lists much like "New features".
However, it lists lists APIs that were previously in the "preview" state and are now
considered stable. The structure is the same as the "New Features" section but
instead of `Add [api]` write `The [api]`.

The "Deprecations and Removals" section can consist of API lists like the "New
features" section when individual API calls are being deprecated. Lines take
the form `Deprecate [ApiName`. It can also consist of free form paragraphs
deprecating something like a ceph release or explaining a deprecation and it's
expected replacements.

The "Other" section is a bulleted list containing short descriptions of changes
to the codebase that don't fit the other categories. This include bugfixes,
workflow/process changes, and minor non-API-visible improvements. Because most of
these items are not visisble to codebases that make use of go-ceph, these lines
can and should be short and avoid going into details. Eight commits to the code
that make many changes can be summarized as "organizational changes to the code
layout" for example. Parties interested in more can use git to explore the
project history in detail.

### Authoring process - A personal note

Generally, I use `git log --reverse --oneline vX.Y.Z...` to generate a summary
of changes from the last release. I concatenate that to a copy of the previous
release notes and then remove all the "old stuff". Then I go line by line
through the "change log" removing the lines from git and adding a release note
equivalent.  One I've filled in the new features, deprecations, and now stable
sections, I write the highlight. I then send it, as a draft, to the other
maintainers and a few trusted colleagues for a brief review. I usually try to
write the notes the day before the release.

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
