#!/usr/bin/python3
"""
apiage.py - a quick and dirty tool for tracking when apis become stable
and deprecated apis are to be removed.

PDX-License-Identifier: MIT
"""

import argparse
import copy
import json
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


def format_markdown(tracked, outfh):
    print("<!-- GENERATED FILE: DO NOT EDIT DIRECTLY -->", file=outfh)
    print("", file=outfh)
    print("# go-ceph API Stability", file=outfh)
    print("", file=outfh)
    for pkg, pkg_api in tracked.items():
        print(f"## Package: {pkg}", file=outfh)
        print("", file=outfh)
        if "preview_api" in pkg_api:
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
        if "deprecated_api" in pkg_api:
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
        choices=("compare", "update", "write-doc"),
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
    cli = parser.parse_args()

    api_src = read_json(cli.source) if cli.source else {}
    api_tracked = read_json(cli.current) if cli.current else {}

    if not api_src:
        print(
            f"error: no source data found (path: {cli.source})", file=sys.stderr
        )
        sys.exit(1)

    if cli.current_tag:
        tag_to_versions(cli, cli.current_tag)

    if cli.mode == "compare":
        # just compare the json files. useful for CI
        pcount = api_compare(api_tracked, api_src)
        if pcount:
            print(f"error: {pcount} problems detected", file=sys.stderr)
            sys.exit(1)
    elif cli.mode == "update":
        # update the current/tracked apis with those from the source
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
    elif cli.mode == "write-doc":
        write_markdown(cli.document, api_tracked)


if __name__ == "__main__":
    main()
