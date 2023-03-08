#!/bin/bash

DOCKER=${DOCKER:-docker}

"$DOCKER" run -it \
  --name ximager-redis \
  -p 6379:6379 -d --rm \
  --health-cmd "redis-cli -a ximager ping || exit 1" \
  --health-interval 10s \
  --health-timeout 5s \
  --health-retries 10 \
  redis:7.0-alpine --requirepass ximager
