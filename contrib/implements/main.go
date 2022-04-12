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
//   ./implements --list --json rbd

import (
	"flag"
	"implements/internal/implements"
	"log"
	"os"
)

var (
	verbose    bool
	list       bool
	reportJSON bool
	outputJSON string
	outputText string

	// verbose logger
	logger = log.New(os.Stderr, "(implements/verbose) ", log.LstdFlags)
)

func abort(msg string) {
	log.Fatalf("error: %v\n", msg)
}

func init() {
	flag.BoolVar(&verbose, "verbose", false, "be more verbose (for debugging)")
	flag.BoolVar(&list, "list", false, "list functions")
	flag.BoolVar(&reportJSON, "json", false,
		"use JSON output format (this is a shortcut for '--report-json -')")

	flag.StringVar(&outputJSON, "report-json", "", "filename for JSON report")
	flag.StringVar(&outputText, "report-text", "", "filename for plain-text report")
}

// TODO: this is a stub implementation that doesn't do anything.
// the intent is to eventually be able to tell a go-ceph sub-package
// like "rbd" or "cephfs/admin" apart from a real path like "/home/foo/go-ceph/rbd".
func splitPkg(s string) (string, string) {
	return "", s
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

	rpts := []implements.Reporter{}
	// always annotate for now, leave the option of disabling it someday if it
	// gets costly
	o := implements.ReportOptions{
		List:     list,
		Annotate: true,
	}
	switch {
	case reportJSON:
		rpts = append(rpts, implements.NewJSONReport(o, os.Stdout))
	case outputJSON == "-":
		rpts = append(rpts, implements.NewJSONReport(o, os.Stdout))
	case outputJSON != "":
		f, err := os.Create(outputJSON)
		if err != nil {
			abort(err.Error())
		}
		defer func() { _ = f.Close() }()
		rpts = append(rpts, implements.NewJSONReport(o, f))
	}
	switch {
	case outputText == "-":
		rpts = append(rpts, implements.NewTextReport(o, os.Stdout))
	case outputText != "":
		f, err := os.Create(outputText)
		if err != nil {
			abort(err.Error())
		}
		defer func() { _ = f.Close() }()
		rpts = append(rpts, implements.NewTextReport(o, f))
	}
	if len(rpts) == 0 {
		// no report types were explicitly selected. Use text report to stdout.
		rpts = append(rpts, implements.NewTextReport(o, os.Stdout))
	}

	for _, pkgref := range args[0:] {
		source, pkg := splitPkg(pkgref)
		checkCLang := false
		switch pkg {
		case "cephfs", "rados", "rbd":
			checkCLang = true
			if verbose {
				logger.Printf("Processing package (with C): %s\n", pkg)
			}
		default:
			if verbose {
				logger.Printf("Processing package: %s\n", pkg)
			}
		}
		if source == "" {
			source = "."
		}
		ii := implements.NewInspector()
		if checkCLang {
			if err := implements.CephCFunctions(pkg, ii); err != nil {
				abort(err.Error())
			}
		}
		if err := implements.CephGoFunctions(source, pkg, ii); err != nil {
			abort(err.Error())
		}
		for _, r := range rpts {
			if err := r.Report(pkg, ii); err != nil {
				abort(err.Error())
			}
		}
	}
	for _, r := range rpts {
		_ = r.Done()
	}
}
