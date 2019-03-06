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

package protostrs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCSharpNamespace(t *testing.T) {
	assert.Equal(t, "", CSharpNamespace(""))
	assert.Equal(t, "Foo", CSharpNamespace("foo"))
	assert.Equal(t, "Foo.Bar", CSharpNamespace("foo.bar"))
	assert.Equal(t, "Foo.BAr", CSharpNamespace("foo.bAr"))
	assert.Equal(t, "Foo.Bar.V1", CSharpNamespace("foo.bar.v1"))
	assert.Equal(t, "Foo.Bar.V1betaa1", CSharpNamespace("foo.bar.v1betaa1"))
	assert.Equal(t, "Foo.Bar.V1Beta1", CSharpNamespace("foo.bar.v1beta1"))
}

func TestPHPNamespace(t *testing.T) {
	assert.Equal(t, ``, PHPNamespace(""))
	assert.Equal(t, `Foo`, PHPNamespace("foo"))
	assert.Equal(t, `Foo\\Bar`, PHPNamespace("foo.bar"))
	assert.Equal(t, `Foo\\BAr`, PHPNamespace("foo.bAr"))
	assert.Equal(t, `Foo\\Bar\\V1`, PHPNamespace("foo.bar.v1"))
	assert.Equal(t, `Foo\\Bar\\V1betaa1`, PHPNamespace("foo.bar.v1betaa1"))
	assert.Equal(t, `Foo\\Bar\\V1Beta1`, PHPNamespace("foo.bar.v1beta1"))
}

func TestGoPackage(t *testing.T) {
	assert.Equal(t, "", GoPackage(""))
	assert.Equal(t, "foopb", GoPackage("foo"))
	assert.Equal(t, "barpb", GoPackage("foo.bar"))
}

func TestGoPackageV2(t *testing.T) {
	assert.Equal(t, "", GoPackageV2(""))
	assert.Equal(t, "foopb", GoPackageV2("foo"))
	assert.Equal(t, "barpb", GoPackageV2("foo.bar"))
	assert.Equal(t, "barpb", GoPackageV2("first.foo.bar"))
	assert.Equal(t, "barv1", GoPackageV2("foo.bar.v1"))
	assert.Equal(t, "barv1", GoPackageV2("first.foo.bar.v1"))
	assert.Equal(t, "barv1beta1", GoPackageV2("foo.bar.v1beta1"))
	assert.Equal(t, "barv1beta1", GoPackageV2("first.foo.bar.v1beta1"))
	assert.Equal(t, "v1betaa1pb", GoPackageV2("foo.bar.v1betaa1"))
}

func TestJavaOuterClassname(t *testing.T) {
	assert.Equal(t, "", JavaOuterClassname(""))
	assert.Equal(t, "FileProto", JavaOuterClassname("file.proto"))
	assert.Equal(t, "FileProto", JavaOuterClassname("file.txt"))
	assert.Equal(t, "FileProto", JavaOuterClassname("a/file.proto"))
	assert.Equal(t, "FileProto", JavaOuterClassname("a/b/file.proto"))
	assert.Equal(t, "FileOneProto", JavaOuterClassname("a/b/file_one.proto"))
	assert.Equal(t, "FileOneProto", JavaOuterClassname("a/b/file-one.proto"))
	assert.Equal(t, "FileOneProto", JavaOuterClassname("a/b/file one.proto"))
	assert.Equal(t, "FiLeOneTwoProto", JavaOuterClassname("a/b/fiLe_One_two.proto"))
	assert.Equal(t, "FileOneProto", JavaOuterClassname("a/b/file one.txt"))
}

func TestJavaPackage(t *testing.T) {
	assert.Equal(t, "", JavaPackage(""))
	assert.Equal(t, "com.foo", JavaPackage("foo"))
	assert.Equal(t, "com.foo.bar", JavaPackage("foo.bar"))
}

func TestJavaPackagePrefixOverride(t *testing.T) {
	assert.Equal(t, "", JavaPackagePrefixOverride("", ""))
	assert.Equal(t, "com.foo", JavaPackagePrefixOverride("foo", ""))
	assert.Equal(t, "com.foo.bar", JavaPackagePrefixOverride("foo.bar", ""))
	assert.Equal(t, "au.com.foo", JavaPackagePrefixOverride("foo", "au.com"))
	assert.Equal(t, "au.com.foo.bar", JavaPackagePrefixOverride("foo.bar", "au.com"))
}

