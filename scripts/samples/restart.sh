#!/bin/sh

"$DOCKER" exec -it sigma-dind \
  cd /app && docker load -i sigma.tar.gz && \
  docker compose --project-name sigma down && \
  docker compose --project-name sigma -f /app/docker-compose.yml up --detach
