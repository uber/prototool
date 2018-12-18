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

package protostrs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGoPackage(t *testing.T) {
	assert.Equal(t, "", GoPackage(""))
	assert.Equal(t, "foopb", GoPackage("foo"))
	assert.Equal(t, "barpb", GoPackage("foo.bar"))
}

func TestGoPackageLastTwo(t *testing.T) {
	assert.Equal(t, "", GoPackageLastTwo(""))
	assert.Equal(t, "foo", GoPackageLastTwo("foo"))
	assert.Equal(t, "foobar", GoPackageLastTwo("foo.bar"))
	assert.Equal(t, "foobar", GoPackageLastTwo("first.foo.bar"))
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

func TestMajorVersion(t *testing.T) {
	version, ok := MajorVersion("foo.v0")
	assert.True(t, ok)
	assert.Equal(t, 0, int(version))
	version, ok = MajorVersion("foo.v1")
	assert.True(t, ok)
	assert.Equal(t, 1, int(version))
	version, ok = MajorVersion("foo.bar.v1")
	assert.True(t, ok)
	assert.Equal(t, 1, int(version))
	version, ok = MajorVersion("foo.bar.v18")
	assert.True(t, ok)
	assert.Equal(t, 18, int(version))
	version, ok = MajorVersion("foo.bar.v180")
	assert.True(t, ok)
	assert.Equal(t, 180, int(version))
	_, ok = MajorVersion("foo.barv1")
	assert.False(t, ok)
	_, ok = MajorVersion("barv1")
	assert.False(t, ok)
	_, ok = MajorVersion("v1")
	assert.False(t, ok)
	_, ok = MajorVersion("foo.barv")
	assert.False(t, ok)
	_, ok = MajorVersion("barv")
	assert.False(t, ok)
	_, ok = MajorVersion("v")
	assert.False(t, ok)
	_, ok = MajorVersion("foo.bar.v-1")
	assert.False(t, ok)
}
