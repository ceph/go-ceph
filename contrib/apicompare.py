#!/usr/bin/env python3

import argparse
import glob
import os
import re
import subprocess
import sys
import tempfile
import xml.etree.ElementTree


USAGE = """
Compare go-ceph bindings coverage of the ceph api to available C
code.
Note that this script is not comprehensive and does not claim to be.
It only takes functions into account and does not care about overlap
in functionality (function x is a superset of y and thus y does
not need to be implemented) or deprecated functions.

Example: python3 apicompare.py --check=rbd --list .
Example: python3 apicompare.py -crbd -crados -ccephfs .
"""


RBD_C = """
#include <rbd/librbd.h>
#include <rbd/features.h>

#define GNDN
GNDN int foo(int x) {
    return x;
}
"""

RADOS_C = """
#include <rados/librados.h>

#define GNDN
GNDN int foo(int x) {
    return x;
}
"""

CEPHFS_C = """
#define _FILE_OFFSET_BITS 64
#include <stdlib.h>
#include <cephfs/libcephfs.h>

#define GNDN
GNDN int foo(int x) {
    return x;
}
"""


class Results(object):
    def __init__(self):
        self.total = 0
        self.referenced = []
        self.documented = []
        self.missing = []


def go_sources(srcdir):
    pat = '/'.join((srcdir, '*.go'))
    for fn in glob.glob(pat):
        if fn.endswith("_test.go"):
            continue
        yield fn


def check_go(path, funcs):
    gotext = []
    for fn in go_sources(path):
        with open(fn) as fh:
            gotext.append(fh.read())
    gotext = '\n'.join(gotext)
    res = Results()
    res.total = len(funcs)
    for fname in funcs:
        regex = re.compile('\\bC\\.{}\\b'.format(fname))
        m = regex.findall(gotext)
        if m:
            res.referenced.append(fname)
        else:
            res.missing.append(fname)
    return res


def api_funcs(code_stub, prefix, dump_xml=False):
    funcs = []
    with tempfile.TemporaryDirectory() as tempdir:
        tmpc = os.path.join(tempdir, 'temp.c')
        tmpxml = os.path.join(tempdir, 'temp_c.xml')
        with open(tmpc, 'w') as fh:
            fh.write(code_stub)
        subprocess.check_call(
            ['castxml', '--castxml-output=1', '-o', tmpxml, tmpc])
        if dump_xml:
            with open(tmpxml) as fh:
                print(fh.read())
        x = xml.etree.ElementTree.parse(tmpxml)
        for fe in x.getroot().findall('Function'):
            name = fe.attrib.get('name', '')
            if not name.startswith(prefix):
                continue
            funcs.append(name)
    return funcs


def check_for_castxml():
    try:
        with open(os.devnull) as nullfh:
            subprocess.check_call(
                ['castxml'],
                stdin=nullfh,
                stderr=nullfh,
                stdout=nullfh)
    except subprocess.CalledProcessError:
        sys.stderr.write('error: "castxml" binary must be installed\n')
        sys.exit(2)
    except FileNotFoundError:
        sys.stderr.write('error: "castxml" binary must be installed\n')
        sys.exit(2)


def report(label, res, list_functions=False):
    print("{} functions covered: {}/{} : {:.2f}%".format(
          label,
          len(res.referenced),
          res.total,
          (100.0 * len(res.referenced)) / res.total))
    print("{} functions remain: {}/{} : {:.2f}%".format(
          label,
          len(res.missing),
          res.total,
          (100.0 * len(res.missing)) / res.total))

    if list_functions:
        for n in sorted(res.referenced):
            print("  Covered: {}".format(n))
        for n in sorted(res.missing):
            print("  Missing: {}".format(n))


def main():
    ap = argparse.ArgumentParser(
        description='compare api coverage extent',
        usage=USAGE)
    ap.add_argument(
        '--check', '-c',
        choices=['rbd', 'rados', 'cephfs'],
        action='append')
    ap.add_argument(
        '--list-functions',
        action='store_true')
    ap.add_argument('GO_SRC')
    cli = ap.parse_args()

    check_for_castxml()
    os.chdir(cli.GO_SRC)

    if 'rbd' in (cli.check or []):
        funcs = api_funcs(RBD_C, 'rbd_')
        res = check_go('rbd', funcs)
        report('RBD', res, cli.list_functions)
    if 'rados' in (cli.check or []):
        funcs = api_funcs(RADOS_C, 'rados_')
        res = check_go('rados', funcs)
        report('RADOS', res, cli.list_functions)
    if 'cephfs' in (cli.check or []):
        funcs = api_funcs(CEPHFS_C, 'ceph_')
        res = check_go('cephfs', funcs)
        report('CephFS', res, cli.list_functions)
    return


if __name__ == '__main__':
    main()
