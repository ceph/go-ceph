#!/usr/bin/python3
"""
apiage.py - a quick and dirty tool for tracking when apis become stable
and deprecated apis are to be removed.

PDX-License-Identifier: MIT
"""

import argparse
import copy
import json
import re
import sys


def read_json(path):
    try:
        with open(path, "r") as fh:
            data = json.load(fh)
    except FileNotFoundError:
        return {}
    return data


def write_json(path, data):
    if path is None:
        raise ValueError("a valid path is required")
    with open(path, "w") as fh:
        json.dump(data, fh, indent=2)


def write_markdown(path, data):
    if path is None:
        return
    with open(path, "w") as fh:
        format_markdown(data, fh)


def copy_api(tracked, keys, src, defaults=None):
    dst = tracked
    for key in keys[:-1]:
        dst = dst.setdefault(key, {})
    dst = dst.setdefault(keys[-1], [])
    added = []
    for gfunc in src:
        name = gfunc["name"]
        if name in [d["name"] for d in dst]:
            continue
        gfunc.update(defaults or {})
        dst.append(gfunc)
        added.append(gfunc)
    return added


def compare_and_update(tracked, pkg, pkg_api, defaults=None):
    if defaults is None:
        defaults = {}
    new_deprecated = new_preview = new_stable = []
    if "deprecated_api" in pkg_api:
        new_deprecated = copy_api(
            tracked=tracked,
            keys=[pkg, "deprecated_api"],
            src=pkg_api["deprecated_api"],
            defaults={
                "deprecated_in_version": defaults.get(
                    "deprecated_in_version", ""
                ),
                "expected_remove_version": defaults.get(
                    "expected_remove_version", ""
                ),
            },
        )
    if "preview_api" in pkg_api:
        new_preview = copy_api(
            tracked=tracked,
            keys=[pkg, "preview_api"],
            src=pkg_api["preview_api"],
            defaults={
                "added_in_version": defaults.get("added_in_version", ""),
                "expected_stable_version": defaults.get(
                    "expected_stable_version", ""
                ),
            },
        )
    if "stable_api" in pkg_api:
        new_stable = copy_api(
            tracked=tracked,
            keys=[pkg, "stable_api"],
            src=pkg_api["stable_api"],
        )
    return new_deprecated, new_preview, new_stable


def api_update(tracked, src, copy_stable=False, defaults=None):
    for pkg, pkg_api in src.items():
        _, _, new_stable = compare_and_update(
            tracked, pkg, pkg_api, defaults=defaults
        )
        if new_stable and not copy_stable:
            return len(new_stable)
    return 0


def api_compare(tracked, src):
    problems = 0
    tmp = copy.deepcopy(tracked)
    for pkg, pkg_api in src.items():
        new_deprecated, new_preview, new_stable = compare_and_update(
            tmp, pkg, pkg_api
        )
        for dapi in new_deprecated:
            print("not tracked (deprecated):", pkg, dapi["name"])
            problems += 1
        for papi in new_preview:
            print("not tracked (preview):", pkg, papi["name"])
            problems += 1
        for sapi in new_stable:
            print("not tracked (stable):", pkg, sapi["name"])
            problems += 1
    for pkg, pkg_api in tmp.items():
        for api in pkg_api.get("deprecated_api", []):
            if not api.get("deprecated_in_version"):
                print("no deprecated_in_version set:", pkg, api["name"])
                problems += 1
        for api in pkg_api.get("preview_api", []):
            if not api.get("added_in_version"):
                print("no added_in_version set:", pkg, api["name"])
                problems += 1
            if not api.get("expected_stable_version"):
                print("no expected_stable_version set:", pkg, api["name"])
                problems += 1
    return problems


def api_fix_versions(tracked, values, pred=None):
    """Walks through tracked API and fixes any placeholder versions to real version numbers.
    """
    for pkg, pkg_api in tracked.items():
        for api in pkg_api.get("deprecated_api", []):
            if pred and not pred(pkg, api["name"]):
                print(f"Skipping {pkg}:{api['name']} due to filter")
                continue
            _vfix(pkg, "deprecated_in_version", api, values)
            _vfix(pkg, "expected_remove_version", api, values)
        for api in pkg_api.get("preview_api", []):
            if pred and not pred(pkg, api["name"]):
                print(f"Skipping {pkg}:{api['name']} due to filter")
                continue
            _vfix(pkg, "added_in_version", api, values)
            _vfix(pkg, "expected_stable_version", api, values)


