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
	"encoding/json"
	"io"

	"github.com/fullstorydev/grpcurl"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/desc"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var _ grpcurl.InvocationEventHandler = &invocationEventHandler{}

type invocationEventHandler struct {
	jsonpbMarshaler *jsonpb.Marshaler
	output          io.Writer
	logger          *zap.Logger
	details         bool
	err             error
}

func newInvocationEventHandler(anyResolver jsonpb.AnyResolver, output io.Writer, logger *zap.Logger, details bool) *invocationEventHandler {
	return &invocationEventHandler{
		jsonpbMarshaler: &jsonpb.Marshaler{
			AnyResolver: anyResolver,
		},
		output:  output,
		logger:  logger,
		details: details,
	}
}

func (i *invocationEventHandler) OnResolveMethod(*desc.MethodDescriptor) {}

func (i *invocationEventHandler) OnSendHeaders(metadata.MD) {}

func (i *invocationEventHandler) OnReceiveHeaders(headers metadata.MD) {
	if !i.details {
		return
	}
	i.printMetadata(headers, "headers")
}

func (i *invocationEventHandler) OnReceiveResponse(message proto.Message) {
	if !i.details {
		i.printProtoMessage(message, "")
		return
	}
	i.printProtoMessage(message, "response")
}

func (i *invocationEventHandler) OnReceiveTrailers(s *status.Status, trailers metadata.MD) {
	if err := s.Err(); err != nil {
		i.err = err
	}
	if !i.details {
		return
	}
	i.printProtoMessage(s.Proto(), "status")
	i.printMetadata(trailers, "trailers")
}

func (i *invocationEventHandler) Err() error {
	return i.err
}

func (i *invocationEventHandler) printProtoMessage(input proto.Message, detailsKey string) {
	if input == nil {
		return
	}
	s, err := i.jsonpbMarshaler.MarshalToString(input)
	if err != nil {
		i.logger.Error("marshal error", zap.Error(err))
		return
	}
	i.println(i.marshalSanitize(s, detailsKey))
}

func (i *invocationEventHandler) printMetadata(input metadata.MD, detailsKey string) {
	if len(input) == 0 {
		return
	}
	data, err := json.Marshal(input)
	if err != nil {
		i.logger.Error("marshal error", zap.Error(err))
		return
	}
	i.println(i.marshalSanitize(string(data), detailsKey))
}

func (i *invocationEventHandler) marshalSanitize(s string, detailsKey string) string {
	if s == "{}" {
		s = ""
	}
	if i.details && detailsKey != "" && s != "" {
		return `{"` + detailsKey + `":` + s + `}`
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
