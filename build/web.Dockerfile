ARG NODE_VERSION=20-alpine3.19
ARG NGINX_VERSION=1.27.1-alpine

FROM --platform=$BUILDPLATFORM node:${NODE_VERSION} AS web-builder

ARG USE_MIRROR=false

RUN set -eux && \
    if [ "$USE_MIRROR" = true ]; then sed -i "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g" /etc/apk/repositories; fi && \
    apk add --no-cache make bash ncurses build-base

COPY ./web /web

WORKDIR /web

RUN --mount=type=cache,target=/web/node_modules set -eux && corepack enable && yarn install --immutable && yarn build

FROM nginx:1.27.1-alpine

COPY --from=web-builder /web/dist /usr/share/nginx/html
