#!/bin/bash

set -e

TEST_RUN=ALL
PAUSE=no
MICRO_OSD_PATH="/micro-osd.sh"

CLI="$(getopt -o h --long test-run:,test-pkg:,pause,micro-osd:,help -n "$0" -- "$@")"
eval set -- "${CLI}"
while true ; do
    case "$1" in
        --test-pkg)
            TEST_PKG="$2"
            shift
            shift
        ;;
        --test-run)
            TEST_RUN="$2"
            shift
            shift
        ;;
        --pause)
            PAUSE=yes
            shift
        ;;
        --micro-osd)
            MICRO_OSD_PATH="$2"
            shift
            shift
        ;;
        -h|--help)
            echo "Options:"
            echo "  --test-run=VALUE    Run selected test or ALL, NONE"
            echo "                      ALL is the default"
            echo "  --test-pkg=PKG      Run only tests from PKG"
            echo "  --pause             Sleep forever after tests execute"
            echo "  --micro-osd         Specify path to micro-osd script"
            echo "  -h|--help           Display help text"
            echo ""
            exit 0
        ;;
        --)
            shift
            break
        ;;
        *)
            echo "unknown option" >&2
            exit 2
        ;;
    esac
done

show() {
    echo "*** running:" "$@"
    "$@"
}

test_failed() {
    local pkg="$1"
    echo "*** ERROR: ${pkg} tests failed"
    pause_if_needed
    return 1
}

test_pkg() {
    local pkg="$1"
    if [[ "$TEST_PKG" && "$TEST_PKG" != "$pkg" ]]; then
        return 0
    fi
    testargs=(\
        "-covermode=count" \
        "-coverprofile=$pkg.cover.out" \
        "-coverpkg=$P/$pkg")
    # disable caching of tests results
    testargs+=("-count=1")
    if [[ ${TEST_RUN} != ALL ]]; then
        testargs+=("-run" "${TEST_RUN}")
    fi

    show go test -v "${testargs[@]}" "./$pkg"
    ret=$?
    grep -v "^mode: count" "$pkg.cover.out" >> "cover.out"
    return ${ret}
}

pre_all_tests() {
    # Prepare Go code
    go get -t -v ./...
    diff -u <(echo -n) <(gofmt -d -s .)

    # TODO: Consider enabling go vet but it currently fails

    # Reset whole-module coverage file
    echo "mode: count" > "cover.out"
}

post_all_tests() {
    mkdir -p /results/coverage
    show go tool cover -html=cover.out -o /results/coverage/go-ceph.html
}

test_go_ceph() {
    mkdir -p /tmp/ceph
    show "${MICRO_OSD_PATH}" /tmp/ceph
    export CEPH_CONF=/tmp/ceph/ceph.conf

    if [[ ${TEST_RUN} == NONE ]]; then
        echo "skipping test execution"
        return 0
    fi

    P=github.com/ceph/go-ceph
    pkgs=(\
        "cephfs" \
        "errutil" \
        "rados" \
        "rbd" \
        )
    pre_all_tests
    for pkg in "${pkgs[@]}"; do
        test_pkg "$pkg" || test_failed "$pkg"
    done
    post_all_tests
}

pause_if_needed() {
    if [[ ${PAUSE} = yes ]]; then
        echo "*** pausing execution"
        sleep infinity
    fi
}

test_go_ceph
pause_if_needed
