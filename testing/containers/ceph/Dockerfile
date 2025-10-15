ARG CEPH_IMG=quay.io/ceph/ceph
ARG CEPH_TAG=v19
FROM ${CEPH_IMG}:${CEPH_TAG}

# A CEPH_VERSION env var is already set in the base image.
# We save our build arg as GO_CEPH_VERSION and later ensure that
# the values agree to ensure we're building what we meant to build.
ARG GO_CEPH_VERSION
ENV GO_CEPH_VERSION=${GO_CEPH_VERSION:-$CEPH_VERSION}

RUN true \
    && if [ -z "${CEPH_VERSION}" ]; then CEPH_VERSION="${CEPH_REF}"; fi \
    && echo "Check: [ ${CEPH_VERSION} = ${GO_CEPH_VERSION} ]" \
    && [ "${CEPH_VERSION}" = "${GO_CEPH_VERSION}" ] \
    && (. /etc/os-release ; if [ "$ID" = centos -a "$VERSION" = 8 ]; then find /etc/yum.repos.d/ -name '*.repo' -exec sed -i -e 's|^mirrorlist=|#mirrorlist=|g' -e 's|^#baseurl=http://mirror.centos.org|baseurl=https://vault.centos.org|g'  {} \; ; fi ) \
    && if [ ! -f /etc/yum.repos.d/ceph.repo -a "$CEPH_IS_DEVEL" = true ]; then if [ -z "$CEPH_SHA1" ]; then CEPH_SHA1="$(sed -n 's/.*CEPH_GIT_VER *= *"\(.*\)".*/\1/p' /usr/bin/ceph)"; fi ; yum reinstall -y "$(curl -fs "https://shaman.ceph.com/api/search/?project=ceph&distros=centos/9/x86_64&flavor=default&ref=${CEPH_REF}&sha1=${CEPH_SHA1:-latest}" | jq -r .[0].url)/noarch/ceph-release-1-0.el9.noarch.rpm"; fi \
    && if [ ! -f /etc/yum.repos.d/ceph.repo -a "$CEPH_IS_DEVEL" != true ]; then yum reinstall -y "https://download.ceph.com/rpm-${CEPH_REF}/el9/noarch/ceph-release-1-1.el9.noarch.rpm"; fi \
    && yum install -y \
        git wget /usr/bin/curl make \
        /usr/bin/cc /usr/bin/c++ gdb \
        libcephfs-devel \
        librados-devel \
        librbd-devel \
        libradosstriper-devel \
        libcephfs2-debuginfo \
        librados2-debuginfo \
        librbd1-debuginfo \
        libradosstriper1-debuginfo \
    && yum clean all \
    && true

ARG GO_VERSION=1.24.9
ENV GO_VERSION=${GO_VERSION}
ARG GOARCH
ENV GOARCH=${GOARCH}
ARG GOPROXY
ENV GOPROXY=${GOPROXY}
RUN true && \
    gotar=go${GO_VERSION}.linux-${GOARCH}.tar.gz && \
    gourl="https://dl.google.com/go/${gotar}" && \
    curl -o /tmp/${gotar} "${gourl}" && \
    tar -x -C /opt/ -f /tmp/${gotar} && \
    rm -f /tmp/${gotar} && \
    true

ENV PATH="${PATH}:/opt/go/bin"
ENV GOROOT=/opt/go
ENV GO111MODULE=on
ENV GOPATH /go
ENV GOCACHE=/go/cache
WORKDIR /go/src/github.com/ceph/go-ceph
VOLUME /go/src/github.com/ceph/go-ceph

COPY micro-osd.sh /
COPY entrypoint.sh /
ENTRYPOINT ["/entrypoint.sh"]
