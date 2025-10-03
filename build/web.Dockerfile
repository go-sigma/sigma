ARG NODE_VERSION=24-alpine3.22
ARG NGINX_VERSION=1.29.1-alpine3.22-otel

FROM --platform=$BUILDPLATFORM node:${NODE_VERSION} AS web-builder

ARG USE_MIRROR=false

RUN set -eux && \
  if [ "$USE_MIRROR" = true ]; then sed -i "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g" /etc/apk/repositories; fi && \
  apk add --no-cache make bash ncurses build-base

WORKDIR /web

COPY ./web .

RUN --mount=type=cache,target=/web/node_modules set -eux && \
  corepack enable && yarn install --immutable && yarn build

FROM nginxinc/nginx-unprivileged:${NGINX_VERSION}

COPY --from=web-builder --chmod=755 /web/dist /app
COPY --chmod=755 ./swag /app/swag
COPY --chmod=644 ./conf/default.conf.template /etc/nginx/templates/default.conf.template
