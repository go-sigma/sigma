ARG GOLANG_VERSION=1.19.10-alpine3.18
ARG NODE_VERSION=18-alpine3.18
ARG ALPINE_VERSION=3.18

FROM node:${NODE_VERSION} as web-builder

RUN set -eux && apk add --no-cache make bash ncurses build-base

COPY ./web /web

WORKDIR /web

RUN set -eux && yarn install --frozen-lockfile && yarn build

FROM alpine:${ALPINE_VERSION} as syft

ARG SYFT_VERSION=0.84.1

ARG TARGETARCH

RUN set -eux && \
  sed -i "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g" /etc/apk/repositories && \
  apk add --no-cache wget && \
  wget -q -O syft_"${SYFT_VERSION}"_linux_"${TARGETARCH}".tar.gz https://github.com/anchore/syft/releases/download/v"${SYFT_VERSION}"/syft_"${SYFT_VERSION}"_linux_"${TARGETARCH}".tar.gz && \
  tar -xzf syft_"${SYFT_VERSION}"_linux_"${TARGETARCH}".tar.gz && \
  mv syft /usr/local/bin/syft && \
  rm syft_"${SYFT_VERSION}"_linux_"${TARGETARCH}".tar.gz

FROM alpine:${ALPINE_VERSION} as trivy

ARG TRIVY_VERSION=0.43.1
ARG ORAS_VERSION=1.0.0

ARG TARGETARCH

RUN set -eux && \
  sed -i "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g" /etc/apk/repositories && \
  apk add --no-cache wget && \
  case "${TARGETARCH}" in \
		amd64) export TRIVYARCH='64bit' ;; \
		arm64) export TRIVYARCH='ARM64' ;; \
	esac; \
  wget -q -O trivy_"${TRIVY_VERSION}"_Linux-"${TRIVYARCH}".tar.gz https://github.com/aquasecurity/trivy/releases/download/v"${TRIVY_VERSION}"/trivy_"${TRIVY_VERSION}"_Linux-"${TRIVYARCH}".tar.gz && \
  tar -xzf trivy_"${TRIVY_VERSION}"_Linux-"${TRIVYARCH}".tar.gz && \
  mv trivy /usr/local/bin/trivy && \
  rm trivy_"${TRIVY_VERSION}"_Linux-"${TRIVYARCH}".tar.gz && \
  wget -q -O oras_"${ORAS_VERSION}"_linux_"${TARGETARCH}".tar.gz https://github.com/oras-project/oras/releases/download/v"${ORAS_VERSION}"/oras_"${ORAS_VERSION}"_linux_"${TARGETARCH}".tar.gz && \
  tar -xzf oras_"${ORAS_VERSION}"_linux_"${TARGETARCH}".tar.gz && \
  mv oras /usr/local/bin/oras && \
  rm oras_"${ORAS_VERSION}"_linux_"${TARGETARCH}".tar.gz && \
  oras pull ghcr.io/aquasecurity/trivy-db:2 && \
  mkdir -p /opt/trivy/ && \
  mv ./db.tar.gz /opt/trivy/db.tar.gz && \
  cd /opt/trivy && \
  tar -xzf db.tar.gz && \
  rm db.tar.gz

FROM golang:${GOLANG_VERSION} as builder

RUN set -eux && \
  apk add --no-cache make bash ncurses build-base git git-lfs

COPY . /go/src/github.com/ximager/ximager
COPY --from=web-builder /web/dist /go/src/github.com/ximager/ximager/web/dist

WORKDIR /go/src/github.com/ximager/ximager

RUN make build-release

FROM alpine:${ALPINE_VERSION}

COPY --from=syft /usr/local/bin/syft /usr/local/bin/syft
COPY --from=trivy /usr/local/bin/trivy /usr/local/bin/trivy
COPY --from=trivy /opt/trivy/trivy.db /opt/trivy/db/trivy.db
COPY ./conf/ximager.yaml /etc/ximager/ximager.yaml
COPY --from=builder /go/src/github.com/ximager/ximager/bin/ximager /usr/local/bin/ximager

ENTRYPOINT ["/usr/local/bin/ximager"]
