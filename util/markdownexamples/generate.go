package main

import (
	"bytes"
	"os"
	"regexp"

	"github.com/pborman/getopt"
	"github.com/raviqqe/gherkin2markdown/convert"
)

const (
	parameters = "some.feature [...]"
)

var (
	showHelp        = getopt.BoolLong("help", 'h', "show help/usage")
	sourceURL       = getopt.StringLong("source-url", 's', "./", "project source base URL")
	workdirLinkRe   = regexp.MustCompile(`"([^"]+)" (as a working directory)`)
	blubberConfigRe = regexp.MustCompile("(?ms)^```\n(version:.+?)\n```$")
)

func main() {
	getopt.SetParameters(parameters)
	getopt.Parse()

	if *showHelp {
		getopt.Usage()
		os.Exit(1)
	}

	args := getopt.Args()

	if len(args) < 1 {
		getopt.Usage()
		os.Exit(1)
	}

	for i, featureFile := range args {
		if i > 0 {
			os.Stdout.Write([]byte("\n"))
		}

		buf := new(bytes.Buffer)
		convert.FeatureFile(featureFile, buf)
		md := buf.Bytes()

		// Link paths in `Given "examples/foo" as a working directory` examples
		md = workdirLinkRe.ReplaceAll(md, []byte(`[$1](`+(*sourceURL)+`$1) $2`))

		// Add yaml syntax highlighting to example blubber config
		md = blubberConfigRe.ReplaceAll(md, []byte("```yaml\n$1\n```"))

		os.Stdout.Write(md)
	}
}
