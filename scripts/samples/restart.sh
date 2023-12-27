#!/bin/sh

DOCKER=${DOCKER:-docker}
SIGMA_VOLUME=${SIGMA_VOLUME:-$PWD}

if [ "$SIGMA_VOLUME" != "$PWD" ]; then
  cp -rf sigma.tar.gz "$SIGMA_VOLUME"
  cp -rf docker-compose.yml "$SIGMA_VOLUME"
  cp -rf samples "$SIGMA_VOLUME"
fi

"$DOCKER" exec -it sigma-dind \
  sh -c 'cd /app && docker load -i sigma.tar.gz && \
  docker compose --project-name sigma down && \
  docker compose --project-name sigma -f /app/docker-compose.yml up --detach'
