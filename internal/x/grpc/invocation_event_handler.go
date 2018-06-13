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
	"encoding/json"
	"io"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/desc"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var jsonpbMarshaler = &jsonpb.Marshaler{}
var jsonpbPrettyMarshaler = &jsonpb.Marshaler{Indent: "  "}

type invocationEventHandler struct {
	output     io.Writer
	logger     *zap.Logger
	jsonOutput bool
	err        error
}

func newInvocationEventHandler(output io.Writer, logger *zap.Logger, jsonOutput bool) *invocationEventHandler {
	return &invocationEventHandler{
		output:     output,
		logger:     logger,
		jsonOutput: jsonOutput,
	}
}

func (i *invocationEventHandler) OnResolveMethod(*desc.MethodDescriptor) {}

func (i *invocationEventHandler) OnSendHeaders(metadata.MD) {}

func (i *invocationEventHandler) OnReceiveHeaders(headers metadata.MD) {
	if !i.jsonOutput {
		return
	}
	i.println(i.marshalJSONOutputData(&jsonOutputData{Headers: headers}))
}

func (i *invocationEventHandler) OnReceiveResponse(message proto.Message) {
	if !i.jsonOutput {
		i.println(i.marshal(message, true))
		return
	}
	i.println(i.marshalJSONOutputData(&jsonOutputData{Response: i.marshal(message, false)}))
}

func (i *invocationEventHandler) OnReceiveTrailers(s *status.Status, trailers metadata.MD) {
	if err := s.Err(); err != nil {
		// TODO(pedge): not great for streaming
		i.err = err
	}
	if !i.jsonOutput {
		return
	}
	i.println(i.marshalJSONOutputData(&jsonOutputData{Status: i.marshal(s.Proto(), false), Trailers: trailers}))
}

func (i *invocationEventHandler) Err() error {
	return i.err
}

func (i *invocationEventHandler) marshal(message proto.Message, pretty bool) string {
	if message == nil {
		return ""
	}
	jsonMarshaler := jsonpbMarshaler
	if pretty {
		jsonMarshaler = jsonpbPrettyMarshaler
	}
	s, err := jsonMarshaler.MarshalToString(message)
	if err != nil {
		i.logger.Error("marshal error", zap.Error(err))
		return ""
	}
	return s
}

func (i *invocationEventHandler) marshalJSONOutputData(jsonOutputData *jsonOutputData) string {
	if jsonOutputData == nil {
		return ""
	}
	data, err := json.Marshal(jsonOutputData)
	if err != nil {
		i.logger.Error("marshal error", zap.Error(err))
		return ""
	}
	return string(data)
}

func (i *invocationEventHandler) println(s string) {
	if s == "" {
		return
	}
	if _, err := i.output.Write([]byte(s + "\n")); err != nil {
		i.logger.Error("write error", zap.Error(err))
	}
}

type jsonOutputData struct {
	Response string              `json:"response,omitempty"`
	Headers  map[string][]string `json:"headers,omitempty"`
	Trailers map[string][]string `json:"trailers,omitempty"`
	Status   string              `json:"status,omitempty"`
}
