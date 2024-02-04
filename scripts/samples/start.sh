#!/bin/sh

# First of all you must install docker, there are two ways:
#  1. apt-get install docker.io
#  2. sh -c "$(curl -fsSL https://get.docker.com)"

DOCKER=${DOCKER:-docker}
SIGMA_VOLUME=${SIGMA_VOLUME:-$PWD}

"$DOCKER" run --privileged --name sigma-dind -d \
  -v "$SIGMA_VOLUME":/app \
  -p 443:3000 \
  -p 3306:3306 \
  -p 6379:6379 \
  -p 9000:9000 \
  -p 9001:9001 \
  docker:25-dind

if [ "$SIGMA_VOLUME" != "$PWD" ]; then
  cp -rf sigma.tar.gz "$SIGMA_VOLUME"
  cp -rf docker-compose.yml "$SIGMA_VOLUME"
  cp -rf samples "$SIGMA_VOLUME"
fi

"$DOCKER" exec -it sigma-dind \
  timeout 180 sh -c 'while ! docker info >/dev/null 2>&1; do echo "Waiting for docker daemon ready..."; sleep 3; done && \
  cd /app && docker load -i sigma.tar.gz && \
  docker compose --project-name sigma -f /app/docker-compose.yml up --detach'
