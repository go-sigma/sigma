---
sidebar_position: 1
---

# Quick Start

Let's discover **sigma in less than 3 minutes**.

## Getting Started

Run sigma in Docker. If you want to use the builder, Docker version should be used the latest version.
The builder will push image to server, so we should link the together, you should create network for them.
And set the network name `daemon.builder.docker.network` in `config.yaml`.

``` bash
# create network
docker network create sigma
# run the sigma in docker
docker run --name sigma -v /home/admin/config:/etc/sigma \
  -v /home/admin/storage:/var/lib/sigma \
  -v /var/run/docker.sock:/var/run/docker.sock -p 443:3000 \
  -d --net sigma ghcr.io/go-sigma/sigma:nightly-alpine
```

That's enough, now you got a service like docker hub or harbor.

Push image to sigma.

``` bash
docker pull redis:7
docker tag redis:7 127.0.0.1:3000/library/redis:7
docker push 127.0.0.1:3000/library/redis:7
```
