---
layout: home
markdownStyles: false
title: sigma

hero:
  name: Self-hosted
  text: OCI registry
  tagline: '<span class="hero-prompt">&gt;_</span><span class="hero-rotator" aria-hidden="true"><span class="hero-rotator__sizer">Scan and sign your artifacts</span><span class="hero-rotator__item">Deploy in one command</span><span class="hero-rotator__item">Scan and sign your artifacts</span><span class="hero-rotator__item">Pull through any upstream</span><span class="hero-rotator__item">Replicate across clusters</span></span><span class="hero-sr">Deploy in one command. Scan and sign your artifacts. Pull through any upstream. Replicate across clusters.</span>'
  image:
    src: /img/logo.svg
    alt: sigma
  actions:
    - theme: brand
      text: Get started
      link: /quickstart
    - theme: alt
      text: Live demo
      link: https://sigma.tosone.cn

features:
  - icon: OCI
    title: Store artifacts
    details: OCI images, Helm charts, SBOMs, and any custom artifact type behind one registry.
  - icon: BUILD
    title: Build images
    details: Build on Docker, Podman, or Kubernetes and push straight into sigma.
  - icon: SCAN
    title: Scan for CVEs
    details: Trivy-backed scanning on push, with a severity report for every artifact.
  - icon: SIGN
    title: Sign and verify
    details: Cosign key and keyless signing, verified from the web console.
  - icon: PROXY
    title: Proxy upstreams
    details: Cache Docker Hub or any registry with per-namespace pull-through rules.
  - icon: SYNC
    title: Distribute
    details: Replicate across sigma instances, with Dragonfly P2P for large clusters.
---

<HomeQuickStart />
<HomeArchitecture />
<HomeCta />
