// Copyright (c) 2021 Uber Technologies, Inc.
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

	foov1 "github.com/uber/prototool/example/gen/go/uber/foo/v1"
	"google.golang.org/grpc"
)

func main() {
	if err := do(); err != nil {
		log.Fatal(err.Error())
	}
}

func do() error {
	grpcServer := grpc.NewServer()
	foov1.RegisterExcitedAPIServer(grpcServer, newServer())
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

func (s *server) Exclamation(ctx context.Context, request *foov1.ExclamationRequest) (*foov1.ExclamationResponse, error) {
	return &foov1.ExclamationResponse{
		Value: request.Value + "!",
	}, nil
}

func (s *server) ExclamationClientStream(streamServer foov1.ExcitedAPI_ExclamationClientStreamServer) error {
	value := ""
	for request, err := streamServer.Recv(); err != io.EOF; request, err = streamServer.Recv() {
		if err != nil {
			return err
		}
		value += request.Value
	}
	return streamServer.SendAndClose(&foov1.ExclamationResponse{
		Value: value + "!",
	})
}

func (s *server) ExclamationServerStream(request *foov1.ExclamationRequest, streamServer foov1.ExcitedAPI_ExclamationServerStreamServer) error {
	for _, c := range request.Value {
		if err := streamServer.Send(&foov1.ExclamationResponse{
			Value: string(c),
		}); err != nil {
			return err
		}
	}
	return streamServer.Send(&foov1.ExclamationResponse{
		Value: "!",
	})
}

func (s *server) ExclamationBidiStream(streamServer foov1.ExcitedAPI_ExclamationBidiStreamServer) error {
	for request, err := streamServer.Recv(); err != io.EOF; request, err = streamServer.Recv() {
		if err != nil {
			return err
		}
		if err := streamServer.Send(&foov1.ExclamationResponse{
			Value: request.Value + "!",
		}); err != nil {
			return err
		}
	}
	return nil
}
