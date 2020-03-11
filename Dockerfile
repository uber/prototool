ARG GOLANG_PROTOBUF_VERSION=1.3.1
ARG GOGO_PROTOBUF_VERSION=1.2.1
ARG GRPC_GATEWAY_VERSION=1.8.5
ARG GRPC_WEB_VERSION=1.0.4
ARG YARPC_VERSION=1.37.3
ARG TWIRP_VERSION=5.7.0
ARG SWIFT_PROTO_VERSION=1.8.0
ARG GRPC_VERSION=1.25.0
ARG ALPINE_GRPC_VERSION_SUFFIX=r1
ARG PROTOBUF_VERSION=3.11.2
ARG ALPINE_PROTOBUF_VERSION_SUFFIX=r1

# Base image contains all the protoc plugins included with prototool
FROM golang:1.13.8-alpine3.11 as base

RUN apk add --update --no-cache build-base curl git upx

ARG GOLANG_PROTOBUF_VERSION
ARG GOGO_PROTOBUF_VERSION
RUN GO111MODULE=on go get \
  github.com/golang/protobuf/protoc-gen-go@v${GOLANG_PROTOBUF_VERSION} \
  github.com/gogo/protobuf/protoc-gen-gofast@v${GOGO_PROTOBUF_VERSION} \
  github.com/gogo/protobuf/protoc-gen-gogo@v${GOGO_PROTOBUF_VERSION} \
  github.com/gogo/protobuf/protoc-gen-gogofast@v${GOGO_PROTOBUF_VERSION} \
  github.com/gogo/protobuf/protoc-gen-gogofaster@v${GOGO_PROTOBUF_VERSION} \
  github.com/gogo/protobuf/protoc-gen-gogoslick@v${GOGO_PROTOBUF_VERSION} && \
  mv /go/bin/protoc-gen-go* /usr/local/bin/

ARG GRPC_GATEWAY_VERSION
RUN curl -sSL \
  https://github.com/grpc-ecosystem/grpc-gateway/releases/download/v${GRPC_GATEWAY_VERSION}/protoc-gen-grpc-gateway-v${GRPC_GATEWAY_VERSION}-linux-x86_64 \
  -o /usr/local/bin/protoc-gen-grpc-gateway && \
  curl -sSL \
  https://github.com/grpc-ecosystem/grpc-gateway/releases/download/v${GRPC_GATEWAY_VERSION}/protoc-gen-swagger-v${GRPC_GATEWAY_VERSION}-linux-x86_64 \
  -o /usr/local/bin/protoc-gen-swagger && \
  chmod +x /usr/local/bin/protoc-gen-grpc-gateway && \
  chmod +x /usr/local/bin/protoc-gen-swagger

ARG GRPC_WEB_VERSION
RUN curl -sSL \
  https://github.com/grpc/grpc-web/releases/download/${GRPC_WEB_VERSION}/protoc-gen-grpc-web-${GRPC_WEB_VERSION}-linux-x86_64 \
  -o /usr/local/bin/protoc-gen-grpc-web && \
  chmod +x /usr/local/bin/protoc-gen-grpc-web

ARG YARPC_VERSION
RUN git clone --depth 1 -b v${YARPC_VERSION} https://github.com/yarpc/yarpc-go.git /go/src/go.uber.org/yarpc && \
  cd /go/src/go.uber.org/yarpc && \
  GO111MODULE=on go mod init && \
  GO111MODULE=on go install ./encoding/protobuf/protoc-gen-yarpc-go && \
  mv /go/bin/protoc-gen-yarpc-go /usr/local/bin/

ARG TWIRP_VERSION
RUN curl -sSL \
  https://github.com/twitchtv/twirp/releases/download/v${TWIRP_VERSION}/protoc-gen-twirp-Linux-x86_64 \
  -o /usr/local/bin/protoc-gen-twirp && \
  curl -sSL \
  https://github.com/twitchtv/twirp/releases/download/v${TWIRP_VERSION}/protoc-gen-twirp_python-Linux-x86_64 \
  -o /usr/local/bin/protoc-gen-twirp_python && \
  chmod +x /usr/local/bin/protoc-gen-twirp && \
  chmod +x /usr/local/bin/protoc-gen-twirp_python

