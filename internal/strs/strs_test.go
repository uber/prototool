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

func TestToLowerSnakeCase(t *testing.T) {
	assert.Equal(t, "", ToLowerSnakeCase(""))
	assert.Equal(t, "camel_case", ToLowerSnakeCase("CamelCase"))
	assert.Equal(t, "camel_case", ToLowerSnakeCase("  CamelCase"))
	assert.Equal(t, "camel_case", ToLowerSnakeCase("CamelCase  "))
	assert.Equal(t, "camel_case", ToLowerSnakeCase("camelCase"))
	assert.Equal(t, "camel_case_", ToLowerSnakeCase("CamelCase_"))
	assert.Equal(t, "_camel_case", ToLowerSnakeCase("_CamelCase"))
	assert.Equal(t, "camel_case__hello", ToLowerSnakeCase("CamelCase__Hello"))
	assert.Equal(t, "abbr_camel", ToLowerSnakeCase("ABBRCamel"))
	assert.Equal(t, "foo_abbr_camel", ToLowerSnakeCase("FooABBRCamel"))
	assert.Equal(t, "abbr_camel_abbr", ToLowerSnakeCase("ABBRCamelABBR"))
}

func TestSplitCamelCaseWord(t *testing.T) {
	assert.Equal(t, []string(nil), SplitSnakeCaseWord(""))
	assert.Equal(t, []string{"Camel", "Case"}, SplitCamelCaseWord("CamelCase"))
	assert.Equal(t, []string{"Camel", "Case"}, SplitCamelCaseWord("  CamelCase"))
	assert.Equal(t, []string{"Camel", "Case"}, SplitCamelCaseWord("CamelCase  "))
	assert.Equal(t, []string{"camel", "Case"}, SplitCamelCaseWord("camelCase"))
	assert.Equal(t, []string{"camel", "Case", "Hello"}, SplitCamelCaseWord("camelCaseHello"))
	assert.Equal(t, []string(nil), SplitCamelCaseWord("CamelCase_"))
	assert.Equal(t, []string(nil), SplitCamelCaseWord("_CamelCase"))
	assert.Equal(t, []string(nil), SplitCamelCaseWord("CamelCase__Hello"))
	assert.Equal(t, []string{"ABBR", "Camel"}, SplitCamelCaseWord("ABBRCamel"))
	assert.Equal(t, []string{"Foo", "ABBR", "Camel"}, SplitCamelCaseWord("FooABBRCamel"))
	assert.Equal(t, []string{"ABBR", "Camel", "ABBR"}, SplitCamelCaseWord("ABBRCamelABBR"))
	assert.Equal(t, []string(nil), SplitCamelCaseWord("Camel_Case"))
}

func TestSplitSnakeCaseWord(t *testing.T) {
	assert.Equal(t, []string(nil), SplitSnakeCaseWord(""))
	assert.Equal(t, []string{"Snake", "Case"}, SplitSnakeCaseWord("Snake_Case"))
	assert.Equal(t, []string{"Snake", "Case"}, SplitSnakeCaseWord("  Snake_Case"))
	assert.Equal(t, []string{"Snake", "Case"}, SplitSnakeCaseWord("Snake_Case  "))
	assert.Equal(t, []string{"snake", "Case"}, SplitSnakeCaseWord("snake_Case"))
	assert.Equal(t, []string(nil), SplitSnakeCaseWord("Snake_Case_"))
	assert.Equal(t, []string(nil), SplitSnakeCaseWord("_Snake_Case"))
	assert.Equal(t, []string{"Snake", "Case", "Hello"}, SplitSnakeCaseWord("Snake_Case__Hello"))
	assert.Equal(t, []string{"ABBR", "Snake"}, SplitSnakeCaseWord("ABBR_Snake"))
	assert.Equal(t, []string{"ABBR", "snake"}, SplitSnakeCaseWord("ABBR_snake"))
	assert.Equal(t, []string{"Foo", "ABBR", "Snake"}, SplitSnakeCaseWord("Foo_ABBR_Snake"))
	assert.Equal(t, []string{"Foo", "ABBR", "snake"}, SplitSnakeCaseWord("Foo_ABBR_snake"))
	assert.Equal(t, []string{"ABBR", "Snake", "ABBR"}, SplitSnakeCaseWord("ABBR_Snake_ABBR"))
	assert.Equal(t, []string{"ABBR", "snake", "ABBR"}, SplitSnakeCaseWord("ABBR_snake_ABBR"))
	assert.Equal(t, []string{"SNAKE", "CASE"}, SplitSnakeCaseWord("SNAKE_CASE"))
	assert.Equal(t, []string{"snake", "case"}, SplitSnakeCaseWord("snake_case"))
	assert.Equal(t, []string(nil), SplitSnakeCaseWord("SNAKE_CASE_"))
	assert.Equal(t, []string(nil), SplitSnakeCaseWord("_snake_case"))
}

func TestToUpperSnakeCase(t *testing.T) {
	assert.Equal(t, "", ToUpperSnakeCase(""))
	assert.Equal(t, "CAMEL_CASE", ToUpperSnakeCase("CamelCase"))
	assert.Equal(t, "CAMEL_CASE", ToUpperSnakeCase("camelCase"))
	assert.Equal(t, "CAMEL_CASE_", ToUpperSnakeCase("CamelCase_"))
	assert.Equal(t, "_CAMEL_CASE", ToUpperSnakeCase("_CamelCase"))
	assert.Equal(t, "CAMEL_CASE__HELLO", ToUpperSnakeCase("CamelCase__Hello"))
	assert.Equal(t, "ABBR_CAMEL", ToUpperSnakeCase("ABBRCamel"))
	assert.Equal(t, "FOO_ABBR_CAMEL", ToUpperSnakeCase("FooABBRCamel"))
}

