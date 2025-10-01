variable "REGISTRY" {
  default = "registry:5000/blubber"
  description = "The registry/repo namespace under which to publish images."
}

variable "TAG" {
  default = "latest"
  description = "The tag to use for published images."
}

target "common" {
  context = "."
  dockerfile = ".pipeline/blubber.yaml"

  args = {
    # Declare the syntax as a build arg so bake on trusted runners works
    # without proxying through the dockerfile frontend. This should be kept in
    # sync with the syntax declared in .pipeline/blubber.yaml
    BUILDKIT_SYNTAX = "docker-registry.wikimedia.org/repos/releng/blubber/buildkit:v1.5.0"
    BUILDKIT_CONTEXT_KEEP_GIT_DIR = "1"
  }
}

group "default" {
  targets = [
    "buildkit",
  ]
}

group "test" {
  targets = [
    "lint",
    "unit",
    "ensure-docs",
  ]
}

target "make" {
  inherits = ["common"]
  name = make_target
  target = "make"
  args = {
    MAKE_TARGET = make_target
  }
  matrix = {
    make_target = ["lint", "unit", "ensure-docs"]
  }
}

target "acceptance" {
  inherits = ["common"]
  target = "acceptance"
  platforms = ["linux/amd64"]
  tags = ["${REGISTRY}/acceptance:${TAG}"]
  output = [ "type=registry" ]
}

target "buildkit" {
  inherits = ["common"]
  target = "buildkit"
  platforms = ["linux/amd64", "linux/arm64"]
  attest = ["type=provenance,mode=max"]
  tags = ["${REGISTRY}/buildkit:${TAG}"]
  output = [ "type=registry" ]
}

target "docs" {
  inherits = ["common"]
  target = "docs-for-publishing"
  output = [ "type=local,dest=dist/docs" ]
}
