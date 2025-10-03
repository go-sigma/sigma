ARG BUILDKIT_VERSION=v0.25.0-rootless

FROM server AS server

FROM base AS base

FROM moby/buildkit:${BUILDKIT_VERSION}

ARG USE_MIRROR=false

USER root
RUN set -eux && \
  if [ "$USE_MIRROR" = true ]; then sed -i "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g" /etc/apk/repositories; fi && \
  apk add --no-cache git-lfs && \
  mkdir -p /code/ && \
  chown -R 1000:1000 /opt/ && \
  chown -R 1000:1000 /code/

COPY --from=base /usr/local/bin/cosign /usr/local/bin/cosign
COPY --from=server /go/src/github.com/go-sigma/sigma/bin/sigma /usr/local/bin/sigma

WORKDIR /code

USER 1000:1000