func TestToUpperCamelCase(t *testing.T) {
	assert.Equal(t, "", ToUpperCamelCase(""))
	assert.Equal(t, "", ToUpperCamelCase("  "))
	assert.Equal(t, "CamelCase", ToUpperCamelCase("camel_case"))
	assert.Equal(t, "CamelCase", ToUpperCamelCase("  camel_case"))
	assert.Equal(t, "CamelCase", ToUpperCamelCase("  camel_case  "))
	assert.Equal(t, "CamelCase", ToUpperCamelCase("camel_case  "))
	assert.Equal(t, "CamelCase", ToUpperCamelCase("Camel_case"))
	assert.Equal(t, "CamelCase", ToUpperCamelCase("__Camel___case"))
	assert.Equal(t, "CamelCase", ToUpperCamelCase("__Camel___case__"))
	assert.Equal(t, "CamelCase", ToUpperCamelCase("Camel___case__"))
	assert.Equal(t, "CamelCase", ToUpperCamelCase("Camel-case"))
	assert.Equal(t, "CamelCase", ToUpperCamelCase("Camel case"))
	assert.Equal(t, "CamelCase", ToUpperCamelCase("  Camel case"))
	assert.Equal(t, "CAMELCase", ToUpperCamelCase("CAMEL_case"))
	assert.Equal(t, "CAMELCase", ToUpperCamelCase("__CAMEL___case"))
	assert.Equal(t, "CAMELCase", ToUpperCamelCase("__CAMEL___case__"))
	assert.Equal(t, "CAMELCase", ToUpperCamelCase("CAMEL___case__"))
	assert.Equal(t, "CAMELCase", ToUpperCamelCase("CAMEL-case"))
	assert.Equal(t, "CAMELCase", ToUpperCamelCase("CAMEL case"))
	assert.Equal(t, "CAMELCase", ToUpperCamelCase("  CAMEL case"))
}

func TestSortUniq(t *testing.T) {
	assert.Equal(t, []string{"A", "b", "c"}, SortUniq([]string{"b", "A", "c"}))
	assert.Equal(t, []string{"A", "a", "b", "c"}, SortUniq([]string{"b", "A", "c", "a"}))
	assert.Equal(t, []string{"A", "B", "b", "c"}, SortUniq([]string{"b", "A", "c", "A", "B"}))
	assert.Equal(t, []string{"A", "B", "b", "c"}, SortUniq([]string{"b", "A", "c", "", "A", "B"}))
}

func TestSortUniqModify(t *testing.T) {
	assert.Equal(t, []string{"a", "b", "c"}, SortUniqModify([]string{"b", "A", "c"}, strings.ToLower))
	assert.Equal(t, []string{"a", "b", "c"}, SortUniqModify([]string{"b", "A", "c", "a"}, strings.ToLower))
	assert.Equal(t, []string{"a", "b", "c"}, SortUniqModify([]string{"b", "A", "c", "a", "B"}, strings.ToLower))
	assert.Equal(t, []string{"a", "b", "c"}, SortUniqModify([]string{"b", "A", "c", "a", "B", "b"}, strings.ToLower))
	assert.Equal(t, []string{"A", "b", "c"}, SortUniqModify([]string{"b", "A", "c"}, nil))
	assert.Equal(t, []string{"A", "a", "b", "c"}, SortUniqModify([]string{"b", "A", "c", "a"}, nil))
	assert.Equal(t, []string{"A", "B", "b", "c"}, SortUniqModify([]string{"b", "A", "c", "A", "B"}, nil))
	assert.Equal(t, []string{"A", "B", "b", "c"}, SortUniqModify([]string{"b", "A", "c", "", "A", "B"}, nil))
}

func TestIntersection(t *testing.T) {
	assert.Equal(t, []string{}, Intersection([]string{"1", "2", "3"}, []string{"4", "5", "6"}))
	assert.Equal(t, []string{"1"}, Intersection([]string{"1", "2", "3"}, []string{"1", "5", "6"}))
	assert.Equal(t, []string{"1", "2"}, Intersection([]string{"1", "2", "3"}, []string{"1", "5", "2"}))
	assert.Equal(t, []string{}, Intersection([]string{"1"}, []string{"4"}))
	assert.Equal(t, []string{"1"}, Intersection([]string{"1"}, []string{"1"}))
	assert.Equal(t, []string{}, Intersection([]string{}, []string{"1"}))
	assert.Equal(t, []string{}, Intersection([]string{"1"}, []string{}))
	assert.Equal(t, []string{}, Intersection([]string{}, []string{}))
	assert.Equal(t, []string{"1", "2"}, Intersection([]string{"1", "2", "3"}, []string{"1", "5", "", "2"}))
	assert.Equal(t, []string{"1", "2"}, Intersection([]string{"1", "2", "", "3"}, []string{"1", "5", "2"}))
	assert.Equal(t, []string{"1", "2"}, Intersection([]string{"1", "2", "", "3"}, []string{"1", "5", "", "2"}))
}
