<p align="center">
  <a href="https://github.com/go-sigma/sigma">
    <img alt="sigma" src="https://raw.githubusercontent.com/go-sigma/sigma/dev/assets/sigma.svg" width="220"/>
  </a>
</p>
<h1 align="center">sigma</h1>

![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/go-sigma/sigma/test.yml?style=for-the-badge) ![Codecov](https://img.shields.io/codecov/c/github/go-sigma/sigma?style=for-the-badge) ![GitHub repo size](https://img.shields.io/github/repo-size/go-sigma/sigma?style=for-the-badge)

Sigma 是一个极容易部署和维护的镜像仓库，并且自主完整实现了 [OCI Distribution Specification 1.1](https://github.com/opencontainers/distribution-spec/tree/v1.1.0) 的协议，除了支持 docker 客户端以外，还支持其他类型的各种客户端，例如 [oras](https://github.com/oras-project/oras)，[apptainer](https://github.com/apptainer/apptainer)，[helm](https://github.com/helm/helm), [nerdctl](https://github.com/containerd/nerdctl) 等。在部署层面上完全可以做到 all-in-one 的部署，启动单个容器即可将整体的镜像仓库的所有服务启动起来，在高可用层面也可以将各个模块分别单独部署。

## 快速开始

你可以用以下的一个简单的命令来运行起来 Sigma 镜像仓库:

```bash
docker run --name sigma -p 3000:3000 --rm tosone/sigma:nightly-alpine
```

默认的用户名密码是: sigma/Admin@123, 如果你想在启动的时候初始化其他的用户名密码, 请根据[这里](https://docs.sigma.tosone.cn/docs/configuration)的配置说明来修改配置文件。

## Demo Server

演示服务器部署在一个运行 Debian 12.1 Linux 发行版的 AWS EC2 实例上(2核4G内存, 40G磁盘)，使用的 Docker 版本是 25.0.3，演示服务器是参照[这里](https://github.com/go-sigma/demo-server)说明搭建的。

访问: <https://sigma.tosone.cn>, username/password: sigma/Admin@123

## Architecture

<img alt="sigma" src="https://raw.githubusercontent.com/go-sigma/sigma/main/assets/architecture.png" width="100%" />

## Features

- [x] 支持 [OCI Distribution Specification 1.1](https://github.com/opencontainers/distribution-spec/tree/v1.1.0) 的协议。
- [x] 支持 OCI Image v1 格式和 OCI Image Index v1 格式。
- [x] 支持 OCI artifacts，例如：helm charts，apptainer 等等。
- [x] 支持获取 OCI 制品的 sbom。
- [x] 支持 OCI 制品的安全扫描。
- [x] 支持镜像代理。
- [x] 支持 Namespace 级别的配额限制。
- [x] 支持镜像的自动垃圾回收。
- [x] 支持镜像签名。
- [x] 支持镜像构建，构建的镜像自动签名，构建的驱动支持 Docker，Podman，Kubernetes。
- [ ] 支持镜像复制。

## AI 客户端与自动化

Sigma 为 AI 客户端和外部自动化工具提供了两种集成方式：

- 内置 MCP server：Sigma 可以通过现有的 server 进程暴露镜像仓库管理工具。该能力默认关闭，可以通过 `mcp` 配置项启用。详细说明见 [MCP Server](docs/mcp.md)。
- 外部 REST API skill：`tools/skill/sigma-api-operator` 打包了可复用的 skill 和 `swagger.yaml`，方便 Claude Code、Copilot CLI 等 CLI agent 发现 Sigma API，并通过 HTTP 请求操作资源。

## Alternatives

- [Distribution](https://distribution.github.io/distribution/)
- [Harbor](https://goharbor.io/)
- [zot](https://zotregistry.io/)