func TestOBJCClassPrefix(t *testing.T) {
	assert.Equal(t, "", OBJCClassPrefix(""))
	assert.Equal(t, "FXX", OBJCClassPrefix("foo"))
	assert.Equal(t, "FBX", OBJCClassPrefix("foo.bar"))
	assert.Equal(t, "FBB", OBJCClassPrefix("foo.bar.baz"))
	assert.Equal(t, "FBB", OBJCClassPrefix("foo.bar.baz.v1"))
	assert.Equal(t, "FBB", OBJCClassPrefix("foo.bar.baz.v1beta1"))
	assert.Equal(t, "FBB", OBJCClassPrefix("Foo.bAr.baz"))
	assert.Equal(t, "FBBV", OBJCClassPrefix("foo.bar.baz.v1betaa1"))
	assert.Equal(t, "GPP", OBJCClassPrefix("goo.par.paz"))
	assert.Equal(t, "GPP", OBJCClassPrefix("goo.par.paz.v1"))
	assert.Equal(t, "GPP", OBJCClassPrefix("goo.par.paz.v1beta1"))
	assert.Equal(t, "GPX", OBJCClassPrefix("goo.par.baz"))
	assert.Equal(t, "GPX", OBJCClassPrefix("goo.par.baz.v1"))
	assert.Equal(t, "GPX", OBJCClassPrefix("goo.par.baz.v1beta1"))
}

func TestMajorBetaVersion(t *testing.T) {
	testMajorBetaVersionValid(t, "foo.v1", 1, 0)
	testMajorBetaVersionValid(t, "foo.bar.v1", 1, 0)
	testMajorBetaVersionValid(t, "foo.bar.v18", 18, 0)
	testMajorBetaVersionValid(t, "foo.bar.v180", 180, 0)
	testMajorBetaVersionInvalid(t, "")
	testMajorBetaVersionInvalid(t, "foo.v")
	testMajorBetaVersionInvalid(t, "foo.v0")
	testMajorBetaVersionInvalid(t, "foo.V1")
	testMajorBetaVersionInvalid(t, "foo.barv1")
	testMajorBetaVersionInvalid(t, "barv1")
	testMajorBetaVersionInvalid(t, "v1")
	testMajorBetaVersionInvalid(t, "foo.barv")
	testMajorBetaVersionInvalid(t, "barv")
	testMajorBetaVersionInvalid(t, "v")
	testMajorBetaVersionInvalid(t, "foo.bar.v-1")
	testMajorBetaVersionValid(t, "foo.v1beta1", 1, 1)
	testMajorBetaVersionValid(t, "foo.bar.v1beta1", 1, 1)
	testMajorBetaVersionValid(t, "foo.bar.v18beta18", 18, 18)
	testMajorBetaVersionValid(t, "foo.bar.v180beta180", 180, 180)
	testMajorBetaVersionInvalid(t, "foo.v0beta0")
	testMajorBetaVersionInvalid(t, "foo.v1beta0")
	testMajorBetaVersionInvalid(t, "foo.v0beta1")
	testMajorBetaVersionInvalid(t, "foo.V1beta1")
	testMajorBetaVersionInvalid(t, "foo.barv1beta1")
	testMajorBetaVersionInvalid(t, "barv1beta1")
	testMajorBetaVersionInvalid(t, "v1beta1")
	testMajorBetaVersionInvalid(t, "foo.barvbeta1")
	testMajorBetaVersionInvalid(t, "barvbeta1")
	testMajorBetaVersionInvalid(t, "vbeta1")
	testMajorBetaVersionInvalid(t, "foo.bar.v1beta-1")
	testMajorBetaVersionInvalid(t, "foo.bar.v-1beta-1")
	testMajorBetaVersionInvalid(t, "foo.v1beta")
	testMajorBetaVersionInvalid(t, "foo.beta1")
	testMajorBetaVersionInvalid(t, "foo.v1beta1beta1")
}

func testMajorBetaVersionValid(t *testing.T, packageName string, expectedMajorBetaVersion uint64, expectedBetaVersion uint64) {
	majorVersion, betaVersion, ok := MajorBetaVersion(packageName)
	assert.True(t, ok, packageName)
	assert.Equal(t, expectedMajorBetaVersion, majorVersion, packageName)
	assert.Equal(t, expectedBetaVersion, betaVersion, packageName)
}

func testMajorBetaVersionInvalid(t *testing.T, packageName string) {
	_, _, ok := MajorBetaVersion(packageName)
	assert.False(t, ok, packageName)
}
