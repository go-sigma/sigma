<p align="center">
  <a href="https://github.com/go-sigma/sigma">
    <img alt="sigma" src="https://raw.githubusercontent.com/go-sigma/sigma/main/assets/sigma.svg" width="220"/>
  </a>
</p>
<h1 align="center">sigma</h1>

![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/go-sigma/sigma/test.yml?style=for-the-badge) ![Codecov](https://img.shields.io/codecov/c/github/go-sigma/sigma?style=for-the-badge) ![GitHub repo size](https://img.shields.io/github/repo-size/go-sigma/sigma?style=for-the-badge)

Yet another OCI artifact manager. [Harbor](https://goharbor.io/) is a great product, but it's not easy to use, it is so complex. So I want to make a simple artifact manager, that never depends on [distribution](https://github.com/distribution/distribution) like [harbor](https://goharbor.io/).

## Demo Server

It runs on AWS ec2 (2C2G, Disk 20G), Linux distribution is Debian 12.1, Docker version 24.0.6.

``` sh
# Install Docker from get.docker.com
sh ./scripts/samples/init.sh

# If your docker running with rootless mode,
# make sure add net.ipv4.ip_unprivileged_port_start=0 to /etc/sysctl.conf and run sudo sysctl --system.
docker network create sigma
docker run --name sigma -v /home/admin/config:/etc/sigma \
  -v /var/run/docker.sock:/var/run/docker.sock -p 443:3000 \
  -d --net sigma ghcr.io/go-sigma/sigma:nightly-alpine

# Add sample data
sh ./scripts/samples/samples.sh
```

Visit: <https://sigma.tosone.cn>, username/password: sigma/Admin@123

I will periodically reboot the container, and since the container doesn't have any disk mount, every reboot will clear all the data.

## Architecture

I was hoping you could wait for me to finish drawing the architecture.

## Quick Start

Now sigma is under very early development, so it's not easy to use. But you can try it.

``` bash
cd web && yarn && yarn build && cd .. && make build && ./scripts/run_all.sh
./bin/sigma server -c ./conf/config.yaml
```

## Features

- [x] Support docker registry v2 protocol.
- [x] Support OCI Image v1 Format and OCI Image Index v1 Format.
- [x] Support OCI artifacts such as helm and so on.
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

## Release Schedule

- There will be a major release in the middle and end of each year.
- A minor release is released at the beginning of each month.
- There will probably be a bugfix release in the middle of each week.
