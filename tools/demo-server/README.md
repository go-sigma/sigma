<p align="center">
  <a href="https://github.com/go-sigma/sigma">
    <img alt="sigma" src="https://raw.githubusercontent.com/go-sigma/sigma/main/assets/sigma.svg" width="220"/>
  </a>
</p>
<h1 align="center">sigma</h1>

Sigma is an image registry that is extremely easy to deploy and maintain, and it adheres to the interface standards defined by the [OCI Distribution Specification 1.1](https://github.com/opencontainers/distribution-spec/tree/v1.1.0), it can also support any other client programs that follow the interface definition of the OCI Distribution Specification, such as [oras](https://github.com/oras-project/oras), [apptainer](https://github.com/apptainer/apptainer), [helm](https://github.com/helm/helm), and [nerdctl](https://github.com/containerd/nerdctl).

## Demo Server

It is deployed on an AWS EC2 instance (2C4G, 40G disk) running Debian 12.1 as the Linux distribution. The Docker version used is 25.0.3. The demo server was set up with the files in this directory.

Visit: <https://sigma.tosone.cn>, username/password: sigma/Admin@123

Status check here: [https://grafana.sigma.tosone.cn/public-dashboards/ebf12be7e7f44b59bfe4096f0f51ab88](https://grafana.sigma.tosone.cn/public-dashboards/ebf12be7e7f44b59bfe4096f0f51ab88)

### Deploy Your Own Demo Server

You can change the `.env` file and change these lines in `compose.yaml` with your own domain:

```yaml
- traefik.http.routers.traefik.rule=Host(`traefik.sigma${HOST_DOMAIN}`)
# ...
- traefik.http.routers.sigma.rule=Host(`sigma${HOST_DOMAIN}`)
# ...
- traefik.http.routers.grafana.rule=Host(`grafana.sigma${HOST_DOMAIN}`)
```

## Commands

```bash
# Install docker
make install

# Start or restart all of the docker services
make start

# Stop all of the docker services
make stop

# Show this help
make help
```
