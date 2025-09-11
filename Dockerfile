FROM --platform=$BUILDPLATFORM public.ecr.aws/docker/library/golang:1.25 AS build

ARG TARGETARCH
WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download
RUN GOARCH=${TARGETARCH} go build all

COPY --exclude=argocd/* --exclude=Dockerfile --exclude=argocd/* --exclude=go.mod --exclude=go.sum --exclude=Makefile . /usr/src/app
RUN GOARCH=${TARGETARCH} go build ./cmd/odin

FROM --platform=$TARGETPLATFORM rockylinux/rockylinux:9
RUN useradd -U -m odin &&  useradd -U -r -m -u 999 argocd
RUN dnf install -y jq
COPY --from=build --chown=root:root /usr/src/app/odin /usr/local/bin/odin
COPY argocd/*.sh /home/argocd/bin/
COPY argocd/plugin.yml /home/argocd/cmp-server/config/plugin.yaml

RUN chown -R argocd:argocd /home/argocd/cmp-server
USER odin
