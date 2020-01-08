DOCKER_CI_IMAGE = go-ceph-ci
CONTAINER_CMD := docker
CONTAINER_OPTS := --security-opt $(shell grep -q selinux /sys/kernel/security/lsm && echo "label=disabled" || echo "apparmor:unconfined")
VOLUME_FLAGS := 

SELINUX := $(shell getenforce 2>/dev/null)
ifeq ($(SELINUX),Enforcing)
	VOLUME_FLAGS = :z
endif

build:
	go build -v $(shell go list ./... | grep -v /contrib)
fmt:
	go fmt ./...
test:
	go test -v ./...

test-docker: .build-docker
	$(CONTAINER_CMD) run --device /dev/fuse --cap-add SYS_ADMIN $(CONTAINER_OPTS) --rm -it -v $(CURDIR):/go/src/github.com/ceph/go-ceph$(VOLUME_FLAGS) $(DOCKER_CI_IMAGE)

.build-docker: Dockerfile entrypoint.sh
	$(CONTAINER_CMD) build -t $(DOCKER_CI_IMAGE) .
	@$(CONTAINER_CMD) inspect -f '{{.Id}}' $(DOCKER_CI_IMAGE) > .build-docker

check:
	# Configure project's revive checks using .revive.toml
	# See: https://github.com/mgechev/revive
	@for d in $$(go list ./... | grep -v /vendor/); do revive -config .revive.toml $${d}; done
