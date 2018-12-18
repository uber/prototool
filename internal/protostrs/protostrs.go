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
	"strconv"
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

// GoPackageLastTwo returns the value for the file option "go_package" given
// a package name. This will be equal to the last two values of the package
// separated by "."s. If packageName is empty, this will return an empty string.
// If packageName has only one value when separated by "."s, this will be that
// value.
func GoPackageLastTwo(packageName string) string {
	if packageName == "" {
		return ""
	}
	split := strings.Split(packageName, ".")
	if len(split) == 1 {
		return split[0]
	}
	return split[len(split)-2] + split[len(split)-1]
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

// MajorVersion extracts the major version number from the package name, if
// present. A package must be of the form "foo.vMAJORVERSION". Returns the
// version and true if the package is of this form, 0 and false otherwise.
func MajorVersion(packageName string) (uint64, bool) {
	if packageName == "" {
		return 0, false
	}
	split := strings.Split(packageName, ".")
	// A package named "vX" should not count as it is just a single package name,
	// not a package name and a version.
	if len(split) < 2 {
		return 0, false
	}
	versionPart := split[len(split)-1]
	// Must be 'v' along with at least one number.
	if len(versionPart) < 2 {
		return 0, false
	}
	if versionPart[0] != 'v' {
		return 0, false
	}
	versionString := versionPart[1:]
	version, err := strconv.ParseUint(versionString, 10, 64)
	if err != nil {
		return 0, false
	}
	return version, true
}
