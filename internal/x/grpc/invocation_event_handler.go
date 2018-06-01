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
	"fmt"
	"io"

	"github.com/fullstorydev/grpcurl"
	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/desc"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type invocationEventHandler struct {
	output io.Writer
	logger *zap.Logger

	res *response
}

type response struct {
	Headers []string      `json:"headers,omitempty"`
	Body    proto.Message `json:"body,omitempty"`
	Err     error         `json:"error,omitempty"`
}

func newInvocationEventHandler(output io.Writer, logger *zap.Logger) *invocationEventHandler {
	return &invocationEventHandler{
		output: output,
		logger: logger,
		res:    &response{},
	}
}

var _ grpcurl.InvocationEventHandler = (*invocationEventHandler)(nil)

func (i *invocationEventHandler) OnResolveMethod(*desc.MethodDescriptor) {}
func (i *invocationEventHandler) OnSendHeaders(metadata.MD)              {}

func (i *invocationEventHandler) OnReceiveHeaders(m metadata.MD) {
	for k, v := range m {
		i.res.Headers = append(i.res.Headers, fmt.Sprintf("%s:%s", k, v))
	}
}

func (i *invocationEventHandler) OnReceiveResponse(message proto.Message) {
	i.res.Body = message
	i.println()
}

func (i *invocationEventHandler) OnReceiveTrailers(s *status.Status, _ metadata.MD) {
	if err := s.Err(); err != nil {
		// TODO(pedge): not great for streaming
		i.res.Err = err
	}
}

func (i *invocationEventHandler) Err() error {
	return i.res.Err
}

func (i *invocationEventHandler) println() {
	bytes, err := json.MarshalIndent(i.res, "", "  ")
	if err != nil {
		i.logger.Error("marhsal error", zap.Error(err))
	}
	if _, err := i.output.Write(bytes); err != nil {
		i.logger.Error("write error", zap.Error(err))
	}
}