def api_find_updates(tracked, values):
    found = {"preview": [], "deprecated": []}
    for pkg, pkg_api in tracked.items():
        for api in pkg_api.get("deprecated_api", []):
            erversion = api.get("expected_remove_version", "")
            if erversion == values.get("next_version"):
                found["deprecated"].append({
                    "package": pkg,
                    "name": api.get("name", ""),
                    "expected_remove_version": erversion,
                })
        for api in pkg_api.get("preview_api", []):
            esversion = api.get("expected_stable_version", "")
            if esversion == values.get("next_version"):
                found["preview"].append({
                    "package": pkg,
                    "name": api.get("name", ""),
                    "expected_stable_version": esversion,
                })
    return found


def api_promote(tracked, src, values):
    changes = problems = 0
    for pkg, pkg_api in src.items():
        src_stable = pkg_api.get("stable_api", [])
        src_preview = pkg_api.get("preview_api", [])
        new_tracked_stable = new_tracked_preview = False
        try:
            tracked_stable = tracked.get(pkg, {})["stable_api"]
        except KeyError:
            tracked_stable = []
            new_tracked_stable = True
        try:
            tracked_preview = tracked.get(pkg, {})["preview_api"]
        except KeyError:
            tracked_preview = []
            new_tracked_preview = True
        for api in src_stable:
            indexed_stable = {a.get("name", ""): a for a in tracked_stable}
            indexed_preview = {a.get("name", ""): a for a in tracked_preview}
            name = api.get("name", "")
            if name in indexed_preview and name not in indexed_stable:
                # need to promote this api
                if values.get("added_in_version"):
                    # track some metadata. why not right?
                    api["added_in_version"] = indexed_preview[name].get("added_in_version", "")
                    api["became_stable_version"] = values["added_in_version"]
                tracked_preview[:] = [a for n, a in indexed_preview.items() if n != name]
                tracked_stable.append(api)
                print("promoting to stable: {}:{}".format(pkg, name))
                changes += 1
            elif name in indexed_preview and name in indexed_preview:
                print("bad state: {}:{} found in both preview and stable"
                      .format(pkg, name))
                problems += 1
            elif name not in indexed_preview and name not in indexed_stable:
                print("api not found in preview: {}:{}".format(pkg, name))
                problems += 1
            # else api is already stable. do nothing.
        if new_tracked_stable and tracked_stable:
            tracked[pkg]["stable_api"] = tracked_stable
        if new_tracked_preview and tracked_preview:
            tracked[pkg]["preview_api"] = tracked_preview
    return changes, problems



def format_markdown(tracked, outfh):
    print("<!-- GENERATED FILE: DO NOT EDIT DIRECTLY -->", file=outfh)
    print("", file=outfh)
    print("# go-ceph API Stability", file=outfh)
    print("", file=outfh)
    for pkg, pkg_api in tracked.items():
        print(f"## Package: {pkg}", file=outfh)
        print("", file=outfh)
        if "preview_api" in pkg_api and pkg_api["preview_api"]:
            print("### Preview APIs", file=outfh)
            print("", file=outfh)
            _table(
                pkg_api["preview_api"],
                columns=[
                    ("Name", "name"),
                    ("Added in Version", "added_in_version"),
                    ("Expected Stable Version", "expected_stable_version"),
                ],
                outfh=outfh,
            )
            print("", file=outfh)
        if "deprecated_api" in pkg_api and pkg_api["deprecated_api"]:
            print("### Deprecated APIs", file=outfh)
            print("", file=outfh)
            _table(
                pkg_api["deprecated_api"],
                columns=[
                    ("Name", "name"),
                    ("Deprecated in Version", "deprecated_in_version"),
                    ("Expected Removal Version", "expected_remove_version"),
                ],
                outfh=outfh,
            )
            print("", file=outfh)
        if (all(x not in pkg_api for x in ("preview_api", "deprecated_api")) or
            all(x in pkg_api and not pkg_api[x] for x in ("preview_api", "deprecated_api"))):
            print("No Preview/Deprecated APIs found. "
                  "All APIs are considered stable.", file=outfh)
            print("", file=outfh)


