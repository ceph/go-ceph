ARG CEPH_IMG=quay.io/ceph/ceph
ARG CEPH_TAG=v19
FROM ${CEPH_IMG}:${CEPH_TAG}

# A CEPH_VERSION env var is already set in the base image.
# We save our build arg as GO_CEPH_VERSION and later ensure that
# the values agree to ensure we're building what we meant to build.
ARG GO_CEPH_VERSION
ENV GO_CEPH_VERSION=${GO_CEPH_VERSION:-$CEPH_VERSION}

COPY setup-pkgs.sh /tmp/setup-pkgs.sh
RUN /tmp/setup-pkgs.sh && rm /tmp/setup-pkgs.sh

ARG GO_VERSION=1.25.8
ENV GO_VERSION=${GO_VERSION}
ARG GOARCH
ENV GOARCH=${GOARCH}
ARG GOPROXY
ENV GOPROXY=${GOPROXY}
COPY setup-go.sh /tmp/setup-go.sh
RUN /tmp/setup-go.sh && rm /tmp/setup-go.sh

ENV PATH="${PATH}:/opt/go/bin"
ENV GOROOT=/opt/go
ENV GO111MODULE=on
ENV GOPATH=/go
ENV GOCACHE=/go/cache
WORKDIR /go/src/github.com/ceph/go-ceph
VOLUME /go/src/github.com/ceph/go-ceph

COPY micro-osd.sh /
COPY entrypoint.sh /
ENTRYPOINT ["/entrypoint.sh"]
