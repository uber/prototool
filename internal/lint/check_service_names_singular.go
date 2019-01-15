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
	"strings"

	"github.com/emicklei/proto"
	"github.com/gobuffalo/flect"
	"github.com/uber/prototool/internal/file"
	"github.com/uber/prototool/internal/strs"
	"github.com/uber/prototool/internal/text"
)

var serviceNamesNoPluralsLinter = NewSuppressableLinter(
	"SERVICE_NAMES_NO_PLURALS",
	`Verifies that all CamelCase service names do not contain plural components.`,
	"plurals",
	checkServiceNamesNoPlurals,
)

var allowedServiceNamePlurals = map[string]struct{}{
	"data": struct{}{},
}

func checkServiceNamesNoPlurals(add func(*file.ProtoSet, *proto.Comment, *text.Failure), dirPath string, descriptors []*FileDescriptor) error {
	return runVisitor(&serviceNamesNoPluralsVisitor{baseAddSuppressableVisitor: newBaseAddSuppressableVisitor(add)}, descriptors)
}

type serviceNamesNoPluralsVisitor struct {
	*baseAddSuppressableVisitor
}

func (v *serviceNamesNoPluralsVisitor) VisitService(service *proto.Service) {
	for _, word := range strs.SplitCamelCaseWord(service.Name) {
		if singular := flect.New(word).Singularize().String(); singular != word {
			if _, ok := allowedServiceNamePlurals[word]; !ok {
				v.AddFailuref(service.Comment, service.Position, `Service name %q contains plural component %q, consider using %q instead.`, service.Name, strings.Title(word), strings.Title(singular))
			}
		}
	}
}
