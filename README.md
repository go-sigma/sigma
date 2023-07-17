<p align="center">
  <a href="https://github.com/go-sigma/sigma">
    <img alt="sigma" src="https://media.githubusercontent.com/media/go-sigma/sigma/main/assets/sigma.svg" width="220"/>
  </a>
</p>
<h1 align="center">sigma</h1>

![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/go-sigma/sigma/test.yml?style=for-the-badge) ![Codecov](https://img.shields.io/codecov/c/github/go-sigma/sigma?style=for-the-badge) ![GitHub repo size](https://img.shields.io/github/repo-size/go-sigma/sigma?style=for-the-badge)

Yet another harbor OCI artifact manager. Harbor is a great product, but it's not easy to use, it is so complex. So I want to make a simple artifact manager and well tested. And it never depends on [distribution](https://github.com/distribution/distribution) like harbor.

## Introduction

<https://user-images.githubusercontent.com/5346506/229798487-798225b1-e2bf-40a2-b5ab-588003c02f7b.mp4>

## Architecture

Wait for me to complete draw the architecture.

## Quick Start

Now sigma is under very early development, so it's not easy to use. But you can try it.

``` bash
cd web && yarn && yarn build && cd .. && make build && ./scripts/run_all.sh
./bin/sigma server -c ./conf/config.yaml
```

## Features

- [x] Support docker registry v2 protocol.
- [x] Support OCI Image v1 Format and OCI Image Index v1 Format.
- [x] Support OCI artifact such as helm and so on.
- [x] Support OCI sbom.
- [x] Support Image security scan.
- [x] Support registry proxy.
- [x] Support Namespace quota.
- [x] Support Image automatic garbage collection.
- [x] Support Multi-tenancy.
- [ ] Support Image replication.
- [ ] Support Image build in docker, podman and kubernetes.
- [ ] Support Image sign.
- [ ] Support helm chart search and index.json.
