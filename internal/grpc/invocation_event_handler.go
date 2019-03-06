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

var jsonpbMarshaler = &jsonpb.Marshaler{}
var jsonpbPrettyMarshaler = &jsonpb.Marshaler{Indent: "  "}

var _ grpcurl.InvocationEventHandler = &invocationEventHandler{}

type invocationEventHandler struct {
	output io.Writer
	logger *zap.Logger
	json   bool
	err    error
}

func newInvocationEventHandler(output io.Writer, logger *zap.Logger, json bool) *invocationEventHandler {
	return &invocationEventHandler{
		output: output,
		logger: logger,
		json:   json,
	}
}

func (i *invocationEventHandler) OnResolveMethod(*desc.MethodDescriptor) {}

func (i *invocationEventHandler) OnSendHeaders(metadata.MD) {}

func (i *invocationEventHandler) OnReceiveHeaders(headers metadata.MD) {
	if !i.json {
		return
	}
	if len(headers) != 0 {
		if s := i.marshalJSON(headers); s != "" {
			i.println(`{"headers":` + s + "}")
		}
	}
}

func (i *invocationEventHandler) OnReceiveResponse(message proto.Message) {
	if !i.json {
		i.println(i.marshal(message, true))
		return
	}
	if s := i.marshal(message, false); s != "" && s != "{}" {
		i.println(`{"response":` + s + `}`)
	}
}

func (i *invocationEventHandler) OnReceiveTrailers(s *status.Status, trailers metadata.MD) {
	if err := s.Err(); err != nil {
		i.err = err
	}
	if !i.json {
		return
	}
	ss := i.marshal(s.Proto(), false)
	if ss == "{}" {
		ss = ""
	}
	if ss != "" {
		ss = `"status":` + ss
	}
	var st string
	if len(trailers) != 0 {
		if st = i.marshalJSON(trailers); st != "" {
			st = `"trailers":` + st
			if ss != "" {
				st = `,` + st
			}
		}
	}
	if ss != "" || st != "" {
		i.println(`{` + ss + st + `}`)
	}
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

func (i *invocationEventHandler) marshalJSON(input interface{}) string {
	if input == nil {
		return ""
	}
	data, err := json.Marshal(input)
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
