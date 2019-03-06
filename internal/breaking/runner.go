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

package breaking

import (
	"github.com/uber/prototool/internal/extract"
	"github.com/uber/prototool/internal/settings"
	"github.com/uber/prototool/internal/text"
	"go.uber.org/zap"
)

type runner struct {
	logger   *zap.Logger
	checkers []Checker
}

func newRunner(options ...RunnerOption) *runner {
	runner := &runner{
		logger:   zap.NewNop(),
		checkers: AllCheckers,
	}
	for _, option := range options {
		option(runner)
	}
	return runner
}

func (r *runner) Run(config settings.BreakConfig, from *extract.PackageSet, to *extract.PackageSet) ([]*text.Failure, error) {
	var err error
	if !config.IncludeBeta {
		from, err = from.WithoutBeta()
		if err != nil {
			return nil, err
		}
		to, err = to.WithoutBeta()
		if err != nil {
			return nil, err
		}
	}
	checkers := r.checkers
	// if includeBeta, do not do the check
	// else if not including beta, unless allow beta deps, do not do the check
	if !config.IncludeBeta && !config.AllowBetaDeps {
		checkers = append(checkers, PackagesNoBetaDepsChecker)
	}
	var failures []*text.Failure
	for _, checker := range checkers {
		if err := checker.Check(
			func(failure *text.Failure) {
				failure.LintID = checker.ID
				failures = append(failures, failure)
			},
			from,
			to,
		); err != nil {
			return nil, err
		}
	}
	return failures, nil
}
