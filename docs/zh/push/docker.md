---
title: Docker
description: 推送 Docker 镜像到 sigma
---

# 推送 Docker 镜像到 sigma

### 推送单架构镜像

``` bash
docker pull redis:7
docker tag redis:7 127.0.0.1:3000/library/redis:7
docker push 127.0.0.1:3000/library/redis:7
```

### 推送多架构镜像

创建 `buildkit.toml`：

``` toml
[registry."10.3.201.221:3000"] # replace it with your host ip
  http = true
```

创建 buildx 实例：

``` bash
docker buildx create --use --config ./buildkit.toml
```

创建 Dockerfile：

``` dockerfile
FROM alpine:3.18
```

推送多架构镜像：

``` bash
docker buildx build --platform linux/amd64,linux/arm64 --tag 10.3.201.221:3000/library/alpine:3.18.0 --file alpine.Dockerfile --push .
```
