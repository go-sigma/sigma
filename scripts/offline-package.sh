#!/bin/bash

DOCKER=${DOCKER:-docker}

"$DOCKER" pull quay.io/minio/minio:RELEASE.2024-01-31T20-20-33Z
"$DOCKER" pull ghcr.io/go-sigma/sigma-builder:nightly
"$DOCKER" pull ghcr.io/go-sigma/sigma:nightly-alpine
"$DOCKER" pull redis:7.0-alpine
"$DOCKER" pull mysql:8.0

"$DOCKER" save quay.io/minio/minio:RELEASE.2024-01-31T20-20-33Z ghcr.io/go-sigma/sigma:nightly-alpine ghcr.io/go-sigma/sigma-builder:nightly redis:7.0-alpine mysql:8.0 | gzip > sigma.tar.gz

if [ -d package ]; then
  rm -rf package
fi

mkdir -p package/sigma/conf

cp -r scripts/samples ./package/sigma
mv ./package/sigma/samples/start.sh ./package/sigma/
mv ./package/sigma/samples/restart.sh ./package/sigma/
cp docker-compose.yml ./package/sigma
cp conf/config-compose.yaml ./package/sigma/conf/config.yaml
cp conf/sigma.test.io.crt ./package/sigma/conf/
cp conf/sigma.test.io.key ./package/sigma/conf/
mv ./sigma.tar.gz ./package/sigma

if [ -f sigma-offline.tar.gz ]; then
  rm sigma-offline.tar.gz
fi

tar zcf sigma-offline.tar.gz -C ./package sigma

rm -rf ./package
