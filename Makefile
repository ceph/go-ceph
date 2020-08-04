CI_IMAGE_NAME = go-ceph-ci
CONTAINER_CMD ?=
CONTAINER_OPTS := --security-opt $(shell grep -q selinux /sys/kernel/security/lsm && echo "label=disable" || echo "apparmor:unconfined")
CONTAINER_CONFIG_DIR := testing/containers/ceph
VOLUME_FLAGS :=
CEPH_VERSION := nautilus
RESULTS_DIR :=
CHECK_GOFMT_FLAGS := -e -s -l
IMPLEMENTS_OPTS :=

ifeq ($(CONTAINER_CMD),)
	CONTAINER_CMD:=$(shell docker version >/dev/null 2>&1 && echo docker)
endif
ifeq ($(CONTAINER_CMD),)
	CONTAINER_CMD:=$(shell podman version >/dev/null 2>&1 && echo podman)
endif

# the full name of the marker file including the ceph version
BUILDFILE=.build.$(CEPH_VERSION)

# the name of the image plus ceph version as tag
CI_IMAGE_TAG=$(CI_IMAGE_NAME):$(CEPH_VERSION)

SELINUX := $(shell getenforce 2>/dev/null)
ifeq ($(SELINUX),Enforcing)
	VOLUME_FLAGS = :z
endif

ifdef RESULTS_DIR
	RESULTS_VOLUME := -v $(RESULTS_DIR):/results$(VOLUME_FLAGS)
endif

build:
	go build -v -tags $(CEPH_VERSION) $(shell go list ./... | grep -v /contrib)
fmt:
	go fmt ./...
test:
	go test -v -tags $(CEPH_VERSION) ./...

.PHONY: test-docker test-container
test-docker: test-container
test-container: $(BUILDFILE) $(RESULTS_DIR)
	$(CONTAINER_CMD) run --device /dev/fuse --cap-add SYS_ADMIN $(CONTAINER_OPTS) --rm -v $(CURDIR):/go/src/github.com/ceph/go-ceph$(VOLUME_FLAGS) $(RESULTS_VOLUME) $(CI_IMAGE_TAG)

ifdef RESULTS_DIR
$(RESULTS_DIR):
	mkdir -p $(RESULTS_DIR)
endif

.PHONY: ci-image
ci-image: $(BUILDFILE)
$(BUILDFILE): $(CONTAINER_CONFIG_DIR)/Dockerfile entrypoint.sh micro-osd.sh
	$(CONTAINER_CMD) build --build-arg CEPH_VERSION=$(CEPH_VERSION) -t $(CI_IMAGE_TAG) -f $(CONTAINER_CONFIG_DIR)/Dockerfile .
	@$(CONTAINER_CMD) inspect -f '{{.Id}}' $(CI_IMAGE_TAG) > $(BUILDFILE)
	echo $(CEPH_VERSION) >> $(BUILDFILE)

check: check-revive check-format

check-format:
	! gofmt $(CHECK_GOFMT_FLAGS) . | sed 's,^,formatting error: ,' | grep 'go$$'

check-revive:
	# Configure project's revive checks using .revive.toml
	# See: https://github.com/mgechev/revive
	revive -config .revive.toml $$(go list ./... | grep -v /vendor/)

# Do a quick compile only check of the tests and impliclity the
# library code as well.
test-binaries: \
	cephfs.test \
	cephfs/admin.test \
	internal/callbacks.test \
	internal/cutil.test \
	internal/errutil.test \
	internal/retry.test \
	rados.test \
	rbd.test
test-bins: test-binaries

%.test: % force_go_build
	go test -c -tags $(CEPH_VERSION) ./$<

implements:
	go build -o implements ./contrib/implements

check-implements: implements
	./implements $(IMPLEMENTS_OPTS) ./cephfs ./rados ./rbd

# force_go_build is phony and builds nothing, can be used for forcing
# go toolchain commands to always run
.PHONY: build fmt test test-docker check test-binaries test-bins force_go_build check-implements
