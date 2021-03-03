CI_IMAGE_NAME = go-ceph-ci
CONTAINER_CMD ?=
CONTAINER_OPTS := --security-opt $(shell grep -q selinux /sys/kernel/security/lsm 2>/dev/null && echo "label=disable" || echo "apparmor:unconfined")
CONTAINER_CONFIG_DIR := testing/containers/ceph
VOLUME_FLAGS :=
CEPH_VERSION := octopus
RESULTS_DIR :=
CHECK_GOFMT_FLAGS := -e -s -l
IMPLEMENTS_OPTS :=
BUILD_TAGS := $(CEPH_VERSION)

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

ifneq ($(USE_PTRGUARD),)
	CONTAINER_OPTS += -e USE_PTRGUARD=true
	BUILD_TAGS := $(BUILD_TAGS),ptrguard
endif

SELINUX := $(shell getenforce 2>/dev/null)
ifeq ($(SELINUX),Enforcing)
	VOLUME_FLAGS = :z
endif

ifdef RESULTS_DIR
	RESULTS_VOLUME := -v $(RESULTS_DIR):/results$(VOLUME_FLAGS)
endif

build:
	go build -v -tags $(BUILD_TAGS) $(shell go list ./... | grep -v /contrib)
fmt:
	go fmt ./...
test:
	go test -v -tags $(BUILD_TAGS) ./...

.PHONY: test-docker test-container test-multi-container
test-docker: test-container
test-container: $(BUILDFILE) $(RESULTS_DIR)
	$(CONTAINER_CMD) run $(CONTAINER_OPTS) --rm --hostname test_ceph_aio -v $(CURDIR):/go/src/github.com/ceph/go-ceph$(VOLUME_FLAGS) $(RESULTS_VOLUME) $(CI_IMAGE_TAG) $(ENTRYPOINT_ARGS)
test-multi-container: $(BUILDFILE) $(RESULTS_DIR)
	$(CONTAINER_CMD) kill test_ceph_a test_ceph_b 2>/dev/null || true
	$(CONTAINER_CMD) volume remove test_ceph_a_data test_ceph_b_data 2>/dev/null || true
	$(CONTAINER_CMD) network create test_ceph_net 2>/dev/null || true
	$(CONTAINER_CMD) run $(CONTAINER_OPTS) --rm -d --name test_ceph_a --hostname test_ceph_a --net test_ceph_net \
		-v test_ceph_a_data:/tmp/ceph $(CI_IMAGE_TAG) --test-run=NONE --pause
	$(CONTAINER_CMD) run $(CONTAINER_OPTS) --rm -d --name test_ceph_b --hostname test_ceph_b --net test_ceph_net \
		-v test_ceph_b_data:/tmp/ceph $(CI_IMAGE_TAG) --test-run=NONE --pause
	$(CONTAINER_CMD) run $(CONTAINER_OPTS) --rm \
		--net test_ceph_net -v test_ceph_a_data:/ceph_a -v test_ceph_b_data:/ceph_b \
		-v $(CURDIR):/go/src/github.com/ceph/go-ceph$(VOLUME_FLAGS) $(RESULTS_VOLUME) \
		$(CI_IMAGE_TAG) --wait-for=/ceph_a/.ready:/ceph_b/.ready --ceph-conf=/ceph_a/ceph.conf \
		--mirror=/ceph_b/ceph.conf $(ENTRYPOINT_ARGS)
	$(CONTAINER_CMD) kill test_ceph_a test_ceph_b

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
	rbd.test \
	rbd/admin.test
test-bins: test-binaries

%.test: % force_go_build
	go test -c -tags $(BUILD_TAGS) ./$<

implements:
	go build -o implements ./contrib/implements

check-implements: implements
	./implements $(IMPLEMENTS_OPTS) ./cephfs ./rados ./rbd

# force_go_build is phony and builds nothing, can be used for forcing
# go toolchain commands to always run
.PHONY: build fmt test test-docker check test-binaries test-bins force_go_build check-implements
