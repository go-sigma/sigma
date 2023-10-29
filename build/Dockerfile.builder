ARG GOLANG_VERSION=1.21.3-alpine3.18
ARG BUILDKIT_VERSION=v0.12.3-rootless

FROM golang:${GOLANG_VERSION} as cosign

WORKDIR /go/src/github.com/sigstore

RUN set -eux && \
  apk add --no-cache make bash ncurses build-base git git-lfs && \
  git clone https://github.com/go-sigma/cosign.git && \
  cd cosign && \
  make

FROM golang:${GOLANG_VERSION} as builder

COPY . /go/src/github.com/go-sigma/sigma

WORKDIR /go/src/github.com/go-sigma/sigma

RUN set -eux && \
  apk add --no-cache make bash ncurses build-base git git-lfs && \
  make build-builder

FROM moby/buildkit:${BUILDKIT_VERSION}

USER root
RUN set -eux && \
  apk add --no-cache git-lfs && \
  mkdir -p /code/ && \
  chown -R 1000:1000 /opt/ && \
  chown -R 1000:1000 /code/

COPY --from=cosign /go/src/github.com/sigstore/cosign/cosign /usr/local/bin/cosign
COPY --from=builder /go/src/github.com/go-sigma/sigma/bin/sigma-builder /usr/local/bin/sigma-builder

WORKDIR /code

USER 1000:1000
