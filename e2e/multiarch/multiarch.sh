#!/bin/sh
# docker buildx create --use --config ./buildkit.toml

docker buildx build --platform linux/amd64,linux/arm64 --tag 192.168.31.200:3000/library/alpine:3.18.2 --file alpine.Dockerfile --push .
