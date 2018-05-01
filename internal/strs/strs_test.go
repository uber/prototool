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

package strs

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsCapitalized(t *testing.T) {
	assert.False(t, IsCapitalized(""))
	assert.False(t, IsCapitalized("hello"))
	assert.True(t, IsCapitalized("Hello"))
	assert.False(t, IsCapitalized("hELLO"))
	assert.True(t, IsCapitalized("HeLLo"))
}

func TestIsCamelCase(t *testing.T) {
	assert.False(t, IsCamelCase(""))
	assert.True(t, IsCamelCase("hello"))
	assert.True(t, IsCamelCase("helloWorld"))
	assert.False(t, IsCamelCase("hello_world"))
	assert.False(t, IsCamelCase("hello.World"))
	assert.True(t, IsCamelCase("hello.World", '.'))
	assert.True(t, IsCamelCase("ABBRCamel"))
}

func TestIsLowerSnakeCase(t *testing.T) {
	assert.False(t, IsLowerSnakeCase(""))
	assert.True(t, IsLowerSnakeCase("hello"))
	assert.False(t, IsLowerSnakeCase("helloWorld"))
	assert.True(t, IsLowerSnakeCase("hello_world"))
	assert.False(t, IsLowerSnakeCase("Hello_world"))
	assert.False(t, IsLowerSnakeCase("_hello_world"))
	assert.False(t, IsLowerSnakeCase("hello_world_"))
	assert.False(t, IsLowerSnakeCase("hello_world.foo"))
	assert.True(t, IsLowerSnakeCase("hello_world.foo", '.'))
}

func TestIsUpperSnakeCase(t *testing.T) {
	assert.False(t, IsUpperSnakeCase(""))
	assert.False(t, IsUpperSnakeCase("hello"))
	assert.True(t, IsUpperSnakeCase("HELLO"))
	assert.False(t, IsUpperSnakeCase("helloWorld"))
	assert.False(t, IsUpperSnakeCase("hello_world"))
	assert.True(t, IsUpperSnakeCase("HELLO_WORLD"))
	assert.False(t, IsUpperSnakeCase("Hello_world"))
	assert.False(t, IsUpperSnakeCase("_HELLO_WORLD"))
	assert.False(t, IsUpperSnakeCase("HELLO_WORLD_"))
	assert.False(t, IsUpperSnakeCase("HELLO_WORLD.FOO"))
	assert.True(t, IsUpperSnakeCase("HELLO_WORLD.FOO", '.'))
}

func TestIsLowercase(t *testing.T) {
	assert.False(t, IsLowercase(""))
	assert.True(t, IsLowercase("hello"))
	assert.False(t, IsLowercase("hEllo"))
	assert.False(t, IsLowercase("HELLO"))
}

func TestIsUppercase(t *testing.T) {
	assert.False(t, IsUppercase(""))
	assert.False(t, IsUppercase("hello"))
	assert.False(t, IsUppercase("hEllo"))
	assert.True(t, IsUppercase("HELLO"))
}

func TestToSnakeCase(t *testing.T) {
	assert.Equal(t, "", ToSnakeCase(""))
	assert.Equal(t, "Camel_Case", ToSnakeCase("CamelCase"))
	assert.Equal(t, "camel_Case", ToSnakeCase("camelCase"))
	assert.Equal(t, "Camel_Case_", ToSnakeCase("CamelCase_"))
	assert.Equal(t, "_Camel_Case", ToSnakeCase("_CamelCase"))
	assert.Equal(t, "Camel_Case__Hello", ToSnakeCase("CamelCase__Hello"))
	assert.Equal(t, "ABBR_Camel", ToSnakeCase("ABBRCamel"))
}

