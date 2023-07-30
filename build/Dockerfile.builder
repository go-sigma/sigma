ARG GOLANG_VERSION=1.20.6-alpine3.18

FROM golang:${GOLANG_VERSION} as builder

COPY . /go/src/github.com/go-sigma/sigma

WORKDIR /go/src/github.com/go-sigma/sigma

RUN set -eux && \
  # sed -i "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g" /etc/apk/repositories && \
  apk add --no-cache make bash ncurses build-base git git-lfs

RUN make build-builder-release

FROM moby/buildkit:v0.12.0-rootless

USER root
RUN set -eux && \
  # sed -i "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g" /etc/apk/repositories && \
  apk add --no-cache git-lfs && \
  mkdir -p /code/ && \
  chown -R 1000:1000 /opt/ && \
  chown -R 1000:1000 /code/

COPY --from=builder /go/src/github.com/go-sigma/sigma/bin/sigma-builder /usr/local/bin/sigma-builder

WORKDIR /code

USER 1000:1000
