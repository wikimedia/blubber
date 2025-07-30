@build-arguments
Feature: Build arguments
  Not all information can or should be hardcoded in the configuration file.
  Build arguments let you pass information to builder processes at build time
  which can reduce redundancy in your configuration and make the build more
  flexible. Note that build arguments become environment variables in the
  resulting image configuration and should not be used to store sensitive
  values.

  @set1
  Scenario: Using an argument to generalize a builder/variant
    Given "examples/web-app" as a working directory
    And this "blubber.yaml"
      """
      version: v4
      variants:
        make:
          base: debian:stable
          arguments:
            MAKE_TARGET: foo
          apt:
            packages: [make]
          builders:
            - custom:
                requirements: [Makefile]
                command: "make ${MAKE_TARGET}"
      """
    When you build the "make" variant with the following build arguments
      | MAKE_TARGET | bar |
    Then the image will have the following files in the default working directory
      | bar |
