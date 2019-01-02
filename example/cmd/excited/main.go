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

package main

import (
	"context"
	"io"
	"log"
	"net"

	foopb "github.com/uber/prototool/example/gen/proto/go/foo"
	"google.golang.org/grpc"
)

func main() {
	if err := do(); err != nil {
		log.Fatal(err.Error())
	}
}

func do() error {
	grpcServer := grpc.NewServer()
	foopb.RegisterExcitedServiceServer(grpcServer, newServer())
	listener, err := net.Listen("tcp", "0.0.0.0:8080")
	if err != nil {
		return err
	}
	return grpcServer.Serve(listener)
}

type server struct{}

func newServer() *server {
	return &server{}
}

func (s *server) Exclamation(ctx context.Context, request *foopb.ExclamationRequest) (*foopb.ExclamationResponse, error) {
	return &foopb.ExclamationResponse{
		Value: request.Value + "!",
	}, nil
}

func (s *server) ExclamationClientStream(streamServer foopb.ExcitedService_ExclamationClientStreamServer) error {
	value := ""
	for request, err := streamServer.Recv(); err != io.EOF; request, err = streamServer.Recv() {
		if err != nil {
			return err
		}
		value += request.Value
	}
	return streamServer.SendAndClose(&foopb.ExclamationResponse{
		Value: value + "!",
	})
}

func (s *server) ExclamationServerStream(request *foopb.ExclamationRequest, streamServer foopb.ExcitedService_ExclamationServerStreamServer) error {
	for _, c := range request.Value {
		if err := streamServer.Send(&foopb.ExclamationResponse{
			Value: string(c),
		}); err != nil {
			return err
		}
	}
	return streamServer.Send(&foopb.ExclamationResponse{
		Value: "!",
	})
}

func (s *server) ExclamationBidiStream(streamServer foopb.ExcitedService_ExclamationBidiStreamServer) error {
	for request, err := streamServer.Recv(); err != io.EOF; request, err = streamServer.Recv() {
		if err != nil {
			return err
		}
		if err := streamServer.Send(&foopb.ExclamationResponse{
			Value: request.Value + "!",
		}); err != nil {
			return err
		}
	}
	return nil
}
