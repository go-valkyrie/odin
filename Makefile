.PHONY: container push-container container-local test
REPOSITORY = ghcr.io/go-valkyrie/odin
TAG = $(shell git describe --tags)

all: container-local

test:
	go test ./...

push-container: container
ifeq (,$(REPOSITORY))
	$(error Invalid repository $(REPOSITORY))
endif
ifeq (,$(TAG))
	$(error Invalid image tag $(TAG))
endif
	podman manifest rm "$(REPOSITORY):$(TAG)" || exit 0
	podman manifest create "$(REPOSITORY):$(TAG)"
	podman manifest add "$(REPOSITORY):$(TAG)" "odin:amd64"
	podman manifest add "$(REPOSITORY):$(TAG)" "odin:arm64"
	podman manifest push "$(REPOSITORY):$(TAG)"

container:
	podman build --pull=missing --platform=linux/amd64 -t "odin:amd64" .
	podman build --pull=missing --platform=linux/arm64 -t "odin:arm64" .

container-local:
	podman build -t odin:latest --pull=missing .