
# The 'implements' tool

implements is a small-ish tool created to compare the Ceph C APIs with
go-ceph implmeents.

## Build

In the go-ceph repository run `make implmeents` to create a standalone
binary for the `implements` cli tool.

## Run

```
./implements [--verbose] [--json] [--list] [pkg...]
```

The --verbose option causes verbose details about the source scan to be
printed.

The tool can produce either plain-text output, or JSON with the --json option.

The --list option produces a list of all covered and missing functions from
the Ceph library. The listing also provides information about each function's
status.

`DIR` should be a directory containing go-ceph sources. If running the command from the root of the go-ceph git checkout, `.` is sufficient.

`pkg` is one or more package names such as: "cephfs", "rados", or "rbd".
The packages may be indicated by directory, such as "./cephfs".
The tool will output a section pertaining to each named package.


Examples:

```
# Quickly summarize all packages
./implements cephfs rados rbd

# List missing and present functions in rbd
./implements --list ./rbd

# Print debugging info while processing rados
./implements --verbose rados

# Full analysis of everything in JSON
./implements --json --list ./cephfs ./rados ./rbd

```
