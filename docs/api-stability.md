# API stability

This library offers Go API bindings for ceph libraries and interfaces. In order
to provide a stable exported API and at the same time be able to get early
feedback for new and possibly immature API designs, we maintain three levels of
API stablity:

## Stable

This is the default level. Everything that is exported and not annotated
otherwise is considered stable. As long as we release 0.x versions, this is
still no 100% guarantee, but we try to avoid breaking changes as much as
possible. Once we reached version 1.x, this level provides a guarantee that no
breaking changes will be introduced until the next major release, as it is
convention in the Go community.

## Deprecated

This is a level for APIs that should not be used for new code. These are marked
as deprecated according to Go conventions in the documentation (that is, a
paragraph beginning with _Deprecated:_). During 0.x releases these APIs
might get removed in a future release, especially the 1.0 release, so we
recommend refactoring the code at the earliest convenience. After the 1.0
release, deprecated APIs will not be removed, however they are still deprecated
and only in maintanence mode. We usually don't make improvements for these APIs
and we can't guarantee optimal performance.

## Preview

This is a level for APIs that are fresh and might need further refinements in
following releases. These are not included in the documentation and are
only available, if the build tag `ceph_preview` is set. There might be breaking
changes in future releases regarding preview APIs. Usually new exported APIs are
introduced with this level first and become stable when there were no major
changes to the API for two releases. The schedule for preview APIs becoming
stable is tracked in a [separate document](./api-status.md).

Please note that while these APIs are still considered "unstable", this is not
true for the quality of their implementations, which we regard as stable and
error free, at least with the quality of beta code. Therefore we highly
encourage the use of these APIs and providing feedback to us, if a possible
breaking change in the API in future releases is feasible for your project.
