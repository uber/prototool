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
	assert.Equal(t, false, IsCapitalized("hello"))
	assert.Equal(t, true, IsCapitalized("Hello"))
	assert.Equal(t, false, IsCapitalized("hELLO"))
	assert.Equal(t, true, IsCapitalized("HeLLo"))
}

func TestIsCamelCase(t *testing.T) {
	assert.Equal(t, true, IsCamelCase("hello"))
	assert.Equal(t, true, IsCamelCase("helloWorld"))
	assert.Equal(t, false, IsCamelCase("hello_world"))
	assert.Equal(t, false, IsCamelCase("hello.World"))
	assert.Equal(t, true, IsCamelCase("hello.World", '.'))
	assert.Equal(t, true, IsCamelCase("ABBRCamel"))
}

func TestIsLowerSnakeCase(t *testing.T) {
	assert.Equal(t, true, IsLowerSnakeCase("hello"))
	assert.Equal(t, false, IsLowerSnakeCase("helloWorld"))
	assert.Equal(t, true, IsLowerSnakeCase("hello_world"))
	assert.Equal(t, false, IsLowerSnakeCase("Hello_world"))
	assert.Equal(t, false, IsLowerSnakeCase("_hello_world"))
	assert.Equal(t, false, IsLowerSnakeCase("hello_world_"))
	assert.Equal(t, false, IsLowerSnakeCase("hello_world.foo"))
	assert.Equal(t, true, IsLowerSnakeCase("hello_world.foo", '.'))
}

func TestToSnakeCase(t *testing.T) {
	assert.Equal(t, "Camel_Case", ToSnakeCase("CamelCase"))
	assert.Equal(t, "camel_Case", ToSnakeCase("camelCase"))
	assert.Equal(t, "Camel_Case_", ToSnakeCase("CamelCase_"))
	assert.Equal(t, "_Camel_Case", ToSnakeCase("_CamelCase"))
	assert.Equal(t, "Camel_Case__Hello", ToSnakeCase("CamelCase__Hello"))
	assert.Equal(t, "ABBR_Camel", ToSnakeCase("ABBRCamel"))
}

func TestDedupeSlice(t *testing.T) {
	assert.Equal(t, []string{"b", "a", "c"}, DedupeSlice([]string{"b", "A", "c"}, strings.ToLower))
	assert.Equal(t, []string{"b", "a", "c"}, DedupeSlice([]string{"b", "A", "c", "a"}, strings.ToLower))
	assert.Equal(t, []string{"b", "a", "c"}, DedupeSlice([]string{"b", "A", "c", "a", "B"}, strings.ToLower))
	assert.Equal(t, []string{"b", "a", "c"}, DedupeSlice([]string{"b", "A", "c", "a", "B", "b"}, strings.ToLower))
	assert.Equal(t, []string{"b", "A", "c"}, DedupeSlice([]string{"b", "A", "c"}, nil))
	assert.Equal(t, []string{"b", "A", "c", "a"}, DedupeSlice([]string{"b", "A", "c", "a"}, nil))
	assert.Equal(t, []string{"b", "A", "c", "B"}, DedupeSlice([]string{"b", "A", "c", "A", "B"}, nil))
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
}
