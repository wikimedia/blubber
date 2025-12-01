# Blubber configuration (v4)

## version
`.version` _string_ (required)

Blubber configuration version. Currently `v4`.

## apt
`.apt` _object_


### packages
`.apt.packages` _array&lt;string&gt;_

Packages to install from APT sources of base image.

For example:

```yaml
apt:
  sources:
    - url: http://apt.wikimedia.org/wikimedia
      distribution: buster-wikimedia
      components: [thirdparty/confluent]
  packages: [ ca-certificates, confluent-kafka-2.11, curl ]
```

#### packages[]
`.apt.packages[]` _string_

### apt object
`.apt.packages` _object_

Key-Value pairs of target release and packages to install from APT sources.

#### apt array
`.apt.packages.*` _array&lt;string&gt;_

The packages to install using the target release.

#### *[]
`.apt.packages.*[]` _string_

### proxies
`.apt.proxies` _array&lt;object|string&gt;_

HTTP/HTTPS proxies to use during package installation.


#### proxies[]
`.apt.proxies[]` _string_

Shorthand configuration of a proxy that applies to all sources of its protocol.

#### proxies[]
`.apt.proxies[]` _object_

Proxy for either all sources of a given protocol or a specific source.

#### source
`.apt.proxies[].source` _string_

APT source to which this proxy applies.

#### url
`.apt.proxies[].url` _string_ (required)

HTTP/HTTPS proxy URL.

### sources
`.apt.sources` _array&lt;object&gt;_

Additional APT sources to configure prior to package installation.

#### APT sources object
`.apt.sources[]` _object_

APT source URL, distribution/release name, and components.

#### components
`.apt.sources[].components` _array&lt;string&gt;_