def format_updates_markdown(updates, outfh):
    print("## Preview APIs due to become stable", file=outfh)
    if not updates.get("preview"):
        print("n/a", file=outfh)
    for api in updates.get("preview", []):
        print(f"* {api['package']}: {api['name']}", file=outfh)
    print("", file=outfh)
    print("", file=outfh)
    print("## Deprecated APIs due to be removed", file=outfh)
    if not updates.get("deprecated"):
        print("n/a", file=outfh)
    for api in updates.get("deprecated", []):
        print(f"* {api['package']}/{api['name']}", file=outfh)
    print("", file=outfh)
    print("", file=outfh)


def _table(data, columns, outfh):
    for key, _ in columns:
        outfh.write(key)
        outfh.write(" | ")
    outfh.write("\n")
    for key, _ in columns:
        outfh.write("-" * len(key))
        outfh.write(" | ")
    outfh.write("\n")
    for entry in data:
        for _, dname in columns:
            outfh.write(entry[dname])
            outfh.write(" | ")
        outfh.write("\n")


def _setif(dct, key, value):
    if value:
        dct[key] = value


def _vfmt(x, y, z):
    return f"v{x}.{y}.{z}"


def _vfix(pkg, key, api, values):
    if api.get(key, "").startswith("$"):
        try:
            val = values[key]
        except KeyError:
            raise ValueError(f"missing {key} in values: {key} must be provided to fix apis")
        api[key] = val
        print(f"Updated {pkg}:{api['name']} {key}={values[key]}")


def _make_fix_filter(cli):
    pkgre = namere = None
    if cli.fix_filter_pkg:
        pkgre = re.compile(cli.fix_filter_pkg)
    if cli.fix_filter_func:
        namere = re.compile(cli.fix_filter_func)

    def f(pkg, fname):
        if pkgre and not pkgre.match(pkg):
            return False
        if namere and not namere.match(fname):
            return False
        return True

    return f


def tag_to_versions(cli, version_tag):
    # first: parse the tag
    if not version_tag.startswith("v"):
        raise ValueError(f"unexpected tag: {version_tag}")
    try:
        x, y, z = [int(val) for val in version_tag[1:].split(".")]
    except ValueError:
        raise ValueError(f"unexpected tag: {version_tag}")
    # set values according to the simple policy:
    # where version is X.Y.Z
    #  * added in: X+1
    #  * expected stable in: X+1+2
    #  * deprecated in: X+1
    # if they weren't manually specified
    if not cli.added_in_version:
        cli.added_in_version = _vfmt(x, y + 1, z)
    if not cli.stable_in_version:
        cli.stable_in_version = _vfmt(x, y + 3, z)
    if not cli.deprecated_in_version:
        cli.deprecated_in_version = _vfmt(x, y + 1, z)


