ARG ALPINE_VERSION=3.19
ARG GOLANG_VERSION=1.23.1-alpine3.19

FROM --platform=$BUILDPLATFORM golang:${GOLANG_VERSION} AS fetcher

ARG USE_MIRROR=false
ARG SKOPEO_VERSION=1.16.0
ARG TRIVY_VERSION=0.55.2
ARG SYFT_VERSION=1.8.0
ARG TARGETOS TARGETARCH

RUN set -eux && \
  if [ "$USE_MIRROR" = true ]; then sed -i "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g" /etc/apk/repositories; fi && \
  apk add --no-cache make git wget curl file && \
  git clone --branch v"${SKOPEO_VERSION}" https://github.com/containers/skopeo /go/src/github.com/containers/skopeo && \
  cd /go/src/github.com/containers/skopeo && \
  DISABLE_CGO=1 make bin/skopeo."${TARGETOS}"."${TARGETARCH}" && \
  cp bin/skopeo."${TARGETOS}"."${TARGETARCH}" /tmp/skopeo && \
  case "${TARGETARCH}" in \
		amd64) export TRIVYARCH='64bit' ;; \
		arm64) export TRIVYARCH='ARM64' ;; \
	esac; \
  export TRIVYOS=$(echo "${TARGETOS}" | awk '{print toupper(substr($0, 1, 1)) substr($0, 2)}') && \
  wget --progress=dot:giga -O trivy_"${TRIVY_VERSION}"_"${TRIVYOS}"-"${TRIVYARCH}".tar.gz https://github.com/aquasecurity/trivy/releases/download/v"${TRIVY_VERSION}"/trivy_"${TRIVY_VERSION}"_"${TRIVYOS}"-"${TRIVYARCH}".tar.gz && \
  tar -xzf trivy_"${TRIVY_VERSION}"_"${TRIVYOS}"-"${TRIVYARCH}".tar.gz && \
  mv trivy /usr/local/bin/trivy && \
  rm trivy_"${TRIVY_VERSION}"_"${TRIVYOS}"-"${TRIVYARCH}".tar.gz && \
  wget --progress=dot:giga -O syft_"${SYFT_VERSION}"_"${TARGETOS}"_"${TARGETARCH}".tar.gz https://github.com/anchore/syft/releases/download/v"${SYFT_VERSION}"/syft_"${SYFT_VERSION}"_"${TARGETOS}"_"${TARGETARCH}".tar.gz && \
  tar -xzf syft_"${SYFT_VERSION}"_"${TARGETOS}"_"${TARGETARCH}".tar.gz && \
  mv syft /usr/local/bin/syft && \
  rm syft_"${SYFT_VERSION}"_"${TARGETOS}"_"${TARGETARCH}".tar.gz && \
  mkdir -p /opt/trivy/ && \
  trivy --cache-dir /opt/trivy/ image --download-db-only --no-progress

FROM alpine:${ALPINE_VERSION}

ARG USE_MIRROR=false

RUN set -eux && \
  if [ "$USE_MIRROR" = true ]; then sed -i "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g" /etc/apk/repositories; fi && \
  apk add --no-cache curl

RUN adduser --disabled-password -h /home/sigma -s /bin/sh -u 1001 sigma

USER sigma

WORKDIR /home/sigma

COPY --from=fetcher /tmp/skopeo /usr/local/bin/skopeo
COPY --from=fetcher /usr/local/bin/syft /usr/local/bin/syft
COPY --from=fetcher /usr/local/bin/trivy /usr/local/bin/trivy
COPY --from=fetcher /opt/trivy/ /opt/trivy/
COPY ./bin/*.tar /baseimages/
COPY ./conf/config.yaml /etc/sigma/config.yaml
COPY ./bin/sigma /usr/local/bin/sigma

VOLUME /var/lib/sigma
VOLUME /etc/sigma

CMD ["sigma", "server"]
