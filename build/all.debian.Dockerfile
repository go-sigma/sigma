ARG GOLANG_VERSION=1.23.1-bookworm
ARG NODE_VERSION=20-alpine3.19
ARG ALPINE_VERSION=3.19
ARG DEBIAN_VERSION=bookworm-slim

FROM --platform=$BUILDPLATFORM node:${NODE_VERSION} AS web-builder

ARG USE_MIRROR=false

RUN set -eux && \
  if [ "$USE_MIRROR" = true ]; then sed -i "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g" /etc/apk/repositories; fi && \
  apk add --no-cache make bash ncurses build-base

COPY ./web /web

WORKDIR /web

RUN --mount=type=cache,target=/web/node_modules set -eux && corepack enable && yarn install --immutable && yarn build

FROM alpine:${ALPINE_VERSION} AS syft

ARG USE_MIRROR=false
ARG SYFT_VERSION=1.8.0
ARG TARGETOS TARGETARCH

RUN set -eux && \
  if [ "$USE_MIRROR" = true ]; then sed -i "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g" /etc/apk/repositories; fi && \
  apk add --no-cache wget && \
  wget -q -O syft_"${SYFT_VERSION}"_"${TARGETOS}"_"${TARGETARCH}".tar.gz https://github.com/anchore/syft/releases/download/v"${SYFT_VERSION}"/syft_"${SYFT_VERSION}"_"${TARGETOS}"_"${TARGETARCH}".tar.gz && \
  tar -xzf syft_"${SYFT_VERSION}"_"${TARGETOS}"_"${TARGETARCH}".tar.gz && \
  mv syft /usr/local/bin/syft && \
  rm syft_"${SYFT_VERSION}"_"${TARGETOS}"_"${TARGETARCH}".tar.gz

FROM alpine:${ALPINE_VERSION} AS trivy

ARG USE_MIRROR=false
ARG WITH_TRIVY_DB=false
ARG TRIVY_VERSION=0.55.1
ARG TARGETOS TARGETARCH

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
  if [ "$WITH_TRIVY_DB" = true ]; then trivy --cache-dir /opt/trivy/ image --download-java-db-only --no-progress; fi && \
  trivy --cache-dir /opt/trivy/ image --download-db-only --no-progress

FROM --platform=$BUILDPLATFORM golang:${GOLANG_VERSION} AS skopeo

ARG USE_MIRROR=false
ARG SKOPEO_VERSION=1.16.0
ARG TARGETOS TARGETARCH

RUN set -eux && \
  if [ "$USE_MIRROR" = true ]; then sed -i "s/deb.debian.org/mirrors.aliyun.com/g" /etc/apt/sources.list.d/debian.sources; fi && \
  apt-get update && \
	apt-get install -y --no-install-recommends \
		build-essential \
		git \
    make \
	&& \
	rm -rf /var/lib/apt/lists/* && \
  git clone --branch v"${SKOPEO_VERSION}" https://github.com/containers/skopeo /go/src/github.com/containers/skopeo && \
  cd /go/src/github.com/containers/skopeo && \
  DISABLE_CGO=1 make bin/skopeo."${TARGETOS}"."${TARGETARCH}" && \
  cp bin/skopeo."${TARGETOS}"."${TARGETARCH}" /tmp/skopeo

FROM --platform=$BUILDPLATFORM golang:${GOLANG_VERSION} AS builder

ARG USE_MIRROR=false
ARG BUILDARCH

RUN set -eux && \
  if [ "$USE_MIRROR" = true ]; then sed -i "s/deb.debian.org/mirrors.aliyun.com/g" /etc/apt/sources.list.d/debian.sources; fi && \
  apt-get update && \
	apt-get install -y --no-install-recommends \
		git \
	&& \
	rm -rf /var/lib/apt/lists/*

COPY . /go/src/github.com/go-sigma/sigma
COPY --from=web-builder /web/dist /go/src/github.com/go-sigma/sigma/web/dist

WORKDIR /go/src/github.com/go-sigma/sigma

ARG TARGETOS TARGETARCH

RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
  GOOS=$TARGETOS GOARCH=$TARGETARCH make build

FROM debian:${DEBIAN_VERSION}

ARG USE_MIRROR=false

RUN set -eux && \
  if [ "$USE_MIRROR" = true ]; then sed -i "s/deb.debian.org/mirrors.aliyun.com/g" /etc/apt/sources.list.d/debian.sources; fi && \
  apt-get update && \
	apt-get install -y --no-install-recommends \
    ca-certificates \
		curl \
		netbase \
    gnupg \
	  dirmngr \
	&& \
	rm -rf /var/lib/apt/lists/*

COPY --from=syft /usr/local/bin/syft /usr/local/bin/syft
COPY --from=trivy /usr/local/bin/trivy /usr/local/bin/trivy
COPY --from=trivy /opt/trivy/ /opt/trivy/
COPY --from=skopeo /tmp/skopeo /usr/local/bin/skopeo
COPY ./bin/*.tar /baseimages/
COPY ./conf/config.yaml /etc/sigma/config.yaml
COPY --from=builder /go/src/github.com/go-sigma/sigma/bin/sigma /usr/local/bin/sigma

VOLUME /var/lib/sigma
VOLUME /etc/sigma

CMD ["sigma", "server"]
