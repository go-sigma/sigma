#!/bin/bash

DOCKER=${DOCKER:-docker}

"$DOCKER" pull quay.io/minio/minio:RELEASE.2023-08-04T17-40-21Z
"$DOCKER" pull ghcr.io/go-sigma/sigma-builder:nightly
"$DOCKER" pull ghcr.io/go-sigma/sigma:nightly-alpine
"$DOCKER" pull redis:7.0-alpine
"$DOCKER" pull mysql:8.0
"$DOCKER" pull postgres:15-alpine

"$DOCKER" save quay.io/minio/minio:RELEASE.2023-08-04T17-40-21Z ghcr.io/go-sigma/sigma:nightly-alpine ghcr.io/go-sigma/sigma-builder:nightly redis:7.0-alpine mysql:8.0 postgres:15-alpine | gzip > sigma.tar.gz

if [ -d package ]; then
  rm -rf package
fi

mkdir -p package/sigma/conf

cp -r scripts/samples ./package/sigma
mv ./package/sigma/samples/start.sh ./package/sigma/
cp docker-compose.yml ./package/sigma
cp conf/config-compose.yaml ./package/sigma/conf/config.yaml
cp conf/sigma.test.io.crt ./package/sigma/config/
cp conf/sigma.test.io.key ./package/sigma/config/
mv ./sigma.tar.gz ./package/sigma

tar zcvf sigma-offline.tar.gz -C ./package sigma