package implements

import (
	"fmt"
	"regexp"

	"modernc.org/cc/v3"
)

var (
	// Individual "package" stubs. Add the needed headers to pick up the
	// ceph lib<whatever> content.

	cephfsCStub = `
#include "cephfs/libcephfs.h"
`
	radosCStub = `
#include "rados/librados.h"
`
	rbdCStub = `
#include "rbd/librbd.h"
#include "rbd/features.h"
`
	typeStubs = `
#define int8_t int
#define int16_t int
#define int32_t int
#define int64_t int
#define uint8_t int
#define uint16_t int
#define uint32_t int
#define uint64_t int
#define dev_t int
#define size_t int
#define ssize_t int
#define mode_t int
#define uid_t int
#define gid_t int
#define off_t int
#define time_t int
#define bool int
#define __GNUC__ 4
#define __x86_64__ 1
#define __linux__ 1
`
	stubs = map[string]string{
		"cephfs": cephfsCStub,
		"rados":  radosCStub,
		"rbd":    rbdCStub,
	}
	funcPrefix = map[string]string{
		"cephfs": "ceph_",
		"rados":  "rados_",
		"rbd":    "rbd_",
	}
)

func stubCFunctions(libname string) (CFunctions, error) {
	cstub := stubs[libname]
	if cstub == "" {
		return nil, fmt.Errorf("no C stub available for '%s'", libname)
	}
	var conf cc.Config
	conf.PreprocessOnly = true
	conf.IgnoreInclude = regexp.MustCompile(`^<.+>$`)
	src := []cc.Source{
		{Name: "typestubs", Value: typeStubs, DoNotCache: true},
		{Name: libname, Value: cstub, DoNotCache: true},
	}
	inc := []string{"@", "/usr/local/include", "/usr/include"}
	cAST, err := cc.Parse(&conf, inc, nil, src)
	if err != nil {
		return nil, err
	}
	cfMap := map[cc.StringID]CFunction{}
	deprecated := cc.String("deprecated")
	for i := range cAST.Scope {
		for _, n := range cAST.Scope[i] {
			if n, ok := n.(*cc.Declarator); ok &&
				!n.IsTypedefName &&
				n.DirectDeclarator != nil &&
				(n.DirectDeclarator.Case == cc.DirectDeclaratorFuncParam ||
					n.DirectDeclarator.Case == cc.DirectDeclaratorFuncIdent) {
				name := n.Name()
				if _, exists := cfMap[name]; !exists {
					_, isDeprecated := n.AttributeSpecifierList.Has(deprecated)
					cf := CFunction{
						Name:         name.String(),
						IsDeprecated: isDeprecated,
					}
					cfMap[name] = cf
				}
			}
		}
	}
	cfs := make(CFunctions, 0, len(cfMap))
	for _, cf := range cfMap {
		cfs = append(cfs, cf)
	}
	return cfs, nil
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
