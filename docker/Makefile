PROTOTOOL_VERSION := $(shell git rev-list -1 HEAD)
DOCKER_IMAGE := uber/prototool:$(PROTOTOOL_VERSION)

.PHONY: build
build:
	docker build --build-arg PROTOTOOL_VERSION=$(PROTOTOOL_VERSION) -t $(DOCKER_IMAGE) .

.PHONY: run
run: build
	docker run -it $(DOCKER_IMAGE) /bin/bash