func TestToUpperSnakeCase(t *testing.T) {
	assert.Equal(t, "", ToUpperSnakeCase(""))
	assert.Equal(t, "CAMEL_CASE", ToUpperSnakeCase("CamelCase"))
	assert.Equal(t, "CAMEL_CASE", ToUpperSnakeCase("camelCase"))
	assert.Equal(t, "CAMEL_CASE_", ToUpperSnakeCase("CamelCase_"))
	assert.Equal(t, "_CAMEL_CASE", ToUpperSnakeCase("_CamelCase"))
	assert.Equal(t, "CAMEL_CASE__HELLO", ToUpperSnakeCase("CamelCase__Hello"))
	assert.Equal(t, "ABBR_CAMEL", ToUpperSnakeCase("ABBRCamel"))
}

func TestDedupeSlice(t *testing.T) {
	assert.Equal(t, []string{"b", "a", "c"}, DedupeSlice([]string{"b", "A", "c"}, strings.ToLower))
	assert.Equal(t, []string{"b", "a", "c"}, DedupeSlice([]string{"b", "A", "c", "a"}, strings.ToLower))
	assert.Equal(t, []string{"b", "a", "c"}, DedupeSlice([]string{"b", "A", "c", "a", "B"}, strings.ToLower))
	assert.Equal(t, []string{"b", "a", "c"}, DedupeSlice([]string{"b", "A", "c", "a", "B", "b"}, strings.ToLower))
	assert.Equal(t, []string{"b", "A", "c"}, DedupeSlice([]string{"b", "A", "c"}, nil))
	assert.Equal(t, []string{"b", "A", "c", "a"}, DedupeSlice([]string{"b", "A", "c", "a"}, nil))
	assert.Equal(t, []string{"b", "A", "c", "B"}, DedupeSlice([]string{"b", "A", "c", "A", "B"}, nil))
	assert.Equal(t, []string{"b", "A", "c", "B"}, DedupeSlice([]string{"b", "A", "c", "", "A", "B"}, nil))
}

func TestDedupeSortSlice(t *testing.T) {
	assert.Equal(t, []string{"a", "b", "c"}, DedupeSortSlice([]string{"b", "A", "c"}, strings.ToLower))
	assert.Equal(t, []string{"a", "b", "c"}, DedupeSortSlice([]string{"b", "A", "c", "a"}, strings.ToLower))
	assert.Equal(t, []string{"a", "b", "c"}, DedupeSortSlice([]string{"b", "A", "c", "a", "B"}, strings.ToLower))
	assert.Equal(t, []string{"a", "b", "c"}, DedupeSortSlice([]string{"b", "A", "c", "a", "B", "b"}, strings.ToLower))
	assert.Equal(t, []string{"A", "b", "c"}, DedupeSortSlice([]string{"b", "A", "c"}, nil))
	assert.Equal(t, []string{"A", "a", "b", "c"}, DedupeSortSlice([]string{"b", "A", "c", "a"}, nil))
	assert.Equal(t, []string{"A", "B", "b", "c"}, DedupeSortSlice([]string{"b", "A", "c", "A", "B"}, nil))
	assert.Equal(t, []string{"A", "B", "b", "c"}, DedupeSortSlice([]string{"b", "A", "c", "", "A", "B"}, nil))
}

func TestIntersectionSlice(t *testing.T) {
	assert.Equal(t, []string{}, IntersectionSlice([]string{"1", "2", "3"}, []string{"4", "5", "6"}))
	assert.Equal(t, []string{"1"}, IntersectionSlice([]string{"1", "2", "3"}, []string{"1", "5", "6"}))
	assert.Equal(t, []string{"1", "2"}, IntersectionSlice([]string{"1", "2", "3"}, []string{"1", "5", "2"}))
	assert.Equal(t, []string{}, IntersectionSlice([]string{"1"}, []string{"4"}))
	assert.Equal(t, []string{"1"}, IntersectionSlice([]string{"1"}, []string{"1"}))
	assert.Equal(t, []string{}, IntersectionSlice([]string{}, []string{"1"}))
	assert.Equal(t, []string{}, IntersectionSlice([]string{"1"}, []string{}))
	assert.Equal(t, []string{}, IntersectionSlice([]string{}, []string{}))
	assert.Equal(t, []string{"1", "2"}, IntersectionSlice([]string{"1", "2", "3"}, []string{"1", "5", "", "2"}))
}
