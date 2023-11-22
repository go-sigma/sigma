#!/bin/sh

# First of all you must install docker, there are two ways:
#  1. apt-get install docker.io
#  2. sh -c "$(curl -fsSL https://get.docker.com)"

DOCKER=${DOCKER:-docker}

"$DOCKER" run --privileged --name sigma-dind -d \
  -v "$PWD":/app \
  -p 443:3000 \
  -p 3306:3306 \
  -p 6379:6379 \
  -p 9000:9000 \
  -p 9001:9001 \
  docker:24-dind

"$DOCKER" exec -it sigma-dind \
  timeout 180 sh -c 'while ! docker info >/dev/null 2>&1; do echo "Waiting for docker daemon ready..."; sleep 3; done && \
  cd /app && docker load -i sigma.tar.gz && \
  docker compose --project-name sigma -f /app/docker-compose.yml up --detach'
