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

package location

import (
	"strconv"
	"strings"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
)

// Finder is used to map location paths to their corresponding
// locations. The Finder is meant to be used on a per-file basis;
// each file can access the Finder to determine the span for
// their individual field types.
//
// The locations map is keyed based on a unique representation
// of location paths.
type Finder struct {
	locations map[string]Location
}

// NewFinder constructs a *Finder for a specific file.
func NewFinder(src *descriptor.SourceCodeInfo) *Finder {
	locations := src.GetLocation()
	m := make(map[string]Location, len(locations))
	for _, l := range locations {
		m[getKey(l.GetPath())] = NewLocation(l)
	}
	return &Finder{locations: m}
}

// Find returns the Location for the corresponding Path
// based on its unique key. If not found, false is
// returned.
func (f *Finder) Find(p Path) (Location, bool) {
	l, ok := f.locations[getKey(p)]
	return l, ok
}

// getKey constructs a key from the given proto
// path representation.
//
// A proto path is composed of any number of elements,
// where each element represents a field type, index
// or target component (e.g. The first message's name field).
//
// The path is uniquely represented by a dot-delimited
// string containing all of the path's type, index, and
// target contents.
func getKey(p []int32) string {
	parts := make([]string, len(p))
	for i := range p {
		parts[i] = strconv.Itoa(int(p[i]))
	}
	return strings.Join(parts, ".")
}
