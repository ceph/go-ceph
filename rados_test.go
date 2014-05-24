package rados_test

import "testing"
//import "bytes"
import "github.com/noahdesu/rados"
import "github.com/stretchr/testify/assert"

func TestVersion(t *testing.T) {
    var major, minor, patch = rados.Version()
    assert.False(t, major < 0 || major > 1000, "invalid major")
    assert.False(t, minor < 0 || minor > 1000, "invalid minor")
    assert.False(t, patch < 0 || patch > 1000, "invalid patch")
}

func TestOpen(t *testing.T) {
    _, err := rados.Open("admin")
    assert.Equal(t, err, nil, "error")
}

func TestConnect(t *testing.T) {
    conn, _ := rados.Open("admin")
    conn.ReadDefaultConfigFile()
    err := conn.Connect()
    assert.Equal(t, err, nil)
}

func TestPingMonitor(t *testing.T) {
    conn, _ := rados.Open("admin")
    conn.ReadDefaultConfigFile()
    conn.Connect()
    reply, err := conn.PingMonitor("kyoto")
    assert.Equal(t, err, nil)
    assert.True(t, len(reply) > 0)
}

//func TestConnect(t *testing.T) {
//    conn, _ := rados.Open("admin")
//    conn.ReadConfigFile("/home/nwatkins/ceph/ceph/src/ceph.conf")
//    conn.Connect()
//    pool, _ := conn.OpenPool("data")
//
//    data_in := []byte("blah");
//    data_out := make([]byte, 10)
//
//    pool.Write("xyz", data_in, 0)
//    pool.Read("xyz", data_out[:4], 0)
//
//    if !bytes.Equal(data_in, data_out[:4]) {
//        t.Errorf("yuk")
//    }
//}
