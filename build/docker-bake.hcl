group "all" {
  targets = ["sigma-server", "sigma-builder"]
}

target "base" {
  dockerfile = "./build/Dockerfile.base"
}

target "web" {
  dockerfile = "./build/Dockerfile.web"
}

target "server" {
  dockerfile = "./build/Dockerfile.server"
}

target "sigma-server" {
  dockerfile = "./build/Dockerfile"
  contexts = {
    base = "target:base"
    web = "target:web"
    server = "target:server"
  }
}

target "sigma-builder" {
  dockerfile = "./build/Dockerfile.builder"
  contexts = {
    base = "target:base"
    server = "target:server"
  }
}
