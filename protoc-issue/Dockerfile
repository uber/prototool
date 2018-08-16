FROM alpine:3.8

RUN \
  mkdir /tmp/protoc && \
  wget -O /tmp/protoc/protoc.zip https://github.com/google/protobuf/releases/download/v3.6.1/protoc-3.6.1-linux-x86_64.zip && \
  cd /tmp/protoc && \
  unzip protoc.zip && \
  mv bin/protoc /protoc && \
  cd / && \
  rm -rf /tmp/protoc
