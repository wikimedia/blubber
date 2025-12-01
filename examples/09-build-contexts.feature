Feature: Build contexts
  In addition to copying files from `local` or another variant's filesystem,
  Blubber supports client provided [build contexts][contexts] as sources for
  `copies` operations, builder `requirements`, or even as the `base` of a
  variant.

  Build contexts become particularly useful when Blubber configuration is used
  in conjunction with tools like [docker buildx build][buildx-build-contexts]
  or [docker buildx bake][buildx-bake-contexts]. The following examples will
  provide local directories as the sources for additional build contexts, but
  remote sources such as Git repos or OCI images are also possible. See the
  [upstream documentation][contexts] for details.

  [contexts]: https://docs.docker.com/build/concepts/context/
  [buildx-build-contexts]: https://docs.docker.com/reference/cli/docker/buildx/build/#build-context
  [buildx-bake-contexts]: https://docs.docker.com/build/bake/contexts/

  @set1
  Scenario: Copying files from a build context
    Given "examples/hello-world-wrapper" as a working directory
    And this "blubber.yaml"
      """
      version: v4
      variants:
        build:
          base: debian:trixie
          apt:
            packages:
              - gcc
              - libc6-dev
          builder:
            requirements:
              - from: hello-src  # copy from a named build context below
                source: main.c
            command: "gcc -o hello main.c"
        hello:
          base: debian:trixie
          runs:
            environment:
              HELLO_WORLD_BIN: /usr/local/bin/hello
          copies:
            - from: local
              source: hello.sh
            - from: build
              source: hello
              destination: /usr/local/bin/hello
          entrypoint: [./hello.sh]
      """
    When you build and run the "hello" variant with the following build contexts
      | hello-src | ./examples/hello-world-c |
    Then the image will have the following files in the default working directory
      | hello.sh |
    And the entrypoint will have run successfully

  @set2
  Scenario: Providing a base image using a build context
    Given "examples/hello-world" as a working directory
    And this "blubber.yaml"
      """
      version: v4
      variants:
        hello:
          base: debian-ish
          copies: [local]
          entrypoint: [./hello.sh]
      """
    When you build and run the "hello" variant with the following build contexts
      | debian-ish | docker-image://docker-registry.wikimedia.org/trixie:latest |
    Then the entrypoint will have run successfully
