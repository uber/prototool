FROM golang:1.12.4-alpine3.9 as builder

RUN apk add --update --no-cache build-base curl git upx && \
  rm -rf /var/cache/apk/*

ENV GOLANG_PROTOBUF_VERSION=1.3.1 \
  GOGO_PROTOBUF_VERSION=1.2.1
RUN GO111MODULE=on go get \
  github.com/golang/protobuf/protoc-gen-go@v${GOLANG_PROTOBUF_VERSION} \
  github.com/gogo/protobuf/protoc-gen-gofast@v${GOGO_PROTOBUF_VERSION} \
  github.com/gogo/protobuf/protoc-gen-gogo@v${GOGO_PROTOBUF_VERSION} \
  github.com/gogo/protobuf/protoc-gen-gogofast@v${GOGO_PROTOBUF_VERSION} \
  github.com/gogo/protobuf/protoc-gen-gogofaster@v${GOGO_PROTOBUF_VERSION} \
  github.com/gogo/protobuf/protoc-gen-gogoslick@v${GOGO_PROTOBUF_VERSION} && \
  mv /go/bin/protoc-gen-go* /usr/local/bin/

ENV GRPC_GATEWAY_VERSION=1.8.5
RUN curl -sSL \
  https://github.com/grpc-ecosystem/grpc-gateway/releases/download/v${GRPC_GATEWAY_VERSION}/protoc-gen-grpc-gateway-v${GRPC_GATEWAY_VERSION}-linux-x86_64 \
  -o /usr/local/bin/protoc-gen-grpc-gateway && \
  curl -sSL \
  https://github.com/grpc-ecosystem/grpc-gateway/releases/download/v${GRPC_GATEWAY_VERSION}/protoc-gen-swagger-v${GRPC_GATEWAY_VERSION}-linux-x86_64 \
  -o /usr/local/bin/protoc-gen-swagger && \
  chmod +x /usr/local/bin/protoc-gen-grpc-gateway && \
  chmod +x /usr/local/bin/protoc-gen-swagger

ENV GRPC_WEB_VERSION=1.0.4
RUN curl -sSL \
  https://github.com/grpc/grpc-web/releases/download/${GRPC_WEB_VERSION}/protoc-gen-grpc-web-${GRPC_WEB_VERSION}-linux-x86_64 \
  -o /usr/local/bin/protoc-gen-grpc-web && \
  chmod +x /usr/local/bin/protoc-gen-grpc-web

ENV YARPC_VERSION=1.37.3
RUN git clone --depth 1 -b v${YARPC_VERSION} https://github.com/yarpc/yarpc-go.git /go/src/go.uber.org/yarpc && \
    cd /go/src/go.uber.org/yarpc && \
    GO111MODULE=on go mod init && \
    GO111MODULE=on go install ./encoding/protobuf/protoc-gen-yarpc-go && \
    mv /go/bin/protoc-gen-yarpc-go /usr/local/bin/

ENV TWIRP_VERSION=5.7.0
RUN curl -sSL \
  https://github.com/twitchtv/twirp/releases/download/v${TWIRP_VERSION}/protoc-gen-twirp-Linux-x86_64 \
  -o /usr/local/bin/protoc-gen-twirp && \
  curl -sSL \
  https://github.com/twitchtv/twirp/releases/download/v${TWIRP_VERSION}/protoc-gen-twirp_python-Linux-x86_64 \
  -o /usr/local/bin/protoc-gen-twirp_python && \
  chmod +x /usr/local/bin/protoc-gen-twirp && \
  chmod +x /usr/local/bin/protoc-gen-twirp_python

ENV PROTOBUF_VERSION=3.6.1
RUN mkdir -p /tmp/protoc && \
  curl -sSL \
  https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOBUF_VERSION}/protoc-${PROTOBUF_VERSION}-linux-x86_64.zip \
  -o /tmp/protoc/protoc.zip && \
  cd /tmp/protoc && \
  unzip protoc.zip && \
  mv /tmp/protoc/include /usr/local/include

RUN mkdir -p /tmp/prototool
COPY go.mod go.sum /tmp/prototool/
RUN cd /tmp/prototool && go mod download
COPY cmd /tmp/prototool/cmd
COPY internal /tmp/prototool/internal
RUN cd /tmp/prototool && \
  go install ./cmd/prototool && \
  mv /go/bin/prototool /usr/local/bin/prototool

RUN upx --lzma /usr/local/bin/*

FROM swift:5.1 AS swift-builder
ENV SWIFT_PROTO_VERSION=1.8.0
RUN set -ex; \
  mkdir -p /tmp/swift-proto; \
  git clone -c advice.detachedHead=false --depth 1 -b ${SWIFT_PROTO_VERSION} \
    https://github.com/apple/swift-protobuf.git /tmp/swift-proto; \
  cd /tmp/swift-proto; \
  swift build -c release; \
  mv .build/release/protoc-gen-swift /usr/local/bin/protoc-gen-swift

FROM scratch AS swift-dist
# This list of files is generated using `ldd /usr/local/bin/protoc-gen-swift`
# Static linking is not yet supported by swiftc https://bugs.swift.org/browse/SR-648
COPY --from=swift-builder /usr/local/bin/protoc-gen-swift /usr/local/bin/protoc-gen-swift
COPY --from=swift-builder /lib/x86_64-linux-gnu/libc.so.6 /lib/x86_64-linux-gnu/libc.so.6
COPY --from=swift-builder /lib/x86_64-linux-gnu/libdl.so.2 /lib/x86_64-linux-gnu/libdl.so.2
COPY --from=swift-builder /lib/x86_64-linux-gnu/libgcc_s.so.1 /lib/x86_64-linux-gnu/libgcc_s.so.1
COPY --from=swift-builder /lib/x86_64-linux-gnu/libm.so.6 /lib/x86_64-linux-gnu/libm.so.6
COPY --from=swift-builder /lib/x86_64-linux-gnu/libpthread.so.0 /lib/x86_64-linux-gnu/libpthread.so.0
COPY --from=swift-builder /lib/x86_64-linux-gnu/librt.so.1 /lib/x86_64-linux-gnu/librt.so.1
COPY --from=swift-builder /lib/x86_64-linux-gnu/libutil.so.1 /lib/x86_64-linux-gnu/libutil.so.1
COPY --from=swift-builder /lib64/ld-linux-x86-64.so.2 /lib64/ld-linux-x86-64.so.2
COPY --from=swift-builder /usr/lib/swift/linux/libBlocksRuntime.so /usr/lib/swift/linux/libBlocksRuntime.so
COPY --from=swift-builder /usr/lib/swift/linux/libdispatch.so /usr/lib/swift/linux/libdispatch.so
COPY --from=swift-builder /usr/lib/swift/linux/libFoundation.so /usr/lib/swift/linux/libFoundation.so
COPY --from=swift-builder /usr/lib/swift/linux/libicudataswift.so.61 /usr/lib/swift/linux/libicudataswift.so.61
COPY --from=swift-builder /usr/lib/swift/linux/libicui18nswift.so.61 /usr/lib/swift/linux/libicui18nswift.so.61
COPY --from=swift-builder /usr/lib/swift/linux/libicuucswift.so.61 /usr/lib/swift/linux/libicuucswift.so.61
COPY --from=swift-builder /usr/lib/swift/linux/libswiftCore.so /usr/lib/swift/linux/libswiftCore.so
COPY --from=swift-builder /usr/lib/swift/linux/libswiftDispatch.so /usr/lib/swift/linux/libswiftDispatch.so
COPY --from=swift-builder /usr/lib/swift/linux/libswiftGlibc.so /usr/lib/swift/linux/libswiftGlibc.so
COPY --from=swift-builder /usr/lib/x86_64-linux-gnu/libatomic.so.1 /usr/lib/x86_64-linux-gnu/libatomic.so.1
COPY --from=swift-builder /usr/lib/x86_64-linux-gnu/libstdc++.so.6 /usr/lib/x86_64-linux-gnu/libstdc++.so.6

FROM alpine:latest

WORKDIR /work

ENV \
  PROTOTOOL_PROTOC_BIN_PATH=/usr/bin/protoc \
  PROTOTOOL_PROTOC_WKT_PATH=/usr/include \
  GRPC_VERSION=1.21.3 \
  PROTOBUF_VERSION=3.9.2 \
  ALPINE_GRPC_VERSION_SUFFIX=r1 \
  ALPINE_PROTOBUF_VERSION_SUFFIX=r0

RUN echo 'http://dl-cdn.alpinelinux.org/alpine/edge/testing' >> /etc/apk/repositories && \
  apk add --update --no-cache bash curl git grpc=${GRPC_VERSION}-${ALPINE_GRPC_VERSION_SUFFIX} protobuf=${PROTOBUF_VERSION}-${ALPINE_PROTOBUF_VERSION_SUFFIX} && \
  rm -rf /var/cache/apk/*

COPY --from=builder /usr/local/bin /usr/local/bin
COPY --from=builder /usr/local/include /usr/include
COPY --from=swift-dist / /

ENV GOGO_PROTOBUF_VERSION=1.2.1 \
  GOLANG_PROTOBUF_VERSION=1.3.1 \
  GRPC_GATEWAY_VERSION=1.8.5 \
  GRPC_WEB_VERSION=1.0.4 \
  TWIRP_VERSION=5.7.0 \
  YARPC_VERSION=1.37.3
