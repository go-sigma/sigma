#!/bin/bash

DOCKER=${DOCKER:-docker}

"$DOCKER" run -it \
  --name sigma-postgres \
  -e POSTGRES_PASSWORD=sigma \
  -e POSTGRES_USER=sigma \
  -e POSTGRES_DB=sigma \
  -p 5432:5432 -d --rm \
  --health-cmd "pg_isready -U sigma -d sigma || exit 1" \
  --health-interval 10s \
  --health-timeout 5s \
  --health-retries 10 \
  postgres:15-alpine
