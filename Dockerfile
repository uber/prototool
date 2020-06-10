FROM golang:1.14-alpine3.11 as builder

RUN apk add --update --no-cache build-base curl git upx && \
  rm -rf /var/cache/apk/*

ENV GOLANG_PROTOBUF_VERSION=1.4.2 \
  GOGO_PROTOBUF_VERSION=1.3.1
RUN GO111MODULE=on go get \
  github.com/golang/protobuf/protoc-gen-go@v${GOLANG_PROTOBUF_VERSION} \
  github.com/gogo/protobuf/protoc-gen-gofast@v${GOGO_PROTOBUF_VERSION} \
  github.com/gogo/protobuf/protoc-gen-gogo@v${GOGO_PROTOBUF_VERSION} \
  github.com/gogo/protobuf/protoc-gen-gogofast@v${GOGO_PROTOBUF_VERSION} \
  github.com/gogo/protobuf/protoc-gen-gogofaster@v${GOGO_PROTOBUF_VERSION} \
  github.com/gogo/protobuf/protoc-gen-gogoslick@v${GOGO_PROTOBUF_VERSION} && \
  mv /go/bin/protoc-gen-go* /usr/local/bin/

ENV GRPC_GATEWAY_VERSION=1.14.6
RUN curl -sSL \
  https://github.com/grpc-ecosystem/grpc-gateway/releases/download/v${GRPC_GATEWAY_VERSION}/protoc-gen-grpc-gateway-v${GRPC_GATEWAY_VERSION}-linux-x86_64 \
  -o /usr/local/bin/protoc-gen-grpc-gateway && \
  curl -sSL \
  https://github.com/grpc-ecosystem/grpc-gateway/releases/download/v${GRPC_GATEWAY_VERSION}/protoc-gen-swagger-v${GRPC_GATEWAY_VERSION}-linux-x86_64 \
  -o /usr/local/bin/protoc-gen-swagger && \
  chmod +x /usr/local/bin/protoc-gen-grpc-gateway && \
  chmod +x /usr/local/bin/protoc-gen-swagger

ENV GRPC_WEB_VERSION=1.1.0
RUN curl -sSL \
  https://github.com/grpc/grpc-web/releases/download/${GRPC_WEB_VERSION}/protoc-gen-grpc-web-${GRPC_WEB_VERSION}-linux-x86_64 \
  -o /usr/local/bin/protoc-gen-grpc-web && \
  chmod +x /usr/local/bin/protoc-gen-grpc-web

ENV YARPC_VERSION=1.46.0
RUN git clone --depth 1 -b v${YARPC_VERSION} https://github.com/yarpc/yarpc-go.git /go/src/go.uber.org/yarpc && \
    cd /go/src/go.uber.org/yarpc && \
    GO111MODULE=on go install ./encoding/protobuf/protoc-gen-yarpc-go && \
    mv /go/bin/protoc-gen-yarpc-go /usr/local/bin/

ENV TWIRP_VERSION=5.11.0
RUN curl -sSL \
  https://github.com/twitchtv/twirp/releases/download/v${TWIRP_VERSION}/protoc-gen-twirp-Linux-x86_64 \
  -o /usr/local/bin/protoc-gen-twirp && \
  curl -sSL \
  https://github.com/twitchtv/twirp/releases/download/v${TWIRP_VERSION}/protoc-gen-twirp_python-Linux-x86_64 \
  -o /usr/local/bin/protoc-gen-twirp_python && \
  chmod +x /usr/local/bin/protoc-gen-twirp && \
  chmod +x /usr/local/bin/protoc-gen-twirp_python

ENV PROTOBUF_VERSION=3.12.2
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

FROM alpine:latest

WORKDIR /work

ENV \
  PROTOTOOL_PROTOC_BIN_PATH=/usr/bin/protoc \
  PROTOTOOL_PROTOC_WKT_PATH=/usr/include \
  GRPC_VERSION=1.28.1 \
  PROTOBUF_VERSION=3.12.2 \
  ALPINE_GRPC_VERSION_SUFFIX=r1 \
  ALPINE_PROTOBUF_VERSION_SUFFIX=r0

RUN echo 'http://dl-cdn.alpinelinux.org/alpine/edge/testing' >> /etc/apk/repositories && \
  apk add --update --no-cache bash curl git grpc=${GRPC_VERSION}-${ALPINE_GRPC_VERSION_SUFFIX} protobuf=${PROTOBUF_VERSION}-${ALPINE_PROTOBUF_VERSION_SUFFIX} && \
  rm -rf /var/cache/apk/*

COPY --from=builder /usr/local/bin /usr/local/bin
COPY --from=builder /usr/local/include /usr/include

ENV GOGO_PROTOBUF_VERSION=1.3.1 \
  GOLANG_PROTOBUF_VERSION=1.4.2 \
  GRPC_GATEWAY_VERSION=1.14.6 \
  GRPC_WEB_VERSION=1.1.0 \
  TWIRP_VERSION=5.11.0 \
  YARPC_VERSION=1.46.0
