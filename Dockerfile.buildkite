FROM golang:1.12.4-stretch

ENV GO111MODULE=on \
  GOPATH=
RUN apt-get update -yq && \
  apt-get install -yq build-essential unzip && \
  rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY go.mod go.sum /app/
RUN go mod download
COPY Makefile /app
RUN make deps
COPY . /app/
