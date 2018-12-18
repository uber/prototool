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

// Path represents a Proto location path.
type Path []int32

// Scope adds field-specific scope to this Path.
// The path is scoped by adding a specific proto
// identifier, along with its index.
//
// Adding scope to a proto file's first
// message type, for example, is done by,
//
//  p.Scope(Message, 0)
func (p Path) Scope(typ ID, idx int) Path {
	return p.copyWith(int32(typ), int32(idx))
}

// Target adds the given target to this Path.
// The path receives a target to a specific
// proto type's component by explicitly
// specifying the target identifier.
//
// Targetting a proto file's package
// definition, for example, is done by,
//
//  p.Target(Package)
func (p Path) Target(target ID) Path {
	return p.copyWith(int32(target))
}

// copyWith creates a new Path, appending the
// given elements to those already found in p.
//
// We purposefully make a copy of the Path
// so that we don't accidentally truncate
// previously added elements.
//
// For details, read up at:
// https://blog.golang.org/go-slices-usage-and-internals
func (p Path) copyWith(elems ...int32) Path {
	path := make(Path, len(p)+len(elems))
	for i, e := range p {
		path[i] = e
	}
	for i, e := range elems {
		path[len(p)+i] = e
	}
	return path
}
