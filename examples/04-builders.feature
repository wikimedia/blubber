@builders
Feature: Builders
  Use cases will often involve more than simply copying files to the image.
  Many will require a build step: some process that creates additional files
  in the image's local filesystem needed at runtime.

  @set3
  Scenario: Compiling an application from source
    Given "examples/hello-world-c" as a working directory
    And this "blubber.yaml"
      """
      version: v4
      variants:
        build:
          base: debian:bullseye
          apt:
            packages:
              - gcc
              - libc6-dev
          builder:
            requirements: [main.c]
            command: "gcc -o hello main.c"
          entrypoint: [./hello]
      """
    When you build the "build" variant
    Then the image will have the following files in the default working directory
      | hello |

  @set4
  Scenario: Compiling an application using multiple builders
    Given "examples/hello-world-go" as a working directory
    And this "blubber.yaml"
      """
      version: v4
      variants:
        build:
          base: golang:1.18
          builders:
            - custom:
                requirements: [go.mod, go.sum]
                command: "go mod download"
            - custom:
                requirements: [main.go]
                command: "go build ."
          entrypoint: [./hello-world-go]
      """
    When you build the "build" variant
    Then the image will have the following files in the default working directory
      | hello-world-go |

  @set1
  Scenario: Defining inline builder scripts
    Given "examples/hello-world-go" as a working directory
    And this "blubber.yaml"
      """
      version: v4
      variants:
        build:
          base: golang:1.18
          runs:
            environment:
              CGO_ENABLED: "0"
          builders:
            - custom:
                requirements: [go.mod, go.sum]
                command: "go mod download"
            - custom:
                requirements: [main.go]
                script: |
                  #!/bin/bash
                  if ! [[ "$CGO_ENABLED" == "0" ]]; then
                    echo "you must set CGO_ENABLED=0 for this build"
                    exit 1
                  fi
                  go build $(go list)
        application:
          copies:
            - from: build
              source: ./hello-world-go
              destination: /hello
          entrypoint: [/hello]
      """
    When you build the "application" variant
    Then the image will have the following files in "/"
      | hello |

  @set2
  Scenario: Excluding files from requirements by glob pattern
    Given "examples/hello-world-c" as a working directory
    And this "blubber.yaml"
      """
      version: v4
      variants:
        build:
          base: debian:bullseye
          apt:
            packages:
              - gcc
              - libc6-dev
          builder:
            requirements:
              - from: local
                exclude: ["*.md"]
            command: "gcc -o hello main.c"
          entrypoint: [./hello]
      """
    When you build the "build" variant
    Then the image will have the following files in the default working directory
      | hello |
    Then the image will not have the following files in the default working directory
      | README.md |

  @set1
  Scenario: Using caches to speed up builds
    Given "examples/hello-world-go" as a working directory
    And this "blubber.yaml"
      """
      version: v4
      variants:
        build:
          base: golang:1.18
          runs:
            environment:
              GOCACHE: /var/cache/go
          builders:
            - custom:
                requirements: [ go.mod, go.sum ]
                command: "go mod download"
                caches: [ "${GOPATH}/pkg" ]
            - custom:
                requirements: [main.go]
                command: "go build ."
                caches: [ "${GOCACHE}" ]
          entrypoint: [./hello-world-go]
      """
    When you build the "build" variant
    Then the image will have the following files in the default working directory
      | hello-world-go |

  @set4
  Scenario: Using a mount to read files from another variant
    Given "examples/web-app" as a working directory
    And this "blubber.yaml"
      """
      version: v4
      variants:
        assets:
          base: node:22-bookworm
          lives:
            in: /src
          builders:
            - node:
                requirements: [package.json, package-lock.json]
            - custom:
                requirements:
                  - webpack.config.js
                  - ./src/
                command: "npm run build"
        build:
          base: golang:1.23
          runs:
            environment:
              CGO_ENABLED: "0"
              GOCACHE: /var/cache/go
          builders:
            - custom:
                requirements: [go.mod, main.go]
                command: "go build"
                caches: [ "${GOCACHE}" ]
                mounts:
                  - from: assets
                    source: /src/dist
                    destination: ./dist
        webserver:
          copies:
            - from: build
              source: ./web-app
              destination: /usr/bin/web-app
          entrypoint: [/usr/bin/web-app]
      """
    When you build the "webserver" variant
    Then the image will have the following files in "/usr/bin"
      | web-app |
    And the image entrypoint will be "/usr/bin/web-app"
