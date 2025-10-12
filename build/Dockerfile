ARG NGINX_VERSION=1.29.1-alpine3.22-otel

FROM web AS web

FROM server AS server

FROM base AS base

FROM nginx:${NGINX_VERSION}

LABEL org.opencontainers.image.title="sigma" \
  org.opencontainers.image.description="sigma is an OCI artifact storage and distribution system, which is designed to be a lightweight, easy-to-use, and easy-to-deploy, and can be used as a private registry or a public registry. sigma is a cloud-native, distributed, and highly available system, which can be deployed on any cloud platform or on-premises." \
  org.opencontainers.image.url="https://hub.docker.com/r/sigmago/sigma" \
  org.opencontainers.image.source="https://github.com/go-sigma/sigma" \
  org.opencontainers.image.licenses="Apache-2.0" \
  org.opencontainers.image.version="v1.0.0"

ARG USE_MIRROR=false

RUN set -eux && \
  if [ "$USE_MIRROR" = true ]; then sed -i "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g" /etc/apk/repositories; fi && \
  apk add --no-cache curl tini

COPY --chmod=0755 ./swag /app/swag
COPY --chmod=0644 ./conf/default.conf.template /etc/nginx/templates/default.conf.template
COPY --chmod=0644 ./conf/config.yaml /etc/sigma/config.yaml

COPY --from=base --chmod=0755 /usr/local/bin/syft /usr/local/bin/syft
COPY --from=base --chmod=0755 /usr/local/bin/cosign /usr/local/bin/cosign
COPY --from=base --chmod=0755 /usr/local/bin/trivy /usr/local/bin/trivy
COPY --from=base --chmod=0755 /opt/trivy/ /opt/trivy/
COPY --from=server --chmod=0755 /go/src/github.com/go-sigma/sigma/bin/sigma /usr/local/bin/sigma
COPY --from=web --chmod=755 /web/dist /app

RUN adduser --disabled-password -h /home/sigma -s /bin/sh -u 1001 sigma && \
  mkdir -p /var/lib/sigma && \
  mkdir -p /code/ && \
  chown -R 1001:1001 /var/lib/sigma && \
  chown -R 1001:1001 /opt/trivy && \
  chown -R 1001:1001 /code/

WORKDIR /home/sigma

USER sigma

CMD ["sigma", "server"]
