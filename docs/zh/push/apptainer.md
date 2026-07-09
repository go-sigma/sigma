---
title: Apptainer
description: 推送 Apptainer SIF 到 sigma
---

# 推送 Apptainer SIF 到 sigma

### 运行 Apptainer 容器

Apptainer 主要运行在 Linux 系统上，也可以通过容器方式启动：

``` bash
docker run -it --rm ghcr.io/apptainer/apptainer:1.3.0-rc.1 bash
```

### 拉取镜像并保存为 SIF 文件

``` bash
apptainer build alpine.sif docker://alpine
```

### 登录并推送 SIF 文件

``` bash
apptainer registry login -u sigma docker://192.168.31.112:3000
apptainer push alpine.sif oras://192.168.31.112:3000/library/alpine:tosone
```
