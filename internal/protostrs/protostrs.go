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

// Package protostrs contains common string manipulation functionality for
// Protobuf packages and files.
//
// This is used in the format, lint, and create packages. The Java rules in this
// package roughly follow https://cloud.google.com/apis/design/file_structure.
package protostrs

import (
	"path/filepath"
	"strings"

	"github.com/uber/prototool/internal/strs"
)

// GoPackage returns the value for the file option "go_package" given
// a package name. This will be equal to the last value of the package
// separated by "."s, followed by "pb". If packageName is empty,
// this will return an empty string.
func GoPackage(packageName string) string {
	if packageName == "" {
		return ""
	}
	split := strings.Split(packageName, ".")
	return split[len(split)-1] + "pb"
}

// JavaOuterClassname returns the value for the file option
// "java_outer_classname" given a file name. This will be equal to the
// basename of the file with it's extension stripped, UpperCamelCased,
// followed by "Proto". If filename is empty, this will return an empty
// string.
func JavaOuterClassname(filename string) string {
	if filename == "" {
		return ""
	}
	filename = filepath.Base(filename)
	filename = strings.TrimSuffix(filename, filepath.Ext(filename))
	return strs.ToUpperCamelCase(filename) + "Proto"
}

// JavaPackage returns the value for the file option "java_package" given
// a package name. This will be equal to "com." followed by the package.
// If packageName is empty, this will return an empty string.
func JavaPackage(packageName string) string {
	if packageName == "" {
		return ""
	}
	return "com." + packageName
}
