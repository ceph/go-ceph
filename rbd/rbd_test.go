package rbd_test

import (
    "testing"
    "github.com/noahdesu/go-ceph/rados"
    "github.com/noahdesu/go-ceph/rbd"
    "github.com/stretchr/testify/assert"
    "os/exec"
    "sort"
)

func GetUUID() string {
    out, _ := exec.Command("uuidgen").Output()
    return string(out[:36])
}

func TestVersion(t *testing.T) {
    var major, minor, patch = rbd.Version()
    assert.False(t, major < 0 || major > 1000, "invalid major")
    assert.False(t, minor < 0 || minor > 1000, "invalid minor")
    assert.False(t, patch < 0 || patch > 1000, "invalid patch")
}

func TestGetImageNames(t *testing.T) {
    conn, _ := rados.NewConn()
    conn.ReadDefaultConfigFile()
    conn.Connect()

    poolname := GetUUID()
    err := conn.MakePool(poolname)
    assert.NoError(t, err)

    ioctx, err := conn.OpenIOContext(poolname)
    assert.NoError(t, err)

    createdList := []string{}
    for i := 0; i < 10; i++ {
        name := GetUUID()
        err = rbd.Create(ioctx, name, 1<<22)
        assert.NoError(t, err)
        createdList = append(createdList, name)
    }

    imageNames, err := rbd.GetImageNames(ioctx)
    assert.NoError(t, err)

    sort.Strings(createdList)
    sort.Strings(imageNames)
    assert.Equal(t, createdList, imageNames)

    ioctx.Destroy()
    conn.Shutdown()
}
