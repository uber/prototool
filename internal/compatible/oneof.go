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

package compatible

import (
	"fmt"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/uber/prototool/internal/location"
)

type oneofs map[string]*oneof

var _ descriptorProtoGroup = (oneofs)(nil)

func (os oneofs) Items() map[string]descriptorProto {
	items := make(map[string]descriptorProto)
	for i, o := range os {
		items[i] = o
	}
	return items
}

// oneof represents a *descriptor.OneofDescriptorProto.
type oneof struct {
	path location.Path
	name string
}

var _ descriptorProto = (*oneof)(nil)

func (o *oneof) Name() string        { return o.name }
func (o *oneof) Path() location.Path { return o.path }
func (o *oneof) Type() string        { return fmt.Sprintf("Oneof %q", o.name) }

func newOneof(od *descriptor.OneofDescriptorProto, p location.Path) *oneof {
	return &oneof{
		path: p,
		name: od.GetName(),
	}
}

// checkOneofs verifies that none of the oneof declarations were removed.
func (c *fileChecker) checkOneofs(original, updated oneofs) {
	c.checkRemovedItems(original, updated, location.Name)
}