def placeholder_versions(cli):
    if not cli.added_in_version:
        cli.added_in_version = "$NEXT_RELEASE"
    if not cli.stable_in_version:
        cli.stable_in_version = "$NEXT_RELEASE_STABLE"
    if not cli.deprecated_in_version:
        cli.deprecated_in_version = "$NEXT_RELEASE"


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "--source",
        "-s",
        default="./_results/implements.json",
        help="json describing state of code",
    )
    parser.add_argument(
        "--current",
        "-c",
        default="./docs/api-status.json",
        help="json tracking current apis",
    )
    parser.add_argument(
        "--document",
        "-d",
        default="./docs/api-status.md",
        help="markdown file describing current apis",
    )
    parser.add_argument(
        "--mode",
        choices=(
            "compare",
            "update",
            "write-doc",
            "fix-versions",
            "find-updates",
            "updates-to-markdown",
            "promote",
        ),
        default="compare",
        help="either update current state or compare current state to source",
    )
    parser.add_argument(
        "--copy-stable-apis",
        action="store_true",
        help="allow copying of pre-existing stable APIs",
    )
    parser.add_argument(
        "--added-in-version",
        "-A",
        help="specify an added-in version for all new preview apis",
    )
    parser.add_argument(
        "--stable-in-version",
        "-S",
        help="specify a stable-in version for all new preview apis",
    )
    parser.add_argument(
        "--deprecated-in-version",
        "-D",
        help="specify a deprecated-in version for all newly deprecated apis",
    )
    parser.add_argument(
        "--remove-in-version",
        "-R",
        help="specify a version that this deprecated api is expected to be removed",
    )
    parser.add_argument(
        "--current-tag",
        "-t",
        help=(
            "Specify the current VCS tag. This will be used to automatically"
            " set version values if not otherwise specified."
        ),
    )
    parser.add_argument(
        "--placeholder-versions",
        action="store_true",
        help="Specify special placeholder values for version numbers.",
    )
    parser.add_argument(
        "--fix-filter-pkg",
        help="Specify a regular expression to filter on package names.",
    )
    parser.add_argument(
        "--fix-filter-func",
        help="Specify a regular expression to filter on function names.",
    )
    cli = parser.parse_args()

    api_tracked = read_json(cli.current) if cli.current else {}

    def _get_api_src():
        api_src = read_json(cli.source) if cli.source else {}
        if not api_src:
            print(
                f"error: no source data found (path: {cli.source})", file=sys.stderr
            )
            sys.exit(1)
        return api_src

    if cli.current_tag:
        tag_to_versions(cli, cli.current_tag)
    elif cli.placeholder_versions:
        if cli.mode == "fix-versions":
            raise ValueError("fix-versions requires real version numbers")
        placeholder_versions(cli)

    if cli.mode == "compare":
        # just compare the json files. useful for CI
        api_src = _get_api_src()
        pcount = api_compare(api_tracked, api_src)
        if pcount:
            print(f"error: {pcount} problems detected", file=sys.stderr)
            sys.exit(1)
    elif cli.mode == "update":
        # update the current/tracked apis with those from the source
        api_src = _get_api_src()
        defaults = {}
        _setif(defaults, "added_in_version", cli.added_in_version)
        _setif(defaults, "expected_stable_version", cli.stable_in_version)
        _setif(defaults, "deprecated_in_version", cli.deprecated_in_version)
        _setif(defaults, "expected_remove_version", cli.remove_in_version)
        pcount = api_update(
            api_tracked,
            api_src,
            copy_stable=cli.copy_stable_apis,
            defaults=defaults,
        )
        if pcount:
            print(f"error: {pcount} problems detected", file=sys.stderr)
            sys.exit(1)
        write_json(cli.current, api_tracked)
        write_markdown(cli.document, api_tracked)
    elif cli.mode == "fix-versions":
        values = {}
        _setif(values, "added_in_version", cli.added_in_version)
        _setif(values, "expected_stable_version", cli.stable_in_version)
        _setif(values, "deprecated_in_version", cli.deprecated_in_version)
        _setif(values, "expected_remove_version", cli.remove_in_version)
        api_fix_versions(api_tracked, values=values, pred=_make_fix_filter(cli))
        write_json(cli.current, api_tracked)
    elif cli.mode == "find-updates":
        values = {}
        _setif(values, "next_version", cli.added_in_version)
        updates_needed = api_find_updates(api_tracked, values=values)
        json.dump(updates_needed, sys.stdout, indent=2)
        print()
        if not (updates_needed.get("preview") or updates_needed.get("deprecated")):
            sys.exit(1)
    elif cli.mode == "promote":
        values = {}
        api_src = _get_api_src()
        _setif(values, "added_in_version", cli.added_in_version)
        ccount, pcount = api_promote(
            api_tracked,
            api_src,
            values,
        )
        print("found {} apis to promote".format(ccount))
        if pcount:
            print(f"error: {pcount} problems detected", file=sys.stderr)
            sys.exit(1)
        write_json(cli.current, api_tracked)
    elif cli.mode == "write-doc":
        write_markdown(cli.document, api_tracked)
    elif cli.mode == "updates-to-markdown":
        updates_needed = json.load(sys.stdin)
        format_updates_markdown(updates_needed, sys.stdout)


if __name__ == "__main__":
    main()
