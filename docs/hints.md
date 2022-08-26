# API Hints & Quick How-Tos

Below you'll find some brief sections that show how some of the API calls
in go-ceph work together. This is not meant to cover every possible use
case but are recorded here as a quick way to get familiar with these
calls.

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
