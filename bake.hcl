variable "REGISTRY" {
  default = "registry:5000/blubber"
  description = "The registry/repo namespace under which to publish images."
}

variable "TAG" {
  default = "latest"
  description = "The tag to use for published images."
}

// It would be better to simply inherit the bake build context, but there is
// currently a bug that prevents preservation of the .git directory when the
// bake context is a remote git repo URL.
//
// See https://github.com/docker/buildx/pull/3338
//
variable "BUILD_CONTEXT" {
  default = "."
  description = "The main build context."
}

target "common" {
  context = BUILD_CONTEXT
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
