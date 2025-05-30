Feature: Installing packages
  Often variants will need to install additional software to satisfy build or
  runtime dependencies. You can have APT packages installed using the `apt`
  directives.

  Background:
    Given "examples/hello-world-c" as a working directory

  @set1
  Scenario: Install additional build dependencies
    Given this "blubber.yaml"
      """
      version: v4
      variants:
        build:
          base: debian:bullseye
          apt:
            packages:
              - gcc
              - libc6-dev
      """
    When you build the "build" variant
    Then the image will have the following files in "/usr/bin"
      | gcc |

  @set2
  Scenario: Install from additional APT sources
    Given this "blubber.yaml"
      """
      version: v4
      variants:
        build:
          base: docker-registry.wikimedia.org/golang1.21:1.21-1-20240609
          apt:
            sources:
              - url: https://apt.wikimedia.org/wikimedia
                distribution: bullseye-wikimedia
                components:
                  - thirdparty/amd-rocm54
            packages:
              bullseye-wikimedia: # you may use an explicit distribution/release name like so
                - fake-libgcc-7-dev
      """
    When you build the "build" variant
    Then the image will have the following files in "/usr/share/doc/fake-libgcc-7-dev"
      | copyright     |

  @set3
  Scenario: Provide a public key for an additional APT source
    Given this "blubber.yaml"
      """
      version: v4
      base: docker-registry.wikimedia.org/bookworm:20250601
      variants:
        build:
          apt:
            sources:
              - url: https://packages.microsoft.com/debian/12/prod
                distribution: bookworm
                components: [main]
                signed-by: |
                  -----BEGIN PGP PUBLIC KEY BLOCK-----
                  Version: GnuPG v1.4.7 (GNU/Linux)

                  mQENBFYxWIwBCADAKoZhZlJxGNGWzqV+1OG1xiQeoowKhssGAKvd+buXCGISZJwT
                  LXZqIcIiLP7pqdcZWtE9bSc7yBY2MalDp9Liu0KekywQ6VVX1T72NPf5Ev6x6DLV
                  7aVWsCzUAF+eb7DC9fPuFLEdxmOEYoPjzrQ7cCnSV4JQxAqhU4T6OjbvRazGl3ag
                  OeizPXmRljMtUUttHQZnRhtlzkmwIrUivbfFPD+fEoHJ1+uIdfOzZX8/oKHKLe2j
                  H632kvsNzJFlROVvGLYAk2WRcLu+RjjggixhwiB+Mu/A8Tf4V6b+YppS44q8EvVr
                  M+QvY7LNSOffSO6Slsy9oisGTdfE39nC7pVRABEBAAG0N01pY3Jvc29mdCAoUmVs
                  ZWFzZSBzaWduaW5nKSA8Z3Bnc2VjdXJpdHlAbWljcm9zb2Z0LmNvbT6JATUEEwEC
                  AB8FAlYxWIwCGwMGCwkIBwMCBBUCCAMDFgIBAh4BAheAAAoJEOs+lK2+EinPGpsH
                  /32vKy29Hg51H9dfFJMx0/a/F+5vKeCeVqimvyTM04C+XENNuSbYZ3eRPHGHFLqe
                  MNGxsfb7C7ZxEeW7J/vSzRgHxm7ZvESisUYRFq2sgkJ+HFERNrqfci45bdhmrUsy
                  7SWw9ybxdFOkuQoyKD3tBmiGfONQMlBaOMWdAsic965rvJsd5zYaZZFI1UwTkFXV
                  KJt3bp3Ngn1vEYXwijGTa+FXz6GLHueJwF0I7ug34DgUkAFvAs8Hacr2DRYxL5RJ
                  XdNgj4Jd2/g6T9InmWT0hASljur+dJnzNiNCkbn9KbX7J/qK1IbR8y560yRmFsU+
                  NdCFTW7wY0Fb1fWJ+/KTsC4=
                  =J6gs
                  -----END PGP PUBLIC KEY BLOCK-----
            packages: [dotnet-sdk-8.0]
          entrypoint: [dotnet, --version]
      """
    When you build and run the "build" variant
    Then the entrypoint will have run successfully
