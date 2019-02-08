FROM golang:1.11.5-alpine3.9 as builder

RUN apk add --no-cache \
  autoconf \
  automake \
  build-base \
  curl \
  libtool \
  git \
  unzip

RUN mkdir -p /builder/bin /tmp/bin
ENV PATH /tmp/bin:/builder/bin:${PATH}

ENV GRPC_VERSION 1.18.0
RUN git clone --depth 1 -b v${GRPC_VERSION} https://github.com/grpc/grpc.git /tmp/grpc
RUN cd /tmp/grpc && git submodule update --init
# https://github.com/grpc/grpc/issues/17781
RUN cd /tmp/grpc && make CFLAGS='-g -O2 -w' CXXFLAGS='-g -O2 -w' -j8 plugins
RUN cp /tmp/grpc/bins/opt/grpc_cpp_plugin /builder/bin/
RUN cp /tmp/grpc/bins/opt/grpc_csharp_plugin /builder/bin/
RUN cp /tmp/grpc/bins/opt/grpc_node_plugin /builder/bin/
RUN cp /tmp/grpc/bins/opt/grpc_objective_c_plugin /builder/bin/
RUN cp /tmp/grpc/bins/opt/grpc_php_plugin /builder/bin/
RUN cp /tmp/grpc/bins/opt/grpc_python_plugin /builder/bin/
RUN cp /tmp/grpc/bins/opt/grpc_ruby_plugin /builder/bin/

ENV GLIDE_VERSION 0.13.2
RUN mkdir -p /tmp/glide
RUN curl -sSL \
  https://github.com/Masterminds/glide/releases/download/v${GLIDE_VERSION}/glide-v${GLIDE_VERSION}-linux-amd64.tar.gz \
  -o /tmp/glide/glide.tar.gz
RUN cd /tmp/glide && tar xvzf glide.tar.gz
RUN cp /tmp/glide/linux-amd64/glide /tmp/bin/glide

ENV DEP_VERSION 0.5.0
RUN curl -sSL https://github.com/golang/dep/releases/download/v${DEP_VERSION}/dep-linux-amd64 -o /tmp/bin/dep
RUN chmod +x /tmp/bin/dep

ENV GOLANG_PROTOBUF_VERSION 1.2.0
RUN git clone --depth 1 -b v${GOLANG_PROTOBUF_VERSION} https://github.com/golang/protobuf.git /go/src/github.com/golang/protobuf
RUN cd /go/src/github.com/golang/protobuf && go get -d -v ./... && \
  go install ./protoc-gen-go
RUN cp /go/bin/protoc-gen-go /builder/bin/

ENV GOGO_PROTOBUF_VERSION 1.2.0
RUN git clone --depth 1 -b v${GOGO_PROTOBUF_VERSION} https://github.com/gogo/protobuf.git /go/src/github.com/gogo/protobuf
RUN cd /go/src/github.com/gogo/protobuf && go get -d -v ./... && \
  go install \
    ./protoc-gen-gofast \
    ./protoc-gen-gogo \
    ./protoc-gen-gogofast \
    ./protoc-gen-gogofaster \
    ./protoc-gen-gogoslick
RUN cp /go/bin/protoc-gen-gofast /builder/bin/
RUN cp /go/bin/protoc-gen-gogo /builder/bin/
RUN cp /go/bin/protoc-gen-gogofast /builder/bin/
RUN cp /go/bin/protoc-gen-gogofaster /builder/bin/
RUN cp /go/bin/protoc-gen-gogoslick /builder/bin/

ENV GRPC_GATEWAY_VERSION 1.7.0
RUN curl -sSL \
  https://github.com/grpc-ecosystem/grpc-gateway/releases/download/v${GRPC_GATEWAY_VERSION}/protoc-gen-grpc-gateway-v${GRPC_GATEWAY_VERSION}-linux-x86_64 \
  -o /builder/bin/protoc-gen-grpc-gateway
RUN chmod +x /builder/bin/protoc-gen-grpc-gateway
RUN curl -sSL \
  https://github.com/grpc-ecosystem/grpc-gateway/releases/download/v${GRPC_GATEWAY_VERSION}/protoc-gen-swagger-v${GRPC_GATEWAY_VERSION}-linux-x86_64 \
  -o /builder/bin/protoc-gen-swagger
RUN chmod +x /builder/bin/protoc-gen-swagger

ENV GRPC_WEB_VERSION 1.0.3
RUN curl -sSL \
  https://github.com/grpc/grpc-web/releases/download/v${GRPC_WEB_VERSION}/protoc-gen-grpc-web-v${GRPC_WEB_VERSION}-linux-x86_64 \
  -o /builder/bin/protoc-gen-grpc-web
RUN chmod +x /builder/bin/protoc-gen-grpc-web

ENV YARPC_VERSION 1.36.1
RUN git clone --depth 1 -b v${YARPC_VERSION} https://github.com/yarpc/yarpc-go.git /go/src/go.uber.org/yarpc
RUN cd /go/src/go.uber.org/yarpc && glide install && \
  go install ./encoding/protobuf/protoc-gen-yarpc-go
RUN cp /go/bin/protoc-gen-yarpc-go /builder/bin/

ENV TWIRP_VERSION 5.5.1
RUN git clone --depth 1 -b v${TWIRP_VERSION} https://github.com/twitchtv/twirp.git /go/src/github.com/twitchtv/twirp
RUN cd /go/src/github.com/twitchtv/twirp && dep ensure -v && \
  go install \
    ./protoc-gen-twirp \
    ./protoc-gen-twirp_python
RUN cp /go/bin/protoc-gen-twirp /builder/bin/
RUN cp /go/bin/protoc-gen-twirp_python /builder/bin/

ENV PROTOC_VERSION 3.6.1
RUN mkdir -p /tmp/protoc
RUN curl -sSL \
  https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-x86_64.zip -o /tmp/protoc/protoc.zip
RUN cd /tmp/protoc && unzip protoc.zip
RUN cp /tmp/protoc/bin/protoc /builder/bin/
RUN cp -R /tmp/protoc/include /builder/bin/include

ARG PROTOTOOL_VERSION
ENV PROTOTOOL_VERSION ${PROTOTOOL_VERSION}
RUN git clone https://github.com/uber/prototool.git /tmp/prototool
RUN cd /tmp/prototool && git checkout ${PROTOTOOL_VERSION}
RUN cd /tmp/prototool && go install ./cmd/prototool
RUN cp /go/bin/prototool /builder/bin/prototool

FROM alpine:3.9

RUN apk add --no-cache libstdc++ libc6-compat
COPY --from=builder /builder /app
ENV PATH /app/bin:${PATH}
ENV PROTOTOOL_PROTOC_BIN_PATH /app/bin/protoc
ENV PROTOTOOL_PROTOC_WKT_PATH /app/include
