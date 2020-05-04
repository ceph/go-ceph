package main

// The "implements" tool uses code analysis and the conventions of the go-ceph
// project to produce a report comparing what go-ceph implements and what
// exists in the C ceph APIs. Only (lib)cephfs, (lib)rados, and (lib)rbd are
// and the equilvaent go-ceph packages are supported.
//
// Examples:
//   # generate a full report on cephfs with a function listing
//   ./implements --list ./cephfs
//
//   # generate a brief summary on all packages
//   ./implements ./cephfs ./rados ./rbd
//
//   # generate a comprehensive report on rbd in json
//   ./implements --list --annotate --json rbd

import (
	"flag"
	"log"
	"os"
	"path"

	"github.com/ceph/go-ceph/contrib/implements/internal/implements"
)

var (
	verbose    bool
	list       bool
	annotate   bool
	reportJSON bool

	// verbose logger
	logger = log.New(os.Stderr, "(implements/verbose) ", log.LstdFlags)
)

func abort(msg string) {
	log.Fatalf("error: %v\n", msg)
}

func init() {
	flag.BoolVar(&verbose, "verbose", false, "be more verbose (for debugging)")
	flag.BoolVar(&list, "list", false, "list functions")
	flag.BoolVar(&annotate, "annotate", false, "annotate functions")
	flag.BoolVar(&reportJSON, "json", false, "use JSON output format")
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		abort("missing package(s) to analyze")
	}
	if verbose {
		implements.SetLogger(logger)
	}

	var r implements.Reporter
	o := implements.ReportOptions{
		List:     list,
		Annotate: annotate,
	}
	switch {
	case reportJSON:
		r = implements.NewJSONReport(o, os.Stdout)
	default:
		r = implements.NewTextReport(o)
	}

	for _, pkgref := range args[0:] {
		source, pkg := path.Split(pkgref)
		switch pkg {
		case "cephfs", "rados", "rbd":
			if verbose {
				logger.Printf("Processing package: %s\n", pkg)
			}
		default:
			abort("unknown package name: " + pkg)
		}
		if source == "" {
			source = "."
		}
		ii := implements.NewInspector()
		if err := implements.CephCFunctions(pkg, ii); err != nil {
			abort(err.Error())
		}
		if err := implements.CephGoFunctions(source, pkg, ii); err != nil {
			abort(err.Error())
		}
		if err := r.Report(pkg, ii); err != nil {
			abort(err.Error())
		}
	}
	r.Done()
}
