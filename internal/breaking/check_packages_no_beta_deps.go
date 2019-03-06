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
	"github.com/uber/prototool/internal/protostrs"
	"github.com/uber/prototool/internal/text"
)

func checkPackagesNoBetaDeps(addFailure func(*text.Failure), from *extract.PackageSet, to *extract.PackageSet) error {
	// this check is only run if WithIncludeBeta is not set
	// so this map will only contain stable packages
	for _, toPackage := range to.PackageNameToPackage() {
		for _, depPackageName := range toPackage.ProtoMessage().DependencyNames {
			if _, betaVersion, ok := protostrs.MajorBetaVersion(depPackageName); ok && betaVersion > 0 {
				addFailure(newPackagesNoBetaDepsFailure(toPackage.FullyQualifiedName(), depPackageName))
			}
		}
	}
	return nil
}

func newPackagesNoBetaDepsFailure(packageName string, depPackageName string) *text.Failure {
	return newTextFailuref(`Package %q depends on beta package %q which is not allowed.`, packageName, depPackageName)
}
