ARG CEPH_VERSION
FROM ceph/daemon-base:latest-${CEPH_VERSION:-octopus}

ENV CEPH_VERSION=${CEPH_VERSION:-octopus}

RUN true && \
    yum update -y && \
    cv="$(rpm -q --queryformat '%{version}-%{release}' ceph-common)" && \
    yum install -y \
        git wget curl make \
        /usr/bin/cc /usr/bin/c++ \
        "libcephfs-devel-${cv}" "librados-devel-${cv}" "librbd-devel-${cv}" && \
    yum clean all && \
    true

ENV GOTAR=go1.15.10.linux-amd64.tar.gz
RUN true && \
    curl -o /tmp/${GOTAR} https://dl.google.com/go/${GOTAR} && \
    tar -x -C /opt/ -f /tmp/${GOTAR} && \
    rm -f /tmp/${GOTAR} && \
    true

ENV PATH="${PATH}:/opt/go/bin"
ENV GOROOT=/opt/go
ENV GO111MODULE=on
ENV GOPATH /go
ENV GODEBUG=cgocheck=2
ENV GOCACHE=/go/cache
WORKDIR /go/src/github.com/ceph/go-ceph
VOLUME /go/src/github.com/ceph/go-ceph

COPY micro-osd.sh /
COPY entrypoint.sh /
ENTRYPOINT ["/entrypoint.sh"]
