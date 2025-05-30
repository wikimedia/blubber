package config

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"

	"gitlab.wikimedia.org/repos/releng/blubber/build"
)

// AptConfig represents configuration pertaining to package installation from
// existing APT sources.
type AptConfig struct {
	// Packages keys are the name of the targeted release, or 'default' to
	// specify no target and use the base image's target release,
	// Packages values are a list of the desired packages
	Packages AptPackages `json:"packages" validate:"dive,keys,omitempty,debianrelease,endkeys,dive,debianpackage"`

	// Proxies provides APT HTTP/HTTPS proxies to use for certain sources
	Proxies []AptProxy `json:"proxies" validate:"dive,omitempty"`

	// Sources provides APT sources for packages
	Sources []AptSource `json:"sources" validate:"dive,omitempty"`
}

const (
	// AptDefaultTargetKeyword defines a special keyword indicating that the
	// packages to be installed should use the default target release
	AptDefaultTargetKeyword = "default"

	// AptSourceConfigurationPath is the file that source configuration will be
	// written to for each defined source.
	AptSourceConfigurationPath = "/etc/apt/sources.list.d/99blubber.list"

	// AptKeyringDir is the directory where [AptSource.SignedBy] key data will
	// be written.
	AptKeyringDir = "/etc/apt/keyrings"

	// AptProxyConfigurationPath is the file that configuration will be written
	// to for each defined proxy.
	AptProxyConfigurationPath = "/etc/apt/apt.conf.d/99blubber-proxies"

	// AptFileMode is the default file mode of APT configuration files.
	AptFileMode = os.FileMode(0o644)
)

// Merge takes another AptConfig and combines the packages declared within
// with the packages of this AptConfig.
func (apt *AptConfig) Merge(apt2 AptConfig) {

	if apt2.Packages != nil {
		if apt.Packages == nil {
			apt.Packages = make(map[string][]string)
		}

		for key, pkgs := range apt2.Packages {
			apt.Packages[key] = append(apt.Packages[key], pkgs...)
		}
	}

	if apt2.Proxies != nil {
		apt.Proxies = append(apt.Proxies, apt2.Proxies...)
	}

	if apt2.Sources != nil {
		apt.Sources = append(apt.Sources, apt2.Sources...)
	}
}

// InstructionsForPhase injects build instructions that will install the
// declared packages during the privileged phase.
//
// # PhasePrivileged
//
// Updates the APT cache, installs configured packages, and cleans up.
func (apt AptConfig) InstructionsForPhase(phase build.Phase) []build.Instruction {
	ins := []build.Instruction{}

	if len(apt.Packages) > 0 || len(apt.Sources) > 0 || len(apt.Proxies) > 0 {
		switch phase {
		case build.PhasePrivileged:
			var (
				runAll  []build.Run
				targets []string
			)

			ins = append(ins, build.Env{map[string]string{
				"DEBIAN_FRONTEND": "noninteractive",
			}})

			// Configure proxies
			var proxies []string
			for _, proxy := range apt.Proxies {
				proxies = append(proxies, proxy.Configuration())
			}

			if len(proxies) > 0 {
				ins = append(ins, build.File{
					Path:    AptProxyConfigurationPath,
					Content: []byte(strings.Join(proxies, "\n") + "\n"),
					Mode:    os.FileMode(AptFileMode),
				})
			}

			// Configure sources
			var sources []string
			for _, source := range apt.Sources {
				sources = append(sources, source.Configuration())

				if source.SignedBy != "" {
					ins = append(ins, build.File{
						Path:    source.KeyringPath(),
						Content: []byte(source.SignedBy),
						Mode:    os.FileMode(AptFileMode),
					})
				}
			}

			if len(sources) > 0 {
				// If we're configuring any additional sources, install
				// ca-certificates first to ensure successful fetching of third-party
				// package lists over https
				ins = append(ins,
					build.RunAll{[]build.Run{
						{"apt-get update", []string{}},
						{"apt-get install -y", []string{"ca-certificates"}},
					}},
					build.File{
						Path:    AptSourceConfigurationPath,
						Content: []byte(strings.Join(sources, "\n") + "\n"),
						Mode:    os.FileMode(AptFileMode),
					},
				)
			}

			// order the targets for the same result each run
			for target := range apt.Packages {
				targets = append(targets, target)
			}

			if len(targets) > 0 {
				runAll = append(runAll, build.Run{"apt-get update", []string{}})
			}

			sort.Strings(targets)

			for _, target := range targets {
				if target == AptDefaultTargetKeyword {
					runAll = append(runAll, build.Run{"apt-get install -y", apt.Packages[target]})
				} else {
					args := append([]string{target}, apt.Packages[target]...)
					runAll = append(runAll, build.Run{"apt-get install -y -t", args})
				}
			}

			runAll = append(runAll, build.Run{"rm -rf /var/lib/apt/lists/*", []string{}})

			if len(apt.Proxies) > 0 {
				runAll = append(runAll, build.Run{"rm -f", []string{AptProxyConfigurationPath}})
			}

			ins = append(ins, build.RunAll{runAll})
		}
	}

	return ins
}

