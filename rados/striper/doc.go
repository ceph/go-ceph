/*
Package striper contains a set of wrappers around Ceph's libradosstriper API.

The Striper type supports synchronous operations to read and write data,
as well as read and manipulate xattrs. Note that a striped object will
consist of one or more objects in RADOS.

There is no object list API in libradosstriper. Listing objects must be done
using the base RADOS APIs. Striped objects will be stored in RADOS using the
provided Striped Object ID (soid) suffixed by a dot (.) and a 16 byte
0-prefixed hex number (for example, "foo.0000000000000000" or
"bar.000000000000000a"). The object suffixed with ".0000000000000000" is the
0-index stripe and will also possess striper specific xattrs (see the [ceph
libradosstriper implementation] for a list) that are hidden from the
libradosstriper xattr APIs.  You can use the name and/or these striper xattrs
to distinguish a striped object from a non-striped RADOS object.

[ceph libradosstriper implementation]: https://github.com/ceph/ceph/blob/2fa0e43b7e714df9811f87cbc5bf862ac503483c/src/libradosstriper/RadosStriperImpl.cc#L94-L97
*/
package striper