List of distribution components (e.g. main, contrib). See [APT repository structure](https://wikitech.wikimedia.org/wiki/APT_repository#Repository_Structure) for more information about our use of the distribution and component fields.

#### components[]
`.apt.sources[].components[]` _string_

#### distribution
`.apt.sources[].distribution` _string_

Debian distribution/release name (e.g. buster). See [APT repository structure](https://wikitech.wikimedia.org/wiki/APT_repository#Repository_Structure) for more information about our use of the distribution and component fields.

#### signed-by
`.apt.sources[].signed-by` _string_

ASCII encoded public key(s) with which to verify the source. See [Debian source.list](https://wiki.debian.org/DebianRepository/UseThirdParty#Sources.list_entry) for more information on the expected format.

#### url
`.apt.sources[].url` _string_ (required)

APT source URL.

## arguments
`.arguments` _object_

Build argument names and default values. Values may be passed in at build time. Final build arguments (defaults merged with build-time values) are exposed as environment variables and effect only certain configuration fields such as builder commands and scripts.

See the description of each field for whether it supports environment and build argument expansion.

Note that build arguments become environment variables in the resulting image configuration and should not be used to store sensitive values.

## base
`.base` _null|string_

Base image on which the new image will be built; a list of available images can be found by querying the [Wikimedia Docker Registry](https://docker-registry.wikimedia.org/).

## builder
`.builder` _object_

Run an arbitrary build command.

### caches
`.builder.caches` _array&lt;object|string&gt;_

Mount a number of caching filesystems that will persist between builds. Each cache mount should specify the a `destination` path, and optionally the level of `access` (`"shared"` by default) and a unique `id` (`destination` is used by default).

Cache mounts are most useful in speeding up build processes that save downloaded files or compilation state in local directories.

Example

```yaml
runs:
  environment: { GOCACHE: "/var/cache/go" }
builders:
  - custom:
      command: "go build"
      requirements: [.]
      caches:
        - destination: "/var/cache/go"
```

Example (shorthand and using an environment variable)

```yaml
runs:
  environment: { GOCACHE: "/var/cache/go" }
builders:
  - custom:
      command: "go build"
      requirements: [.]
      caches: [ "${GOCACHE}" ]
```


#### caches[]
`.builder.caches[]` _string_

#### caches[]
`.builder.caches[]` _object_

#### access
`.builder.caches[].access` __

Level of access between concurrent build processes. This depends on the underlying build process and what kind of locking it employs.

 - A `shared` cache can be written to concurrently.

 - A `private` cache ensures only one writer at a time and is non-blocking (a new filesystem is created for concurrent writers).

 - A `locked` cache ensures only one writer at a time and will block other build processes until the first releases the lock.

#### destination
`.builder.caches[].destination` _string_

Destination path in the build container where the cache filesystem will be mounted.

Supports environment variables and build arguments.

#### id
`.builder.caches[].id` _string_

A unique ID used to persistent the cache filesystem between builds. The `destination` path is used by default.


### command
`.builder.command` _array&lt;string&gt;_

Command and arguments of an arbitrary build command, for example `[make, build]`.

Only one of `script` or `command` may be used for a given builder.

#### command[]
`.builder.command[]` _string_

### command
`.builder.command` _string_

Arbitrary build command, for example `"make build"`.

Only one of `script` or `command` may be used for a given builder.

Supports environment variables and build arguments.

### mounts
`.builder.mounts` _array&lt;object|string&gt;_

Mount a number of filesystems from either the local build context, other variants or images. Each mount should specify the name of the variant or image (or `local` for the local build context), a `destination` path, and optionally the `source` path within the filesystem to use as the root of the mount. Note that changes to files under source mounts are discarded after each builder command completes.

Mounts are most useful when you need files from some other filesystem for a build process but do not want the files in the resulting image.

Example

```yaml
builders:
  - custom:
      command: "make -C /src/foo install"
      mounts:
        - from: foo
          destination: /src/foo
```

Example (shorthand for mounting another variant, image or `local`)

```yaml
builders:
  - custom:
      command: "make install"
      mounts: [ local ]
```


#### mounts[]
`.builder.mounts[]` _string_

#### mounts[]
`.builder.mounts[]` _object_

#### destination
`.builder.mounts[].destination` _string_

Destination path in the build container where the root of the `from` filesystem will be mounted. Defaults to the working directory.

Supports environment variables and build arguments.

#### from
`.builder.mounts[].from` _string_

Variant or image filesystem to mount. Set to `local` to mount the local build context.

#### source
`.builder.mounts[].source` _string_

Path within the `from` filesystem to use as the root directory of the mount.

Supports environment variables and build arguments.

### requirements
`.builder.requirements` _array&lt;object|string&gt;_

Path of files/directories to copy from the local build context. This is done before any of the build steps. Note that there are two possible formats for `requirements`. The first is a simple shorthand notation that means copying a list of source files from the local build context to a destination of the same relative path in the image. The second is a longhand form that gives more control over the source context (local or another variant), source and destination paths.

Example (shorthand)

```yaml
builder:
  command: "some build command"
  requirements: [config.json, Makefile, src/] # copy files/directories to the same paths in the image
```

Example (longhand/advanced)

```yaml
builder:
  command: "some build command"
  requirements:
    - from: local
      source: config.production.json
      destination: config.json
    - Makefile # note that longhand/shorthand can be mixed
    - src/
    - from: other-variant
      source: /srv/some/previous/build/product
      destination: dist/product
```


#### requirements[]
`.builder.requirements[]` _string_

#### requirements[]
`.builder.requirements[]` _object_

#### destination
`.builder.requirements[].destination` _string_

Destination path. Defaults to source path.

#### exclude
`.builder.requirements[].exclude` _array&lt;string&gt;_

Exclude files that match any of these patterns.

#### exclude[]
`.builder.requirements[].exclude[]` _string_

A valid glob pattern (e.g. `**/*.swp` or `!**/only-these`).

#### from
`.builder.requirements[].from` _null|string_

Variant from which to copy files. Set to `local` to copy build-context files that match the `source` pattern, or another variant name to copy files that match the `source` pattern from the variant's filesystem.

#### source
`.builder.requirements[].source` _string_

Path of files/directories to copy.

### script
`.builder.script` _string_

Content of a build script to execute. The script may contain its own shebang (`#!`) line to specify the interpreter. If no shebang is provided, `#!/bin/sh` will be used.

Note that the script will not be included in the resulting image. It is written to its own scratch filesystem and only mounted temporarily during execution.

Example

```yaml
builder:
  requirements: [targets]
  script: |
    #!/bin/bash
    for target in $(cat targets); do
      make $target
    done
```

Only one of `script` or `command` may be used for a given builder. Supports environment variables and build arguments.

## builders
`.builders` _array&lt;object&gt;_

Multiple builders to be executed in an explicit order. You can specify any of the predefined standalone builder keys (node, python and php), but each can only appear once. Additionally, any number of custom keys can appear; their definition and subkeys are the same as the standalone builder key.


### builders[]
`.builders[]` _object_

#### custom
`.builders[].custom` _object_

Run an arbitrary build command.

#### caches
`.builders[].custom.caches` _array&lt;object|string&gt;_

Mount a number of caching filesystems that will persist between builds. Each cache mount should specify the a `destination` path, and optionally the level of `access` (`"shared"` by default) and a unique `id` (`destination` is used by default).

Cache mounts are most useful in speeding up build processes that save downloaded files or compilation state in local directories.

Example

```yaml
runs:
  environment: { GOCACHE: "/var/cache/go" }
builders:
  - custom:
      command: "go build"
      requirements: [.]
      caches:
        - destination: "/var/cache/go"
```

Example (shorthand and using an environment variable)

```yaml
runs:
  environment: { GOCACHE: "/var/cache/go" }
builders:
  - custom:
      command: "go build"
      requirements: [.]
      caches: [ "${GOCACHE}" ]
```


#### caches[]
`.builders[].custom.caches[]` _string_

#### caches[]
`.builders[].custom.caches[]` _object_

#### access
`.builders[].custom.caches[].access` __

Level of access between concurrent build processes. This depends on the underlying build process and what kind of locking it employs.

 - A `shared` cache can be written to concurrently.

 - A `private` cache ensures only one writer at a time and is non-blocking (a new filesystem is created for concurrent writers).

 - A `locked` cache ensures only one writer at a time and will block other build processes until the first releases the lock.

#### destination
`.builders[].custom.caches[].destination` _string_

Destination path in the build container where the cache filesystem will be mounted.

Supports environment variables and build arguments.

#### id
`.builders[].custom.caches[].id` _string_

A unique ID used to persistent the cache filesystem between builds. The `destination` path is used by default.


#### command
`.builders[].custom.command` _array&lt;string&gt;_

Command and arguments of an arbitrary build command, for example `[make, build]`.

Only one of `script` or `command` may be used for a given builder.

#### command[]
`.builders[].custom.command[]` _string_

#### command
`.builders[].custom.command` _string_

Arbitrary build command, for example `"make build"`.

Only one of `script` or `command` may be used for a given builder.

Supports environment variables and build arguments.

#### mounts
`.builders[].custom.mounts` _array&lt;object|string&gt;_

Mount a number of filesystems from either the local build context, other variants or images. Each mount should specify the name of the variant or image (or `local` for the local build context), a `destination` path, and optionally the `source` path within the filesystem to use as the root of the mount. Note that changes to files under source mounts are discarded after each builder command completes.

Mounts are most useful when you need files from some other filesystem for a build process but do not want the files in the resulting image.

Example

```yaml
builders:
  - custom:
      command: "make -C /src/foo install"
      mounts:
        - from: foo
          destination: /src/foo
```

Example (shorthand for mounting another variant, image or `local`)

```yaml
builders:
  - custom:
      command: "make install"
      mounts: [ local ]
```


#### mounts[]
`.builders[].custom.mounts[]` _string_

#### mounts[]
`.builders[].custom.mounts[]` _object_

#### destination
`.builders[].custom.mounts[].destination` _string_

Destination path in the build container where the root of the `from` filesystem will be mounted. Defaults to the working directory.

Supports environment variables and build arguments.

#### from
`.builders[].custom.mounts[].from` _string_

Variant or image filesystem to mount. Set to `local` to mount the local build context.

#### source
`.builders[].custom.mounts[].source` _string_

Path within the `from` filesystem to use as the root directory of the mount.

Supports environment variables and build arguments.

#### requirements
`.builders[].custom.requirements` _array&lt;object|string&gt;_

Path of files/directories to copy from the local build context. This is done before any of the build steps. Note that there are two possible formats for `requirements`. The first is a simple shorthand notation that means copying a list of source files from the local build context to a destination of the same relative path in the image. The second is a longhand form that gives more control over the source context (local or another variant), source and destination paths.

Example (shorthand)

```yaml
builder:
  command: "some build command"
  requirements: [config.json, Makefile, src/] # copy files/directories to the same paths in the image
```

Example (longhand/advanced)

```yaml
builder:
  command: "some build command"
  requirements:
    - from: local
      source: config.production.json
      destination: config.json
    - Makefile # note that longhand/shorthand can be mixed
    - src/
    - from: other-variant
      source: /srv/some/previous/build/product
      destination: dist/product
```


#### requirements[]
`.builders[].custom.requirements[]` _string_

#### requirements[]
`.builders[].custom.requirements[]` _object_

#### destination
`.builders[].custom.requirements[].destination` _string_

Destination path. Defaults to source path.

#### exclude
`.builders[].custom.requirements[].exclude` _array&lt;string&gt;_

Exclude files that match any of these patterns.

#### exclude[]
`.builders[].custom.requirements[].exclude[]` _string_

A valid glob pattern (e.g. `**/*.swp` or `!**/only-these`).

#### from
`.builders[].custom.requirements[].from` _null|string_

Variant from which to copy files. Set to `local` to copy build-context files that match the `source` pattern, or another variant name to copy files that match the `source` pattern from the variant's filesystem.

#### source
`.builders[].custom.requirements[].source` _string_

Path of files/directories to copy.

#### script
`.builders[].custom.script` _string_

Content of a build script to execute. The script may contain its own shebang (`#!`) line to specify the interpreter. If no shebang is provided, `#!/bin/sh` will be used.

Note that the script will not be included in the resulting image. It is written to its own scratch filesystem and only mounted temporarily during execution.

Example

```yaml
builder:
  requirements: [targets]
  script: |
    #!/bin/bash
    for target in $(cat targets); do
      make $target
    done
```

Only one of `script` or `command` may be used for a given builder. Supports environment variables and build arguments.

### builders[]
`.builders[]` _object_

#### node
`.builders[].node` _object_

Configuration related to the NodeJS/NPM environment

#### allow-dedupe-failure
`.builders[].node.allow-dedupe-failure` _boolean_

Whether to allow `npm dedupe` to fail; can be used to temporarily unblock CI while conflicts are resolved.

#### env
`.builders[].node.env` _string_

Node environment (e.g. production, etc.). Sets the environment variable `NODE_ENV`. Will pass `npm install --production` and run `npm dedupe` if set to production.

#### requirements
`.builders[].node.requirements` _array&lt;object|string&gt;_

Path of files/directories to copy from the local build context. This is done before any of the build steps. Note that there are two possible formats for `requirements`. The first is a simple shorthand notation that means copying a list of source files from the local build context to a destination of the same relative path in the image. The second is a longhand form that gives more control over the source context (local or another variant), source and destination paths.

Example (shorthand)

```yaml
builder:
  command: "some build command"
  requirements: [config.json, Makefile, src/] # copy files/directories to the same paths in the image
```

Example (longhand/advanced)

```yaml
builder:
  command: "some build command"
  requirements:
    - from: local
      source: config.production.json
      destination: config.json
    - Makefile # note that longhand/shorthand can be mixed
    - src/
    - from: other-variant
      source: /srv/some/previous/build/product
      destination: dist/product
```


#### requirements[]
`.builders[].node.requirements[]` _string_

#### requirements[]
`.builders[].node.requirements[]` _object_

#### destination
`.builders[].node.requirements[].destination` _string_

Destination path. Defaults to source path.

#### exclude
`.builders[].node.requirements[].exclude` _array&lt;string&gt;_

Exclude files that match any of these patterns.

#### exclude[]
`.builders[].node.requirements[].exclude[]` _string_

A valid glob pattern (e.g. `**/*.swp` or `!**/only-these`).

#### from
`.builders[].node.requirements[].from` _null|string_

Variant from which to copy files. Set to `local` to copy build-context files that match the `source` pattern, or another variant name to copy files that match the `source` pattern from the variant's filesystem.

#### source
`.builders[].node.requirements[].source` _string_

Path of files/directories to copy.

#### use-npm-ci
`.builders[].node.use-npm-ci` _boolean_

Whether to run `npm ci` instead of `npm install`.

### builders[]
`.builders[]` _object_

#### php
`.builders[].php` _object_

#### production
`.builders[].php.production` _boolean_

Whether to inject the --no-dev flag into the install command.

#### requirements
`.builders[].php.requirements` _array&lt;object|string&gt;_

Path of files/directories to copy from the local build context. This is done before any of the build steps. Note that there are two possible formats for `requirements`. The first is a simple shorthand notation that means copying a list of source files from the local build context to a destination of the same relative path in the image. The second is a longhand form that gives more control over the source context (local or another variant), source and destination paths.

Example (shorthand)

```yaml
builder:
  command: "some build command"
  requirements: [config.json, Makefile, src/] # copy files/directories to the same paths in the image
```

Example (longhand/advanced)

```yaml
builder:
  command: "some build command"
  requirements:
    - from: local
      source: config.production.json
      destination: config.json
    - Makefile # note that longhand/shorthand can be mixed
    - src/
    - from: other-variant
      source: /srv/some/previous/build/product
      destination: dist/product
```


#### requirements[]
`.builders[].php.requirements[]` _string_

#### requirements[]
`.builders[].php.requirements[]` _object_

#### destination
`.builders[].php.requirements[].destination` _string_

Destination path. Defaults to source path.

#### exclude
`.builders[].php.requirements[].exclude` _array&lt;string&gt;_

Exclude files that match any of these patterns.

#### exclude[]
`.builders[].php.requirements[].exclude[]` _string_

A valid glob pattern (e.g. `**/*.swp` or `!**/only-these`).

#### from
`.builders[].php.requirements[].from` _null|string_

Variant from which to copy files. Set to `local` to copy build-context files that match the `source` pattern, or another variant name to copy files that match the `source` pattern from the variant's filesystem.

#### source
`.builders[].php.requirements[].source` _string_

Path of files/directories to copy.

### builders[]
`.builders[]` _object_

#### python
`.builders[].python` _object_

Predefined configurations for Python build tools

#### no-deps
`.builders[].python.no-deps` _boolean_

Inject `--no-deps` into the `pip install` command. All transitive requirements thus must be explicitly listed in the requirements file. `pip check` will be run to verify all dependencies are fulfilled.

#### poetry
`.builders[].python.poetry` _object_

Configuration related to installation of pip dependencies using [Poetry](https://python-poetry.org/).

#### devel
`.builders[].python.poetry.devel` _boolean_

Whether to install development dependencies or not when using Poetry 1.x.

#### only
`.builders[].python.poetry.only` _string_

Poetry 2.x dependency groups to install (e.g. `main` or `main,docs`). Translates to `--only main,docs`. Not compatible with `devel` field.

#### version
`.builders[].python.poetry.version` _string_

Version constraint for installing Poetry package.

#### without
`.builders[].python.poetry.without` _string_

Poetry 2.x dependency groups to exclude (e.g. `dev` or `dev,test`). Translates to `--without dev,test`. Not compatible with `devel` field.

#### requirements
`.builders[].python.requirements` _array&lt;object|string&gt;_

Path of files/directories to copy from the local build context. This is done before any of the build steps. Note that there are two possible formats for `requirements`. The first is a simple shorthand notation that means copying a list of source files from the local build context to a destination of the same relative path in the image. The second is a longhand form that gives more control over the source context (local or another variant), source and destination paths.

Example (shorthand)

```yaml
builder:
  command: "some build command"
  requirements: [config.json, Makefile, src/] # copy files/directories to the same paths in the image
```

Example (longhand/advanced)

```yaml
builder:
  command: "some build command"
  requirements:
    - from: local
      source: config.production.json
      destination: config.json
    - Makefile # note that longhand/shorthand can be mixed
    - src/
    - from: other-variant
      source: /srv/some/previous/build/product
      destination: dist/product
```


#### requirements[]
`.builders[].python.requirements[]` _string_

#### requirements[]
`.builders[].python.requirements[]` _object_

#### destination
`.builders[].python.requirements[].destination` _string_

Destination path. Defaults to source path.

#### exclude
`.builders[].python.requirements[].exclude` _array&lt;string&gt;_

Exclude files that match any of these patterns.

#### exclude[]
`.builders[].python.requirements[].exclude[]` _string_

A valid glob pattern (e.g. `**/*.swp` or `!**/only-these`).

#### from
`.builders[].python.requirements[].from` _null|string_

Variant from which to copy files. Set to `local` to copy build-context files that match the `source` pattern, or another variant name to copy files that match the `source` pattern from the variant's filesystem.

#### source
`.builders[].python.requirements[].source` _string_

Path of files/directories to copy.

#### use-system-site-packages
`.builders[].python.use-system-site-packages` _boolean_

Whether to inject the --system-site-packages flag into the venv setup command.

#### version
`.builders[].python.version` _string_

Python binary present in the system (e.g. python3).

## entrypoint
`.entrypoint` _array&lt;string&gt;_

Runtime entry point command and arguments.

### entrypoint[]
`.entrypoint[]` _string_

## lives
`.lives` _object_

### as
`.lives.as` _string_

Owner (name) of application files within the container.

### gid
`.lives.gid` _integer_

Group owner (GID) of application files within the container.

### in
`.lives.in` _string_

Working directory used when the variant is built and at runtime unless `runs.in` is specified.

### uid
`.lives.uid` _integer_

Owner (UID) of application files within the container.

## node
`.node` _object_

Configuration related to the NodeJS/NPM environment

### allow-dedupe-failure
`.node.allow-dedupe-failure` _boolean_

Whether to allow `npm dedupe` to fail; can be used to temporarily unblock CI while conflicts are resolved.

### env
`.node.env` _string_

Node environment (e.g. production, etc.). Sets the environment variable `NODE_ENV`. Will pass `npm install --production` and run `npm dedupe` if set to production.

### requirements
`.node.requirements` _array&lt;object|string&gt;_

Path of files/directories to copy from the local build context. This is done before any of the build steps. Note that there are two possible formats for `requirements`. The first is a simple shorthand notation that means copying a list of source files from the local build context to a destination of the same relative path in the image. The second is a longhand form that gives more control over the source context (local or another variant), source and destination paths.

Example (shorthand)

```yaml
builder:
  command: "some build command"
  requirements: [config.json, Makefile, src/] # copy files/directories to the same paths in the image
```

Example (longhand/advanced)

```yaml
builder:
  command: "some build command"
  requirements:
    - from: local
      source: config.production.json
      destination: config.json
    - Makefile # note that longhand/shorthand can be mixed
    - src/
    - from: other-variant
      source: /srv/some/previous/build/product
      destination: dist/product
```


#### requirements[]
`.node.requirements[]` _string_

#### requirements[]
`.node.requirements[]` _object_

#### destination
`.node.requirements[].destination` _string_

Destination path. Defaults to source path.

#### exclude
`.node.requirements[].exclude` _array&lt;string&gt;_

Exclude files that match any of these patterns.

#### exclude[]
`.node.requirements[].exclude[]` _string_

A valid glob pattern (e.g. `**/*.swp` or `!**/only-these`).

#### from
`.node.requirements[].from` _null|string_

Variant from which to copy files. Set to `local` to copy build-context files that match the `source` pattern, or another variant name to copy files that match the `source` pattern from the variant's filesystem.

#### source
`.node.requirements[].source` _string_

Path of files/directories to copy.

### use-npm-ci
`.node.use-npm-ci` _boolean_

Whether to run `npm ci` instead of `npm install`.

## php
`.php` _object_

### production
`.php.production` _boolean_

Whether to inject the --no-dev flag into the install command.

### requirements
`.php.requirements` _array&lt;object|string&gt;_

Path of files/directories to copy from the local build context. This is done before any of the build steps. Note that there are two possible formats for `requirements`. The first is a simple shorthand notation that means copying a list of source files from the local build context to a destination of the same relative path in the image. The second is a longhand form that gives more control over the source context (local or another variant), source and destination paths.

Example (shorthand)

```yaml
builder:
  command: "some build command"
  requirements: [config.json, Makefile, src/] # copy files/directories to the same paths in the image
```

Example (longhand/advanced)

```yaml
builder:
  command: "some build command"
  requirements:
    - from: local
      source: config.production.json
      destination: config.json
    - Makefile # note that longhand/shorthand can be mixed
    - src/
    - from: other-variant
      source: /srv/some/previous/build/product
      destination: dist/product
```


#### requirements[]
`.php.requirements[]` _string_

#### requirements[]
`.php.requirements[]` _object_

#### destination
`.php.requirements[].destination` _string_

Destination path. Defaults to source path.

#### exclude
`.php.requirements[].exclude` _array&lt;string&gt;_

Exclude files that match any of these patterns.

#### exclude[]
`.php.requirements[].exclude[]` _string_

A valid glob pattern (e.g. `**/*.swp` or `!**/only-these`).

#### from
`.php.requirements[].from` _null|string_

Variant from which to copy files. Set to `local` to copy build-context files that match the `source` pattern, or another variant name to copy files that match the `source` pattern from the variant's filesystem.

#### source
`.php.requirements[].source` _string_

Path of files/directories to copy.

## python
`.python` _object_

Predefined configurations for Python build tools

### no-deps
`.python.no-deps` _boolean_

Inject `--no-deps` into the `pip install` command. All transitive requirements thus must be explicitly listed in the requirements file. `pip check` will be run to verify all dependencies are fulfilled.

### poetry
`.python.poetry` _object_

Configuration related to installation of pip dependencies using [Poetry](https://python-poetry.org/).

#### devel
`.python.poetry.devel` _boolean_

Whether to install development dependencies or not when using Poetry 1.x.

#### only
`.python.poetry.only` _string_

Poetry 2.x dependency groups to install (e.g. `main` or `main,docs`). Translates to `--only main,docs`. Not compatible with `devel` field.

#### version
`.python.poetry.version` _string_

Version constraint for installing Poetry package.

#### without
`.python.poetry.without` _string_

Poetry 2.x dependency groups to exclude (e.g. `dev` or `dev,test`). Translates to `--without dev,test`. Not compatible with `devel` field.

### requirements
`.python.requirements` _array&lt;object|string&gt;_

Path of files/directories to copy from the local build context. This is done before any of the build steps. Note that there are two possible formats for `requirements`. The first is a simple shorthand notation that means copying a list of source files from the local build context to a destination of the same relative path in the image. The second is a longhand form that gives more control over the source context (local or another variant), source and destination paths.

Example (shorthand)

```yaml
builder:
  command: "some build command"
  requirements: [config.json, Makefile, src/] # copy files/directories to the same paths in the image
```

Example (longhand/advanced)

```yaml
builder:
  command: "some build command"
  requirements:
    - from: local
      source: config.production.json
      destination: config.json
    - Makefile # note that longhand/shorthand can be mixed
    - src/
    - from: other-variant
      source: /srv/some/previous/build/product
      destination: dist/product
```


#### requirements[]
`.python.requirements[]` _string_

#### requirements[]
`.python.requirements[]` _object_

#### destination
`.python.requirements[].destination` _string_

Destination path. Defaults to source path.

#### exclude
`.python.requirements[].exclude` _array&lt;string&gt;_

Exclude files that match any of these patterns.

#### exclude[]
`.python.requirements[].exclude[]` _string_

A valid glob pattern (e.g. `**/*.swp` or `!**/only-these`).

#### from
`.python.requirements[].from` _null|string_

Variant from which to copy files. Set to `local` to copy build-context files that match the `source` pattern, or another variant name to copy files that match the `source` pattern from the variant's filesystem.

#### source
`.python.requirements[].source` _string_

Path of files/directories to copy.

### use-system-site-packages
`.python.use-system-site-packages` _boolean_

Whether to inject the --system-site-packages flag into the venv setup command.

### version
`.python.version` _string_

Python binary present in the system (e.g. python3).

## runs
`.runs` _object_

Settings for things run in the container.

### as
`.runs.as` _string_

Runtime process owner (name) of application entrypoint.

### environment
`.runs.environment` _object_

Environment variables and values to be set before entrypoint execution.

### gid
`.runs.gid` _integer_

Runtime process group (GID) of application entrypoint.

### in
`.runs.in` _string_

Runtime working directory. The `lives.in` directory is used by default.

### insecurely
`.runs.insecurely` _boolean_

Skip dropping of privileges to the runtime process owner before entrypoint execution. Production variants should have this set to `false`, but other variants may set it to `true` in some circumstances, for example when enabling [caching for ESLint](https://eslint.org/docs/user-guide/command-line-interface#caching).

### uid
`.runs.uid` _integer_

Runtime process owner (UID) of application entrypoint.

## variants
`.variants` _object_

Configuration variants (e.g. development, test, production).

Blubber can build several variants of an image from the same specification file. The variants are named and described under the `variants` top level item. Typically, there are variants for development versus production: the development variant might have more debugging tools, while the production variant should have no extra software installed to minimize the risk of security issues and other problems.

A variant is built using the top level items, combined with the items for the variant. So if the top level `apt` installed some packages, and the variant's `apt` some other packages, both sets of packages get installed in that variant.


### variant
`.variants.*` _object_

#### apt
`.variants.*.apt` _object_


#### packages
`.variants.*.apt.packages` _array&lt;string&gt;_

Packages to install from APT sources of base image.

For example:

```yaml
apt:
  sources:
    - url: http://apt.wikimedia.org/wikimedia
      distribution: buster-wikimedia
      components: [thirdparty/confluent]
  packages: [ ca-certificates, confluent-kafka-2.11, curl ]
```

#### packages[]
`.variants.*.apt.packages[]` _string_

#### apt object
`.variants.*.apt.packages` _object_

Key-Value pairs of target release and packages to install from APT sources.

#### apt array
`.variants.*.apt.packages.*` _array&lt;string&gt;_

The packages to install using the target release.

#### *[]
`.variants.*.apt.packages.*[]` _string_

#### proxies
`.variants.*.apt.proxies` _array&lt;object|string&gt;_

HTTP/HTTPS proxies to use during package installation.


#### proxies[]
`.variants.*.apt.proxies[]` _string_

Shorthand configuration of a proxy that applies to all sources of its protocol.

#### proxies[]
`.variants.*.apt.proxies[]` _object_

Proxy for either all sources of a given protocol or a specific source.

#### source
`.variants.*.apt.proxies[].source` _string_

APT source to which this proxy applies.

#### url
`.variants.*.apt.proxies[].url` _string_ (required)

HTTP/HTTPS proxy URL.

#### sources
`.variants.*.apt.sources` _array&lt;object&gt;_

Additional APT sources to configure prior to package installation.

#### APT sources object
`.variants.*.apt.sources[]` _object_

APT source URL, distribution/release name, and components.

#### components
`.variants.*.apt.sources[].components` _array&lt;string&gt;_

List of distribution components (e.g. main, contrib). See [APT repository structure](https://wikitech.wikimedia.org/wiki/APT_repository#Repository_Structure) for more information about our use of the distribution and component fields.

#### components[]
`.variants.*.apt.sources[].components[]` _string_

#### distribution
`.variants.*.apt.sources[].distribution` _string_

Debian distribution/release name (e.g. buster). See [APT repository structure](https://wikitech.wikimedia.org/wiki/APT_repository#Repository_Structure) for more information about our use of the distribution and component fields.

#### signed-by
`.variants.*.apt.sources[].signed-by` _string_

ASCII encoded public key(s) with which to verify the source. See [Debian source.list](https://wiki.debian.org/DebianRepository/UseThirdParty#Sources.list_entry) for more information on the expected format.

#### url
`.variants.*.apt.sources[].url` _string_ (required)

APT source URL.

#### arguments
`.variants.*.arguments` _object_

Build argument names and default values. Values may be passed in at build time. Final build arguments (defaults merged with build-time values) are exposed as environment variables and effect only certain configuration fields such as builder commands and scripts.

See the description of each field for whether it supports environment and build argument expansion.

Note that build arguments become environment variables in the resulting image configuration and should not be used to store sensitive values.

#### base
`.variants.*.base` _null|string_

Base image on which the new image will be built; a list of available images can be found by querying the [Wikimedia Docker Registry](https://docker-registry.wikimedia.org/).

#### builder
`.variants.*.builder` _object_

Run an arbitrary build command.

#### caches
`.variants.*.builder.caches` _array&lt;object|string&gt;_

Mount a number of caching filesystems that will persist between builds. Each cache mount should specify the a `destination` path, and optionally the level of `access` (`"shared"` by default) and a unique `id` (`destination` is used by default).

Cache mounts are most useful in speeding up build processes that save downloaded files or compilation state in local directories.

Example

```yaml
runs:
  environment: { GOCACHE: "/var/cache/go" }
builders:
  - custom:
      command: "go build"
      requirements: [.]
      caches:
        - destination: "/var/cache/go"
```

Example (shorthand and using an environment variable)

```yaml
runs:
  environment: { GOCACHE: "/var/cache/go" }
builders:
  - custom:
      command: "go build"
      requirements: [.]
      caches: [ "${GOCACHE}" ]
```


#### caches[]
`.variants.*.builder.caches[]` _string_

#### caches[]
`.variants.*.builder.caches[]` _object_

#### access
`.variants.*.builder.caches[].access` __

Level of access between concurrent build processes. This depends on the underlying build process and what kind of locking it employs.

 - A `shared` cache can be written to concurrently.

 - A `private` cache ensures only one writer at a time and is non-blocking (a new filesystem is created for concurrent writers).

 - A `locked` cache ensures only one writer at a time and will block other build processes until the first releases the lock.

#### destination
`.variants.*.builder.caches[].destination` _string_

Destination path in the build container where the cache filesystem will be mounted.

Supports environment variables and build arguments.

#### id
`.variants.*.builder.caches[].id` _string_

A unique ID used to persistent the cache filesystem between builds. The `destination` path is used by default.


#### command
`.variants.*.builder.command` _array&lt;string&gt;_

Command and arguments of an arbitrary build command, for example `[make, build]`.

Only one of `script` or `command` may be used for a given builder.

#### command[]
`.variants.*.builder.command[]` _string_

#### command
`.variants.*.builder.command` _string_

Arbitrary build command, for example `"make build"`.

Only one of `script` or `command` may be used for a given builder.

Supports environment variables and build arguments.

#### mounts
`.variants.*.builder.mounts` _array&lt;object|string&gt;_

Mount a number of filesystems from either the local build context, other variants or images. Each mount should specify the name of the variant or image (or `local` for the local build context), a `destination` path, and optionally the `source` path within the filesystem to use as the root of the mount. Note that changes to files under source mounts are discarded after each builder command completes.

Mounts are most useful when you need files from some other filesystem for a build process but do not want the files in the resulting image.

Example

```yaml
builders:
  - custom:
      command: "make -C /src/foo install"
      mounts:
        - from: foo
          destination: /src/foo
```

Example (shorthand for mounting another variant, image or `local`)

```yaml
builders:
  - custom:
      command: "make install"
      mounts: [ local ]
```


#### mounts[]
`.variants.*.builder.mounts[]` _string_

#### mounts[]
`.variants.*.builder.mounts[]` _object_

#### destination
`.variants.*.builder.mounts[].destination` _string_

Destination path in the build container where the root of the `from` filesystem will be mounted. Defaults to the working directory.

Supports environment variables and build arguments.

#### from
`.variants.*.builder.mounts[].from` _string_

Variant or image filesystem to mount. Set to `local` to mount the local build context.

#### source
`.variants.*.builder.mounts[].source` _string_

Path within the `from` filesystem to use as the root directory of the mount.

Supports environment variables and build arguments.

#### requirements
`.variants.*.builder.requirements` _array&lt;object|string&gt;_

Path of files/directories to copy from the local build context. This is done before any of the build steps. Note that there are two possible formats for `requirements`. The first is a simple shorthand notation that means copying a list of source files from the local build context to a destination of the same relative path in the image. The second is a longhand form that gives more control over the source context (local or another variant), source and destination paths.

Example (shorthand)

```yaml
builder:
  command: "some build command"
  requirements: [config.json, Makefile, src/] # copy files/directories to the same paths in the image
```

Example (longhand/advanced)

```yaml
builder:
  command: "some build command"
  requirements:
    - from: local
      source: config.production.json
      destination: config.json
    - Makefile # note that longhand/shorthand can be mixed
    - src/
    - from: other-variant
      source: /srv/some/previous/build/product
      destination: dist/product
```


#### requirements[]
`.variants.*.builder.requirements[]` _string_

#### requirements[]
`.variants.*.builder.requirements[]` _object_

#### destination
`.variants.*.builder.requirements[].destination` _string_

Destination path. Defaults to source path.

#### exclude
`.variants.*.builder.requirements[].exclude` _array&lt;string&gt;_

Exclude files that match any of these patterns.

#### exclude[]
`.variants.*.builder.requirements[].exclude[]` _string_

A valid glob pattern (e.g. `**/*.swp` or `!**/only-these`).

#### from
`.variants.*.builder.requirements[].from` _null|string_

Variant from which to copy files. Set to `local` to copy build-context files that match the `source` pattern, or another variant name to copy files that match the `source` pattern from the variant's filesystem.

#### source
`.variants.*.builder.requirements[].source` _string_

Path of files/directories to copy.

#### script
`.variants.*.builder.script` _string_

Content of a build script to execute. The script may contain its own shebang (`#!`) line to specify the interpreter. If no shebang is provided, `#!/bin/sh` will be used.

Note that the script will not be included in the resulting image. It is written to its own scratch filesystem and only mounted temporarily during execution.

Example

```yaml
builder:
  requirements: [targets]
  script: |
    #!/bin/bash
    for target in $(cat targets); do
      make $target
    done
```

Only one of `script` or `command` may be used for a given builder. Supports environment variables and build arguments.

#### builders
`.variants.*.builders` _array&lt;object&gt;_

Multiple builders to be executed in an explicit order. You can specify any of the predefined standalone builder keys (node, python and php), but each can only appear once. Additionally, any number of custom keys can appear; their definition and subkeys are the same as the standalone builder key.


#### builders[]
`.variants.*.builders[]` _object_

#### custom
`.variants.*.builders[].custom` _object_

Run an arbitrary build command.

#### caches
`.variants.*.builders[].custom.caches` _array&lt;object|string&gt;_

Mount a number of caching filesystems that will persist between builds. Each cache mount should specify the a `destination` path, and optionally the level of `access` (`"shared"` by default) and a unique `id` (`destination` is used by default).

Cache mounts are most useful in speeding up build processes that save downloaded files or compilation state in local directories.

Example

```yaml
runs:
  environment: { GOCACHE: "/var/cache/go" }
builders:
  - custom:
      command: "go build"
      requirements: [.]
      caches:
        - destination: "/var/cache/go"
```

Example (shorthand and using an environment variable)

```yaml
runs:
  environment: { GOCACHE: "/var/cache/go" }
builders:
  - custom:
      command: "go build"
      requirements: [.]
      caches: [ "${GOCACHE}" ]
```


#### caches[]
`.variants.*.builders[].custom.caches[]` _string_

#### caches[]
`.variants.*.builders[].custom.caches[]` _object_

#### access
`.variants.*.builders[].custom.caches[].access` __

Level of access between concurrent build processes. This depends on the underlying build process and what kind of locking it employs.

 - A `shared` cache can be written to concurrently.

 - A `private` cache ensures only one writer at a time and is non-blocking (a new filesystem is created for concurrent writers).

 - A `locked` cache ensures only one writer at a time and will block other build processes until the first releases the lock.

#### destination
`.variants.*.builders[].custom.caches[].destination` _string_

Destination path in the build container where the cache filesystem will be mounted.

Supports environment variables and build arguments.

#### id
`.variants.*.builders[].custom.caches[].id` _string_

A unique ID used to persistent the cache filesystem between builds. The `destination` path is used by default.


#### command
`.variants.*.builders[].custom.command` _array&lt;string&gt;_

Command and arguments of an arbitrary build command, for example `[make, build]`.

Only one of `script` or `command` may be used for a given builder.

#### command[]
`.variants.*.builders[].custom.command[]` _string_

#### command
`.variants.*.builders[].custom.command` _string_

Arbitrary build command, for example `"make build"`.

Only one of `script` or `command` may be used for a given builder.

Supports environment variables and build arguments.

#### mounts
`.variants.*.builders[].custom.mounts` _array&lt;object|string&gt;_

Mount a number of filesystems from either the local build context, other variants or images. Each mount should specify the name of the variant or image (or `local` for the local build context), a `destination` path, and optionally the `source` path within the filesystem to use as the root of the mount. Note that changes to files under source mounts are discarded after each builder command completes.

Mounts are most useful when you need files from some other filesystem for a build process but do not want the files in the resulting image.

Example

```yaml
builders:
  - custom:
      command: "make -C /src/foo install"
      mounts:
        - from: foo
          destination: /src/foo
```

Example (shorthand for mounting another variant, image or `local`)

```yaml
builders:
  - custom:
      command: "make install"
      mounts: [ local ]
```


#### mounts[]
`.variants.*.builders[].custom.mounts[]` _string_

#### mounts[]
`.variants.*.builders[].custom.mounts[]` _object_

#### destination
`.variants.*.builders[].custom.mounts[].destination` _string_

Destination path in the build container where the root of the `from` filesystem will be mounted. Defaults to the working directory.

Supports environment variables and build arguments.

#### from
`.variants.*.builders[].custom.mounts[].from` _string_

Variant or image filesystem to mount. Set to `local` to mount the local build context.

#### source
`.variants.*.builders[].custom.mounts[].source` _string_

Path within the `from` filesystem to use as the root directory of the mount.

Supports environment variables and build arguments.

#### requirements
`.variants.*.builders[].custom.requirements` _array&lt;object|string&gt;_

Path of files/directories to copy from the local build context. This is done before any of the build steps. Note that there are two possible formats for `requirements`. The first is a simple shorthand notation that means copying a list of source files from the local build context to a destination of the same relative path in the image. The second is a longhand form that gives more control over the source context (local or another variant), source and destination paths.

Example (shorthand)

```yaml
builder:
  command: "some build command"
  requirements: [config.json, Makefile, src/] # copy files/directories to the same paths in the image
```

Example (longhand/advanced)

```yaml
builder:
  command: "some build command"
  requirements:
    - from: local
      source: config.production.json
      destination: config.json
    - Makefile # note that longhand/shorthand can be mixed
    - src/
    - from: other-variant
      source: /srv/some/previous/build/product
      destination: dist/product
```


#### requirements[]
`.variants.*.builders[].custom.requirements[]` _string_

#### requirements[]
`.variants.*.builders[].custom.requirements[]` _object_

#### destination
`.variants.*.builders[].custom.requirements[].destination` _string_

Destination path. Defaults to source path.

#### exclude
`.variants.*.builders[].custom.requirements[].exclude` _array&lt;string&gt;_

Exclude files that match any of these patterns.

#### exclude[]
`.variants.*.builders[].custom.requirements[].exclude[]` _string_

A valid glob pattern (e.g. `**/*.swp` or `!**/only-these`).

#### from
`.variants.*.builders[].custom.requirements[].from` _null|string_

Variant from which to copy files. Set to `local` to copy build-context files that match the `source` pattern, or another variant name to copy files that match the `source` pattern from the variant's filesystem.

#### source
`.variants.*.builders[].custom.requirements[].source` _string_

Path of files/directories to copy.

#### script
`.variants.*.builders[].custom.script` _string_

Content of a build script to execute. The script may contain its own shebang (`#!`) line to specify the interpreter. If no shebang is provided, `#!/bin/sh` will be used.

Note that the script will not be included in the resulting image. It is written to its own scratch filesystem and only mounted temporarily during execution.

Example

```yaml
builder:
  requirements: [targets]
  script: |
    #!/bin/bash
    for target in $(cat targets); do
      make $target
    done
```

Only one of `script` or `command` may be used for a given builder. Supports environment variables and build arguments.

#### builders[]
`.variants.*.builders[]` _object_

#### node
`.variants.*.builders[].node` _object_

Configuration related to the NodeJS/NPM environment

#### allow-dedupe-failure
`.variants.*.builders[].node.allow-dedupe-failure` _boolean_

Whether to allow `npm dedupe` to fail; can be used to temporarily unblock CI while conflicts are resolved.

#### env
`.variants.*.builders[].node.env` _string_

Node environment (e.g. production, etc.). Sets the environment variable `NODE_ENV`. Will pass `npm install --production` and run `npm dedupe` if set to production.

#### requirements
`.variants.*.builders[].node.requirements` _array&lt;object|string&gt;_

Path of files/directories to copy from the local build context. This is done before any of the build steps. Note that there are two possible formats for `requirements`. The first is a simple shorthand notation that means copying a list of source files from the local build context to a destination of the same relative path in the image. The second is a longhand form that gives more control over the source context (local or another variant), source and destination paths.

Example (shorthand)

```yaml
builder:
  command: "some build command"
  requirements: [config.json, Makefile, src/] # copy files/directories to the same paths in the image
```

Example (longhand/advanced)

```yaml
builder:
  command: "some build command"
  requirements:
    - from: local
      source: config.production.json
      destination: config.json
    - Makefile # note that longhand/shorthand can be mixed
    - src/
    - from: other-variant
      source: /srv/some/previous/build/product
      destination: dist/product
```


#### requirements[]
`.variants.*.builders[].node.requirements[]` _string_

#### requirements[]
`.variants.*.builders[].node.requirements[]` _object_

#### destination
`.variants.*.builders[].node.requirements[].destination` _string_

Destination path. Defaults to source path.

#### exclude
`.variants.*.builders[].node.requirements[].exclude` _array&lt;string&gt;_

Exclude files that match any of these patterns.

#### exclude[]
`.variants.*.builders[].node.requirements[].exclude[]` _string_

A valid glob pattern (e.g. `**/*.swp` or `!**/only-these`).

#### from
`.variants.*.builders[].node.requirements[].from` _null|string_

Variant from which to copy files. Set to `local` to copy build-context files that match the `source` pattern, or another variant name to copy files that match the `source` pattern from the variant's filesystem.

#### source
`.variants.*.builders[].node.requirements[].source` _string_

Path of files/directories to copy.

#### use-npm-ci
`.variants.*.builders[].node.use-npm-ci` _boolean_

Whether to run `npm ci` instead of `npm install`.

#### builders[]
`.variants.*.builders[]` _object_

#### php
`.variants.*.builders[].php` _object_

#### production
`.variants.*.builders[].php.production` _boolean_

Whether to inject the --no-dev flag into the install command.

#### requirements
`.variants.*.builders[].php.requirements` _array&lt;object|string&gt;_

Path of files/directories to copy from the local build context. This is done before any of the build steps. Note that there are two possible formats for `requirements`. The first is a simple shorthand notation that means copying a list of source files from the local build context to a destination of the same relative path in the image. The second is a longhand form that gives more control over the source context (local or another variant), source and destination paths.

Example (shorthand)

```yaml
builder:
  command: "some build command"
  requirements: [config.json, Makefile, src/] # copy files/directories to the same paths in the image
```

Example (longhand/advanced)

```yaml
builder:
  command: "some build command"
  requirements:
    - from: local
      source: config.production.json
      destination: config.json
    - Makefile # note that longhand/shorthand can be mixed
    - src/
    - from: other-variant
      source: /srv/some/previous/build/product
      destination: dist/product
```


#### requirements[]
`.variants.*.builders[].php.requirements[]` _string_

#### requirements[]
`.variants.*.builders[].php.requirements[]` _object_

#### destination
`.variants.*.builders[].php.requirements[].destination` _string_

Destination path. Defaults to source path.

#### exclude
`.variants.*.builders[].php.requirements[].exclude` _array&lt;string&gt;_

Exclude files that match any of these patterns.

#### exclude[]
`.variants.*.builders[].php.requirements[].exclude[]` _string_

A valid glob pattern (e.g. `**/*.swp` or `!**/only-these`).

#### from
`.variants.*.builders[].php.requirements[].from` _null|string_

Variant from which to copy files. Set to `local` to copy build-context files that match the `source` pattern, or another variant name to copy files that match the `source` pattern from the variant's filesystem.

#### source
`.variants.*.builders[].php.requirements[].source` _string_

Path of files/directories to copy.

#### builders[]
`.variants.*.builders[]` _object_

#### python
`.variants.*.builders[].python` _object_

Predefined configurations for Python build tools

#### no-deps
`.variants.*.builders[].python.no-deps` _boolean_

Inject `--no-deps` into the `pip install` command. All transitive requirements thus must be explicitly listed in the requirements file. `pip check` will be run to verify all dependencies are fulfilled.

#### poetry
`.variants.*.builders[].python.poetry` _object_

Configuration related to installation of pip dependencies using [Poetry](https://python-poetry.org/).

#### devel
`.variants.*.builders[].python.poetry.devel` _boolean_

Whether to install development dependencies or not when using Poetry 1.x.

#### only
`.variants.*.builders[].python.poetry.only` _string_

Poetry 2.x dependency groups to install (e.g. `main` or `main,docs`). Translates to `--only main,docs`. Not compatible with `devel` field.

#### version
`.variants.*.builders[].python.poetry.version` _string_

Version constraint for installing Poetry package.

#### without
`.variants.*.builders[].python.poetry.without` _string_

Poetry 2.x dependency groups to exclude (e.g. `dev` or `dev,test`). Translates to `--without dev,test`. Not compatible with `devel` field.

#### requirements
`.variants.*.builders[].python.requirements` _array&lt;object|string&gt;_

Path of files/directories to copy from the local build context. This is done before any of the build steps. Note that there are two possible formats for `requirements`. The first is a simple shorthand notation that means copying a list of source files from the local build context to a destination of the same relative path in the image. The second is a longhand form that gives more control over the source context (local or another variant), source and destination paths.

Example (shorthand)

```yaml
builder:
  command: "some build command"
  requirements: [config.json, Makefile, src/] # copy files/directories to the same paths in the image
```

Example (longhand/advanced)

```yaml
builder:
  command: "some build command"
  requirements:
    - from: local
      source: config.production.json
      destination: config.json
    - Makefile # note that longhand/shorthand can be mixed
    - src/
    - from: other-variant
      source: /srv/some/previous/build/product
      destination: dist/product
```


#### requirements[]
`.variants.*.builders[].python.requirements[]` _string_

#### requirements[]
`.variants.*.builders[].python.requirements[]` _object_

#### destination
`.variants.*.builders[].python.requirements[].destination` _string_

Destination path. Defaults to source path.

#### exclude
`.variants.*.builders[].python.requirements[].exclude` _array&lt;string&gt;_

Exclude files that match any of these patterns.

#### exclude[]
`.variants.*.builders[].python.requirements[].exclude[]` _string_

A valid glob pattern (e.g. `**/*.swp` or `!**/only-these`).

#### from
`.variants.*.builders[].python.requirements[].from` _null|string_

Variant from which to copy files. Set to `local` to copy build-context files that match the `source` pattern, or another variant name to copy files that match the `source` pattern from the variant's filesystem.

#### source
`.variants.*.builders[].python.requirements[].source` _string_

Path of files/directories to copy.

#### use-system-site-packages
`.variants.*.builders[].python.use-system-site-packages` _boolean_

Whether to inject the --system-site-packages flag into the venv setup command.

#### version
`.variants.*.builders[].python.version` _string_

Python binary present in the system (e.g. python3).

#### copies
`.variants.*.copies` _array&lt;object|string&gt;_


#### copies[]
`.variants.*.copies[]` _string_

Variant from which to copy application and library files. Note that prior to v4, copying of local build-context files was implied by the omission of `copies`. With v4, the configuration must always be explicit. Omitting the field will result in no `COPY` instructions whatsoever, which may be helpful in building very minimal utility images.

#### copies[]
`.variants.*.copies[]` _object_

#### destination
`.variants.*.copies[].destination` _string_

Destination path. Defaults to source path.

#### exclude
`.variants.*.copies[].exclude` _array&lt;string&gt;_

Exclude files that match any of these patterns.

#### exclude[]
`.variants.*.copies[].exclude[]` _string_

A valid glob pattern (e.g. `**/*.swp` or `!**/only-these`).

#### from
`.variants.*.copies[].from` _null|string_

Variant from which to copy files. Set to `local` to copy build-context files that match the `source` pattern, or another variant name to copy files that match the `source` pattern from the variant's filesystem.

#### source
`.variants.*.copies[].source` _string_

Path of files/directories to copy.

#### entrypoint
`.variants.*.entrypoint` _array&lt;string&gt;_

Runtime entry point command and arguments.

#### entrypoint[]
`.variants.*.entrypoint[]` _string_

#### includes
`.variants.*.includes` _array&lt;string&gt;_

Names of other variants to inherit configuration from. Inherited configuration will be combined with this variant's configuration according to key merge rules.

When a Variant configuration overrides the Common configuration the configurations are merged. The way in which configuration is merged depends on whether the type of the configuration is a compound type; e.g., a map or sequence, or a scalar type; e.g., a string or integer.

In general, configuration that is a compound type is appended, whereas configuration that is of a scalar type is overridden.

For example in this Blubberfile:
```yaml
version: v4
base: scratch
apt: { packages: [cowsay] }
variants:
  test:
    base: nodejs
    apt: { packages: [libcaca] }
```

The `base` scalar will be overwritten, whereas the `apt[packages]` sequence will be appended so that both `cowsay` and `libcaca` are installed in the image produced from the `test` Blubberfile variant.


#### includes[]
`.variants.*.includes[]` _string_

Variant name.

#### lives
`.variants.*.lives` _object_

#### as
`.variants.*.lives.as` _string_

Owner (name) of application files within the container.

#### gid
`.variants.*.lives.gid` _integer_

Group owner (GID) of application files within the container.

#### in
`.variants.*.lives.in` _string_

Working directory used when the variant is built and at runtime unless `runs.in` is specified.

#### uid
`.variants.*.lives.uid` _integer_

Owner (UID) of application files within the container.

#### node
`.variants.*.node` _object_

Configuration related to the NodeJS/NPM environment

#### allow-dedupe-failure
`.variants.*.node.allow-dedupe-failure` _boolean_

Whether to allow `npm dedupe` to fail; can be used to temporarily unblock CI while conflicts are resolved.

#### env
`.variants.*.node.env` _string_

Node environment (e.g. production, etc.). Sets the environment variable `NODE_ENV`. Will pass `npm install --production` and run `npm dedupe` if set to production.

#### requirements
`.variants.*.node.requirements` _array&lt;object|string&gt;_

Path of files/directories to copy from the local build context. This is done before any of the build steps. Note that there are two possible formats for `requirements`. The first is a simple shorthand notation that means copying a list of source files from the local build context to a destination of the same relative path in the image. The second is a longhand form that gives more control over the source context (local or another variant), source and destination paths.

Example (shorthand)

```yaml
builder:
  command: "some build command"
  requirements: [config.json, Makefile, src/] # copy files/directories to the same paths in the image
```

Example (longhand/advanced)

```yaml
builder:
  command: "some build command"
  requirements:
    - from: local
      source: config.production.json
      destination: config.json
    - Makefile # note that longhand/shorthand can be mixed
    - src/
    - from: other-variant
      source: /srv/some/previous/build/product
      destination: dist/product
```


#### requirements[]
`.variants.*.node.requirements[]` _string_

#### requirements[]
`.variants.*.node.requirements[]` _object_

#### destination
`.variants.*.node.requirements[].destination` _string_

Destination path. Defaults to source path.

#### exclude
`.variants.*.node.requirements[].exclude` _array&lt;string&gt;_

Exclude files that match any of these patterns.

#### exclude[]
`.variants.*.node.requirements[].exclude[]` _string_

A valid glob pattern (e.g. `**/*.swp` or `!**/only-these`).

#### from
`.variants.*.node.requirements[].from` _null|string_

Variant from which to copy files. Set to `local` to copy build-context files that match the `source` pattern, or another variant name to copy files that match the `source` pattern from the variant's filesystem.

#### source
`.variants.*.node.requirements[].source` _string_

Path of files/directories to copy.

#### use-npm-ci
`.variants.*.node.use-npm-ci` _boolean_

Whether to run `npm ci` instead of `npm install`.

#### php
`.variants.*.php` _object_

#### production
`.variants.*.php.production` _boolean_

Whether to inject the --no-dev flag into the install command.

#### requirements
`.variants.*.php.requirements` _array&lt;object|string&gt;_

Path of files/directories to copy from the local build context. This is done before any of the build steps. Note that there are two possible formats for `requirements`. The first is a simple shorthand notation that means copying a list of source files from the local build context to a destination of the same relative path in the image. The second is a longhand form that gives more control over the source context (local or another variant), source and destination paths.

Example (shorthand)

```yaml
builder:
  command: "some build command"
  requirements: [config.json, Makefile, src/] # copy files/directories to the same paths in the image
```

Example (longhand/advanced)

```yaml
builder:
  command: "some build command"
  requirements:
    - from: local
      source: config.production.json
      destination: config.json
    - Makefile # note that longhand/shorthand can be mixed
    - src/
    - from: other-variant
      source: /srv/some/previous/build/product
      destination: dist/product
```


#### requirements[]
`.variants.*.php.requirements[]` _string_

#### requirements[]
`.variants.*.php.requirements[]` _object_

#### destination
`.variants.*.php.requirements[].destination` _string_

Destination path. Defaults to source path.

#### exclude
`.variants.*.php.requirements[].exclude` _array&lt;string&gt;_

Exclude files that match any of these patterns.

#### exclude[]
`.variants.*.php.requirements[].exclude[]` _string_

A valid glob pattern (e.g. `**/*.swp` or `!**/only-these`).

#### from
`.variants.*.php.requirements[].from` _null|string_

Variant from which to copy files. Set to `local` to copy build-context files that match the `source` pattern, or another variant name to copy files that match the `source` pattern from the variant's filesystem.

#### source
`.variants.*.php.requirements[].source` _string_

Path of files/directories to copy.

#### python
`.variants.*.python` _object_

Predefined configurations for Python build tools

#### no-deps
`.variants.*.python.no-deps` _boolean_

Inject `--no-deps` into the `pip install` command. All transitive requirements thus must be explicitly listed in the requirements file. `pip check` will be run to verify all dependencies are fulfilled.

#### poetry
`.variants.*.python.poetry` _object_

Configuration related to installation of pip dependencies using [Poetry](https://python-poetry.org/).

#### devel
`.variants.*.python.poetry.devel` _boolean_

Whether to install development dependencies or not when using Poetry 1.x.

#### only
`.variants.*.python.poetry.only` _string_

Poetry 2.x dependency groups to install (e.g. `main` or `main,docs`). Translates to `--only main,docs`. Not compatible with `devel` field.

#### version
`.variants.*.python.poetry.version` _string_

Version constraint for installing Poetry package.

#### without
`.variants.*.python.poetry.without` _string_

Poetry 2.x dependency groups to exclude (e.g. `dev` or `dev,test`). Translates to `--without dev,test`. Not compatible with `devel` field.

#### requirements
`.variants.*.python.requirements` _array&lt;object|string&gt;_

Path of files/directories to copy from the local build context. This is done before any of the build steps. Note that there are two possible formats for `requirements`. The first is a simple shorthand notation that means copying a list of source files from the local build context to a destination of the same relative path in the image. The second is a longhand form that gives more control over the source context (local or another variant), source and destination paths.

Example (shorthand)

```yaml
builder:
  command: "some build command"
  requirements: [config.json, Makefile, src/] # copy files/directories to the same paths in the image
```

Example (longhand/advanced)

```yaml
builder:
  command: "some build command"
  requirements:
    - from: local
      source: config.production.json
      destination: config.json
    - Makefile # note that longhand/shorthand can be mixed
    - src/
    - from: other-variant
      source: /srv/some/previous/build/product
      destination: dist/product
```


#### requirements[]
`.variants.*.python.requirements[]` _string_

#### requirements[]
`.variants.*.python.requirements[]` _object_

#### destination
`.variants.*.python.requirements[].destination` _string_

Destination path. Defaults to source path.

#### exclude
`.variants.*.python.requirements[].exclude` _array&lt;string&gt;_

Exclude files that match any of these patterns.

#### exclude[]
`.variants.*.python.requirements[].exclude[]` _string_

A valid glob pattern (e.g. `**/*.swp` or `!**/only-these`).

#### from
`.variants.*.python.requirements[].from` _null|string_

Variant from which to copy files. Set to `local` to copy build-context files that match the `source` pattern, or another variant name to copy files that match the `source` pattern from the variant's filesystem.

#### source
`.variants.*.python.requirements[].source` _string_

Path of files/directories to copy.

#### use-system-site-packages
`.variants.*.python.use-system-site-packages` _boolean_

Whether to inject the --system-site-packages flag into the venv setup command.

#### version
`.variants.*.python.version` _string_

Python binary present in the system (e.g. python3).

#### runs
`.variants.*.runs` _object_

Settings for things run in the container.

#### as
`.variants.*.runs.as` _string_

Runtime process owner (name) of application entrypoint.

#### environment
`.variants.*.runs.environment` _object_

Environment variables and values to be set before entrypoint execution.

#### gid
`.variants.*.runs.gid` _integer_

Runtime process group (GID) of application entrypoint.

#### in
`.variants.*.runs.in` _string_

Runtime working directory. The `lives.in` directory is used by default.

#### insecurely
`.variants.*.runs.insecurely` _boolean_

Skip dropping of privileges to the runtime process owner before entrypoint execution. Production variants should have this set to `false`, but other variants may set it to `true` in some circumstances, for example when enabling [caching for ESLint](https://eslint.org/docs/user-guide/command-line-interface#caching).

#### uid
`.variants.*.runs.uid` _integer_

Runtime process owner (UID) of application entrypoint.
