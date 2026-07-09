---
title: Docker
description: Push image to sigma
---

# Push image to sigma

### Push single image to sigma

``` bash
docker pull redis:7
docker tag redis:7 127.0.0.1:3000/library/redis:7
docker push 127.0.0.1:3000/library/redis:7
```

### Push multiarch image to sigma

Create 'buildkit.toml' with content:

``` toml
[registry."10.3.201.221:3000"] # replace it with your host ip
  http = true
```

Create buildx instance:

``` bash
docker buildx create --use --config ./buildkit.toml
```

Create Dockerfile with content:

``` dockerfile
FROM alpine:3.18
```

Push multiarch image to sigma:

``` bash
docker buildx build --platform linux/amd64,linux/arm64 --tag 10.3.201.221:3000/library/alpine:3.18.0 --file alpine.Dockerfile --push .
```
