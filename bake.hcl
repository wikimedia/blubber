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

target "lint" {
  inherits = ["common"]
  target = "lint"
}

target "unit" {
  inherits = ["common"]
  target = "unit"
}

target "ensure-docs" {
  inherits = ["common"]
  target = "ensure-docs"
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
  tags = ["${REGISTRY}/buildkit:${TAG}"]
  output = [ "type=registry" ]
}

target "docs" {
  inherits = ["common"]
  target = "docs-for-publishing"
  output = [ "type=local,dest=dist/docs" ]
}
