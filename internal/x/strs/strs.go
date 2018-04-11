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
	"sort"
	"strings"
)

// IsCapitalized returns true if is is not empty and the first letter is
// between 'A' and 'Z'.
func IsCapitalized(s string) bool {
	if s == "" {
		return false
	}
	firstLetter := s[0]
	return firstLetter >= 'A' && firstLetter <= 'Z'
}

// IsCamelCase returns false if s is empty or contains any character that is
// not between 'A' and 'Z' or 'a' and 'z'. It does not care about lowercase
// or uppercase.
func IsCamelCase(s string, extraRunes ...rune) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')) && !containsRune(c, extraRunes) {
			return false
		}
	}
	return true
}

// IsLowerSnakeCase returns false if s is empty or contains any character that is
// not between 'a' and 'z' or '0' and '9' or '_', or if s begins or ends
// with '_'.
func IsLowerSnakeCase(s string, extraRunes ...rune) bool {
	if s == "" {
		return false
	}
	if s[0] == '_' {
		return false
	}
	if s[len(s)-1] == '_' {
		return false
	}
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_') && !containsRune(c, extraRunes) {
			return false
		}
	}
	return true
}

// IsUpperSnakeCase returns false if s is empty or contains any character that is
// not between 'A' and 'Z' or '0' and '9' or '_', or if s begins or ends
// with '_'.
func IsUpperSnakeCase(s string, extraRunes ...rune) bool {
	if s == "" {
		return false
	}
	if s[0] == '_' {
		return false
	}
	if s[len(s)-1] == '_' {
		return false
	}
	for _, c := range s {
		if !((c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') && !containsRune(c, extraRunes) {
			return false
		}
	}
	return true
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

// ToUpperSnakeCase converts s to UPPER_SNAKE_CASE.
func ToUpperSnakeCase(s string) string {
	return strings.ToUpper(ToSnakeCase(s))
}

// ToSnakeCase converts s to Snake_case.
// It is assumed s has no spaces.
func ToSnakeCase(s string) string {
	output := ""
	for i, c := range s {
		if i > 0 && isUppercaseRune(c) && output[len(output)-1] != '_' && i < len(s)-1 && !isUppercaseRune(rune(s[i+1])) {
			output += "_" + string(c)
		} else {
			output += string(c)
		}
	}
	return output
}

// DedupeSlice returns s with no duplicates, in the same order.
// If modifier is not nil, modifier will be applied to each element in s.
func DedupeSlice(s []string, modifier func(string) string) []string {
	m := make(map[string]struct{})
	for _, e := range s {
		if e == "" {
			continue
		}
		key := e
		if modifier != nil {
			key = modifier(e)
		}
		m[key] = struct{}{}
	}
	o := make([]string, 0, len(m))
	for _, e := range s {
		if e == "" {
			continue
		}
		key := e
		if modifier != nil {
			key = modifier(e)
		}
		if _, ok := m[key]; ok {
			o = append(o, key)
			delete(m, key)
		}
	}
	return o
}

// DedupeSortSlice returns s with no duplicates, sorted.
// If modifier is not nil, modifier will be applied to each element in s.
func DedupeSortSlice(s []string, modifier func(string) string) []string {
	o := DedupeSlice(s, modifier)
	sort.Strings(o)
	return o
}

// IntersectionSlice return the intersection between one and two, sorted.
func IntersectionSlice(one []string, two []string) []string {
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

func containsRune(r rune, s []rune) bool {
	for _, e := range s {
		if e == r {
			return true
		}
	}
	return false
}

func isUppercaseRune(c rune) bool {
	return c >= 'A' && c <= 'Z'
}
