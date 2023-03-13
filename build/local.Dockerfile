ARG ALPINE_VERSION=3.17

FROM alpine:${ALPINE_VERSION}

RUN set -eux && apk add --no-cache bash

COPY ./conf/ximager.yaml /etc/ximager/ximager.yaml
COPY ./bin/ximager /usr/local/bin/ximager

ENTRYPOINT ["/usr/local/bin/ximager"]
