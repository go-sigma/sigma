---
layout: home
markdownStyles: false
title: sigma

hero:
  name: 自托管
  text: OCI 制品仓库
  tagline: '<span class="hero-prompt">&gt;_</span><span class="hero-rotator" aria-hidden="true"><span class="hero-rotator__sizer">代理任意上游仓库</span><span class="hero-rotator__item">一条命令完成部署</span><span class="hero-rotator__item">自动扫描与签名</span><span class="hero-rotator__item">代理任意上游仓库</span><span class="hero-rotator__item">跨集群分发制品</span></span><span class="hero-sr">一条命令完成部署。自动扫描与签名。代理任意上游仓库。跨集群分发制品。</span>'
  image:
    src: /img/logo.svg
    alt: sigma
  actions:
    - theme: brand
      text: 快速开始
      link: /zh/quickstart
    - theme: alt
      text: 体验演示
      link: https://sigma.tosone.cn

features:
  - icon: OCI
    title: 存储制品
    details: OCI 镜像、Helm Chart、SBOM，以及任意自定义制品类型，统一存放在一个仓库。
  - icon: BUILD
    title: 构建镜像
    details: 支持 Docker、Podman 和 Kubernetes 构建，构建完成后直接推送到 sigma。
  - icon: SCAN
    title: 漏洞扫描
    details: 推送时通过 Trivy 扫描，并为每个制品生成漏洞等级报告。
  - icon: SIGN
    title: 签名校验
    details: 支持 Cosign 密钥与免密钥签名，可直接在控制台完成校验。
  - icon: PROXY
    title: 代理上游
    details: 为 Docker Hub 或任意仓库提供按 namespace 配置的代理缓存。
  - icon: SYNC
    title: 分发同步
    details: 在多个 sigma 实例之间复制制品，大规模集群可结合 Dragonfly P2P。
---

<HomeQuickStart />
<HomeArchitecture />
<HomeCta />
