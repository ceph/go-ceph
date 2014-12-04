package rados_test

import "testing"
//import "bytes"
import "github.com/noahdesu/go-rados"
import "github.com/stretchr/testify/assert"
import "os"
import "os/exec"
import "io"
import "io/ioutil"
import "time"
import "net"
import "fmt"
import "crypto/rand"

func GetUUID() string {
    out, _ := exec.Command("uuidgen").Output()
    return string(out[:36])
}

func TestVersion(t *testing.T) {
    var major, minor, patch = rados.Version()
    assert.False(t, major < 0 || major > 1000, "invalid major")
    assert.False(t, minor < 0 || minor > 1000, "invalid minor")
    assert.False(t, patch < 0 || patch > 1000, "invalid patch")
}

func TestGetSetConfigOption(t *testing.T) {
    conn, _ := rados.NewConn()

    // rejects invalid options
    err := conn.SetConfigOption("wefoijweojfiw", "welfkwjelkfj")
    assert.Error(t, err, "Invalid option")

    // verify SetConfigOption changes a values
    log_file_val, err := conn.GetConfigOption("log_file")
    assert.NotEqual(t, log_file_val, "/dev/null")

    err = conn.SetConfigOption("log_file", "/dev/null")
    assert.NoError(t, err, "Invalid option")

    log_file_val, err = conn.GetConfigOption("log_file")
    assert.Equal(t, log_file_val, "/dev/null")
}

func TestParseDefaultConfigEnv(t *testing.T) {
    conn, _ := rados.NewConn()

    log_file_val, _ := conn.GetConfigOption("log_file")
    assert.NotEqual(t, log_file_val, "/dev/null")

    err := os.Setenv("CEPH_ARGS", "--log-file /dev/null")
    assert.NoError(t, err)

    err = conn.ParseDefaultConfigEnv()
    assert.NoError(t, err)

    log_file_val, _ = conn.GetConfigOption("log_file")
    assert.Equal(t, log_file_val, "/dev/null")
}

func TestParseCmdLineArgs(t *testing.T) {
    conn, _ := rados.NewConn()
    conn.ReadDefaultConfigFile()

    mon_host_val, _ := conn.GetConfigOption("mon_host")
    assert.NotEqual(t, mon_host_val, "1.1.1.1")

    args := []string{ "--mon-host", "1.1.1.1" }
    err := conn.ParseCmdLineArgs(args)
    assert.NoError(t, err)

    mon_host_val, _ = conn.GetConfigOption("mon_host")
    assert.Equal(t, mon_host_val, "1.1.1.1")
}

func TestGetClusterStats(t *testing.T) {
    conn, _ := rados.NewConn()
    conn.ReadDefaultConfigFile()
    conn.Connect()

    poolname := GetUUID()
    err := conn.MakePool(poolname)
    assert.NoError(t, err)

    pool, err := conn.OpenPool(poolname)
    assert.NoError(t, err)

    // grab current stats
    prev_stat, err := conn.GetClusterStats()
    fmt.Printf("prev_stat: %+v\n", prev_stat)
    assert.NoError(t, err)

    // make some changes to the cluster
    buf := make([]byte, 1<<20)
    for i := 0; i < 10; i++ {
        objname := GetUUID()
        pool.Write(objname, buf, 0)
    }

    // wait a while for the stats to change
    for i := 0; i < 30; i++ {
        stat, err := conn.GetClusterStats()
        assert.NoError(t, err)

        // wait for something to change
        if stat == prev_stat {
            fmt.Printf("curr_stat: %+v (trying again...)\n", stat)
            time.Sleep(time.Second)
        } else {
            // success
            fmt.Printf("curr_stat: %+v (change detected)\n", stat)
            conn.Shutdown()
            return
        }
    }

    conn.Shutdown()
    t.Error("Cluster stats aren't changing")
}

func TestGetFSID(t *testing.T) {
    conn, _ := rados.NewConn()
    conn.ReadDefaultConfigFile()
    conn.Connect()

    fsid, err := conn.GetFSID()
    assert.NoError(t, err)
    assert.NotEqual(t, fsid, "")

    conn.Shutdown()
}

func TestGetInstanceID(t *testing.T) {
    conn, _ := rados.NewConn()
    conn.ReadDefaultConfigFile()
    conn.Connect()

    id := conn.GetInstanceID()
    assert.NotEqual(t, id, 0)

    conn.Shutdown()
}

