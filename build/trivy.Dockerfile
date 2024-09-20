ARG ALPINE_VERSION=3.19

FROM alpine:${ALPINE_VERSION} AS trivy

ARG USE_MIRROR=false
ARG WITH_TRIVY_DB=false
ARG TRIVY_VERSION=0.55.2
ARG TARGETOS TARGETARCH

RUN set -eux && \
  if [ "$USE_MIRROR" = true ]; then sed -i "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g" /etc/apk/repositories; fi && \
  apk add --no-cache wget && \
  case "${TARGETARCH}" in \
		amd64) export TRIVYARCH='64bit' ;; \
		arm64) export TRIVYARCH='ARM64' ;; \
	esac; \
  export TRIVYOS=$(echo "${TARGETOS}" | awk '{print toupper(substr($0, 1, 1)) substr($0, 2)}') && \
  wget --progress=dot:giga -O trivy_"${TRIVY_VERSION}"_"${TRIVYOS}"-"${TRIVYARCH}".tar.gz https://github.com/aquasecurity/trivy/releases/download/v"${TRIVY_VERSION}"/trivy_"${TRIVY_VERSION}"_"${TRIVYOS}"-"${TRIVYARCH}".tar.gz && \
  tar -xzf trivy_"${TRIVY_VERSION}"_"${TRIVYOS}"-"${TRIVYARCH}".tar.gz && \
  mv trivy /usr/local/bin/trivy && \
  rm trivy_"${TRIVY_VERSION}"_"${TRIVYOS}"-"${TRIVYARCH}".tar.gz && \
  mkdir -p /opt/trivy/ && \
  trivy --cache-dir /opt/trivy/ image --download-java-db-only --no-progress && \
  trivy --cache-dir /opt/trivy/ image --download-db-only --no-progress

FROM scratch

COPY --from=trivy /opt/trivy/ /
