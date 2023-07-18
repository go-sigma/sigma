#!/bin/sh
# docker buildx create --use --config ./buildkit.toml

docker buildx build --platform linux/amd64,linux/arm64 --tag 10.3.201.221:3000/library/alpine:3.18.0 --file alpine.Dockerfile --push .