func TestMakeDeletePool(t *testing.T) {
    conn, _ := rados.NewConn()
    conn.ReadDefaultConfigFile()
    conn.Connect()

    // get current list of pool
    pools, err := conn.ListPools()
    assert.NoError(t, err)

    // check that new pool name is unique
    new_name := GetUUID()
    for _, poolname := range pools {
        if new_name == poolname {
            t.Error("Random pool name exists!")
            return
        }
    }

    // create pool
    err = conn.MakePool(new_name)
    assert.NoError(t, err)

    // get updated list of pools
    pools, err = conn.ListPools()
    assert.NoError(t, err)

    // verify that the new pool name exists
    found := false
    for _, poolname := range pools {
        if new_name == poolname {
            found = true
        }
    }

    if !found {
        t.Error("Cannot find newly created pool")
    }

    // delete the pool
    err = conn.DeletePool(new_name)
    assert.NoError(t, err)

    // verify that it is gone

    // get updated list of pools
    pools, err = conn.ListPools()
    assert.NoError(t, err)

    // verify that the new pool name exists
    found = false
    for _, poolname := range pools {
        if new_name == poolname {
            found = true
        }
    }

    if found {
        t.Error("Deleted pool still exists")
    }

    conn.Shutdown()
}

func TestPingMonitor(t *testing.T) {
    conn, _ := rados.NewConn()
    conn.ReadDefaultConfigFile()
    conn.Connect()

    // mon id that should work with vstart.sh
    reply, err := conn.PingMonitor("a")
    if err == nil {
        assert.NotEqual(t, reply, "")
        return
    }

    // mon id that should work with micro-osd.sh
    reply, err = conn.PingMonitor("0")
    if err == nil {
        assert.NotEqual(t, reply, "")
        return
    }

    // try to use a hostname as the monitor id
    mon_addr, _ := conn.GetConfigOption("mon_host")
    hosts, _ := net.LookupAddr(mon_addr)
    for _, host := range hosts {
        reply, err := conn.PingMonitor(host)
        if err == nil {
            assert.NotEqual(t, reply, "")
            return
        }
    }

    t.Error("Could not find a valid monitor id")

    conn.Shutdown()
}

func TestReadConfigFile(t *testing.T) {
    conn, _ := rados.NewConn()

    // check current log_file value
    log_file_val, err := conn.GetConfigOption("log_file")
    assert.NoError(t, err)
    assert.NotEqual(t, log_file_val, "/dev/null")

    // create a temporary ceph.conf file that changes the log_file conf
    // option.
    file, err := ioutil.TempFile("/tmp", "go-rados")
    assert.NoError(t, err)

    _, err = io.WriteString(file, "[global]\nlog_file = /dev/null\n")
    assert.NoError(t, err)

    // parse the config file
    err = conn.ReadConfigFile(file.Name())
    assert.NoError(t, err)

    // check current log_file value
    log_file_val, err = conn.GetConfigOption("log_file")
    assert.NoError(t, err)
    assert.Equal(t, log_file_val, "/dev/null")

    // cleanup
    file.Close()
    os.Remove(file.Name())
}

func TestWaitForLatestOSDMap(t *testing.T) {
    conn, _ := rados.NewConn()
    conn.ReadDefaultConfigFile()
    conn.Connect()

    err := conn.WaitForLatestOSDMap()
    assert.NoError(t, err)

    conn.Shutdown()
}

func TestReadWrite(t *testing.T) {
    conn, _ := rados.NewConn()
    conn.ReadDefaultConfigFile()
    conn.Connect()

    // make pool
    pool_name := GetUUID()
    err := conn.MakePool(pool_name)
    assert.NoError(t, err)

    pool, err := conn.OpenPool(pool_name)
    assert.NoError(t, err)

    // make random bytes
    bytes := make([]byte, 1<<20)
    n, err := rand.Read(bytes)
    assert.NoError(t, err)

    err = pool.Write("obj", bytes, 0)
    assert.NoError(t, err)

    bytes_out := make([]byte, 1<<20)
    n_out, err := pool.Read("obj", bytes_out, 0)

    assert.Equal(t, n, n_out)
    assert.Equal(t, bytes, bytes_out)
}
