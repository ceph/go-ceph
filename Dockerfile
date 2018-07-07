FROM ubuntu:xenial

RUN apt-get update && apt-get install -y \
  apt-transport-https \
  git \
  golang-go \
  software-properties-common \
  uuid-runtime \
  wget

RUN wget -q -O- 'https://download.ceph.com/keys/release.asc' | apt-key add -
RUN apt-add-repository 'deb https://download.ceph.com/debian-luminous/ xenial main'

RUN apt-get update && apt-get install -y \
  ceph \
  libcephfs-dev \
  librados-dev \
  librbd-dev

ENV GOPATH /go
WORKDIR /go/src/github.com/ceph/go-ceph
VOLUME /go/src/github.com/ceph/go-ceph

COPY micro-osd.sh /
COPY entrypoint.sh /
ENTRYPOINT /entrypoint.sh
