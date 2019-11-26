// issue-109.go: analyze memory leak when rados.Conn.Connect() fails.
//
// build with: go build issue-109.go
// test with: valgrind ./issue-109

package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/ceph/go-ceph/rados"
)

func main() {
	for i := 1; i <= 100; i++ {
		c, err := getConnection("127.0.0.1", "nobody", "secret")
		if err != nil {
			fmt.Printf("getConnection failed: %v\n", err)
		} else {
			c.Shutdown()
		}
	}

	// force running the garbage collector
	runtime.GC()
	time.Sleep(time.Second)
}

func getConnection(monitors, user, key string) (*rados.Conn, error) {
	conn, err := rados.NewConnWithUser(user)
	if err != nil {
		return nil, err
	}
	args := []string{"--client_mount_timeout", "15", "-m", monitors, "--key", key}
	err = conn.ParseCmdLineArgs(args)
	if err != nil {
		return nil, fmt.Errorf("ParseCmdLineArgs: %v", err)
	}
	err = conn.Connect()
	if err != nil {
		return nil, fmt.Errorf("Connect: %v", err)
	}
	return conn, nil
}
