Feature: Python builder with uv
  The Python builder can use the uv package manager for dependency
  installation, either syncing from a lockfile (`uv sync`) or installing
  from requirements files (`uv pip install`).

  Background:
    Given "examples/hello-world-python-uv" as a working directory

  @set3
  Scenario: Installing Python application dependencies via uv sync
    Given this "blubber.yaml"
      """
      version: v4
      variants:
        hello:
          base: python:3.12-trixie
          builders:
            - python:
                version: python3
                uv:
                  version: ==0.11.28
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
        hello:
          base: python:3.12-trixie
          builders:
            - python:
                version: python3
                uv:
                  version: ==0.11.28
                  uvpip: true
                requirements: [requirements.txt]
          copies: [local]
          entrypoint: [uv, run, hello.py]
      """
    When you build and run the "hello" variant
    Then the entrypoint will have run successfully
