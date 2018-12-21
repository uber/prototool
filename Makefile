SRCS := $(shell find . -name '*.go' | grep -v ^\.\/vendor\/ | grep -v ^\.\/example\/ | grep -v \/gen\/grpcpb\/)
PKGS := $(shell go list ./... | grep -v github.com\/uber\/prototool\/example | grep -v \/gen\/grpcpb)
BINS := ./cmd/prototool

DOCKER_IMAGE := golang:1.11.4

SHELL := /bin/bash -o pipefail
UNAME_OS := $(shell uname -s)
UNAME_ARCH := $(shell uname -m)

TMP_BASE := .tmp
TMP := $(TMP_BASE)/$(UNAME_OS)/$(UNAME_ARCH)
TMP_LIB := $(TMP)/lib
TMP_BIN = $(TMP)/bin

unexport GOPATH
export GO111MODULE := on
export GOBIN := $(abspath $(TMP_BIN))
export PATH := $(GOBIN):$(PATH)

.PHONY: all
all: lint cover

.PHONY: ci
ci: lint codecov

.PHONY: init
init:
	go mod download

.PHONY: vendor
vendor:
	go mod tidy -v

.PHONY: install
install:
	go install $(BINS)

.PHONY: license
license:
	@go install go.uber.org/tools/update-license
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

.PHONY: example
example: install
	@go install github.com/golang/protobuf/protoc-gen-go
	rm -rf example/gen
	prototool all example/idl/uber
	go build ./example/gen/proto/go/foo
	go build ./example/gen/proto/go/sub
	go build ./example/cmd/excited/main.go
	prototool lint etc/style

.PHONY: internalgen
internalgen: install
	prototool generate internal/cmd/testdata/grpc
	rm -f etc/config/example/prototool.yaml
	prototool config init etc/config/example --uncomment

.PHONY: generate
generate: license golden example internalgen

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
golint:
	@go install github.com/golang/lint/golint
	for file in $(SRCS); do \
		golint $${file}; \
		if [ -n "$$(golint $${file})" ]; then \
			exit 1; \
		fi; \
	done

.PHONY: vet
vet:
	go vet $(PKGS)

.PHONY:
errcheck:
	@go install github.com/kisielk/errcheck
	errcheck -ignoretests $(PKGS)

.PHONY: staticcheck
staticcheck:
	@go install honnef.co/go/tools/cmd/staticcheck
	staticcheck --tests=false $(PKGS)

.PHONY: unused
unused:
	@go install honnef.co/go/tools/cmd/unused
	unused --tests=false $(PKGS)

.PHONY: checklicense
checklicense: install
	@go install go.uber.org/tools/update-license
	@echo update-license --dry $(SRCS)
	@if [ -n "$$(update-license --dry $(SRCS))" ]; then \
		echo "These files need to have their license updated by running make license:"; \
		update-license --dry $(SRCS); \
		exit 1; \
	fi

.PHONY: lint
# TODO: re-add errcheck staticcheck unused when they are fixed for Golang modules
#lint: checknodiffgenerated golint vet errcheck staticcheck unused checklicense
lint: checknodiffgenerated golint vet checklicense

.PHONY: test
test:
	go test -race $(PKGS)

.PHONY: cover
cover:
	@go install golang.org/x/tools/cmd/cover
	@go install github.com/wadey/gocovmerge
	./etc/bin/cover.sh $(PKGS)
	go tool cover -html=coverage.txt -o cover.html
	go tool cover -func=coverage.txt | grep total

.PHONY: codecov
codecov: SHELL := /bin/bash
codecov: cover
	bash <(curl -s https://codecov.io/bash) -c -f coverage.txt

.PHONY: releasegen
releasegen: internalgen
	docker run \
		--volume "$(CURDIR):/app" \
		--workdir "/app" \
		$(DOCKER_IMAGE) \
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

.PHONY: dockerall
dockerall:
	docker run \
		--volume "$(CURDIR):/app" \
		--workdir "/app" \
		$(DOCKER_IMAGE) \
		make all
