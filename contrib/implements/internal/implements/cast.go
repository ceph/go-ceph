package implements

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

var (
	// CastXMLBin is the name/location of the castxml binary.
	CastXMLBin = "castxml"

	// Add a stub C function that goes nowhere and does nothing. Just
	// to give castxml something to chew on. It may not be strictly
	// needed but it worked for me.
	// TODO Cleanup - The macro part is probably totally unnecessary but I wanted
	// to see how much "extra stuff" castxml picked up.

	gndn = `

#define GNDN
GNDN int foo(int x) {
    return x;
}
`

	// Individual "package" stubs. Add the needed headers to pick up the
	// ceph lib<whatever> content plus the code stub for castxml.

	cephfs_c_stub = `
#define FILE_OFFSET_BITS 64
#include <stdlib.h>
#define __USE_FILE_OFFSET64
#include <cephfs/libcephfs.h>
` + gndn
	rados_c_stub = `
#include <rados/librados.h>
` + gndn
	rbd_c_stub = `
#include <rbd/librbd.h>
#include <rbd/features.h>
` + gndn

	stubs = map[string]string{
		"cephfs": cephfs_c_stub,
		"rados":  rados_c_stub,
		"rbd":    rbd_c_stub,
	}
	funcPrefix = map[string]string{
		"cephfs": "ceph_",
		"rados":  "rados_",
		"rbd":    "rbd_",
	}
)

type allCFunctions struct {
	Functions CFunctions `xml:"Function"`
}

func parseCFunctions(xmlData []byte) ([]CFunction, error) {
	cf := allCFunctions{}
	if err := xml.Unmarshal(xmlData, &cf); err != nil {
		return nil, err
	}
	return cf.Functions.ensure()
}

func parseCFunctionsFromFile(fname string) ([]CFunction, error) {
	cf := allCFunctions{}

	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	xdec := xml.NewDecoder(f)
	err = xdec.Decode(&cf)
	if err != nil {
		return nil, err
	}
	return cf.Functions.ensure()
}

func parseCFunctionsFromCmd(args []string) (CFunctions, error) {
	cf := allCFunctions{}

	cmd := exec.Command(args[0], args[1:]...)
	logger.Printf("will call: %v", cmd)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	xdec := xml.NewDecoder(stdout)
	parseErr := xdec.Decode(&cf)

	err = cmd.Wait()
	if err != nil {
		return nil, err
	}

	if parseErr != nil {
		return nil, parseErr
	}
	return cf.Functions.ensure()
}

func stubCFunctions(libname string) (CFunctions, error) {
	cstub := stubs[libname]
	if cstub == "" {
		return nil, fmt.Errorf("no C stub available for '%s'", libname)
	}

	tfile, err := ioutil.TempFile("", "*-"+libname+".c")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tfile.Name())

	_, err = tfile.Write([]byte(cstub))
	if err != nil {
		return nil, err
	}

	cmd := []string{
		CastXMLBin,
		"--castxml-output=1",
		"-o", "-",
		tfile.Name(),
	}
	return parseCFunctionsFromCmd(cmd)
}

// CephCFunctions will extract C functions from the supplied package name
// and update the results within the code inspector.
func CephCFunctions(pkg string, ii *Inspector) error {
	logger.Printf("getting C AST for %s", pkg)
	f, err := stubCFunctions(pkg)
	if err != nil {
		return err
	}
	return ii.SetExpected(funcPrefix[pkg], f)
}
