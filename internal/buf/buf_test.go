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

package buf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasic(t *testing.T) {
	testPrinter(
		t,
		"onetwothree\n",
		func(p *Printer) {
			p.P(`one`, `two`, `three`)
		},
	)
	testPrinter(
		t,
		"one two three\n",
		func(p *Printer) {
			p.P(`one `, `two `, `three`)
		},
	)
	testPrinter(
		t,
		"one\n  twothree\n   four\n\nfive\n\n",
		func(p *Printer) {
			// purposefully adding space
			p.P(`one `)
			p.In()
			p.P(`two`, `three`)
			// purposefully adding space
			p.P(` four`)
			p.Out()
			p.P()
			p.P(`five`)
			p.P()
		},
	)
}

func testPrinter(t *testing.T, expected string, f func(*Printer)) {
	printer := NewPrinter("  ")
	f(printer)
	assert.Equal(t, expected, printer.String())
}
