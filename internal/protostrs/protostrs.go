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

// CSharpNamespace returns the value for the file option "csharp_namespace"
// given a package name. It will capitalize each part of the package name
// taking special care for beta packages.
func CSharpNamespace(packageName string) string {
	if packageName == "" {
		return ""
	}
	// strings.Title would just work across the split but we need to take
	// care of beta packages
	majorVersion, betaVersion, ok := MajorBetaVersion(packageName)
	if !ok || majorVersion == 0 || betaVersion == 0 {
		return strings.Title(packageName)
	}
	split := strings.Split(packageName, ".")
	return strings.Title(strings.Join(split[:len(split)-1], ".")) + ".V" + strconv.Itoa(int(majorVersion)) + "Beta" + strconv.Itoa(int(betaVersion))
}

// PHPNamespace returns the value for the file option "php_namespace"
// given a package name. It will capitalize each part of the package name
// taking special care for beta packages.
func PHPNamespace(packageName string) string {
	return strings.Replace(CSharpNamespace(packageName), ".", `\\`, -1)
}

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

// GoPackageV2 returns the value for the file option "go_package" given
// a package name. This will be equal to the last two values of the package
// separated by "."s if the package is a MajorBetaPackage, or GoPackage otherwise.
func GoPackageV2(packageName string) string {
	if packageName == "" {
		return ""
	}
	if _, _, ok := MajorBetaVersion(packageName); !ok {
		return GoPackage(packageName)
	}
	split := strings.Split(packageName, ".")
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

// JavaPackagePrefixOverride returns the value for the file option "java_package" given
// a package name. This will be equal to "com." followed by the package.
// If prefixOverride is set, this will be equal to prefixOveride.package.
// If packageName is empty, this will return an empty string.
func JavaPackagePrefixOverride(packageName string, prefixOverride string) string {
	if packageName == "" {
		return ""
	}
	if prefixOverride != "" {
		return prefixOverride + "." + packageName
	}
	return "com." + packageName
}

// OBJCClassPrefix returns the value for the file option "objc_class_prefix"
// given a package name. It takes the first letter of each package part
// and capitalizes it, then concatenates these. If the length is 2, an "X"
// is added. If the length is 1, "XX" is added. If the length is 0, this
// returns empty. If the name is "GPB", this returns "GPX". The version part
// is dropped before all operations.
func OBJCClassPrefix(packageName string) string {
	if packageName == "" {
		return ""
	}
	split := strings.Split(packageName, ".")
	if _, _, ok := MajorBetaVersion(packageName); ok {
		// this is guaranteed to still have something since
		// len(split) must be >=3.
		split = split[:len(split)-1]
	}
	// just for safety
	if len(split) == 0 {
		return ""
	}
	s := ""
	for _, element := range split {
		// just for safety
		if element != "" {
			s = s + strings.ToUpper(element[0:1])
		}
	}
	switch len(s) {
	case 0:
		// just for safety
		return ""
	case 1:
		return s + "XX"
	case 2:
		return s + "X"
	case 3:
		if s == "GPB" {
			return "GPX"
		}
		return s
	default:
		return s
	}
}

// MajorBetaVersion extracts the major and beta version number from the package
// name, if present. A package must be of the form "foo.vMAJORVERSION" or
// "foo.vMAJORVERSIONbetaBETAVERSION" . Returns the major version, beta version
// and true if the package is of this form, 0 and false otherwise. If there is
// no beta version, 0 is returned. Valid versions are >=1.
func MajorBetaVersion(packageName string) (uint64, uint64, bool) {
	if packageName == "" {
		return 0, 0, false
	}
	split := strings.Split(packageName, ".")
	// A package named "vX" should not count as it is just a single package name,
	// not a package name and a version.
	if len(split) < 2 {
		return 0, 0, false
	}

	versionPart := split[len(split)-1]
	// Must be 'v' along with at least one number.
	if len(versionPart) < 2 {
		return 0, 0, false
	}
	if versionPart[0] != 'v' {
		return 0, 0, false
	}

	versionString := versionPart[1:]
	versionStringSplit := strings.Split(versionString, "beta")
	switch len(versionStringSplit) {
	case 1:
		version, err := strconv.ParseUint(versionString, 10, 64)
		if err != nil {
			return 0, 0, false
		}
		if version == 0 {
			return 0, 0, false
		}
		return version, 0, true
	case 2:
		majorVersionString := versionStringSplit[0]
		betaVersionString := versionStringSplit[1]
		if majorVersionString == "" || betaVersionString == "" {
			return 0, 0, false
		}
		majorVersion, err := strconv.ParseUint(majorVersionString, 10, 64)
		if err != nil {
			return 0, 0, false
		}
		betaVersion, err := strconv.ParseUint(betaVersionString, 10, 64)
		if err != nil {
			return 0, 0, false
		}
		if majorVersion == 0 || betaVersion == 0 {
			return 0, 0, false
		}
		return majorVersion, betaVersion, true
	default:
		return 0, 0, false
	}
}
