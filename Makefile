SRCS := $(shell find . -name '*.go' | grep -v ^\.\/vendor\/ | grep -v ^\.\/example\/ | grep -v \/gen\/grpcpb\/ | grep -v \/gen\/uber\/proto\/reflect\/ | grep -v ^\/.\/internal\/cmd\/gen-prototool)
PKGS := $(shell go list ./... | grep -v github.com\/uber\/prototool\/example | grep -v \/gen\/grpcpb | grep -v \/gen\/uber\/proto\/reflect\/ | grep -v \/internal\/cmd\/gen-prototool)
BINS := ./cmd/prototool

DOCKER_IMAGE := uber/prototool:latest
DOCKER_RELEASE_IMAGE := golang:1.12.0-stretch

SHELL := /bin/bash -o pipefail
UNAME_OS := $(shell uname -s)
UNAME_ARCH := $(shell uname -m)

TMP_BASE := .tmp
TMP := $(TMP_BASE)/$(UNAME_OS)/$(UNAME_ARCH)
TMP_BIN = $(TMP)/bin
TMP_ETC := $(TMP)/etc
TMP_LIB := $(TMP)/lib

BAZEL_VERSION := 0.23.0
BAZEL_LIB := $(TMP_LIB)/bazel-$(BAZEL_VERSION)
BAZEL := $(abspath $(BAZEL_LIB)/bin/bazel)
ifeq ($(UNAME_OS),Darwin)
BAZEL_OS := darwin
else
BAZEL_OS = linux
endif
BAZEL_ARCH := $(UNAME_ARCH)
$(BAZEL):
	$Q rm -f $(TMP_BIN)/bazel
	$Q mkdir -p $(TMP_BIN)
	$Q rm -rf $(BAZEL_LIB)
	$Q mkdir -p $(BAZEL_LIB)
	$Q curl -SSL https://github.com/bazelbuild/bazel/releases/download/$(BAZEL_VERSION)/bazel-$(BAZEL_VERSION)-installer-$(BAZEL_OS)-$(BAZEL_ARCH).sh -o $(BAZEL_LIB)/bazel-installer.sh
	$Q chmod +x $(BAZEL_LIB)/bazel-installer.sh
	$Q $(BAZEL_LIB)/bazel-installer.sh --base=$(abspath $(BAZEL_LIB)) --bin=$(abspath $(TMP_BIN))

GOLINT_VERSION := 8f45f776aaf18cebc8d65861cc70c33c60471952
GOLINT := $(TMP_BIN)/golint
$(GOLINT):
	$(eval GOLINT_TMP := $(shell mktemp -d))
	@cd $(GOLINT_TMP); go get github.com/golang/lint/golint@$(GOLINT_VERSION)
	@rm -rf $(GOLINT_TMP)

ERRCHECK_VERSION := v1.2.0
ERRCHECK := $(TMP_BIN)/errcheck
$(ERRCHECK):
	$(eval ERRCHECK_TMP := $(shell mktemp -d))
	@cd $(ERRCHECK_TMP); go get github.com/kisielk/errcheck@$(ERRCHECK_VERSION)
	@rm -rf $(ERRCHECK_TMP)

STATICCHECK_VERSION := c2f93a96b099cbbec1de36336ab049ffa620e6d7
STATICCHECK := $(TMP_BIN)/staticcheck
$(STATICCHECK):
	$(eval STATICCHECK_TMP := $(shell mktemp -d))
	@cd $(STATICCHECK_TMP); go get honnef.co/go/tools/cmd/staticcheck@$(STATICCHECK_VERSION)
	@rm -rf $(STATICCHECK_TMP)

UPDATE_LICENSE_VERSION := ce2550dad7144b81ae2f67dc5e55597643f6902b
UPDATE_LICENSE := $(TMP_BIN)/update-license
$(UPDATE_LICENSE):
	$(eval UPDATE_LICENSE_TMP := $(shell mktemp -d))
	@cd $(UPDATE_LICENSE_TMP); go get go.uber.org/tools/update-license@$(UPDATE_LICENSE_VERSION)
	@rm -rf $(UPDATE_LICENSE_TMP)

PROTOC_GEN_GO_VERSION := v1.3.0
PROTOC_GEN_GO := $(TMP_BIN)/protoc-gen-go
$(PROTOC_GEN_GO):
	$(eval PROTOC_GEN_GO_TMP := $(shell mktemp -d))
	@cd $(PROTOC_GEN_GO_TMP); go get github.com/golang/protobuf/protoc-gen-go@$(PROTOC_GEN_GO_VERSION)
	@rm -rf $(PROTOC_GEN_GO_TMP)

unexport GOPATH
export GO111MODULE := on
export GOBIN := $(abspath $(TMP_BIN))
export PATH := $(GOBIN):$(PATH)

.DEFAULT_GOAL := all

.PHONY: all
all: lint cover bazelbuild

.PHONY: install
install:
	go install $(BINS)

.PHONY: license
license: $(UPDATE_LICENSE)
	update-license $(SRCS)

.PHONY: golden
golden: install
	for file in $(shell find internal/cmd/testdata/format -name '*.proto.golden'); do \
		rm -f $${file}; \
	done
	for file in $(shell find internal/cmd/testdata/format -name '*.proto'); do \
		prototool format $${file} > $${file}.golden || true; \
	done
	for file in $(shell find internal/cmd/testdata/format-fix -name '*.proto.golden'); do \
		rm -f $${file}; \
	done
	for file in $(shell find internal/cmd/testdata/format-fix -name '*.proto'); do \
		prototool format --fix $${file} > $${file}.golden || true; \
	done
	for file in $(shell find internal/cmd/testdata/format-fix-v2 -name '*.proto.golden'); do \
		rm -f $${file}; \
	done
	for file in $(shell find internal/cmd/testdata/format-fix-v2 -name '*.proto'); do \
		prototool format --fix $${file} > $${file}.golden || true; \
	done

