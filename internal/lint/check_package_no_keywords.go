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
	"github.com/uber/prototool/internal/text"
)

var (
	packageNoKeywordsKeywords = []string{
		"internal",  // Golang
		"public",    // Java, C++, others
		"private",   // Java, C++, others
		"protected", // Java, C++, others
		"std",       // C++ (causes a problem with the std package)
	}

	packageNoKeywordsLinter = NewSuppressableLinter(
		"PACKAGE_NO_KEYWORDS",
		fmt.Sprintf(`Verifies that no packages contain one of the keywords "%s" as part of the name when split on '.'.`, strings.Join(packageNoKeywordsKeywords, ",")),
		"keywords",
		checkPackageNoKeywords,
	)

	packageNoKeywordsKeywordsMap = make(map[string]struct{}, len(packageNoKeywordsKeywords))
)

func init() {
	for _, keyword := range packageNoKeywordsKeywords {
		packageNoKeywordsKeywordsMap[keyword] = struct{}{}
	}

}

func checkPackageNoKeywords(add func(*file.ProtoSet, *proto.Comment, *text.Failure), dirPath string, descriptors []*FileDescriptor) error {
	return runVisitor(&packageNoKeywordsVisitor{baseAddSuppressableVisitor: newBaseAddSuppressableVisitor(add)}, descriptors)
}

type packageNoKeywordsVisitor struct {
	*baseAddSuppressableVisitor
}

func (v *packageNoKeywordsVisitor) VisitPackage(pkg *proto.Package) {
	for _, subPackage := range strings.Split(pkg.Name, ".") {
		potentialKeyword := strings.ToLower(subPackage)
		if _, ok := packageNoKeywordsKeywordsMap[potentialKeyword]; ok {
			v.AddFailuref(pkg.Comment, pkg.Position, `Package %q contains the keyword %q, this could cause problems in generated code.`, pkg.Name, potentialKeyword)
		}
	}
}
