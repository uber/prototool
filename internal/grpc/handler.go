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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/fullstorydev/grpcurl"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	protoreflectdesc "github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/uber/prototool/internal/desc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

type handler struct {
	logger         *zap.Logger
	callTimeout    time.Duration
	connectTimeout time.Duration
	keepaliveTime  time.Duration
	headers        []string
	details        bool
	tls            bool
	insecure       bool
	cacert         string
	cert           string
	key            string
	serverName     string
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
	return handler
}

func (h *handler) Invoke(fileDescriptorSets []*descriptor.FileDescriptorSet, address string, method string, inputReader io.Reader, outputWriter io.Writer) error {
	descriptorSource, err := getDescriptorSourceForMethod(fileDescriptorSets, method)
	if err != nil {
		return err
	}
	anyResolver, err := getAnyResolver(fileDescriptorSets)
	if err != nil {
		return err
	}
	clientConn, err := h.dial(address)
	if err != nil {
		return err
	}
	defer func() { _ = clientConn.Close() }()
	invocationEventHandler := newInvocationEventHandler(anyResolver, outputWriter, h.logger, h.details)
	ctx, cancel := context.WithTimeout(context.Background(), h.callTimeout)
	defer cancel()
	if err := grpcurl.InvokeRPC(
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
	creds, err := h.getClientTransportCredentials()
	if err != nil {
		return nil, fmt.Errorf("grpc credentials: %v", err)
	}
	return grpcurl.BlockingDial(ctx, network, address, creds, h.getDialOptions()...)
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

func (h *handler) getClientTransportCredentials() (credentials.TransportCredentials, error) {
	if !h.tls {
		return nil, nil
	}
	creds, err := grpcurl.ClientTransportCredentials(h.insecure, h.cacert, h.cert, h.key)
	if err != nil {
		return nil, err
	}
	if h.serverName != "" {
		if err := creds.OverrideServerName(h.serverName); err != nil {
			return nil, err
		}
	}
	return creds, nil
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

func getAnyResolver(fileDescriptorSets []*descriptor.FileDescriptorSet) (jsonpb.AnyResolver, error) {
	var fileDescriptors []*protoreflectdesc.FileDescriptor
	for _, fileDescriptorSet := range fileDescriptorSets {
		for _, fileDescriptorProto := range fileDescriptorSet.File {
			iFileDescriptorSet, err := desc.SortFileDescriptorSet(fileDescriptorSet, fileDescriptorProto)
			if err != nil {
				return nil, err
			}
			fileDescriptor, err := protoreflectdesc.CreateFileDescriptorFromSet(iFileDescriptorSet)
			if err != nil {
				return nil, err
			}
			fileDescriptors = append(fileDescriptors, fileDescriptor)
		}
	}
	return dynamic.AnyResolver(
		dynamic.NewMessageFactoryWithDefaults(),
		fileDescriptors...,
	), nil
}

func getDescriptorSourceForMethod(fileDescriptorSets []*descriptor.FileDescriptorSet, method string) (grpcurl.DescriptorSource, error) {
	servicePath, err := getServiceForMethod(method)
	if err != nil {
		return nil, err
	}
	serviceInfo, err := getServiceInfo(fileDescriptorSets, servicePath)
	if err != nil {
		return nil, err
	}
	fileDescriptorSet, err := desc.SortFileDescriptorSet(serviceInfo.FileDescriptorSet, serviceInfo.FileDescriptorProto)
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

func decodeFunc(reader io.Reader) func(proto.Message) error {
	decoder := json.NewDecoder(reader)
	return func(message proto.Message) error {
		var rawMessage json.RawMessage
		if err := decoder.Decode(&rawMessage); err != nil {
			return err
		}
		return jsonpb.Unmarshal(bytes.NewReader(rawMessage), message)
	}
}

type serviceInfo struct {
	ServiceDescriptorProto *descriptor.ServiceDescriptorProto
	FileDescriptorProto    *descriptor.FileDescriptorProto
	FileDescriptorSet      *descriptor.FileDescriptorSet
}

func getServiceInfo(fileDescriptorSets []*descriptor.FileDescriptorSet, path string) (*serviceInfo, error) {
	if len(path) == 0 {
		return nil, fmt.Errorf("empty path")
	}
	if path[0] == '.' {
		path = path[1:]
	}
	var serviceDescriptorProto *descriptor.ServiceDescriptorProto
	var fileDescriptorProto *descriptor.FileDescriptorProto
	var fileDescriptorSet *descriptor.FileDescriptorSet
	for _, iFileDescriptorSet := range fileDescriptorSets {
		for _, iFileDescriptorProto := range iFileDescriptorSet.File {
			iServiceDescriptorProto, err := findServiceDescriptorProto(path, iFileDescriptorProto)
			if err != nil {
				return nil, err
			}
			if iServiceDescriptorProto != nil {
				if serviceDescriptorProto != nil {
					return nil, fmt.Errorf("duplicate services for path %s", path)
				}
				serviceDescriptorProto = iServiceDescriptorProto
				fileDescriptorProto = iFileDescriptorProto
			}
		}
		// return first fileDescriptorSet that matches
		// as opposed to duplicate check within fileDescriptorSet, we easily could
		// have multiple fileDescriptorSets that match
		if serviceDescriptorProto != nil {
			fileDescriptorSet = iFileDescriptorSet
			break
		}
	}
	if serviceDescriptorProto == nil {
		return nil, fmt.Errorf("no service for path %s", path)
	}
	return &serviceInfo{
		ServiceDescriptorProto: serviceDescriptorProto,
		FileDescriptorProto:    fileDescriptorProto,
		FileDescriptorSet:      fileDescriptorSet,
	}, nil
}

func findServiceDescriptorProto(path string, fileDescriptorProto *descriptor.FileDescriptorProto) (*descriptor.ServiceDescriptorProto, error) {
	if fileDescriptorProto.GetPackage() == "" {
		return nil, fmt.Errorf("no package on FileDescriptorProto")
	}
	if !strings.HasPrefix(path, fileDescriptorProto.GetPackage()) {
		return nil, nil
	}
	var foundServiceDescriptorProto *descriptor.ServiceDescriptorProto
	for _, serviceDescriptorProto := range fileDescriptorProto.GetService() {
		if fileDescriptorProto.GetPackage()+"."+serviceDescriptorProto.GetName() == path {
			if foundServiceDescriptorProto != nil {
				return nil, fmt.Errorf("duplicate services for path %s", path)
			}
			foundServiceDescriptorProto = serviceDescriptorProto
		}
	}
	return foundServiceDescriptorProto, nil
}
