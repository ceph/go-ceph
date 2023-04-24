# API Hints & Quick How-Tos

Below you'll find some brief sections that show how some of the API calls
in go-ceph work together. This is not meant to cover every possible use
case but are recorded here as a quick way to get familiar with these
calls.


## General

### Finding an API

The go-ceph project wraps existing APIs that are part of Ceph. There are two
kinds of APIs that are wrapped. The first style of API is based on functions
Ceph exports as client libraries in C. The three packages in go-ceph `cephfs`,
`rados`, and `rbd` map to `libcephfs`, `librados`, and `librbd` respectively.

The go-ceph packages that wrap Ceph C APIs follow a documentation convention
that aims to make it easier to map between the APIs. In functions that are
implemented using a certain C API function a line with the term `Implements:`
will be followed by the C function's declaration - matching what can be found
in the C library's header file. For example, if you knew you wanted to wrap the
rbd function to get an image's metadata, `rbd_metadata_get`, you could search
within the source code or https://pkg.go.dev/github.com/ceph/go-ceph/rbd for
`rbd_metadata_get` which would lead you to the `GetMetadata` method of the
`Image` type.

The second style of API is based on functions implemented within Ceph services
based on Ceph's "command" system. These functions are primarily accessed using
the `ceph` command. Many of the functions within the ceph command
are implemented by sending a structured JSON message to either the Ceph MON or
MGR. Packages in go-ceph that wrap these sorts of APIs are found in
`cephfs/admin`, `rbd/admin`, and `common/admin/manager` for example.

The command/JSON based API packages follow a different, but similar
documentation convention to the C based APIs. Functions that roughly map to a
particular `ceph` CLI command will contain a line with the term `Similar To:`
followed by the `ceph` command it is similar to. For example, if
you knew you wanted to create a CephFS subvolume group and would normally use
the command `ceph fs subvolumegroup create` to do so, you could search within
the source code or https://pkg.go.dev/github.com/ceph/go-ceph/cephfs/admin for
`ceph fs subvolumegroup create` which would lead you to the
`CreateSubVolumeGroup` property of the `FSAdmin` type.

#### Can't find the API you want?

The go-ceph project is maintained separately from Ceph and it is common for
APIs to be added to Ceph that are not present in go-ceph. Sometimes we resolve
the differences quickly but not always.

