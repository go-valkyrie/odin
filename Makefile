.PHONY: container container-local test

all: container-local

test:
	go test ./...

container:
	podman build --pull=newer --platform=linux/amd64 -t "odin:amd64" .
	podman build --pull=newer --platform=linux/arm64 -t "odin:arm64" .

container-local:
	podman build -t odin:latest --pull=missing .