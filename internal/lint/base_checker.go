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

package lint

import (
	"strings"
	"sync"

	"github.com/emicklei/proto"
	"github.com/uber/prototool/internal/text"
)

type baseChecker struct {
	id      string
	purpose string
	check   func(string, []*proto.Proto) ([]*text.Failure, error)
}

func newBaseAddChecker(
	id string,
	purpose string,
	addCheck func(func(*text.Failure), string, []*proto.Proto) error,
) *baseChecker {
	return newBaseChecker(
		id,
		purpose,
		func(dirPath string, descriptors []*proto.Proto) ([]*text.Failure, error) {
			var failures []*text.Failure
			var lock sync.Mutex
			if err := addCheck(
				func(failure *text.Failure) {
					lock.Lock()
					failures = append(failures, failure)
					lock.Unlock()
				},
				dirPath,
				descriptors,
			); err != nil {
				return nil, err
			}
			return failures, nil
		},
	)
}

func newBaseChecker(
	id string,
	purpose string,
	check func(string, []*proto.Proto) ([]*text.Failure, error),
) *baseChecker {
	return &baseChecker{
		id:      strings.ToUpper(id),
		purpose: purpose,
		check:   check,
	}
}

func (c *baseChecker) ID() string {
	return c.id
}

func (c *baseChecker) Purpose() string {
	return c.purpose
}

func (c *baseChecker) Check(dirPath string, descriptors []*proto.Proto) ([]*text.Failure, error) {
	failures, err := c.check(dirPath, descriptors)
	for _, failure := range failures {
		failure.ID = c.id
	}
	return failures, err
}
