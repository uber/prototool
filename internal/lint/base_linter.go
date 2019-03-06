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
	"fmt"
	"strings"

	"github.com/emicklei/proto"
	"github.com/uber/prototool/internal/file"
	"github.com/uber/prototool/internal/settings"
	"github.com/uber/prototool/internal/text"
)

type baseLinter struct {
	id                     string
	purpose                string
	suppressableAnnotation string
	addCheck               func(func(*file.ProtoSet, *proto.Comment, *text.Failure), string, []*FileDescriptor) error
}

func newBaseLinter(
	id string,
	purpose string,
	addCheck func(func(*text.Failure), string, []*FileDescriptor) error,
) *baseLinter {
	return newBaseSuppressableLinter(
		id,
		purpose,
		"",
		func(
			f func(*file.ProtoSet, *proto.Comment, *text.Failure),
			dirPath string,
			descriptors []*FileDescriptor,
		) error {
			return addCheck(
				func(failure *text.Failure) {
					f(nil, nil, failure)
				},
				dirPath,
				descriptors,
			)
		},
	)
}

func newBaseSuppressableLinter(
	id string,
	purpose string,
	suppressableAnnotation string,
	addCheck func(func(*file.ProtoSet, *proto.Comment, *text.Failure), string, []*FileDescriptor) error,
) *baseLinter {
	return &baseLinter{
		id:                     strings.ToUpper(id),
		purpose:                purpose,
		suppressableAnnotation: suppressableAnnotation,
		addCheck:               addCheck,
	}
}

func (c *baseLinter) ID() string {
	return c.id
}

func (c *baseLinter) Purpose(config settings.LintConfig) string {
	if c.suppressableAnnotation != "" && config.AllowSuppression {
		return fmt.Sprintf(`Suppressable with "@suppresswarnings %s". %s`, c.suppressableAnnotation, c.purpose)
	}
	return c.purpose
}

func (c *baseLinter) Check(dirPath string, descriptors []*FileDescriptor) ([]*text.Failure, error) {
	var failures []*text.Failure
	err := c.addCheck(
		func(protoSet *file.ProtoSet, comment *proto.Comment, failure *text.Failure) {
			if !c.isSuppressed(protoSet, comment) {
				if c.allowSuppression(protoSet) && failure.Message != "" {
					suppressionMessage := fmt.Sprintf(`This can be suppressed by adding "@suppresswarnings %s" to the comment.`, c.suppressableAnnotation)
					if failure.Message != "" {
						suppressionMessage = " " + suppressionMessage
					}
					failure.Message = failure.Message + suppressionMessage
				}
				failures = append(failures, failure)
			}
		},
		dirPath,
		descriptors,
	)
	for _, failure := range failures {
		failure.LintID = c.id
	}
	return failures, err
}

func (c *baseLinter) allowSuppression(protoSet *file.ProtoSet) bool {
	return c.suppressableAnnotation != "" && protoSet != nil && protoSet.Config.Lint.AllowSuppression
}

func (c *baseLinter) isSuppressed(protoSet *file.ProtoSet, comment *proto.Comment) bool {
	if !c.allowSuppression(protoSet) {
		return false
	}
	if comment == nil {
		return false
	}
	annotation := "@suppresswarnings " + c.suppressableAnnotation
	for _, line := range comment.Lines {
		if strings.Contains(line, annotation) {
			return true
		}
	}
	return false
}
