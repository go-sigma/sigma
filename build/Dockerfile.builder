ARG BUILDKIT_VERSION=v0.25.0-rootless

FROM server AS server

FROM base AS base

FROM moby/buildkit:${BUILDKIT_VERSION}

LABEL org.opencontainers.image.title="sigma" \
  org.opencontainers.image.description="sigma is an OCI artifact storage and distribution system, which is designed to be a lightweight, easy-to-use, and easy-to-deploy, and can be used as a private registry or a public registry. sigma is a cloud-native, distributed, and highly available system, which can be deployed on any cloud platform or on-premises." \
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

COPY --from=base /usr/local/bin/cosign /usr/local/bin/cosign
COPY --from=server /go/src/github.com/go-sigma/sigma/bin/sigma /usr/local/bin/sigma

WORKDIR /code

USER 1000:1000
