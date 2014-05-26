# go-rados - Go bindings for RADOS distributed object store

## Installation

    go get github.com/noahdesu/go-rados

The native RADOS library and development headers are expected to be installed.

## Documentation

Detailed documentation is available at
<http://godoc.org/github.com/noahdesu/go-rados>.

## Example

Connect to a cluster and list the pools:

```go
conn, _ := rados.NewConn()
conn.ReadDefaultConfigFile()
conn.Connect()
pools, _ := conn.ListPools()
fmt.Println(len(pools), pools)
```

will print:

    3 [data metadata rbd]
