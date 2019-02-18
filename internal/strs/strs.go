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

// Package strs contains common string manipulation functionality.
//
// This functionality is not really centralized anywhere in Golang OSS world,
// and there are some specific requirements we have. This is used mostly
// in the lint package.
package strs

import (
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// IsCapitalized returns true if is not empty and the first letter is
// an uppercase character.
func IsCapitalized(s string) bool {
	if s == "" {
		return false
	}
	r, _ := utf8.DecodeRuneInString(s)
	return isUpper(r)
}

// IsCamelCase returns false if s is empty or contains any character that is
// not between 'A' and 'Z', 'a' and 'z', '0' and '9', or in extraRunes.
// It does not care about lowercase or uppercase.
func IsCamelCase(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if !(isLetter(c) || isDigit(c)) {
			return false
		}
	}
	return true
}

// IsLowerSnakeCase returns true if s only contains lowercase letters,
// digits, and/or underscores. s MUST NOT begin or end with an underscore.
func IsLowerSnakeCase(s string) bool {
	if s == "" || s[0] == '_' || s[len(s)-1] == '_' {
		return false
	}
	for _, r := range s {
		if !(isLower(r) || isDigit(r) || r == '_') {
			return false
		}
	}
	return true
}

// IsUpperSnakeCase returns true if s only contains uppercase letters,
// digits, and/or underscores. s MUST NOT begin or end with an underscore.
func IsUpperSnakeCase(s string) bool {
	if s == "" || s[0] == '_' || s[len(s)-1] == '_' {
		return false
	}
	for _, r := range s {
		if !(isUpper(r) || isDigit(r) || r == '_') {
			return false
		}
	}
	return true
}

// ToLowerSnakeCase converts s to lower_snake_case.
func ToLowerSnakeCase(s string) string {
	return strings.ToLower(toSnake(s))
}

// ToUpperSnakeCase converts s to UPPER_SNAKE_CASE.
func ToUpperSnakeCase(s string) string {
	return strings.ToUpper(toSnake(s))
}

// ToUpperCamelCase converts s to UpperCamelCase.
//
// We use this for files, so any delimiter (_, -, or space) is
// used to denote word boundaries, but we trim spaces from the
// beginning and end of the string first.
//
// If a letter is uppercase, it will stay uppercase regardless,
// this is for cases of abbreviations.
func ToUpperCamelCase(s string) string {
	output := ""
	var previous rune
	for i, c := range strings.TrimSpace(s) {
		if !isDelimiter(c) {
			if i == 0 || isDelimiter(previous) || isUpper(c) {
				output += string(unicode.ToUpper(c))
			} else {
				output += string(unicode.ToLower(c))
			}
		}
		previous = c
	}
	return output
}

// SplitCamelCaseWord splits a CamelCase word into its parts.
//
// If s is empty, returns nil.
// If s is not CamelCase, returns nil.
func SplitCamelCaseWord(s string) []string {
	if s == "" {
		return nil
	}
	s = strings.TrimSpace(s)
	if !IsCamelCase(s) {
		return nil
	}
	return SplitSnakeCaseWord(toSnake(s))
}

// SplitSnakeCaseWord splits a snake_case word into its parts.
//
// If s is empty, returns nil.
// If s is not snake_case, returns nil.
func SplitSnakeCaseWord(s string) []string {
	if s == "" {
		return nil
	}
	s = strings.TrimSpace(s)
	if !isSnake(s) {
		return nil
	}
	var previous rune
	var words []string
	var curWord string
	for _, c := range s {
		if c != '_' {
			if previous == '_' {
				if curWord != "" {
					words = append(words, curWord)
				}
				curWord = ""
			}
			curWord += string(c)
		}
		previous = c
	}
	if curWord != "" {
		words = append(words, curWord)
	}
	return words
}

// SortUniq returns the unique sorted non-empty values of s.
func SortUniq(s []string) []string {
	return SortUniqModify(s, nil)
}

// SortUniqModify returns the unique sorted non-empty values of s.
// If modifier is not nil, modifier will be applied to each element in s.
func SortUniqModify(s []string, modifier func(string) string) []string {
	m := make(map[string]struct{})
	for _, e := range s {
		if modifier != nil {
			e = modifier(e)
		}
		if e != "" {
			m[e] = struct{}{}
		}
	}
	return MapToSortedSlice(m)
}

// MapToSortedSlice returns the sorted keys of m.
func MapToSortedSlice(m map[string]struct{}) []string {
	s := make([]string, 0, len(m))
	for e := range m {
		s = append(s, e)
	}
	sort.Strings(s)
	return s
}

// Intersection return the intersection between one and
// two, sorted and dropping empty strings.
func Intersection(one []string, two []string) []string {
	m1 := make(map[string]struct{})
	for _, e := range one {
		if e == "" {
			continue
		}
		m1[e] = struct{}{}
	}
	m2 := make(map[string]struct{})
	for _, e := range two {
		if e == "" {
			continue
		}
		m2[e] = struct{}{}
	}
	for key := range m1 {
		if _, ok := m2[key]; !ok {
			delete(m1, key)
		}
	}
	s := make([]string, 0, len(m1))
	for key := range m1 {
		s = append(s, key)
	}
	sort.Strings(s)
	return s
}

// IsLowercase returns true if s is not empty and is all lowercase.
func IsLowercase(s string) bool {
	if s == "" {
		return false
	}
	return strings.ToLower(s) == s
}

// IsUppercase returns true if s is not empty and is all uppercase.
func IsUppercase(s string) bool {
	if s == "" {
		return false
	}
	return strings.ToUpper(s) == s
}

// isSnake returns true if s only contains letters, digits, and/or underscores.
// s MUST NOT begin or end with an underscore.
func isSnake(s string) bool {
	if s == "" || s[0] == '_' || s[len(s)-1] == '_' {
		return false
	}
	for _, r := range s {
		if !(isLetter(r) || isDigit(r) || r == '_') {
			return false
		}
	}
	return true
}

// toSnake converts s to snake_case.
// It is assumed s has no spaces.
func toSnake(s string) string {
	output := ""
	s = strings.TrimSpace(s)
	for i, c := range s {
		if i > 0 && isUpper(c) && output[len(output)-1] != '_' && ((i < len(s)-1 && !isUpper(rune(s[i+1]))) || (isLower(rune(s[i-1])))) {
			output += "_" + string(c)
		} else {
			output += string(c)
		}
	}
	return output
}

func isLetter(r rune) bool {
	return isUpper(r) || isLower(r)
}

func isLower(r rune) bool {
	return 'a' <= r && r <= 'z'
}

func isUpper(r rune) bool {
	return 'A' <= r && r <= 'Z'
}

func isDigit(r rune) bool {
	return '0' <= r && r <= '9'
}

func isDelimiter(r rune) bool {
	return r == '-' || r == '_' || r == ' ' || r == '\t' || r == '\n' || r == '\r'
}
