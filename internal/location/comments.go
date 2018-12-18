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

// Comments represents a Proto location's comment details.
type Comments struct {
	// Leading is the comment that exists before
	// a proto type's definition.
	//
	//  /* I'm a leading comment. */
	//  message Foo {}
	//
	Leading string

	// Trailing is the comment that exists
	// immediately after a proto type's definition.
	//
	//  message Foo {} /* I'm a trailing comment. */
	//
	Trailing string

	// LeadingDetached are comments that exists
	// before a proto type's definition, separated
	// by at least one newline.
	//
	//  /* I'm a leading detached comment. */
	//
	//  message Foo {}
	//
	LeadingDetached []string
}