Generally, there is no simple way to access a C based API from Go without
updates to the code. If there's an API that you need that doesn't appear to be
wrapped by go-ceph, please [file an
issue](https://github.com/ceph/go-ceph/issues). If you are comfortable writing
Go and would like to try writing a wrapper function we're more than happy to
[welcome contributions](./development.md#contribution-guidelines) as well.

The command/JSON based APIs can be accessed without directly wrapping them.
The large majority of these functions are based on either
[rados.MgrCommand](https://pkg.go.dev/github.com/ceph/go-ceph@v0.21.0/rados#Conn.MgrCommand)
or
[rados.MonCommand](https://pkg.go.dev/github.com/ceph/go-ceph@v0.21.0/rados#Conn.MonCommand).
Both these functions accept a formatted JSON object that maps to a command line
constant prefix and the variable argument values. Determining the JSON prefix
and accepted arguments can be done using a special JSON-command: `{"prefix":
"get_command_descriptions"}`. This will return a dump of all commands the
service knows about.

You can then use this information to construct your own JSON to send to
the server. For example the prefix `fs subvolumegroup ls` takes an argument `vol_name`
which is annotated as a CephString. Thus you can send the JSON
`{"prefix": "fs subvolumegroup ls", "vol_name": "foobar":, "format": "json"}`.
The last parameter `format` is a special general argument that *suggests* to the
server that you want the reply data to be JSON formatted.

That all said, while you can directly interact with the command/JSON based APIs we
are also very happy to consider feature requests as well as contributions to make
working with these API more Go-idiomatic, convenient, and common.

#### Preview APIs

When a new API is added to go-ceph we consider the API to be a "preview" API.
This means that while we think the API is good enough to distribute we do not
promise not to change it. We assume that most consumers of go-ceph want a
stable API and so the preview APIs are "hidden" behind a go build tag. This
tag, `ceph_preview`, can be passed to the Go build command such that when you
import go-ceph packages the preview APIs will become "visible" to your code.
Do be aware that if you use preview APIs in your code there is the chance
they'll change between go-ceph releases.

Preview APIs do not show up on pkg.go.dev but we do list all of them in
our [API status document](./api-status.md). We track when each API was
added and when the API is expected to become stable.


## rados Package

### Connecting to a cluster

Connect to a Ceph cluster using a configuration file located in the default
search paths.

```go
conn, _ := rados.NewConn()
conn.ReadDefaultConfigFile()
conn.Connect()
```

A connection can be shutdown by calling the `Shutdown` method on the
connection object (e.g. `conn.Shutdown()`). There are also other methods for
configuring the connection. Specific configuration options can be set:

```go
conn.SetConfigOption("log_file", "/dev/null")
```

and command line options can also be used using the `ParseCmdLineArgs` method.

```go
args := []string{ "--mon-host", "1.1.1.1" }
err := conn.ParseCmdLineArgs(args)
```

For other configuration options see the full documentation.

### Object I/O

Object in RADOS can be written to and read from with through an interface very
similar to a standard file I/O interface:

```go
// open a pool handle
ioctx, err := conn.OpenIOContext("mypool")

// write some data
bytesIn := []byte("input data")
err = ioctx.Write("obj", bytesIn, 0)

// read the data back out
bytesOut := make([]byte, len(bytesIn))
_, err := ioctx.Read("obj", bytesOut, 0)

if !bytes.Equal(bytesIn, bytesOut) {
    fmt.Println("Output is not input!")
}
```

### Pool maintenance

The list of pools in a cluster can be retrieved using the `ListPools` method
on the connection object. On a new cluster the following code snippet:

```go
pools, _ := conn.ListPools()
fmt.Println(pools)
```

will produce the output `[data metadata rbd]`, along with any other pools that
might exist in your cluster. Pools can also be created and destroyed. The
following creates a new, empty pool with default settings.

```go
conn.MakePool("new_pool")
```

Deleting a pool is also easy. Call `DeletePool(name string)` on a connection object to
delete a pool with the given name. The following will delete the pool named
`new_pool` and remove all of the pool's data.

```go
conn.DeletePool("new_pool")
```

### Error Handling

As typical of Go codebases, a large number of functions in go-ceph return `error`s.
Some of these errors are based on non-exported types. This is deliberate choice.
However, much of the relevant data these types can contain are available. One
does not have to resort to the somewhat brittle approach of converting errors
to strings and matching on (parts of) said string.

In some cases the errors returned by calls are considered "sentinel" errors.
These errors can be matched to exported values in the package using the
`errors.Is` function from the Go standard library.

Example:
```go
// we want to delete a pool, but oops, conn is disconnected
err := conn.DeletePool("foo")
if err != nil {
    if errors.Is(err, rados.ErrNotConnected) {
        // ... do something specific when not connected ...
    } else {
        // ... handle generic error ...
    }
}
```

Example:
```go
err := rgw.MyAPICall()
if err != nil {
    if errors.Is(err, rgw.ErrInvalidAccessKey) {
       // ... do something specific to access errors ...
    } else if errors.Is(err, rgw.ErrNoSuchUser) {
       // ... do something specific to user not existing ...
    } else {
       // ... handle generic error ...
    }
}
```

In other cases the returned error doesn't match a specific error value but
rather is implemented by a type that may carry additional data. Specifically,
many errors in go-ceph implement an `ErrorCode() int` method. If this is the
case you can use ErrorCode to access a numeric error code provided by calls to
Ceph. Note that the error codes returned by Ceph often match unix/linux
`errno`s - but the exact meaning of the values returned by `ErrorCode()` are
determined by the Ceph APIs and go-ceph is just making them accessible.


Example:
```go
type errorWithCode interface {
    ErrorCode() int
}

err := rados.SomeRadosFunc()
if err != nil {
    var ec errorWithCode
    if errors.As(err, &ec) {
        errCode := ec.ErrorCode()
        // ... do something with errCode ...
    } else {
        // ... handle generic error ...
    }
}
```

Note that Go allows type definitions inline so you can even write:
```go
err := rados.SomeRadosFunc()
if err != nil {
    var ec interface { ErrorCode() int }
    if errors.As(err, &ec) {
        errCode := ec.ErrorCode()
        // ... do something with errCode ...
    } else {
        // ... handle generic error ...
    }
}
```

Newer packages in go-ceph generally prefer to latter approach to avoid creating
lots of sentinels that are only used rarely.
