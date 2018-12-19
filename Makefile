SRCS := $(shell find . -name '*.go' | grep -v ^\.\/vendor\/ | grep -v ^\.\/example\/ | grep -v \/gen\/grpcpb\/)
PKGS := $(shell go list ./... | grep -v github.com\/uber\/prototool\/example | grep -v \/gen\/grpcpb)
BINS := github.com/uber/prototool/internal/cmd/prototool

SHELL := /bin/bash -o pipefail
UNAME_OS := $(shell uname -s)
UNAME_ARCH := $(shell uname -m)

TMP_BASE := .tmp
TMP := $(TMP_BASE)/$(UNAME_OS)/$(UNAME_ARCH)
TMP_LIB := $(TMP)/lib
TMP_BIN = $(TMP)/bin

DOCKER_IMAGE := golang:1.11.4

DEP_VERSION := 0.5.0
DEP := $(TMP_BIN)/dep-$(DEP_VERSION)

DEP_LIB := $(TMP_LIB)/dep-$(DEP_VERSION)
ifeq ($(UNAME_OS),Darwin)
DEP_OS := darwin
else
DEP_OS = linux
endif
ifeq ($(UNAME_ARCH),x86_64)
DEP_ARCH := amd64
endif
$(DEP):
	@rm -rf $(DEP_LIB)
	@mkdir -p $(TMP_BIN) $(DEP_LIB)
	curl -sSL "https://github.com/golang/dep/releases/download/v$(DEP_VERSION)/dep-$(DEP_OS)-$(DEP_ARCH)" -o "$(DEP)"
	chmod +x "$(DEP)"

.PHONY: all
all: lint cover

.PHONY: ci
ci: init lint codecov

.PHONY: init
init: $(DEP)
	rm -rf vendor
	$(DEP) ensure -v

.PHONY: vendor
vendor: $(DEP)
	rm -rf vendor
	$(DEP) ensure -update -v

.PHONY: install
install:
	go install \
		-ldflags "-X 'github.com/uber/prototool/internal/vars.GitCommit=$(shell git rev-list -1 HEAD)' -X 'github.com/uber/prototool/internal/vars.BuiltTimestamp=$(shell date -u)'" \
		$(BINS)

.PHONY: license
license:
	@go install ./vendor/go.uber.org/tools/update-license
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
	@go install ./vendor/github.com/gogo/protobuf/protoc-gen-gogoslick
	@go install ./vendor/github.com/golang/protobuf/protoc-gen-go
	@go install ./vendor/github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
	@go install ./vendor/go.uber.org/yarpc/encoding/protobuf/protoc-gen-yarpc-go
	rm -rf example/gen
	prototool all example/idl/uber
	touch ./example/gen/proto/go/foo/.nocover
	touch ./example/gen/proto/go/sub/.nocover
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
	@go install ./vendor/github.com/golang/lint/golint
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
	@go install ./vendor/github.com/kisielk/errcheck
	errcheck -ignoretests $(PKGS)

.PHONY: staticcheck
staticcheck:
	@go install ./vendor/honnef.co/go/tools/cmd/staticcheck
	staticcheck --tests=false $(PKGS)

.PHONY: unused
unused:
	@go install ./vendor/honnef.co/go/tools/cmd/unused
	unused --tests=false $(PKGS)

.PHONY: checklicense
checklicense: install
	@go install ./vendor/go.uber.org/tools/update-license
	@echo update-license --dry $(SRCS)
	@if [ -n "$$(update-license --dry $(SRCS))" ]; then \
		echo "These files need to have their license updated by running make license:"; \
		update-license --dry $(SRCS); \
		exit 1; \
	fi

.PHONY: lint
lint: checknodiffgenerated golint vet errcheck staticcheck unused checklicense

.PHONY: test
test:
	go test -race $(PKGS)

.PHONY: cover
cover:
	@go install ./vendor/golang.org/x/tools/cmd/cover
	@go install ./vendor/github.com/wadey/gocovmerge
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
		--volume "$(CURDIR):/go/src/github.com/uber/prototool" \
		--workdir "/go/src/github.com/uber/prototool" \
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
	go clean -i $(PKGS)
	git clean -xdf --exclude vendor

.PHONY: cleanall
cleanall: clean releaseclean
