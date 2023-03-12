ARG GOLANG_VERSION=1.19.7-alpine3.17
ARG NODE_VERSION=18-alpine3.17
ARG ALPINE_VERSION=3.17

FROM node:${NODE_VERSION} as web-builder

RUN set -eux && apk add --no-cache make bash ncurses build-base

COPY ./web /web

WORKDIR /web

RUN set -eux && yarn install --frozen-lockfile && yarn build

FROM golang:${GOLANG_VERSION} as builder

RUN set -eux && \
  apk add --no-cache make bash ncurses build-base

COPY . /go/src/github.com/ximager/ximager
COPY --from=web-builder /web/dist /go/src/github.com/ximager/ximager/web/dist

WORKDIR /go/src/github.com/ximager/ximager

RUN make build

FROM alpine:${ALPINE_VERSION}

RUN set -eux && apk add --no-cache bash

COPY ./conf/ximager.yaml /etc/ximager/ximager.yaml
COPY --from=builder /go/src/github.com/ximager/ximager/bin/ximager /usr/local/bin/ximager

ENTRYPOINT ["/usr/local/bin/ximager"]
