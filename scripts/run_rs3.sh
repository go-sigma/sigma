#!/bin/bash

DOCKER=${DOCKER:-docker}

"$DOCKER" run -p 9000:9000 \
  --name sigma-rs3 \
  -e RS3_HOST=0.0.0.0 \
  -e RS3_PORT=9000 \
  -e RS3_DATA_DIR=/var/lib/rs3/data \
  -e RS3_ACCESS_KEY=sigma \
  -e RS3_SECRET_KEY=sigma-sigma \
  -e RS3_ENDPOINT=http://127.0.0.1:9000 \
  --rm -d \
  --entrypoint "" \
  --health-cmd 'curl --fail "http://127.0.0.1:${RS3_PORT}/readyz" || exit 1' \
  --health-interval 30s \
  --health-timeout 3s \
  --health-start-period 5s \
  --health-retries 3 \
  ghcr.io/go-sigma/rs3:latest \
  sh -c 'mkdir -p /var/lib/rs3/data/sigma && exec rs3'
