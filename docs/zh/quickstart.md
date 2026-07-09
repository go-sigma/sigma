---
sidebar_position: 1
---

# 快速开始

## 什么是 sigma？

Sigma 是一个轻量、自托管的一体化 OCI 制品存储与分发系统。它支持 OCI 制品管理、垃圾回收、namespace quota、多架构制品和 OCI 镜像构建。

Sigma 的定位类似 [Harbor](https://goharbor.io/)，但内部分发服务由项目自身实现。服务端、分发进程、worker 和 builder 可以按部署场景组合运行，也可以在简单场景中通过一个命令启动。

## 启动服务

用不到 3 分钟启动一个 sigma 实例。

可以直接用 Docker 运行 sigma。如果需要使用 builder，建议使用较新的 Docker 版本。
builder 会把构建出的镜像推送回 sigma，因此需要让 sigma 和 builder 处在同一个 Docker network 中，并在 `config.yaml` 中配置 `daemon.builder.docker.network`。

```bash
# 创建网络
docker network create sigma

# 运行 sigma
docker run --name sigma -v /home/admin/config:/etc/sigma \
  -v /home/admin/storage:/var/lib/sigma \
  -v /var/run/docker.sock:/var/run/docker.sock -p 443:3000 \
  -d --net sigma ghcr.io/go-sigma/sigma:nightly-alpine
```

启动完成后，你会得到一个类似 Docker Hub 或 Harbor 的私有 registry 服务。

## 推送镜像

```bash
docker pull redis:7
docker tag redis:7 127.0.0.1:3000/library/redis:7
docker push 127.0.0.1:3000/library/redis:7
```
