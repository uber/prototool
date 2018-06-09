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
	"io"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/desc"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var jsonpbMarshaler = &jsonpb.Marshaler{Indent: "  "}

type invocationEventHandler struct {
	output io.Writer
	logger *zap.Logger
	err    error
}

func newInvocationEventHandler(output io.Writer, logger *zap.Logger) *invocationEventHandler {
	return &invocationEventHandler{
		output: output,
		logger: logger,
	}
}

func (i *invocationEventHandler) OnResolveMethod(*desc.MethodDescriptor) {}

func (i *invocationEventHandler) OnSendHeaders(metadata.MD) {}

func (i *invocationEventHandler) OnReceiveHeaders(metadata.MD) {}

func (i *invocationEventHandler) OnReceiveResponse(message proto.Message) {
	i.println(i.marshal(message))
}

func (i *invocationEventHandler) OnReceiveTrailers(s *status.Status, _ metadata.MD) {
	if err := s.Err(); err != nil {
		// TODO(pedge): not great for streaming
		i.err = err
		// printed by returning the error in handler
		//i.println(err.Error())
	}
}

func (i *invocationEventHandler) Err() error {
	return i.err
}

func (i *invocationEventHandler) marshal(message proto.Message) string {
	s, err := jsonpbMarshaler.MarshalToString(message)
	if err != nil {
		i.logger.Error("marshal error", zap.Error(err))
		return ""
	}
	return s
}

func (i *invocationEventHandler) println(s string) {
	if s == "" {
		return
	}
	if _, err := i.output.Write([]byte(s + "\n")); err != nil {
		i.logger.Error("write error", zap.Error(err))
	}
}
