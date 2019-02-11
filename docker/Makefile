PROTOTOOL_VERSION := $(shell git rev-list -1 HEAD)
DOCKER_IMAGE := uber/prototool:latest

.DEFAULT_GOAL := test

.PHONY: build
build:
	docker build --build-arg PROTOTOOL_VERSION=$(PROTOTOOL_VERSION) -t $(DOCKER_IMAGE) .

.PHONY: test
test: build
	docker run -v $(CURDIR):/work $(DOCKER_IMAGE) bash testing/bin/test.sh

.PHONY: shell
shell: build
	docker run -it -v $(CURDIR):/work $(DOCKER_IMAGE) bash
