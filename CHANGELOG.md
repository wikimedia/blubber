
<a name="v0.15.0"></a>
## [v0.15.0](https://gitlab.wikimedia.org/repos/releng/blubber/compare/v0.14.0...v0.15.0)

> 2022-11-10

### Artifacts

* Destination for "local" artifact can be anything
* Add copy dependencies for all artifacts that reference variants


<a name="v0.14.0"></a>
## [v0.14.0](https://gitlab.wikimedia.org/repos/releng/blubber/compare/v0.13.1...v0.14.0)

> 2022-11-04


<a name="v0.13.1"></a>
## [v0.13.1](https://gitlab.wikimedia.org/repos/releng/blubber/compare/v0.13.0...v0.13.1)

> 2022-11-04

### BuildKit

* Specify build platforms based on that of the workers


<a name="v0.13.0"></a>
## [v0.13.0](https://gitlab.wikimedia.org/repos/releng/blubber/compare/v0.12.2...v0.13.0)

> 2022-11-04

### BuildKit

* Refactor multi-platform build process
* Support building for multiple target platforms

### Scripts

* Fix unbound variable in scripts/release.sh
* Fix usage function call in scripts/release.sh
* Avoid pushing directly to the remote branch when releasing
* Fix increment_version to zero the subsequent places
* Provide scripts/release.sh to standardize new releases


<a name="v0.12.2"></a>
## [v0.12.2](https://gitlab.wikimedia.org/repos/releng/blubber/compare/v0.12.1...v0.12.2)

> 2022-10-28


<a name="v0.12.1"></a>
## [v0.12.1](https://gitlab.wikimedia.org/repos/releng/blubber/compare/v0.12.0+gitlab...v0.12.1)

> 2022-10-21

### BuildKit

* disable cache for entrypoints executed on BuildKit

### Gitlab

* Change package name to gitlab.wikimedia.org/repos/releng/blubber


<a name="v0.12.0+gitlab"></a>
## [v0.12.0+gitlab](https://gitlab.wikimedia.org/repos/releng/blubber/compare/v0.12.0...v0.12.0+gitlab)

> 2022-10-19

### BuildKit

* Include given value in ParseExtraOptions error message


<a name="v0.12.0"></a>
## [v0.12.0](https://gitlab.wikimedia.org/repos/releng/blubber/compare/v0.11.1...v0.12.0)

> 2022-10-18

### BuildKit

* allow entrypoint to run in the image building process


<a name="v0.11.1"></a>
## [v0.11.1](https://gitlab.wikimedia.org/repos/releng/blubber/compare/v0.11.0...v0.11.1)

> 2022-10-18

### BuildKit

* Do not require a .dockerignore file


<a name="v0.11.0"></a>
## [v0.11.0](https://gitlab.wikimedia.org/repos/releng/blubber/compare/v0.10.0...v0.11.0)

> 2022-10-14

### BuildKit

* Support builds for specific target platforms


<a name="v0.10.0"></a>
## [v0.10.0](https://gitlab.wikimedia.org/repos/releng/blubber/compare/v0.0.0...v0.10.0)

> 2022-10-12

### APT

* Support user defined APT sources
* Implement merging `apt.proxies` config
* Support configuration of http/https proxies

### BuildKit

* Support target platform in Makefile
* Support Docker's `.dockerignore`
* Support Docker's build-arg options

### Builder

* support cross variant copying for builder.requirements

### Copies

* Allow copying directly from other images

### Feature

* build-time arguments for lives & runs user config

### Macros

* Use numeric gid when creating a user

### Python

* add no-deps flag for pip installation
* install setuptools first when bootstrapping
* Stop using easy_install to bootstrap pip
* ban setuptools==60.9.0 from installing
* Support execution of site package modules in builder
* upgrade pip before installing requirements
* Pin pip package to <21 for Python 2

### Requirements

* Fix regression in short form handling

### User

* Check for existing user/group before creating

### Reverts

* Revert "feature: build-time arguments for lives & runs user config"
* feature: build-time arguments for lives & runs user config


<a name="v0.0.0"></a>
## [v0.0.0](https://gitlab.wikimedia.org/repos/releng/blubber/compare/v0.6.0...v0.0.0)

> 2019-04-16

### Add

* a Blubber file for a Blubberoid service Docker image

### Experimental

* Support buildkit

### Python

* Add support for use-system-flag directive

### PythonConfig

* Change UseSystemFlag to Flag


<a name="v0.6.0"></a>
## [v0.6.0](https://gitlab.wikimedia.org/repos/releng/blubber/compare/v0.5.0...v0.6.0)

> 2018-10-11


<a name="v0.5.0"></a>
## [v0.5.0](https://gitlab.wikimedia.org/repos/releng/blubber/compare/v0.4.0...v0.5.0)

> 2018-08-29


<a name="v0.4.0"></a>
## [v0.4.0](https://gitlab.wikimedia.org/repos/releng/blubber/compare/v0.3.0...v0.4.0)

> 2018-05-24


<a name="v0.3.0"></a>
## [v0.3.0](https://gitlab.wikimedia.org/repos/releng/blubber/compare/v0.2.0...v0.3.0)

> 2018-03-22

### Makefile

* install to global GOPATH with correct -ldflags


<a name="v0.2.0"></a>
## [v0.2.0](https://gitlab.wikimedia.org/repos/releng/blubber/compare/v0.1.0...v0.2.0)

> 2017-11-15


<a name="v0.1.0"></a>
## v0.1.0

> 2017-10-19
