// Copyright (c) 2018 Uber Technologies, Inc.
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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/fullstorydev/grpcurl"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/uber/prototool/internal/desc"
	"github.com/uber/prototool/internal/extract"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type handler struct {
	logger         *zap.Logger
	callTimeout    time.Duration
	connectTimeout time.Duration
	keepaliveTime  time.Duration
	headers        []string

	getter extract.Getter
}

func newHandler(options ...HandlerOption) *handler {
	handler := &handler{
		logger: zap.NewNop(),
	}
	for _, option := range options {
		option(handler)
	}
	if handler.callTimeout == 0 {
		handler.callTimeout = DefaultCallTimeout
	}
	if handler.connectTimeout == 0 {
		handler.connectTimeout = DefaultConnectTimeout
	}
	// TODO(pedge): composition
	handler.getter = extract.NewGetter(
		extract.GetterWithLogger(handler.logger),
	)
	return handler
}

func (h *handler) Invoke(fileDescriptorSets []*descriptor.FileDescriptorSet, address string, method string, inputReader io.Reader, outputWriter io.Writer) error {
	descriptorSource, err := h.getDescriptorSourceForMethod(fileDescriptorSets, method)
	if err != nil {
		return err
	}
	clientConn, err := h.dial(address)
	if err != nil {
		return err
	}
	defer func() { _ = clientConn.Close() }()
	invocationEventHandler := newInvocationEventHandler(outputWriter, h.logger)
	ctx, cancel := context.WithTimeout(context.Background(), h.callTimeout)
	defer cancel()
	if err := grpcurl.InvokeRpc(
		ctx,
		descriptorSource,
		clientConn,
		method,
		h.headers,
		invocationEventHandler,
		decodeFunc(inputReader),
	); err != nil {
		return err
	}
	return invocationEventHandler.Err()
}

func (h *handler) dial(address string) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), h.connectTimeout)
	defer cancel()
	network, address, err := getNetworkAddress(address)
	if err != nil {
		return nil, fmt.Errorf("grpc dial: %v", err)
	}
	return grpcurl.BlockingDial(ctx, network, address, nil, h.getDialOptions()...)
}

func getNetworkAddress(address string) (string, string, error) {
	split := strings.SplitN(address, "://", 2)
	if len(split) != 2 {
		return "tcp", address, nil
	}
	switch split[0] {
	case "tcp", "unix":
		return split[0], split[1], nil
	default:
		return "", "", fmt.Errorf("invalid network, only tcp or unix allowed: %s", split[0])
	}
}

func (h *handler) getDialOptions() []grpc.DialOption {
	var dialOptions []grpc.DialOption
	if h.keepaliveTime != 0 {
		dialOptions = append(
			dialOptions,
			grpc.WithKeepaliveParams(
				keepalive.ClientParameters{
					Time:    h.keepaliveTime,
					Timeout: h.keepaliveTime,
				},
			),
		)
	}
	return dialOptions
}

func (h *handler) getDescriptorSourceForMethod(fileDescriptorSets []*descriptor.FileDescriptorSet, method string) (grpcurl.DescriptorSource, error) {
	servicePath, err := getServiceForMethod(method)
	if err != nil {
		return nil, err
	}
	service, err := h.getter.GetService(fileDescriptorSets, servicePath)
	if err != nil {
		return nil, err
	}
	fileDescriptorSet, err := desc.SortFileDescriptorSet(service.FileDescriptorSet, service.FileDescriptorProto)
	if err != nil {
		return nil, err
	}
	return grpcurl.DescriptorSourceFromFileDescriptorSet(fileDescriptorSet)
}

func getServiceForMethod(method string) (string, error) {
	split := strings.Split(method, "/")
	if len(split) != 2 {
		return "", fmt.Errorf("invalid gRPC method: %s", method)
	}
	return split[0], nil
}

func decodeFunc(reader io.Reader) func() ([]byte, error) {
	decoder := json.NewDecoder(reader)
	return func() ([]byte, error) {
		var rawMessage json.RawMessage
		if err := decoder.Decode(&rawMessage); err != nil {
			return nil, err
		}
		return rawMessage, nil
	}
}
