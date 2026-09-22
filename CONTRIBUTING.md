# Contributing to Blubber

Blubber is an open source project maintained by Wikimedia Foundation's
Release Engineering Team and developed primarily to support a continuous
delivery pipeline for MediaWiki and related applications. We will, however,
consider any contribution that advances the project in a way that is valuable
to both users inside and outside of WMF and our communities.

## Requirements

 1. If you have not yet contributed to a Wikimedia project, head over to the
    [Wikimedia Developer Portal](https://developer.wikimedia.org/contribute/overview/)
    and read the Code of Conduct and other relevant materials.
 2. Create a [developer account](https://wikitech.wikimedia.org/wiki/Help:Create_a_Wikimedia_developer_account)
    which you will use clone the [Blubber repo](https://gitlab.wikimedia.org/repos/releng/blubber)
    and submit your changes as a merge request.
 3. `go` >= 1.23.5 and related tools
    * To install on rpm style systems: `sudo dnf install golang golang-godoc`
    * To install on apt style systems: `sudo apt install golang golang-golang-x-tools`
    * To install on macOS use [Homebrew](https://brew.sh) and run:
      `brew install go`
    * You can run `go version` to check the golang version.
    * If your distro's go package is too old or unavailable,
      [download](https://go.dev/doc/install) a newer golang version.
 4. The `docker` and `docker buildx` clients.
    * See upstream documentation for the [various ways to install buildx](https://github.com/docker/buildx?tab=readme-ov-file#installing).

## Get the source

Clone the repo from `https://gitlab.wikimedia.org/repos/releng/blubber.git`.

```sh
git clone https://gitlab.wikimedia.org/repos/releng/blubber.git
cd blubber
```

## Test build

Verify a working toolchain prior to making changes. The target builds on your
host, then runs the tests and linters in containers.

```sh
make test-build
```

## Make your changes

Blubber's source code is organized into the following directories/packages:

 - `api`: Contains the JSON Schema used to validate and document
   configuration. If you are adding a feature, you will likely need to
   amend the schema with extra fields and their validation rules.
 - `build`: Types that represent generic build instructions and functions that
   compile Blubber build types to BuildKit LLB.
 - `buildkit`: BuildKit frontend gateway implementation responsible for
   handling Blubber based image builds. It contains the main entrypoint for
   the gRPC gateway process (`buildkit.Build`).
 - `cmd`: Main CLI/process entrypoints.
 - `config`: Types for all supported configuration. Each type is responsible
   for implementing `build.PhaseCompileable` to emit build instructions. If
   you are adding a feature, you will likely be making changes here.
 - `docs`: Contains the Vitepress based user documentation portal.
 - `examples`: Examples used by the acceptance test runner and as sources for
   user documentation. New features should have at least one example/scenario.
 - `util`: Util Go packages. Modules developed here should be broken out into
   separate repos as they mature.

## Running tests and linters

After you have made your changes, run the unit tests and linters to ensure
basic correctness.

```sh
docker buildx bake -f bake.hcl test
```

## More thorough testing of the BuildKit frontend

### Setup

To run acceptance tests or test/debug Blubber's BuildKit frontend, you will
need your own `buildkitd` instance and an accessible OCI registry.
The easiest way to achieve this setup is to run both locally using Docker.

```sh
make dev-env
```

`make dev-env-clean` deletes the `buildkitd` and `registry` containers, the
builder, the network and the images that `make acceptance` built.

### Running the acceptance tests

(See above for ensuring a local registry and `buildkitd`.)

If you are developing or testing a new feature that has a corresponding
acceptance test under the `examples` directory, you can run the suite locally
to ensure it passes.

```sh
make acceptance
```

### Manually testing a blubber.yaml against local changes

(See above for ensuring a local registry and `buildkitd`.)

To manually test a `blubber.yaml` configuration against your local changes,
first build and publish the Blubber `buildkit` gateway image.

```sh
make dev-frontend
```

Add a `syntax` line to your `blubber.yaml`.

```yaml
# syntax=registry:5000/blubber/buildkit:latest
version: v4
variants:
  foo:
    # [...]
```

From the directory that holds your `blubber.yaml`, build a variant.

```sh
docker buildx build -f blubber.yaml --target foo .
```

To see debugging information, you can use `docker buildx --debug ...` and/or
tail the `buildkitd` logs.

```sh
docker logs -f buildkitd
```

## Getting your changes reviewed and merged

Push your changes to GitLab for review. Refer to the
[workflow guide](https://www.mediawiki.org/wiki/GitLab/Workflows/Making_a_merge_request)
for details.
