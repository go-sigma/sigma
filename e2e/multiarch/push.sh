#!/bin/bash

# docker buildx create --use --config ./buildkit.toml

docker buildx build --sbom=true --platform linux/amd64,linux/arm64 --tag 192.168.31.198:3000/library/alpine:3.18.0 --file alpine.Dockerfile --push .

cosign generate-key-pair

env COSIGN_PASSWORD= cosign sign --tlog-upload=false --allow-http-registry --key cosign.key --recursive 192.168.31.198:3000/library/alpine@sha256:d780ea42ba737c40b6a7fcadd8b8e6f5dd4365fb2166053f132aafe3b6ac9bcd
