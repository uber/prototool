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

package format

import (
	"bytes"
	"fmt"
	"strings"
)

const defaultIndentString = "  "

// printer is a convenience struct that helps when printing proto files.
// The concept was taken from the golang/protobuf plugin.
type printer struct {
	buffer       *bytes.Buffer
	indentString string
	indentCount  int
}

func newPrinter(indentString string) *printer {
	if indentString == "" {
		indentString = defaultIndentString
	}
	return &printer{bytes.NewBuffer(nil), indentString, 0}
}

// P prints the args concatenated on the same line after printing the current indent and then prints a newline.
//
// TODO: There is a lot of unnecessary memory allocation going on here, optimize
func (p *printer) P(args ...interface{}) {
	lineBuffer := bytes.NewBuffer(nil)
	if p.indentCount > 0 {
		fmt.Fprint(lineBuffer, strings.Repeat(p.indentString, p.indentCount))
	}
	for _, arg := range args {
		fmt.Fprint(lineBuffer, arg)
	}
	line := lineBuffer.Bytes()

	if len(bytes.TrimSpace(line)) != 0 {
		_, _ = p.buffer.Write(line)
	}
	_, _ = p.buffer.WriteRune('\n')
}

// In adds one indent.
func (p *printer) In() {
	p.indentCount++
}

// Out deletes one indent.
func (p *printer) Out() {
	// might want to error
	if p.indentCount > 0 {
		p.indentCount--
	}
}

// Bytes returns the printed bytes.
func (p *printer) Bytes() []byte {
	return p.buffer.Bytes()
}
