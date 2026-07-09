---
title: Helm
description: 推送 Helm Chart 到 sigma
---

# 推送 Helm Chart 到 sigma

### 生成示例 Helm Chart

``` bash
helm create demo
```

该命令会创建 `demo` 目录，并在其中生成一个示例 Helm Chart。

``` bash
helm package demo
```

在 `demo` 目录外执行该命令后，会得到 `demo-0.1.0.tgz` 文件。

### 推送 Helm Chart 到 sigma

注意：Helm v3.13.0 之前不支持通过 HTTP 推送 Helm Chart 到 OCI registry，必须使用 HTTPS。
可以参考 Helm v3.13.0 的 [release note](https://github.com/helm/helm/releases/tag/v3.13.0)。

Helm v3.13.0 之前：

``` bash
helm registry login --insecure -u sigma -p Admin@123 127.0.0.1:3000
helm push demo-0.1.0.tgz oci://127.0.0.1:3000/library/demo --insecure-skip-tls-verify
```

Helm v3.13.0 之后：

``` bash
helm registry login -u sigma -p Admin@123 127.0.0.1:3000
helm push demo-0.1.0.tgz oci://127.0.0.1:3000/library/demo --plain-http
```

### 从 sigma 拉取 Helm Chart

``` bash
# before v3.13.0
# helm pull oci://127.0.0.1:3000/library/demo --version 0.1.0 --insecure-skip-tls-verify
helm pull oci://127.0.0.1:3000/library/demo --version 0.1.0 --plain-http
```
