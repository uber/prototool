.DEFAULT_GOAL = all

.PHONY: all
all: build

.PHONY: build
build:
	docker build -t uber/prototool-protoc-issue .

.PHONY: run
run: build
	docker run -it uber/prototool-protoc-issue /protoc --version

.PHONY: shell
shell: build
	docker run -it uber/prototool-protoc-issue
