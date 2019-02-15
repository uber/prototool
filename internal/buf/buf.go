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
	"bytes"
	"fmt"
	"strings"
	"unicode"
)

// Printer is a convenience struct that helps when printing files.
//
// The concept was taken from the golang/protobuf plugin.
type Printer struct {
	indent      string
	buffer      *bytes.Buffer
	indentCount int
}

// NewPrinter returns a new Printer.
func NewPrinter(indent string) *Printer {
	return &Printer{
		indent:      indent,
		buffer:      bytes.NewBuffer(nil),
		indentCount: 0,
	}
}

// P prints the args concatenated on the same line after printing the current indent and then prints a newline.
func (p *Printer) P(args ...interface{}) {
	if len(args) == 0 {
		_, _ = p.buffer.WriteRune('\n')
		return
	}
	lineBuffer := bytes.NewBuffer(nil)
	if p.indentCount > 0 {
		_, _ = fmt.Fprint(lineBuffer, strings.Repeat(p.indent, p.indentCount))
	}
	for _, arg := range args {
		_, _ = fmt.Fprint(lineBuffer, arg)
	}
	line := bytes.TrimRightFunc(lineBuffer.Bytes(), unicode.IsSpace)
	if len(bytes.TrimSpace(line)) != 0 {
		_, _ = p.buffer.Write(line)
	}
	_, _ = p.buffer.WriteRune('\n')
}

// In adds one indent.
func (p *Printer) In() {
	p.indentCount++
}

// Out deletes one indent.
func (p *Printer) Out() {
	// might want to error
	if p.indentCount > 0 {
		p.indentCount--
	}
}

// String returns the printed string.
func (p *Printer) String() string {
	return p.buffer.String()
}

// Bytes returns the printed bytes.
func (p *Printer) Bytes() []byte {
	return p.buffer.Bytes()
}
