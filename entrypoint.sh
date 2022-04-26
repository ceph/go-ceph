#!/bin/bash

set -e

TEST_RUN=ALL
PAUSE=no
COVERAGE=yes
CPUPROFILE=no
MEMPROFILE=no
MICRO_OSD_PATH="/micro-osd.sh"
BUILD_TAGS=""
RESULTS_DIR=/results
CEPH_CONF=/tmp/ceph/ceph.conf
MIRROR_STATE=/dev/null
PKG_PREFIX=github.com/ceph/go-ceph


# Default env vars that are not currently changed by this script
# but can be used to change the test behavior:
# GO_CEPH_TEST_MDS_NAME

CLI="$(getopt -o h --long test-run:,test-bench:,test-pkg:,pause,cpuprofile,memprofile,no-cover,micro-osd:,wait-for:,results:,ceph-conf:,mirror:,mirror-state:,help -n "${0}" -- "$@")"
eval set -- "${CLI}"
while true ; do
    case "${1}" in
        --test-pkg)
            TEST_PKG="${2}"
            shift
            shift
        ;;
        --test-run)
            TEST_RUN="${2}"
            shift
            shift
        ;;
        --test-bench)
            TEST_BENCH="${2}"
            shift
            shift
        ;;
        --pause)
            PAUSE=yes
            shift
        ;;
        --micro-osd)
            MICRO_OSD_PATH="${2}"
            shift
            shift
        ;;
        --wait-for)
            WAIT_FILES="${2}"
            shift
            shift
        ;;
        --results)
            RESULTS_DIR="${2}"
            shift
            shift
        ;;
        --ceph-conf)
            CEPH_CONF="${2}"
            shift
            shift
        ;;
        --mirror)
            MIRROR_CONF="${2}"
            shift
            shift
        ;;
        --mirror-state)
            MIRROR_STATE="${2}"
            shift
            shift
        ;;
        --cpuprofile)
            CPUPROFILE=yes
            shift
        ;;
        --memprofile)
            MEMPROFILE=yes
            shift
        ;;
        --no-cover)
            COVERAGE=no
            shift
        ;;
        -h|--help)
            echo "Options:"
            echo "  --test-run=VALUE    Run selected test or ALL, NONE,"
            echo "                      IMPLEMENTS"
            echo "                      ALL is the default"
            echo "  --test-bench=VALUE  Run selected benchmarks"
            echo "  --test-pkg=PKG      Run only tests from PKG"
            echo "  --pause             Sleep forever after tests execute"
            echo "  --micro-osd         Specify path to micro-osd script"
            echo "  --wait-for=FILES    Wait for files before starting tests"
            echo "                      (colon separated, disables micro-osd)"
            echo "  --mirror-state=PATH Path to track state of (rbd) mirroring"
            echo "  --results=PATH      Specify path to store test results"
            echo "  --ceph-conf=PATH    Specify path to ceph configuration"
            echo "  --mirror=PATH       Specify path to ceph conf of mirror"
            echo "  --cpuprofile        Run tests with cpu profiling"
            echo "  --memprofile        Run tests with mem profiling"
            echo "  --no-cover          Disable code coverage profiling"
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

add_build_tag() {
    local val="$1"
    if [ -n "$BUILD_TAGS" ]; then
        BUILD_TAGS+=",${val}"
    else
        BUILD_TAGS="${val}"
    fi
}

show() {
    local ret
    echo "*** running:" "$@"
    "$@"
    ret=$?
    if [ ${ret} -ne 0 ] ; then
        echo "*** ERROR: returned ${ret}"
    fi
    return ${ret}
}

wait_for_files() {
    for file in "$@" ; do
        echo -n "*** waiting for $file ..."
        while ! [[ -f $file ]] ; do
            sleep 1
        done
        echo "done"
    done
}

test_failed() {
    local pkg="${1}"
    echo "*** ERROR: ${pkg} tests failed"
    pause_if_needed
    return 1
}

setup_mirroring() {
    mstate="$(cat "${MIRROR_STATE}" 2>/dev/null || true)"
    if [[ "$mstate" = functional ]]; then
        echo "Mirroring already functional"
        return 0
    fi
    echo "Setting up mirroring..."
    local CONF_A=${CEPH_CONF}
    local CONF_B=${MIRROR_CONF}
    ceph -c $CONF_A osd pool create rbd 8
    ceph -c $CONF_B osd pool create rbd 8
    rbd -c $CONF_A pool init
    rbd -c $CONF_B pool init
    rbd -c $CONF_A mirror pool enable rbd image
    rbd -c $CONF_B mirror pool enable rbd image
    rbd -c $CONF_A mirror pool peer bootstrap create --site-name ceph_a rbd > token
    rbd -c $CONF_B mirror pool peer bootstrap import --site-name ceph_b rbd token

    echo "enabled" > "${MIRROR_STATE}"
    rbd -c $CONF_A rm mirror_test 2>/dev/null || true
    rbd -c $CONF_B rm mirror_test 2>/dev/null || true
    (echo "Mirror Test"; dd if=/dev/zero bs=1 count=500K) | rbd -c $CONF_A import - mirror_test
    rbd -c $CONF_A mirror image enable mirror_test snapshot
    echo -n "Waiting for mirroring activation..."
    while ! rbd -c $CONF_A mirror image status mirror_test \
      | grep -q "state: \+up+replaying" ; do
        sleep 1
    done
    echo "done"
    rbd -c $CONF_A mirror image snapshot mirror_test
    echo -n "Waiting for mirror sync..."
    while ! rbd -c $CONF_B export mirror_test - 2>/dev/null | grep -q "Mirror Test" ; do
        sleep 1
    done
    echo "functional" > "${MIRROR_STATE}"
    echo " mirroring functional!"
}

