DOCKER_CI_IMAGE = go-ceph-ci
CONTAINER_CMD := docker
CONTAINER_OPTS := --security-opt apparmor:unconfined
VOLUME_FLAGS := 

SELINUX := $(shell getenforce 2>/dev/null)
ifeq ($(SELINUX),Enforcing)
	VOLUME_FLAGS = :z
endif

build:
	go build -v
fmt:
	go fmt ./...
test:
	go test -v ./...

test-docker: .build-docker
	$(CONTAINER_CMD) run --device /dev/fuse --cap-add SYS_ADMIN $(CONTAINER_OPTS) --rm -it -v $(CURDIR):/go/src/github.com/ceph/go-ceph$(VOLUME_FLAGS) $(DOCKER_CI_IMAGE)

.build-docker:
	$(CONTAINER_CMD) build -t $(DOCKER_CI_IMAGE) .
	@$(CONTAINER_CMD) inspect -f '{{.Id}}' $(DOCKER_CI_IMAGE) > .build-docker

check:
	# TODO: add this when golint is fixed	@for d in $$(go list ./... | grep -v /vendor/); do golint -set_exit_status $${d}; done
	@for d in $$(go list ./... | grep -v /vendor/); do golint $${d}; done
