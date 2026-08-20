# =============================================================
# Global build-arg declarations (available to every stage)
# =============================================================
ARG ALPINE_VERSION=3.24
ARG GOLANG_VERSION=1.27-alpine3.24
ARG BUN_VERSION=1.3.14-alpine
ARG BUILDKIT_VERSION=v0.30.0-rootless
ARG SWAGGER_UI_VERSION=5.32.8

# =============================================================
# base stages  (originally Dockerfile.base)
# =============================================================
FROM alpine:${ALPINE_VERSION} AS syft

ARG SYFT_VERSION=1.42.3
ARG TARGETOS
ARG TARGETARCH
ARG USE_MIRROR=false

RUN set -eux && \
  if [ "$USE_MIRROR" = true ]; then sed -i "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g" /etc/apk/repositories; fi && \
  apk add --no-cache wget upx && \
  wget --progress=dot:giga -O syft_"${SYFT_VERSION}"_"${TARGETOS}"_"${TARGETARCH}".tar.gz https://github.com/anchore/syft/releases/download/v"${SYFT_VERSION}"/syft_"${SYFT_VERSION}"_"${TARGETOS}"_"${TARGETARCH}".tar.gz && \
  tar -xzf syft_"${SYFT_VERSION}"_"${TARGETOS}"_"${TARGETARCH}".tar.gz && \
  upx -9 syft && \
  mv syft /usr/local/bin/syft && \
  rm syft_"${SYFT_VERSION}"_"${TARGETOS}"_"${TARGETARCH}".tar.gz

FROM alpine:${ALPINE_VERSION} AS grype

ARG GRYPE_VERSION=0.110.0
ARG TARGETOS
ARG TARGETARCH
ARG USE_MIRROR=false

RUN set -eux && \
  if [ "$USE_MIRROR" = true ]; then sed -i "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g" /etc/apk/repositories; fi && \
  apk add --no-cache wget jq upx && \
  wget --progress=dot:giga -O grype_"${GRYPE_VERSION}"_"${TARGETOS}"_"${TARGETARCH}".tar.gz https://github.com/anchore/grype/releases/download/v"${GRYPE_VERSION}"/grype_"${GRYPE_VERSION}"_"${TARGETOS}"_"${TARGETARCH}".tar.gz && \
  tar -xzf grype_"${GRYPE_VERSION}"_"${TARGETOS}"_"${TARGETARCH}".tar.gz && \
  upx -9 grype && \
  mv grype /usr/local/bin/grype && \
  rm grype_"${GRYPE_VERSION}"_"${TARGETOS}"_"${TARGETARCH}".tar.gz && \
  mkdir -p /opt/grype/ && \
  GRYPE_DB_PATH="$(wget -qO- https://grype.anchore.io/databases/v6/latest.json | jq -r '.path')" && \
  wget --progress=dot:giga -O /opt/grype/"${GRYPE_DB_PATH##*/}" "https://grype.anchore.io/databases/v6/${GRYPE_DB_PATH}"

FROM alpine:${ALPINE_VERSION} AS cosign

ARG USE_MIRROR=false
ARG COSIGN_VERSION=v2.6.1
ARG TARGETOS
ARG TARGETARCH

RUN set -eux && \
  if [ "$USE_MIRROR" = true ]; then sed -i "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g" /etc/apk/repositories; fi && \
  apk add --no-cache wget upx && \
  wget --progress=dot:giga -O /tmp/cosign https://github.com/sigstore/cosign/releases/download/"${COSIGN_VERSION}"/cosign-"${TARGETOS}"-"${TARGETARCH}" && \
  chmod +x /tmp/cosign && \
  upx -9 /tmp/cosign

# =============================================================
# web stage  (originally Dockerfile.web)
# =============================================================
FROM oven/bun:${BUN_VERSION} AS web

ARG USE_MIRROR=false

RUN set -eux && \
  if [ "$USE_MIRROR" = true ]; then sed -i "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g" /etc/apk/repositories; fi && \
  apk add --no-cache make bash ncurses build-base

WORKDIR /web

COPY ./web .

RUN --mount=type=cache,target=/web/node_modules \
  --mount=type=cache,target=/web/.yarn \
  --mount=type=cache,target=/root/.yarn \
  bun install --frozen-lockfile && \
  bun run build

# =============================================================
# server stage  (originally Dockerfile.server)
# =============================================================
FROM golang:${GOLANG_VERSION} AS server

ARG USE_MIRROR=false

RUN set -eux && \
  if [ "$USE_MIRROR" = true ]; then sed -i "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g" /etc/apk/repositories; fi && \
  apk add --no-cache make bash ncurses git git-lfs openssl build-base linux-headers zig upx

WORKDIR /go/src/github.com/go-sigma/sigma

COPY . .

RUN --mount=type=cache,target=/root/.cache/go-build \
  --mount=type=cache,target=/go/pkg/mod \
  go mod download && make build && \
  upx -9 /go/src/github.com/go-sigma/sigma/bin/sigma