ARG PROTOBUF_VERSION
RUN mkdir -p /tmp/protoc && \
  curl -sSL \
  https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOBUF_VERSION}/protoc-${PROTOBUF_VERSION}-linux-x86_64.zip \
  -o /tmp/protoc/protoc.zip && \
  cd /tmp/protoc && \
  unzip protoc.zip && \
  mv /tmp/protoc/include /usr/local/include

# Swift image must be built separately
FROM swift:5.1 AS swift-builder
ARG SWIFT_PROTO_VERSION
WORKDIR /swift-proto
RUN git clone -c advice.detachedHead=false --depth 1 -b ${SWIFT_PROTO_VERSION} https://github.com/apple/swift-protobuf.git .
RUN swift build --configuration release
# Copy release binary to distribution folder
WORKDIR /dist
RUN install -TD /swift-proto/.build/release/protoc-gen-swift ./usr/local/bin/protoc-gen-swift
# Find all dependencies and copy them to distribution
RUN ldd /dist/usr/local/bin/protoc-gen-swift | awk '{ print $3 }' | sort | uniq | xargs -I {} install -TD {} .{}
# Explicitely copy dynamic linker since it is filtered out by awk
RUN install -TD /lib64/ld-linux-x86-64.so.2 ./lib64/ld-linux-x86-64.so.2

# Build & install prototool
FROM base AS builder
WORKDIR /tmp/prototool
COPY go.mod go.sum ./
RUN  go mod download
COPY cmd cmd
COPY internal internal
RUN go install -mod=readonly ./cmd/prototool
RUN mv /go/bin/prototool /usr/local/bin/prototool
RUN upx --lzma /usr/local/bin/*

# Final image containing only binaries
FROM alpine:3.11

WORKDIR /work

ARG GOLANG_PROTOBUF_VERSION
ARG GOGO_PROTOBUF_VERSION
ARG GRPC_GATEWAY_VERSION
ARG GRPC_WEB_VERSION
ARG YARPC_VERSION
ARG TWIRP_VERSION
ARG PROTOBUF_VERSION
ARG SWIFT_PROTO_VERSION
ARG GRPC_VERSION
ARG ALPINE_GRPC_VERSION_SUFFIX
ARG PROTOBUF_VERSION
ARG ALPINE_PROTOBUF_VERSION_SUFFIX

ENV \
  GOLANG_PROTOBUF_VERSION=${GOLANG_PROTOBUF_VERSION} \
  GOGO_PROTOBUF_VERSION=${GOGO_PROTOBUF_VERSION} \
  GRPC_GATEWAY_VERSION=${GRPC_GATEWAY_VERSION} \
  GRPC_WEB_VERSION=${GRPC_WEB_VERSION} \
  YARPC_VERSION=${YARPC_VERSION} \
  TWIRP_VERSION=${TWIRP_VERSION} \
  PROTOBUF_VERSION=${PROTOBUF_VERSION} \
  SWIFT_PROTO_VERSION=${SWIFT_PROTO_VERSION} \
  GRPC_VERSION=${GRPC_VERSION} \
  PROTOBUF_VERSION=${PROTOBUF_VERSION} \
  PROTOTOOL_PROTOC_BIN_PATH=/usr/bin/protoc \
  PROTOTOOL_PROTOC_WKT_PATH=/usr/include

RUN apk add --update --no-cache bash curl git \
  grpc=${GRPC_VERSION}-${ALPINE_GRPC_VERSION_SUFFIX} \
  protobuf=${PROTOBUF_VERSION}-${ALPINE_PROTOBUF_VERSION_SUFFIX}

COPY --from=builder /usr/local/bin /usr/local/bin
COPY --from=builder /usr/local/include /usr/include
COPY --from=swift-builder /dist /
