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

The list of pools in a cluster can be retreived using the `ListPools` method
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