test_pkg() {
    local pkg="${1}"
    if [[ "${TEST_PKG}" && "${TEST_PKG}" != "${pkg}" ]]; then
        return 0
    fi

    # run go vet and capture the result for the package, but still execute the
    # test suite anyway
    show go vet ${BUILD_TAGS} "./${pkg}"
    ret=$?

    # disable caching of tests results
    testargs=("-count=1"\
            ${BUILD_TAGS})
    if [[ ${TEST_RUN} != ALL ]]; then
        testargs+=("-run" "${TEST_RUN}")
    fi
    if [[ -n ${TEST_BENCH} ]]; then
        testargs+=("-bench" "${TEST_BENCH}")
    fi
    if [[ ${COVERAGE} = yes ]]; then
        testargs+=(\
            "-covermode=count" \
            "-coverprofile=${pkg}.cover.out" \
            "-coverpkg=${PKG_PREFIX}/${pkg}")
    fi
    if [[ ${CPUPROFILE} = yes ]]; then
        testargs+=("-cpuprofile" "${pkg}.cpu.out")
    fi
    if [[ ${MEMPROFILE} = yes ]]; then
        testargs+=("-memprofile" "${pkg}.mem.out")
    fi

    show go test -v "${testargs[@]}" "./${pkg}"
    ret=$(($?+${ret}))
    grep -v "^mode: count" "${pkg}.cover.out" >> "cover.out"
    return ${ret}
}

pre_all_tests() {
    # Prepare Go code
    go get -t -v ${BUILD_TAGS} ./...
    diff -u <(echo -n) <(gofmt -d -s .)
    make clean-implements implements

    # Reset whole-module coverage file
    echo "mode: count" > "cover.out"
}

implements_tool() {
    mkdir -p "${RESULTS_DIR}"
    pkgs=$(go list ${BUILD_TAGS} ./... | sed -e "s,^${PKG_PREFIX}/\?,," | \
        grep -v ^contrib | grep -v ^internal)
    show ./implements --list \
        --report-json "${RESULTS_DIR}/implements.json" \
        --report-text "${RESULTS_DIR}/implements.txt" \
        ${pkgs}
    # output the brief summary info onto stdout
    grep '^[A-Z]' "${RESULTS_DIR}/implements.txt"
}

post_all_tests() {
    if [[ ${COVERAGE} = yes ]]; then
        mkdir -p "${RESULTS_DIR}/coverage"
        show go tool cover -html=cover.out -o "${RESULTS_DIR}/coverage/go-ceph.html"
    fi
    if [[ ${COVERAGE} = yes ]]; then
        implements_tool
    fi
}

test_go_ceph() {
    mkdir -p /tmp/ceph
    if ! [[ ${WAIT_FILES} ]]; then
        show "${MICRO_OSD_PATH}" /tmp/ceph
    fi
    export CEPH_CONF

    if [[ ${TEST_RUN} == NONE ]]; then
        echo "skipping test execution"
        return 0
    fi
    if [[ ${TEST_RUN} == IMPLEMENTS ]]; then
        echo "skipping tests, executing implements tool"
        pre_all_tests
        implements_tool
        return $?
    fi

    pkgs=$(go list ${BUILD_TAGS} ./... | sed -e "s,^${PKG_PREFIX}/\?,," | grep -v ^contrib)
    pre_all_tests
    if [[ ${WAIT_FILES} ]]; then
        wait_for_files ${WAIT_FILES//:/ }
    fi
    if [[ ${MIRROR_CONF} && ${CEPH_VERSION} != nautilus ]]; then
        setup_mirroring
        export MIRROR_CONF
    fi
    for pkg in ${pkgs}; do
        test_pkg "${pkg}" || test_failed "${pkg}"
    done
    post_all_tests
}

pause_if_needed() {
    if [[ ${PAUSE} = yes ]]; then
        echo "*** pausing execution"
        sleep infinity
    fi
}


if [ -n "${CEPH_VERSION}" ]; then
    add_build_tag "${CEPH_VERSION}"
fi

if [ -n "${NO_PTRGUARD}" ]; then
    add_build_tag "no_ptrguard"
fi

if [ -z "${NO_PREVIEW}" ]; then
    add_build_tag "ceph_preview"
fi

if [ -n "${BUILD_TAGS}" ]; then
    BUILD_TAGS="-tags ${BUILD_TAGS}"
fi

test_go_ceph
pause_if_needed

# vim: set ts=4 sw=4 sts=4 et:
