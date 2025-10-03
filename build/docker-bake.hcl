group "all" {
  targets = ["sigma-server", "sigma-builder"]
}

target "base" {
  dockerfile = "./build/Dockerfile.base"
  platforms = ["linux/amd64", "linux/arm64"]
  args = {
    USE_MIRROR = "true"
    WITH_TRIVY_DB = "true"
  }
  output = [{ type = "cacheonly" }]
}

target "web" {
  dockerfile = "./build/Dockerfile.web"
  args = {
    USE_MIRROR = "true"
  }
  output = [{ type = "cacheonly" }]
}

target "server" {
  dockerfile = "./build/Dockerfile.server"
  platforms = ["linux/amd64", "linux/arm64"]
  args = {
    USE_MIRROR = "true"
  }
  output = [{ type = "cacheonly" }]
}

target "sigma-server" {
  dockerfile = "./build/Dockerfile"
  contexts = {
    base = "target:base"
    web = "target:web"
    server = "target:server"
  }
  platforms = ["linux/amd64", "linux/arm64"]
  args = {
    USE_MIRROR = "true"
  }
  labels = {
    "org.opencontainers.image.title" = "sigma"
    "org.opencontainers.image.description" = "sigma is an OCI artifact storage and distribution system, which is designed to be a lightweight, easy-to-use, and easy-to-deploy, and can be used as a private registry or a public registry. sigma is a cloud-native, distributed, and highly available system, which can be deployed on any cloud platform or on-premises."
    "org.opencontainers.image.licenses" = "Apache-2.0"
    "org.opencontainers.image.url" = "https://github.com/go-sigma/sigma"
    "org.opencontainers.image.authors" = "Tosone <i@tosone.cn>"
    "org.opencontainers.image.source" = "https://github.com/go-sigma/sigma"
  }
  tags = ["tosone/sigma:latest"]
}

target "sigma-builder" {
  dockerfile = "./build/Dockerfile.builder"
  contexts = {
    base = "target:base"
    server = "target:server"
  }
  platforms = ["linux/amd64", "linux/arm64"]
  args = {
    USE_MIRROR = "true"
  }
  labels = {
    "org.opencontainers.image.title" = "sigma"
    "org.opencontainers.image.description" = "sigma is an OCI artifact storage and distribution system, which is designed to be a lightweight, easy-to-use, and easy-to-deploy, and can be used as a private registry or a public registry. sigma is a cloud-native, distributed, and highly available system, which can be deployed on any cloud platform or on-premises."
    "org.opencontainers.image.licenses" = "Apache-2.0"
    "org.opencontainers.image.url" = "https://github.com/go-sigma/sigma"
    "org.opencontainers.image.authors" = "Tosone <i@tosone.cn>"
    "org.opencontainers.image.source" = "https://github.com/go-sigma/sigma"
  }
  tags = ["tosone/sigma-builder:latest"]
}
