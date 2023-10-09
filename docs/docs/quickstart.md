---
sidebar_position: 1
---

# Quick Start

Let's discover **sigma in less than 5 minutes**.

## Getting Started

Run sigma in Docker. If you want to use the builder, Docker version should be used latest.

``` sh
docker run --name sigma -v /home/admin/config:/etc/sigma \
  -v /var/run/docker.sock:/var/run/docker.sock -p 443:3000 \
  -d ghcr.io/go-sigma/sigma:nightly-alpine
```

Push image to sigma.

``` sh
docker pull redis:7
docker tag redis:7 127.0.0.1:3000/library/redis:7
docker push 127.0.0.1:3000/library/redis:7
```