// AptPackages represents lists of packages to install. Each entry is keyed by
// the release that should be targetted during installation, i.e. `apt-get
// install -t release package`.
type AptPackages map[string][]string

// UnmarshalJSON implements json.Unmarshaler to handle both shorthand and
// longhand apt packages configuration.
//
// Shorthand packages configuration: ["package1", "package2"]
// Longhand packages configuration: { "release1": ["package1, package2"], "release2": ["package3"]}
func (ap *AptPackages) UnmarshalJSON(unmarshal []byte) error {
	(*ap) = make(AptPackages)

	longhand := make(map[string][]string)
	err := json.Unmarshal(unmarshal, &longhand)

	if err == nil {
		for key, pkgs := range longhand {
			(*ap)[key] = append((*ap)[key], pkgs...)
		}
		return nil
	}

	shorthand := []string{}
	err = json.Unmarshal(unmarshal, &shorthand)

	if err == nil {
		// Input was entirely in short form
		(*ap)[AptDefaultTargetKeyword] = shorthand
		return nil
	}

	return err
}

// AptProxy represents an APT proxy to use for a specific or all sources.
type AptProxy struct {
	// URL of the proxy, e.g. "http://webproxy.example:8080"
	URL string `json:"url" validate:"required,httpurl"`

	// Source is a URL representing the APT source, e.g.
	// "http://security.debian.org/". If none is given
	Source string `json:"source" validate:"omitempty,httpurl"`
}

// UnmarshalJSON implements json.Unmarshaler to handle both shorthand and
// longhand apt proxies configuration.
//
// Shorthand: ["http://proxy.example:8080"]
// Longhand: [
//
//	{
//	  "url": "http://proxy.example:8080",
//	  "source": "http://security.debian.org"
//	}
//
// ]
func (ap *AptProxy) UnmarshalJSON(unmarshal []byte) error {
	err := json.Unmarshal(unmarshal, &(*ap).URL)

	if err != nil {
		if !IsUnmarshalTypeError(err) {
			return err
		}

		// shorthand failed. try longhand
		mock := struct {
			URL    string `json:"url"`
			Source string `json:"source"`
		}{}

		err = json.Unmarshal(unmarshal, &mock)

		if err != nil {
			return err
		}

		*ap = AptProxy(mock)
	}

	return nil
}

// Configuration returns the APT configuration for this proxy.
func (ap AptProxy) Configuration() string {
	var schemeURL string

	if len(ap.Source) > 0 {
		schemeURL = ap.Source
	} else {
		schemeURL = ap.URL
	}

	surl, _ := url.Parse(schemeURL)

	cfg := "Acquire::" + surl.Scheme + "::Proxy"

	if len(ap.Source) > 0 {
		cfg += "::" + surl.Host
	}

	cfg += ` "` + ap.URL + `";`

	return cfg
}

// AptSource represents an APT source to set up prior to package installation.
type AptSource struct {
	// URL of the APT source, e.g. "http://apt.wikimedia.org"
	URL string `json:"url" validate:"required,httpurl"`

	// Distribution is the Debian distribution/release name (e.g. buster)
	Distribution string `json:"distribution" validate:"required,debianrelease"`

	// Components is a list of the source components to index (e.g. main, contrib)
	Components []string `json:"components" validate:"dive,omitempty,debiancomponent"`

	// SignedBy is an encoded set of public keys used to verify the source
	SignedBy string `json:"signed-by" validate:"omitempty"`
}

// Configuration returns the APT list configuration for this source.
func (as AptSource) Configuration() string {
	cfg := "deb"

	if as.SignedBy != "" {
		cfg += " [signed-by=" + as.KeyringPath() + "]"

	}

	return cfg + " " + strings.Join(append([]string{as.URL, as.Distribution}, as.Components...), " ")
}

// KeyringPath returns a unique filename for the [SignedBy] key(s).
func (as AptSource) KeyringPath() string {
	sha := sha256.New()
	sha.Write([]byte(as.SignedBy))
	return path.Join(AptKeyringDir, hex.EncodeToString(sha.Sum(nil))+".asc")
}