# =============================================================
# swagger stage
# =============================================================
FROM golang:${GOLANG_VERSION} AS swagger

ARG USE_MIRROR=false
ARG SWAGGER_UI_VERSION

RUN set -eux && \
  if [ "$USE_MIRROR" = true ]; then sed -i "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g" /etc/apk/repositories; fi && \
  apk add --no-cache make bash ncurses git wget

WORKDIR /go/src/github.com/go-sigma/sigma

COPY . .

RUN --mount=type=cache,target=/root/.cache/go-build \
  --mount=type=cache,target=/go/pkg/mod \
  set -eux && \
  mkdir -p /app/swag && \
  wget --progress=dot:giga -O /tmp/swagger-ui.tar.gz https://github.com/swagger-api/swagger-ui/archive/refs/tags/v"${SWAGGER_UI_VERSION}".tar.gz && \
  tar -xzf /tmp/swagger-ui.tar.gz -C /tmp && \
  cp -R /tmp/swagger-ui-"${SWAGGER_UI_VERSION}"/dist/. /app/swag/ && \
  sed -i 's#url: "https://petstore.swagger.io/v2/swagger.json"#url: "swagger.json"#' /app/swag/swagger-initializer.js && \
  go mod download && \
  go tool swag init --output /app/swag --outputTypes json

# =============================================================
# sigma-server  (originally Dockerfile, bake target: sigma-server)
# Build: docker build --target sigma-server ...
# =============================================================
FROM alpine:${ALPINE_VERSION} AS sigma-server

LABEL org.opencontainers.image.title="sigma" \
  org.opencontainers.image.description="Sigma is a lightweight, easy-to-deploy OCI artifact registry that implements the OCI Distribution Specification 1.1 and can serve as a private or public registry on any cloud or on-premises." \
  org.opencontainers.image.url="https://hub.docker.com/r/sigmago/sigma" \
  org.opencontainers.image.source="https://github.com/go-sigma/sigma" \
  org.opencontainers.image.licenses="Apache-2.0" \
  org.opencontainers.image.version="v1.0.0"

ARG USE_MIRROR=false

RUN set -eux && \
  if [ "$USE_MIRROR" = true ]; then sed -i "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g" /etc/apk/repositories; fi && \
  apk add --no-cache ca-certificates mailcap nginx tini

COPY --chmod=0755 ./build/entrypoint /usr/local/bin/sigma-server-entrypoint
COPY --chmod=0644 ./conf/default.conf.template /etc/nginx/nginx.conf
COPY --chmod=0644 ./conf/config.yaml /etc/sigma/config.yaml

COPY --from=syft --chmod=0755 /usr/local/bin/syft /usr/local/bin/syft
COPY --from=cosign --chmod=0755 /tmp/cosign /usr/local/bin/cosign
COPY --from=grype --chmod=0755 /usr/local/bin/grype /usr/local/bin/grype
COPY --from=grype --chown=1001:1001 --chmod=0755 /opt/grype/ /opt/grype/
COPY --from=server --chmod=0755 /go/src/github.com/go-sigma/sigma/bin/sigma /usr/local/bin/sigma
COPY --from=web --chmod=755 /web/dist /app
COPY --from=swagger --chmod=0755 /app/swag /app/swag

RUN adduser --disabled-password -h /home/sigma -s /bin/sh -u 1001 sigma && \
  install -d -o 1001 -g 1001 /var/lib/sigma /code/

WORKDIR /home/sigma

USER sigma

CMD ["tini", "--", "sigma-server-entrypoint"]

# =============================================================
# sigma-builder  (originally Dockerfile.builder, bake target: sigma-builder)
# Build: docker build --target sigma-builder ...
# =============================================================
FROM moby/buildkit:${BUILDKIT_VERSION} AS sigma-builder

LABEL org.opencontainers.image.title="sigma" \
  org.opencontainers.image.description="Sigma is a lightweight, easy-to-deploy OCI artifact registry that implements the OCI Distribution Specification 1.1 and can serve as a private or public registry on any cloud or on-premises." \
  org.opencontainers.image.url="https://hub.docker.com/r/sigmago/sigma" \
  org.opencontainers.image.source="https://github.com/go-sigma/sigma" \
  org.opencontainers.image.licenses="Apache-2.0" \
  org.opencontainers.image.version="v1.0.0"

ARG USE_MIRROR=false

USER root

RUN set -eux && \
  if [ "$USE_MIRROR" = true ]; then sed -i "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g" /etc/apk/repositories; fi && \
  apk add --no-cache curl git git-lfs && \
  mkdir -p /code/ && \
  chown -R 1000:1000 /opt/ && \
  chown -R 1000:1000 /code/

COPY --from=cosign /tmp/cosign /usr/local/bin/cosign
COPY --from=server /go/src/github.com/go-sigma/sigma/bin/sigma /usr/local/bin/sigma

WORKDIR /code

USER 1000:1000
