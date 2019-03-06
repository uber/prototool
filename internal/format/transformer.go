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

package format

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/emicklei/proto"
	"github.com/uber/prototool/internal/text"
	"go.uber.org/zap"
)

type transformer struct {
	logger            *zap.Logger
	fix               int
	fileHeader        string
	javaPackagePrefix string
}

func newTransformer(options ...TransformerOption) *transformer {
	transformer := &transformer{
		logger: zap.NewNop(),
		fix:    FixNone,
	}
	for _, option := range options {
		option(transformer)
	}
	return transformer
}

func (t *transformer) Transform(filename string, data []byte) ([]byte, []*text.Failure, error) {
	if err := checkFix(t.fix); err != nil {
		return nil, nil, err
	}
	descriptor, err := proto.NewParser(bytes.NewReader(data)).Parse()
	if err != nil {
		return nil, nil, err
	}
	descriptor.Filename = filename

	firstPassVisitor := newFirstPassVisitor(filename, t.fix, t.fileHeader, t.javaPackagePrefix)
	for _, element := range descriptor.Elements {
		element.Accept(firstPassVisitor)
	}
	failures := firstPassVisitor.Do()
	buffer := bytes.NewBuffer(nil)
	buffer.Write(firstPassVisitor.Bytes())

	syntaxVersion := 2
	if firstPassVisitor.Syntax != nil && firstPassVisitor.Syntax.Value != "" {
		switch firstPassVisitor.Syntax.Value {
		case "proto2":
			// nothing
		case "proto3":
			syntaxVersion = 3
		default:
			return nil, nil, fmt.Errorf("unknown syntax: %s", firstPassVisitor.Syntax.Value)
		}
	}

	mainVisitor := newMainVisitor(syntaxVersion == 2)
	for _, element := range descriptor.Elements {
		element.Accept(mainVisitor)
	}
	failures = append(failures, mainVisitor.Do()...)
	buffer.Write(mainVisitor.Bytes())

	text.SortFailures(failures)

	// TODO: expensive
	s := strings.TrimSpace(buffer.String())
	if len(s) > 0 {
		return []byte(s + "\n"), failures, nil
	}
	return nil, failures, nil
}

func checkFix(fix int) error {
	switch fix {
	case FixNone, FixV1, FixV2:
		return nil
	default:
		return fmt.Errorf("unknown format fix value: %d", fix)
	}
}
