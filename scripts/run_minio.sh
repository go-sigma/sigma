#!/bin/bash

DOCKER=${DOCKER:-docker}

"$DOCKER" run -p 9000:9000 -p 9001:9001 \
  --name sigma-minio \
  -e MINIO_ROOT_USER=sigma \
  -e MINIO_ROOT_PASSWORD=sigma-sigma \
  -e MINIO_REGION_NAME=cn-north-1 \
  --rm -d \
  --entrypoint "" \
  --health-cmd "curl -f http://localhost:9000/minio/health/live || exit 1" \
  --health-interval 10s \
  --health-timeout 5s \
  --health-retries 10 \
  quay.io/minio/minio:RELEASE.2023-07-07T07-13-57Z \
  sh -c 'mkdir -p /data/sigma && minio server /data --console-address ":9001"'
