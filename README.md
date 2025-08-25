Blubber is a [BuildKit frontend][buildkit-frontend] for building application
container images from a minimal set of declarative constructs in YAML. Its
focus is on composability, determinism, cache efficiency, and secure default
behaviors.

## Requirements

To use Blubber, you'll need:

 * `buildkitd` (included with Docker version 23.0 or greater)
 * A [BuildKit][buildkit] client, any of the following:
   * Docker [Buildx][buildx] (recommended, see [building](#building-a-variant))
   * Docker [Compose][compose-build] (see [compose usage](#docker-compose))
   * Docker [Legacy Builder][build-legacy] (`docker build`, Docker 23.0 or
     greater)

## Usage

### Configuration

A Blubber configuration starts with a `syntax` line and a `version`
declaration.

```yaml
# syntax = docker-registry.wikimedia.org/repos/releng/blubber/buildkit:v1.4.0
version: v4
variants:
  my-variant:
    base: docker-registry.wikimedia.org/bookworm
```

The `syntax` is a reference to a container image that BuildKit will use to
process the configuration (known as a [BuildKit frontend][buildkit-frontend]).

The `version` is the major version of the configuration schema accepted by
Blubber. It rarely changes (only with breaking changes to the schema) and
Blubber will let you know if it isn't right.

A configuration also includes one or more [variants](#variants) (akin to
dockerfile stages).

Once you are ready to write your own configuration, see the
[examples][doc-examples] for build patterns that are possible with Blubber and
the [configuration reference][doc-reference] for documentation about all
available fields.

### Building a variant

You build a variant from a Blubber configuration using `docker buildx build`
just as you would build a stage from a `Dockerfile`.

```console
$ docker buildx build -f blubber.yaml --target my-variant --load .
```

Note the `--load` option will add the resulting image to Docker's image store
so it can be used with `docker run` and other `docker` subcommands. See the
[docker buildx build][buildx-build] CLI reference for all available output
options including those used to publish images to remote registries.

## Compatibility

### Docker Bake

[Bake][bake-intro] is a feature of Docker [Buildx][buildx] for building
multiple targets at once from any number of build contexts and configurations.
It also lets you codify your outputs, tags, and other parameters otherwised
passed in via individual build commands.

Blubber configuration can be referenced from [Bake][bake-intro] configuration
just as you would reference a dockerfile.

```hcl
# bake.hcl
group "default" {
  targets = ["my-variant"]
}

target "my-variant" {
  context = "."
  dockerfile = "blubber.yaml"
  outputs = ["type=registry"]
  tags = ["an.example/registry/my-project/my-variant:stable"]
}
```

You then build your target(s) with [buildx bake][buildx-bake].

```console
$ docker buildx bake -f bake.hcl
```

### Docker Compose

In the same way Blubber files can be referenced from [Bake](#docker-bake)
configuration, they can also be used with [Docker Compose][compose-intro].

In the [build][compose-build] section of your compose configuration, simply
specify the path to your Blubber configuration file just as you would a
dockerfile.

```yaml
services:
  my-service:
    build:
      context: .
      dockerfile: blubber.yaml
      target: my-variant
```

### Predefined build arguments

Blubber allows the same [predefined build arguments][predefined-build-args]
that Docker allows for setting/determining information about the build
environment such as OS and architecture, and available proxies.

### Building for multiple platforms

Blubber supports building for multiple platforms at once and publishing a
single manifest index for the given platforms (aka a "fat" manifest). See the
[OCI Image Index Specification][oci-image-index] for details.

See the documentation of [buildx build][buildx-build-platform] for details on
how to specify your target platforms.

Note that your build process must be aware of the [environment
variables][multi-platform-env-vars] set for multi-platform builds in order to
perform any cross-compilation needed.

### Image attestations

Blubber supports the creation and export of Software Bill of Materials (SBOM)
and provenance metadata in the form of [in-toto attestations][in-toto].

Attestations are exported in the form of image manifests that live alongside
your images within the same manifest list/index. See the upstream BuildKit
documentation on [image attestation storage][bk-image-attestation-storage] for
details.

See the documentation of [buildx build][buildx-build] for details on how to
enable SBOM and provenance metadata creation during a build.

[buildkit]: https://docs.docker.com/build/buildkit/
[buildkit-frontend]: https://docs.docker.com/build/buildkit/#frontend
[buildx]: https://docs.docker.com/reference/cli/docker/buildx/
[buildx-build]: https://docs.docker.com/reference/cli/docker/buildx/build/
[buildx-build-platform]: https://docs.docker.com/reference/cli/docker/buildx/build/#platform
[buildx-build-platform]: https://docs.docker.com/reference/cli/docker/buildx/build/#platform
[bake-intro]: https://docs.docker.com/build/bake/introduction/
[buildx-bake]: https://docs.docker.com/reference/cli/docker/buildx/bake/
[compose-intro]: https://docs.docker.com/compose/
[compose-build]: https://docs.docker.com/reference/compose-file/build/
[build-legacy]: https://docs.docker.com/reference/cli/docker/build-legacy/
[predefined-build-args]: https://docs.docker.com/build/building/variables/#pre-defined-build-arguments
[multi-platform-env-vars]: https://docs.docker.com/build/building/multi-platform/#building-multi-platform-images
[oci-image-index]: https://github.com/opencontainers/image-spec/blob/main/image-index.md
[in-toto]: https://github.com/in-toto/attestation
[bk-image-attestation-storage]: https://github.com/moby/buildkit/blob/master/docs/attestations/attestation-storage.md
[doc-examples]: https://doc.wikimedia.org/releng/blubber/examples/01-basic-usage.html
[doc-reference]: https://doc.wikimedia.org/releng/blubber/configuration.html