.PHONY: example
example: install $(PROTOC_GEN_GO)
	@mkdir -p $(TMP_ETC)
	rm -rf example/gen
	prototool all --fix example/proto/uber
	go build ./example/gen/go/uber/foo/v1
	go build ./example/gen/go/uber/bar/v1
	go build -o $(TMP_ETC)/excited ./example/cmd/excited/main.go
	prototool lint etc/style/google
	prototool lint etc/style/uber1

.PHONY: internalgen
internalgen: install $(PROTOC_GEN_GO)
	rm -rf internal/cmd/testdata/grpc/gen
	prototool generate internal/cmd/testdata/grpc
	rm -rf internal/reflect/gen
	prototool all --fix internal/reflect/proto
	rm -f etc/config/example/prototool.yaml
	prototool config init etc/config/example --uncomment

.PHONY: bazelgen
bazelgen: $(BAZEL)
	bash etc/bin/bazelgen.sh

.PHONY: generate
generate: license golden example internalgen bazelgen
	go mod tidy -v

.PHONY: checknodiffgenerated
checknodiffgenerated:
	$(eval CHECKNODIFFGENERATED_PRE := $(shell mktemp -t checknodiffgenerated_pre.XXXXX))
	$(eval CHECKNODIFFGENERATED_POST := $(shell mktemp -t checknodiffgenerated_post.XXXXX))
	$(eval CHECKNODIFFGENERATED_DIFF := $(shell mktemp -t checknodiffgenerated_diff.XXXXX))
	git status --short > $(CHECKNODIFFGENERATED_PRE)
	$(MAKE) generate
	git status --short > $(CHECKNODIFFGENERATED_POST)
	@diff $(CHECKNODIFFGENERATED_PRE) $(CHECKNODIFFGENERATED_POST) > $(CHECKNODIFFGENERATED_DIFF) || true
	@[ ! -s "$(CHECKNODIFFGENERATED_DIFF)" ] || (echo "make generate produced a diff, make sure to check these in:" | cat - $(CHECKNODIFFGENERATED_DIFF) && false)

.PHONY: golint
golint: $(GOLINT)
	for file in $(SRCS); do \
		golint $${file}; \
		if [ -n "$$(golint $${file})" ]; then \
			exit 1; \
		fi; \
	done

.PHONY:
errcheck: $(ERRCHECK)
	errcheck -ignoretests $(PKGS)


.PHONY: staticcheck
staticcheck: $(STATICCHECK)
	staticcheck --tests=false $(PKGS)


.PHONY: checklicense
checklicense: $(UPDATE_LICENSE)
	@echo update-license --dry $(SRCS)
	@if [ -n "$$(update-license --dry $(SRCS))" ]; then \
		echo "These files need to have their license updated by running make license:"; \
		update-license --dry $(SRCS); \
		exit 1; \
	fi

.PHONY: lint
lint: checknodiffgenerated golint errcheck staticcheck checklicense

.PHONY: test
test:
	go test -race $(PKGS)

.PHONY: cover
cover:
	@mkdir -p $(TMP_ETC)
	@rm -f $(TMP_ETC)/coverage.txt $(TMP_ETC)/coverage.html
	go test -race -coverprofile=$(TMP_ETC)/coverage.txt -coverpkg=$(shell echo $(PKGS) | tr ' ' ',') $(PKGS)
	@go tool cover -html=$(TMP_ETC)/coverage.txt -o $(TMP_ETC)/coverage.html
	@echo
	@go tool cover -func=$(TMP_ETC)/coverage.txt | grep total
	@echo
	@echo Open the coverage report:
	@echo open $(TMP_ETC)/coverage.html

.PHONY: codecov
codecov:
	bash <(curl -s https://codecov.io/bash) -c -f $(TMP_ETC)/coverage.txt

.PHONY: bazelbuild
bazelbuild: $(BAZEL)
	bazel build //...:all

.PHONY: releasegen
releasegen: internalgen
	docker run \
		--volume "$(CURDIR):/app" \
		--workdir "/app" \
		$(DOCKER_RELEASE_IMAGE) \
		bash -x etc/bin/releasegen.sh

.PHONY: brewgen
brewgen:
	sh etc/bin/brewgen.sh

.PHONY: releaseinstall
releaseinstall: releasegen releaseclean
	tar -C /usr/local --strip-components 1 -xzf release/prototool-$(shell uname -s)-$(shell uname -m).tar.gz

.PHONY: releaseclean
releaseclean:
	rm -f /usr/local/bin/prototool
	rm -f /usr/local/etc/bash_completion.d/prototool
	rm -f /usr/local/etc/zsh_completion.d/prototool
	rm -f /usr/local/share/man/man1/prototool*

.PHONY: clean
clean:
	git clean -xdf

.PHONY: cleanall
cleanall: clean releaseclean

.PHONY: dockerbuild
dockerbuild:
	docker build -t $(DOCKER_IMAGE) .

.PHONY: dockertest
dockertest:
	docker run -v $(CURDIR):/work $(DOCKER_IMAGE) bash etc/docker/testing/bin/test.sh

.PHONY: dockershell
dockershell: dockerbuild
	docker run -it -v $(CURDIR):/work $(DOCKER_IMAGE) bash

.PHONY: dockerall
dockerall: dockerbuild dockertest
