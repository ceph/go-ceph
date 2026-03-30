#!/bin/bash
set -e

if [ -z "${CEPH_VERSION}" ]; then
    CEPH_VERSION="${CEPH_REF}"
fi

echo "Check: [ ${CEPH_VERSION} = ${GO_CEPH_VERSION} ]"
[ "${CEPH_VERSION}" = "${GO_CEPH_VERSION}" ]

# shellcheck disable=SC1091
. /etc/os-release
if [ "$ID" = "centos" ] && [ "$VERSION" = "8" ]; then
    find /etc/yum.repos.d/ -name '*.repo' -exec \
        sed -i \
            -e 's|^mirrorlist=|#mirrorlist=|g' \
            -e 's|^#baseurl=http://mirror.centos.org|baseurl=https://vault.centos.org|g' \
            {} \;
fi

DISTRO_VERSION=${VERSION%%[^0-9]*}
if [ ! -f /etc/yum.repos.d/ceph.repo ]; then
    if [ "$CEPH_IS_DEVEL" = "true" ]; then
        if [ -z "$CEPH_SHA1" ]; then
            CEPH_SHA1="$(sed -n 's/.*CEPH_GIT_VER *= *"\(.*\)".*/\1/p' /usr/bin/ceph)"
        fi
        REPO_URL=$(curl -fs "https://shaman.ceph.com/api/search/?project=ceph&distros=${ID}/${DISTRO_VERSION}/x86_64&flavor=default&ref=${CEPH_REF}&sha1=${CEPH_SHA1:-latest}" | jq -r .[0].url)
        yum reinstall -y "${REPO_URL}/noarch/ceph-release-1-0.el${DISTRO_VERSION}.noarch.rpm"
    else
        yum reinstall -y "https://download.ceph.com/rpm-${CEPH_REF}/el${DISTRO_VERSION}/noarch/ceph-release-1-1.el${DISTRO_VERSION}.noarch.rpm"
    fi
fi

yum install -y \
    git \
    wget \
    /usr/bin/curl \
    make \
    /usr/bin/cc \
    /usr/bin/c++ \
    gdb \
    libcephfs-devel \
    librados-devel \
    librbd-devel \
    libradosstriper-devel \
    libcephfs2-debuginfo \
    librados2-debuginfo \
    librbd1-debuginfo \
    libradosstriper1-debuginfo

yum clean all
