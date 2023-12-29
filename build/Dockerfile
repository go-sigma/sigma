ARG GOLANG_VERSION=1.21.5-alpine3.18
ARG NODE_VERSION=18-alpine3.18
ARG ALPINE_VERSION=3.18

FROM --platform=$BUILDPLATFORM node:${NODE_VERSION} as web-builder

RUN set -eux && \
  # sed -i "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g" /etc/apk/repositories && \
  apk add --no-cache make bash ncurses build-base

COPY ./web /web

WORKDIR /web

RUN --mount=type=cache,target=/web/node_modules set -eux && corepack enable && yarn install --immutable && yarn build

FROM alpine:${ALPINE_VERSION} as syft

ARG SYFT_VERSION=0.99.0
ARG TARGETOS TARGETARCH

RUN set -eux && \
  # sed -i "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g" /etc/apk/repositories && \
  apk add --no-cache wget && \
  wget -q -O syft_"${SYFT_VERSION}"_"${TARGETOS}"_"${TARGETARCH}".tar.gz https://github.com/anchore/syft/releases/download/v"${SYFT_VERSION}"/syft_"${SYFT_VERSION}"_"${TARGETOS}"_"${TARGETARCH}".tar.gz && \
  tar -xzf syft_"${SYFT_VERSION}"_"${TARGETOS}"_"${TARGETARCH}".tar.gz && \
  mv syft /usr/local/bin/syft && \
  rm syft_"${SYFT_VERSION}"_"${TARGETOS}"_"${TARGETARCH}".tar.gz

FROM alpine:${ALPINE_VERSION} as trivy

ARG TRIVY_VERSION=0.48.1
ARG ORAS_VERSION=1.0.0
ARG TARGETOS TARGETARCH

SHELL ["/bin/ash", "-eo", "pipefail", "-c"]

RUN set -eux && \
  # sed -i "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g" /etc/apk/repositories && \
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
  wget -q -O oras_"${ORAS_VERSION}"_"${TARGETOS}"_"${TARGETARCH}".tar.gz https://github.com/oras-project/oras/releases/download/v"${ORAS_VERSION}"/oras_"${ORAS_VERSION}"_"${TARGETOS}"_"${TARGETARCH}".tar.gz && \
  tar -xzf oras_"${ORAS_VERSION}"_"${TARGETOS}"_"${TARGETARCH}".tar.gz && \
  mv oras /usr/local/bin/oras && \
  rm oras_"${ORAS_VERSION}"_"${TARGETOS}"_"${TARGETARCH}".tar.gz && \
  oras pull ghcr.io/aquasecurity/trivy-db:2 && \
  mkdir -p /opt/trivy/ && \
  mv ./db.tar.gz /opt/trivy/db.tar.gz && \
  cd /opt/trivy && \
  tar -xzf db.tar.gz && \
  rm db.tar.gz

FROM --platform=$BUILDPLATFORM golang:${GOLANG_VERSION} as builder

RUN set -eux && \
  # sed -i "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g" /etc/apk/repositories && \
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

RUN set -eux && \
  # sed -i "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g" /etc/apk/repositories && \
  apk add --no-cache curl

COPY --from=syft /usr/local/bin/syft /usr/local/bin/syft
COPY --from=trivy /usr/local/bin/trivy /usr/local/bin/trivy
COPY --from=trivy /opt/trivy/trivy.db /opt/trivy/db/trivy.db
COPY --from=trivy /opt/trivy/metadata.json /opt/trivy/db/metadata.json
COPY ./conf/config.yaml /etc/sigma/config.yaml
COPY --from=builder /go/src/github.com/go-sigma/sigma/bin/sigma /usr/local/bin/sigma

VOLUME /var/lib/sigma
VOLUME /etc/sigma

CMD ["sigma", "server"]
