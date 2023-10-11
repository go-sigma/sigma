#!/bin/sh

set -e

docker login sigma.tosone.cn -u sigma -p sigma

push_image() {
  docker pull "$2"
  docker tag "$2" sigma.tosone.cn/"$1"/"$2"
  docker push sigma.tosone.cn/"$1"/"$2"
}

push_image library redis:7-bookworm
push_image library redis:7-alpine
push_image library redis:6-bookworm
push_image library redis:6-alpine
push_image library nginx:1.25-alpine
push_image library nginx:1.25-bookworm
push_image library debian:bookworm-slim
push_image library debian:buster-slim
push_image library debian:bullseye-slim
push_image library ubuntu:22.04
push_image library ubuntu:23.10
push_image library ubuntu:23.04
push_image library centos:7
push_image library centos:8
push_image library alpine:3.18
push_image library alpine:3.17
push_image library alpine:3.16
push_image library alpine:3.15

curl https://github.com/grafana/k6/releases/download/v0.46.0/k6-v0.46.0-linux-arm64.tar.gz -L | tar xvz --strip-components 1

./k6 run samples.js

push_image test-all alpine:3.18
push_image test-all alpine:3.17
push_image test-all redis:6-alpine

push_image test-repo-cnt-limit redis:6-alpine

push_image test-tag-count-limit redis:6-alpine
push_image test-tag-count-limit redis:7-alpine

push_image test-size-limit centos:8
