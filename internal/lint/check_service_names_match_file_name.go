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

package lint

import (
	"path/filepath"

	"github.com/emicklei/proto"
	"github.com/uber/prototool/internal/strs"
	"github.com/uber/prototool/internal/text"
)

var serviceNamesMatchFileNameLinter = NewLinter(
	"SERVICE_NAMES_MATCH_FILE_NAME",
	"Verifies that there is one service per file and the file name is service_name_lower_snake_case.proto.",
	checkServiceNamesMatchFileName,
)

func checkServiceNamesMatchFileName(add func(*text.Failure), dirPath string, descriptors []*FileDescriptor) error {
	return runVisitor(&serviceNamesMatchFileNameVisitor{baseAddVisitor: newBaseAddVisitor(add)}, descriptors)
}

type serviceNamesMatchFileNameVisitor struct {
	baseAddVisitor
	filename string
	services []*proto.Service
}

func (v *serviceNamesMatchFileNameVisitor) OnStart(descriptor *FileDescriptor) error {
	v.filename = descriptor.Filename
	v.services = nil
	return nil
}

func (v *serviceNamesMatchFileNameVisitor) VisitService(service *proto.Service) {
	v.services = append(v.services, service)
}

func (v *serviceNamesMatchFileNameVisitor) Finally() error {
	if len(v.services) == 0 {
		return nil
	}
	if len(v.services) > 1 {
		for _, service := range v.services {
			v.AddFailuref(service.Position, `Multiple services defined in this file and there should be only one service per file.`)
		}
		return nil
	}
	service := v.services[0]
	filename := filepath.Base(v.filename)
	expectedFilename := strs.ToLowerSnakeCase(service.Name) + ".proto"
	if filename != expectedFilename {
		v.AddFailuref(service.Position, `Expected filename to be %q for file containing service %q but was %q.`, expectedFilename, service.Name, filename)
	}
	return nil
}
