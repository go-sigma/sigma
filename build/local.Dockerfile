ARG ALPINE_VERSION=3.17

FROM alpine:${ALPINE_VERSION}

RUN set -eux && apk add --no-cache bash

COPY ./build/entrypoint.sh /usr/local/bin/entrypoint.sh
COPY ./conf/ximager.yaml /etc/ximager/ximager.yaml
COPY ./bin/ximager /usr/local/bin/ximager

VOLUME ["/var/lib/registry"]

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
