Feature: Python builder
  Blubber supports a specialized Python builder for easy and consistent
  dependency installation and setup for Python projects.

  Background:
    Given "examples/hello-world-python-uv" as a working directory

  @set2
  Scenario: Installing Python application dependencies
    Given this "blubber.yaml"
      """
      version: v4
      variants:
        hello:
          base: python:3.10-bullseye
          builders:
            - python:
                version: python3
                requirements: [requirements.txt]
          copies: [local]
          entrypoint: [python3, hello.py]
      """
    When you build and run the "hello" variant
    Then the entrypoint will have run successfully

  @set3
  Scenario: Installing Python application dependencies via Uv
    Given this "blubber.yaml"
      """
      version: v4
      variants:
        hello_uv:
          base: python:3.10-bullseye
          builders:
            - python:
                version: python3
                uv:
                  version: ==0.5.23
                requirements: [pyproject.toml, uv.lock]
          copies: [local]
          entrypoint: [uv, run, python3, hello.py]
      """
    When you build and run the "hello" variant
    Then the entrypoint will have run successfully

  @set4
  Scenario: Installing Python application dependencies via uv pip install
    Given this "blubber.yaml"
      """
      version: v4
      variants:
        hello_uv:
          base: python:3.10-bullseye
          builders:
            - python:
                version: python3
                uv:
                  version: ==0.5.23
                  variant: pip
                requirements: [requirements.txt]
          copies: [local]
          entrypoint: [uv, run, python3, hello.py]
      """
    When you build and run the "hello" variant
    Then the entrypoint will have run successfully
