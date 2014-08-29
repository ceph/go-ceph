package rados

// #cgo LDFLAGS: -lrados
// #include <stdlib.h>
// #include <rados/librados.h>
import "C"

import "unsafe"
import "bytes"

type ClusterStat struct {
    Kb uint64
    Kb_used uint64
    Kb_avail uint64
    Num_objects uint64
}

type Conn struct {
    cluster C.rados_t
}

func (c *Conn) PingMonitor(id string) (string, error) {
    c_id := C.CString(id)
    defer C.free(unsafe.Pointer(c_id))

    var strlen C.size_t
    var strout *C.char

    ret := C.rados_ping_monitor(c.cluster, c_id, &strout, &strlen)
    defer C.rados_buffer_free(strout)

    if ret == 0 {
        reply := C.GoStringN(strout, (C.int)(strlen))
        return reply, nil
    } else {
        return "", RadosError(int(ret))
    }
}

// Connect establishes a connection to a RADOS cluster. It returns an error,
// if any.
func (c *Conn) Connect() error {
    ret := C.rados_connect(c.cluster)
    if ret == 0 {
        return nil
    } else {
        return RadosError(int(ret))
    }
}

func (c *Conn) Shutdown() {
    C.rados_shutdown(c.cluster)
}

func (c *Conn) ReadConfigFile(path string) error {
    c_path := C.CString(path)
    defer C.free(unsafe.Pointer(c_path))
    ret := C.rados_conf_read_file(c.cluster, c_path)
    if ret == 0 {
        return nil
    } else {
        return RadosError(int(ret))
    }
}

func (c *Conn) ReadDefaultConfigFile() error {
    ret := C.rados_conf_read_file(c.cluster, nil)
    if ret == 0 {
        return nil
    } else {
        return RadosError(int(ret))
    }
}

func (c *Conn) OpenPool(pool string) (*Pool, error) {
    c_pool := C.CString(pool)
    defer C.free(unsafe.Pointer(c_pool))
    ioctx := &Pool{}
    ret := C.rados_ioctx_create(c.cluster, c_pool, &ioctx.ioctx)
    if ret == 0 {
        return ioctx, nil
    } else {
        return nil, RadosError(int(ret))
    }
}

// ListPools returns the current list of pool names. It returns an error, if
// any.
func (c *Conn) ListPools() (names []string, err error) {
    buf := make([]byte, 4096)
    for {
        ret := int(C.rados_pool_list(c.cluster,
            (*C.char)(unsafe.Pointer(&buf[0])), C.size_t(len(buf))))
        if ret < 0 {
            return nil, RadosError(int(ret))
        }

        if ret > len(buf) {
            buf = make([]byte, ret)
            continue
        }

        tmp := bytes.SplitAfter(buf[:ret-1], []byte{0})
        for _, s := range tmp {
            if len(s) > 0 {
                name := C.GoString((*C.char)(unsafe.Pointer(&s[0])))
                names = append(names, name)
            }
        }

        return names, nil
    }
}

// SetConfigOption sets the configuration option named option to have the
// value value. It returns an error, if any.
func (c *Conn) SetConfigOption(option, value string) error {
    c_opt, c_val := C.CString(option), C.CString(value)
    defer C.free(unsafe.Pointer(c_opt))
    defer C.free(unsafe.Pointer(c_val))
    ret := C.rados_conf_set(c.cluster, c_opt, c_val)
    if ret < 0 {
        return RadosError(int(ret))
    } else {
        return nil
    }
}

func (c *Conn) GetConfigOption(name string) (value string, err error) {
    buf := make([]byte, 4096)
    c_name := C.CString(name)
    defer C.free(unsafe.Pointer(c_name))
    ret := int(C.rados_conf_get(c.cluster, c_name,
        (*C.char)(unsafe.Pointer(&buf[0])), C.size_t(len(buf))))
    // FIXME: ret may be -ENAMETOOLONG if the buffer is not large enough. We
    // can handle this case, but we need a reliable way to test for
    // -ENAMETOOLONG constant. Will the syscall/Errno stuff in Go help?
    if ret == 0 {
        value = C.GoString((*C.char)(unsafe.Pointer(&buf[0])))
        return value, nil
    } else {
        return "", RadosError(ret)
    }
}

// WaitForLatestOSDMap blocks the caller until the latest OSD map has been
// retrieved. It returns an error, if any.
func (c *Conn) WaitForLatestOSDMap() error {
    ret := C.rados_wait_for_latest_osdmap(c.cluster)
    if ret < 0 {
        return RadosError(int(ret))
    } else {
        return nil
    }
}

// GetClusterStat returns statistics about the cluster. It returns an error,
// if any.
func (c *Conn) GetClusterStats() (stat ClusterStat, err error) {
    c_stat := C.struct_rados_cluster_stat_t{}
    ret := C.rados_cluster_stat(c.cluster, &c_stat)
    if ret < 0 {
        return ClusterStat{}, RadosError(int(ret))
    } else {
        return ClusterStat{
            Kb: uint64(c_stat.kb),
            Kb_used: uint64(c_stat.kb_used),
            Kb_avail: uint64(c_stat.kb_avail),
            Num_objects: uint64(c_stat.num_objects),
        }, nil
    }
}

func (c *Conn) ParseCmdLineArgs(args []string) error {
    argc := C.int(len(args))
    argv := make([]*C.char, argc)
    for i, arg := range args {
        argv[i] = C.CString(arg)
        defer C.free(unsafe.Pointer(argv[i]))
    }
    ret := C.rados_conf_parse_argv(c.cluster, argc, &argv[0])
    if ret < 0 {
        return RadosError(int(ret))
    } else {
        return nil
    }
}

func (c *Conn) ParseDefaultConfigEnv() error {
    ret := C.rados_conf_parse_env(c.cluster, nil)
    if ret == 0 {
        return nil
    } else {
        return RadosError(int(ret))
    }
}
