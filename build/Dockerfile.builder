ARG GOLANG_VERSION=1.22.2-alpine3.19
ARG BUILDKIT_VERSION=v0.12.4-rootless
ARG ALPINE_VERSION=3.19

FROM alpine:${ALPINE_VERSION} as cosign

ARG USE_MIRROR=false
ARG COSIGN_VERSION=v2.2.2
ARG TARGETOS TARGETARCH

RUN set -eux && \
  if [ "$USE_MIRROR" = true ]; then sed -i "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g" /etc/apk/repositories; fi && \
  apk add --no-cache wget && \
  wget -O /tmp/cosign https://github.com/sigstore/cosign/releases/download/"${COSIGN_VERSION}"/cosign-"${TARGETOS}"-"${TARGETARCH}" && \
  chmod +x /tmp/cosign

FROM --platform=$BUILDPLATFORM golang:${GOLANG_VERSION} as builder

ARG USE_MIRROR=false

RUN set -eux && \
  if [ "$USE_MIRROR" = true ]; then sed -i "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g" /etc/apk/repositories; fi && \
  apk add --no-cache make bash ncurses build-base git openssl && \
  apk add --no-cache zig --repository=https://mirrors.aliyun.com/alpine/edge/testing

COPY . /go/src/github.com/go-sigma/sigma
WORKDIR /go/src/github.com/go-sigma/sigma

ARG TARGETOS TARGETARCH

RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
  case "${TARGETARCH}" in \
		amd64) export CC="zig cc -target x86_64-linux-musl" ;; \
		arm64) export CC="zig cc -target aarch64-linux-musl" ;; \
	esac; \
  case "${TARGETARCH}" in \
		amd64) export CXX="zig c++ -target x86_64-linux-musl" ;; \
		arm64) export CXX="zig c++ -target aarch64-linux-musl" ;; \
	esac; \
  GOOS=$TARGETOS GOARCH=$TARGETARCH CC="${CC}" CXX="${CXX}" make build-builder

FROM moby/buildkit:${BUILDKIT_VERSION}

ARG USE_MIRROR=false

USER root
RUN set -eux && \
  if [ "$USE_MIRROR" = true ]; then sed -i "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g" /etc/apk/repositories; fi && \
  apk add --no-cache git-lfs && \
  mkdir -p /code/ && \
  chown -R 1000:1000 /opt/ && \
  chown -R 1000:1000 /code/

COPY --from=cosign /tmp/cosign /usr/local/bin/cosign
COPY --from=builder /go/src/github.com/go-sigma/sigma/bin/sigma-builder /usr/local/bin/sigma-builder

WORKDIR /code

USER 1000:1000
