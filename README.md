<p align="center">
  <a href="https://github.com/go-sigma/sigma">
    <img alt="sigma" src="https://raw.githubusercontent.com/go-sigma/sigma/main/assets/sigma.svg" width="220"/>
  </a>
</p>
<h1 align="center">sigma</h1>

![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/go-sigma/sigma/test.yml?style=for-the-badge) ![Codecov](https://img.shields.io/codecov/c/github/go-sigma/sigma?style=for-the-badge) ![GitHub repo size](https://img.shields.io/github/repo-size/go-sigma/sigma?style=for-the-badge)

Sigma is an image registry that is extremely easy to deploy and maintain. and it adheres to the interface standards defined by the [OCI Distribution Specification 1.1](https://github.com/opencontainers/distribution-spec/tree/v1.1.0-rc4). It can also support any other client programs that follow the interface definition of the OCI Distribution Specification, such as [oras](https://github.com/oras-project/oras), [apptainer](https://github.com/apptainer/apptainer), [helm](https://github.com/helm/helm), and [nerdctl](https://github.com/containerd/nerdctl).

## Demo Server

It runs on AWS ec2 (2C4G, Disk 40G), Linux distribution is Debian 12.1, Docker version 20.10.24. I will periodically reboot the container, and since the container doesn't have any disk mount, every reboot will clear all the data.

Visit: <https://sigma.tosone.cn>, username/password: sigma/Admin@123

Status check here: [https://stats.uptimerobot.com/PoB1PhvAoR/796311885](https://stats.uptimerobot.com/PoB1PhvAoR/796311885)

## Quick Start

Now you can use this command to run a simple server:

``` bash
docker run --name sigma -p 3000:3000 --rm ghcr.io/go-sigma/sigma:nightly-alpine
```

The default username and password is: sigma/Admin@123, if you want to modify the default password, please refer to the instructions provided [here](https://docs.sigma.tosone.cn/docs/configuration).

## Architecture

I was hoping you could wait for me to finish drawing the architecture.

## Compatibility

The sigma registry implements the OCI Distribution Spec version 1.1.0.

## Features

- [x] Support docker registry v2 protocol.
- [x] Support OCI Image v1 Format and OCI Image Index v1 Format.
- [x] Support OCI artifacts such as helm and so on.
- [x] Support OCI sbom.
- [x] Support Image security scan.
- [x] Support registry proxy.
- [x] Support Namespace quota.
- [x] Support Image automatic garbage collection.
- [x] Support Image sign.
- [x] Support Image build in docker, podman and kubernetes.
- [ ] Support Image replication.

## Release Schedule

- There will be a major release in the middle and end of each year.
- A minor release is released at the beginning of each month.
- There will probably be a bugfix release in the middle of each week.

## Alternatives

- [Distribution](https://distribution.github.io/distribution/)
- [Harbor](https://goharbor.io/)
- [zot](https://zotregistry.io/)
