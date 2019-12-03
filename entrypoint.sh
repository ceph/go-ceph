#!/bin/bash

set -e

TEST_RUN=ALL
PAUSE=no
MICRO_OSD_PATH="/micro-osd.sh"

CLI="$(getopt -o h --long test-run:,pause,micro-osd:,help -n "$0" -- "$@")"
eval set -- "${CLI}"
while true ; do
    case "$1" in
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

mkdir -p /tmp/ceph
"${MICRO_OSD_PATH}" /tmp/ceph
export CEPH_CONF=/tmp/ceph/ceph.conf

export PATH=/usr/lib/go-1.10/bin:$PATH

if [[ ${TEST_RUN} == NONE ]]; then
    echo "skipping test execution"
else
    go get -t -v ./...
    diff -u <(echo -n) <(gofmt -d -s .)
    #go vet ./...
    #go list ./...
    P=github.com/ceph/go-ceph
    testargs=(\
        "-covermode=count" \
        "-coverprofile=cover.out" \
        "-coverpkg=$P/cephfs,$P/rados,$P/rbd")
    # disable caching of tests results
    testargs+=("-count=1")
    if [[ ${TEST_RUN} != ALL ]]; then
        testargs+=("-run" "${TEST_RUN}")
    fi

    go test -v "${testargs[@]}" ./...
    mkdir -p /results/coverage
    go tool cover -html=cover.out -o /results/coverage/go-ceph.html
fi

if [[ ${PAUSE} = yes ]]; then
    sleep infinity
fi
