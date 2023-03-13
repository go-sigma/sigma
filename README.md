# ximager

![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/ximager/ximager/test.yml?style=for-the-badge) ![Codecov](https://img.shields.io/codecov/c/github/ximager/ximager?style=for-the-badge) ![GitHub repo size](https://img.shields.io/github/repo-size/ximager/ximager?style=for-the-badge)

Yet another harbor OCI artifact manager. Harbor is a great product, but it's not easy to use, it is so complex. So I want to make a simple artifact manager and well tested. And it never depends on [distribution](https://github.com/distribution/distribution) like harbor.

## Architecture

Wait for me to complete draw the architecture.

## Quick Start

Now ximager is under very early development, so it's not easy to use. But you can try it.

``` bash
cd web && yarn && yarn build && cd .. && make build
./bin/ximager -c ./conf/ximager.yaml
```

## Features

- [x] Support docker registry v2 protocol.
- [x] Support OCI Image v1 Format and OCI Image Index v1 Format.
- [ ] Support registry proxy.
- [ ] Support OCI artifact such as helm and sbom and so on.
- [ ] Support Namespace quota.
- [ ] Support Image security scan.
- [ ] Support Image replication.
- [ ] Support Image build in docker, podman and kubernetes.
- [ ] Support Image sign.
- [ ] Support Image automatic garbage collection.
- [ ] Support Multi-tenancy.
- [ ] Support helm chart search and index.json.
