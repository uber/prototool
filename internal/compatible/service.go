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

type services map[string]*service

var _ descriptorProtoGroup = (services)(nil)

func (ss services) Items() map[string]descriptorProto {
	items := make(map[string]descriptorProto)
	for i, s := range ss {
		items[i] = s
	}
	return items
}

// service represents a *descriptor.ServiceDescriptorProto.
type service struct {
	path    location.Path
	name    string
	methods methods
}

var _ descriptorProto = (*service)(nil)

func (s *service) Name() string        { return s.name }
func (s *service) Path() location.Path { return s.path }
func (s *service) Type() string        { return fmt.Sprintf("Service %q", s.name) }

func newService(sd *descriptor.ServiceDescriptorProto, p location.Path) *service {
	methods := make(methods, len(sd.GetMethod()))
	for i, m := range sd.GetMethod() {
		methods[m.GetName()] = newMethod(m, p.Scope(location.Method, i))
	}
	return &service{
		path:    p,
		name:    sd.GetName(),
		methods: methods,
	}
}

// checkServices verifies that,
//   - None of the services were removed.
//   - None of the services' methods were inappropriately updated.
func (c *fileChecker) checkServices(original, updated services) {
	c.checkRemovedItems(original, updated, location.Name)
	for i, us := range updated {
		if os, ok := original[i]; ok {
			c.checkMethods(os.methods, us.methods)
		}
	}
}
