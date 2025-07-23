SHELL := /bin/bash
RELEASE_DIR ?= ./_release
TARGETS ?= darwin/amd64 linux/amd64 linux/386 linux/arm linux/arm64 linux/ppc64le windows/amd64 plan9/amd64
VERSION = $(shell cat VERSION)
GIT_COMMIT = $(shell git rev-parse --short HEAD)

PACKAGE := gitlab.wikimedia.org/repos/releng/blubber
SOURCE_TREE_URL := https://$(PACKAGE)/-/tree/main/

GO_LIST_GOFILES := '{{range .GoFiles}}{{printf "%s/%s\n" $$.Dir .}}{{end}}{{range .XTestGoFiles}}{{printf "%s/%s\n" $$.Dir .}}{{end}}'
GO_PACKAGES = $(shell go list ./...)

GO_LDFLAGS = \
  -X $(PACKAGE)/meta.Version=$(VERSION) \
  -X $(PACKAGE)/meta.GitCommit=$(GIT_COMMIT)

# go build/install commands
#
GO_BUILD = go build -v -ldflags "$(GO_LDFLAGS)"
GO_INSTALL = go install -v -ldflags "$(GO_LDFLAGS)"
GO_TEST= go test -ldflags "$(GO_LDFLAGS)"

# Respect TARGET* variables defined by docker
# see https://docs.docker.com/engine/reference/builder/#automatic-platform-args-in-the-global-scope
GOOS = $(TARGETOS)
GOARCH = $(TARGETARCH)
export GOOS
export GOARCH

BINARIES = blubber blubber-buildkit

FEATURE_FILES := $(wildcard examples/*.feature)
FEATURE_DOCS := $(patsubst examples/%.feature,examples/%.md,$(FEATURE_FILES))

all: code $(BINARIES)

$(BINARIES): %: cmd/%
	$(GO_BUILD) -o $@ ./$<

.PHONY: code
code:
	go generate $(GO_PACKAGES)

.PHONY: clean
clean:
	go clean $(GO_PACKAGES) || true
	rm -f $(BINARIES) || true

.PHONY: docs
docs: docs/configuration.md
docs/configuration.md: api/config.schema.json
	go run ./util/markdownschema $< > $@

.PHONY: ensure-docs
ensure-docs:
	$(MAKE) -q docs

.PHONY: example-docs
example-docs: $(FEATURE_DOCS)
$(FEATURE_DOCS): examples/%.md: examples/%.feature
	go run ./util/markdownexamples --source-url=$(SOURCE_TREE_URL) $< > $@

.PHONY: install
install: all
	$(GO_INSTALL) $(GO_PACKAGES)

.PHONY: install-tools
install-tools:
	@cat tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % go install %

.PHONY: release
release:
	gox -output="$(RELEASE_DIR)/{{.OS}}-{{.Arch}}/{{.Dir}}" -osarch='$(TARGETS)' -ldflags '$(GO_LDFLAGS)' $(GO_PACKAGES)
	cp LICENSE "$(RELEASE_DIR)"
	for f in "$(RELEASE_DIR)"/*/blubber; do \
		shasum -a 256 "$${f}" | awk '{print $$1}' > "$${f}.sha256"; \
	done

.PHONY: lint
lint:
	@echo > .lint-gofmt.diff
	@go list -f $(GO_LIST_GOFILES) $(GO_PACKAGES) | while read f; do \
		gofmt -e -d "$${f}" >> .lint-gofmt.diff; \
	done
	@test -z "$$(grep '[^[:blank:]]' .lint-gofmt.diff)" || (echo "gofmt found errors:"; cat .lint-gofmt.diff; exit 1)
	golint -set_exit_status $(GO_PACKAGES)
	go vet -composites=false $(GO_PACKAGES)

.PHONY: unit
unit:
	$(GO_TEST) -cover $(GO_PACKAGES)

.PHONY: blubber-buildkit-docker
blubber-buildkit-docker:
	DOCKER_BUILDKIT=1 docker build --pull=false -f .pipeline/blubber.yaml --target buildkit -t localhost/blubber-buildkit .
	@echo Buildkit Docker image built
	@echo It can be used locally in a .pipeline/blubber.yaml with:
	@echo '   # syntax = localhost/blubber-buildkit'

.PHONY: test
test: unit lint

.PHONY: examples
examples: examples.test
	BLUBBER_RUN_EXAMPLES=1 ./examples.test

examples.test:
	$(GO_TEST) -c -o ./$@ -v -timeout 30m ./examples_test.go
