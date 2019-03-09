// Copyright (c) 2019 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package grpc

import (
	"fmt"
	"io"
	"time"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"go.uber.org/zap"
)

const (
	// DefaultCallTimeout is the default call timeout.
	DefaultCallTimeout = 60 * time.Second
	// DefaultConnectTimeout is the default connect timeout.
	DefaultConnectTimeout = 10 * time.Second
)

// Handler handles gRPC calls.
type Handler interface {
	Invoke(fileDescriptorSets []*descriptor.FileDescriptorSet, address string, method string, inputReader io.Reader, outputWriter io.Writer) error
}

// HandlerOption is an option for a new Handler.
type HandlerOption func(*handler)

// HandlerWithLogger returns a HandlerOption that uses the given logger.
//
// The default is to use zap.NewNop().
func HandlerWithLogger(logger *zap.Logger) HandlerOption {
	return func(handler *handler) {
		handler.logger = logger
	}
}

// HandlerWithDetails returns a HandlerOption that outputs responses
// in a structured JSON message that includes headers, trailers, and statuses.
//
// The default is to just print the responses.
func HandlerWithDetails() HandlerOption {
	return func(handler *handler) {
		handler.details = true
	}
}

// HandlerWithCallTimeout returns a HandlerOption that has the given call timeout.
//
// Each invocation must be completed within this time.
//
// The default is to use DefaultCallTimeout.
func HandlerWithCallTimeout(callTimeout time.Duration) HandlerOption {
	return func(handler *handler) {
		handler.callTimeout = callTimeout
	}
}

// HandlerWithConnectTimeout returns a HandlerOption that has the given connect timeout.
//
// The default is to use DefaultConnectTimeout.
func HandlerWithConnectTimeout(connectTimeout time.Duration) HandlerOption {
	return func(handler *handler) {
		handler.connectTimeout = connectTimeout
	}
}

// HandlerWithKeepaliveTime returns a HandlerOption that has the given keepalive time.
//
// The default is to have no keepalive time.
func HandlerWithKeepaliveTime(keepaliveTime time.Duration) HandlerOption {
	return func(handler *handler) {
		handler.keepaliveTime = keepaliveTime
	}
}

// HandlerWithHeader returns a HandlerOption that adds the given key/value header.
func HandlerWithHeader(key string, value string) HandlerOption {
	return func(handler *handler) {
		// this is for grpcurl
		// it takes care of validation
		// if we switch out grpcurl we will probably change to a metadata.MD object
		handler.headers = append(handler.headers, fmt.Sprintf("%s:%s", key, value))
	}
}

// HandlerWithTLS returns a HandlerOption that enables TLS connections to the remote host with the given configuration attributes.
func HandlerWithTLS(insecure bool, cacert string, cert string, key string, serverName string) HandlerOption {
	return func(handler *handler) {
		handler.tls = true
		handler.insecure = insecure
		handler.cacert = cacert
		handler.cert = cert
		handler.key = key
		handler.serverName = serverName
	}
}

// NewHandler returns a new Handler.
func NewHandler(options ...HandlerOption) Handler {
	return newHandler(options...)
}
