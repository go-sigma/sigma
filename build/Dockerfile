ARG GOLANG_VERSION=1.22.0-alpine3.19
ARG NODE_VERSION=20-alpine3.19
ARG ALPINE_VERSION=3.19

FROM --platform=$BUILDPLATFORM node:${NODE_VERSION} as web-builder

ARG USE_MIRROR=false

RUN set -eux && \
  if [ "$USE_MIRROR" = true ]; then sed -i "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g" /etc/apk/repositories; fi && \
  apk add --no-cache make bash ncurses build-base

COPY ./web /web

WORKDIR /web

RUN --mount=type=cache,target=/web/node_modules set -eux && corepack enable && yarn install --immutable && yarn build

FROM alpine:${ALPINE_VERSION} as syft

ARG SYFT_VERSION=1.0.1
ARG TARGETOS TARGETARCH
ARG USE_MIRROR=false

RUN set -eux && \
  if [ "$USE_MIRROR" = true ]; then sed -i "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g" /etc/apk/repositories; fi && \
  apk add --no-cache wget && \
  wget -q -O syft_"${SYFT_VERSION}"_"${TARGETOS}"_"${TARGETARCH}".tar.gz https://github.com/anchore/syft/releases/download/v"${SYFT_VERSION}"/syft_"${SYFT_VERSION}"_"${TARGETOS}"_"${TARGETARCH}".tar.gz && \
  tar -xzf syft_"${SYFT_VERSION}"_"${TARGETOS}"_"${TARGETARCH}".tar.gz && \
  mv syft /usr/local/bin/syft && \
  rm syft_"${SYFT_VERSION}"_"${TARGETOS}"_"${TARGETARCH}".tar.gz

FROM alpine:${ALPINE_VERSION} as trivy

ARG USE_MIRROR=false
ARG TRIVY_VERSION=0.49.0
ARG TARGETOS TARGETARCH

SHELL ["/bin/ash", "-eo", "pipefail", "-c"]

RUN set -eux && \
  if [ "$USE_MIRROR" = true ]; then sed -i "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g" /etc/apk/repositories; fi && \
  apk add --no-cache wget && \
  case "${TARGETARCH}" in \
		amd64) export TRIVYARCH='64bit' ;; \
		arm64) export TRIVYARCH='ARM64' ;; \
	esac; \
  export TRIVYOS=$(echo "${TARGETOS}" | awk '{print toupper(substr($0, 1, 1)) substr($0, 2)}') && \
  wget -q -O trivy_"${TRIVY_VERSION}"_"${TRIVYOS}"-"${TRIVYARCH}".tar.gz https://github.com/aquasecurity/trivy/releases/download/v"${TRIVY_VERSION}"/trivy_"${TRIVY_VERSION}"_"${TRIVYOS}"-"${TRIVYARCH}".tar.gz && \
  tar -xzf trivy_"${TRIVY_VERSION}"_"${TRIVYOS}"-"${TRIVYARCH}".tar.gz && \
  mv trivy /usr/local/bin/trivy && \
  rm trivy_"${TRIVY_VERSION}"_"${TRIVYOS}"-"${TRIVYARCH}".tar.gz && \
  mkdir -p /opt/trivy/ && \
  trivy --cache-dir /opt/trivy/ image --download-java-db-only --no-progress && \
  trivy --cache-dir /opt/trivy/ image --download-db-only --no-progress

FROM --platform=$BUILDPLATFORM golang:${GOLANG_VERSION} as builder

ARG USE_MIRROR=false

RUN set -eux && \
  if [ "$USE_MIRROR" = true ]; then sed -i "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g" /etc/apk/repositories; fi && \
  apk add --no-cache make bash ncurses build-base git openssl && \
  apk add --no-cache zig --repository=https://mirrors.aliyun.com/alpine/edge/testing

COPY . /go/src/github.com/go-sigma/sigma
COPY --from=web-builder /web/dist /go/src/github.com/go-sigma/sigma/web/dist

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
  GOOS=$TARGETOS GOARCH=$TARGETARCH CC="${CC}" CXX="${CXX}" make build

FROM alpine:${ALPINE_VERSION}

ARG USE_MIRROR=false

RUN set -eux && \
  if [ "$USE_MIRROR" = true ]; then sed -i "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g" /etc/apk/repositories; fi && \
  apk add --no-cache curl

COPY --from=syft /usr/local/bin/syft /usr/local/bin/syft
COPY --from=trivy /usr/local/bin/trivy /usr/local/bin/trivy
COPY --from=trivy /opt/trivy/ /opt/trivy/
COPY ./conf/config.yaml /etc/sigma/config.yaml
COPY --from=builder /go/src/github.com/go-sigma/sigma/bin/sigma /usr/local/bin/sigma

VOLUME /var/lib/sigma
VOLUME /etc/sigma

CMD ["sigma", "server"]
