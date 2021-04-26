ARG CEPH_VERSION
FROM go-ceph-ci:${CEPH_VERSION}

RUN cd /tmp && \
  go get -v \
    golang.org/x/tools/gopls \
    honnef.co/go/tools/... \
    golang.org/x/lint/golint \
    github.com/mgechev/revive \
    github.com/uudashr/gopkgs/v2/cmd/gopkgs \
    github.com/ramya-rao-a/go-outline \
    github.com/go-delve/delve/cmd/dlv \
    github.com/golangci/golangci-lint/cmd/golangci-lint
