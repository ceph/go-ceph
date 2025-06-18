package implements

import (
	"fmt"
	"runtime"
	"strings"

	"modernc.org/cc/v4"
)

var (
	// Individual "package" stubs. Add the needed headers to pick up the
	// ceph lib<whatever> content.

	cephfsCStub = `
#define _FILE_OFFSET_BITS 64
#include "cephfs/libcephfs.h"
`
	radosCStub = `
#include "rados/librados.h"
`
	radosStriperCStub = `
#include "radosstriper/libradosstriper.h"
`
	rbdCStub = `
#include "rbd/librbd.h"
#include "rbd/features.h"
`
	stubs = map[string]string{
		"cephfs":        cephfsCStub,
		"rados":         radosCStub,
		"rados/striper": radosStriperCStub,
		"rbd":           rbdCStub,
	}
	funcPrefix = map[string]string{
		"cephfs":        "ceph_",
		"rados":         "rados_",
		"rados/striper": "rados_striper_",
		"rbd":           "rbd_",
	}
)

func stubCFunctions(libname string) (CFunctions, error) {
	cstub := stubs[libname]
	if cstub == "" {
		return nil, fmt.Errorf("no C stub available for '%s'", libname)
	}
	conf, err := cc.NewConfig(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return nil, err
	}

	src := []cc.Source{
		{Name: "<predefined>", Value: conf.Predefined},
		{Name: "<builtin>", Value: cc.Builtin},
		{Name: libname, Value: cstub},
	}
	cAST, err := cc.Translate(conf, src)
	if err != nil {
		return nil, err
	}
	cfMap := map[string]CFunction{}
	for i := range cAST.Scope.Nodes {
		for _, n := range cAST.Scope.Nodes[i] {
			if n, ok := n.(*cc.Declarator); ok &&
				!n.IsTypename() &&
				strings.HasPrefix(n.Name(), funcPrefix[libname]) &&
				n.DirectDeclarator != nil &&
				(n.DirectDeclarator.Case == cc.DirectDeclaratorFuncParam ||
					n.DirectDeclarator.Case == cc.DirectDeclaratorFuncIdent) {
				name := n.Name()
				if _, exists := cfMap[name]; !exists {
					isDeprecated := false
					if attrs := n.Type().Attributes(); attrs != nil {
						isDeprecated = attrs.IsAttrSet("deprecated")
					}
					cf := CFunction{
						Name:         name,
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
