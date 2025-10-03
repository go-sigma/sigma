ARG NGINX_VERSION=1.29.1-alpine3.22-otel

FROM web AS web

FROM server AS server

FROM base AS base

FROM nginx:${NGINX_VERSION}

ARG USE_MIRROR=false

RUN set -eux && \
  if [ "$USE_MIRROR" = true ]; then sed -i "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g" /etc/apk/repositories; fi && \
  apk add --no-cache curl tini

COPY --from=base --chmod=0755 /usr/local/bin/syft /usr/local/bin/syft
COPY --from=base --chmod=0755 /usr/local/bin/cosign /usr/local/bin/cosign
COPY --from=base --chmod=0755 /usr/local/bin/trivy /usr/local/bin/trivy
COPY --from=base /opt/trivy/ /opt/trivy/
COPY ./conf/config.yaml /etc/sigma/config.yaml
COPY --from=server /go/src/github.com/go-sigma/sigma/bin/sigma /usr/local/bin/sigma
COPY --from=web --chmod=755 /web/dist /app
COPY --chmod=755 ./swag /app/swag
COPY --chmod=644 ./conf/default.conf.template /etc/nginx/templates/default.conf.template

VOLUME /var/lib/sigma
VOLUME /etc/sigma

RUN adduser --disabled-password -h /home/sigma -s /bin/sh -u 1001 sigma && \
  mkdir -p /var/lib/sigma && \
  mkdir -p /code/ && \
  chown -R 1001:1001 /var/lib/sigma && \
  chown -R 1001:1001 /opt/trivy && \
  chown -R 1001:1001 /code/

WORKDIR /home/sigma

USER sigma

CMD ["sigma", "server"]
