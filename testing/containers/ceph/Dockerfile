ARG CEPH_VERSION
FROM ceph/daemon-base:latest-${CEPH_VERSION:-nautilus}

ENV CEPH_VERSION=${CEPH_VERSION:-nautilus}

RUN true && \
    yum clean all && \
    ( if [ "${CEPH_VERSION}" = "octopus" ]; then dnf install -y --repo BaseOS 'dnf-command(config-manager)' && dnf config-manager --set-disabled epel && dnf update -y; fi ) && \
    yum install -y git wget curl libcephfs-devel librados-devel librbd-devel /usr/bin/cc /usr/bin/c++ make && \
    (yum install -y /usr/bin/castxml || true) && \
    true

ENV GOTAR=go1.14.7.linux-amd64.tar.gz
RUN true && \
    curl -o /tmp/${GOTAR} https://dl.google.com/go/${GOTAR} && \
    tar -x -C /opt/ -f /tmp/${GOTAR} && \
    rm -f /tmp/${GOTAR} && \
    true

ENV PATH="${PATH}:/opt/go/bin"
ENV GOROOT=/opt/go
ENV GO111MODULE=on
ENV GOPATH /go
WORKDIR /go/src/github.com/ceph/go-ceph
VOLUME /go/src/github.com/ceph/go-ceph

COPY micro-osd.sh /
COPY entrypoint.sh /
ENTRYPOINT ["/entrypoint.sh"]
