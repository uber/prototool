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

package failure

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"text/scanner"
)

const (
	// Filename references the Filename field of a Failure.
	Filename Field = iota
	// Line references the Line field of a Failure.
	Line
	// Column references the Column field of a Failure.
	Column
	// ID references the ID field of a Failure.
	ID
	// Message references the Message field of a Failure.
	Message
)

var (
	_defaultFields = []Field{
		Filename,
		Line,
		Column,
		Message,
	}
	_failureFieldToString = map[Field]string{
		Filename: "filename",
		Line:     "line",
		Column:   "column",
		ID:       "id",
		Message:  "message",
	}
	_stringToField = map[string]Field{
		"filename": Filename,
		"line":     Line,
		"column":   Column,
		"id":       ID,
		"message":  Message,
	}
)

// Field references a field of a Failure.
type Field int

// String implements fmt.Stringer.
func (f Field) String() string {
	if s, ok := _failureFieldToString[f]; ok {
		return s
	}
	return strconv.Itoa(int(f))
}

// ParseFields parses Fields from the given string.
// Fields are expected to be colon-separated in the given string.
// Input is case-insensitive. If the string is empty, _defaultFields
// will be returned.
func ParseFields(s string) ([]Field, error) {
	if len(s) == 0 {
		return _defaultFields, nil
	}
	fields := strings.Split(s, ":")
	failureFields := make([]Field, len(fields))
	for i, f := range fields {
		ff, err := parseField(f)
		if err != nil {
			return nil, err
		}
		failureFields[i] = ff
	}
	return failureFields, nil
}

// Failure is a failure with a position in text.
type Failure struct {
	Filename string
	Line     int
	Column   int
	ID       string
	Message  string
}

// Writer is a writer that Failure.Println can accept.
//
// Both bytes.Buffer and bufio.Writer implement this.
type Writer interface {
	WriteRune(rune) (int, error)
	WriteString(string) (int, error)
}

// Fprintln prints the Failure to the writer with the given ordered fields.
func (f *Failure) Fprintln(writer Writer, fields ...Field) error {
	if len(fields) == 0 {
		fields = _defaultFields
	}
	written := false
	for i, field := range fields {
		printColon := true
		switch field {
		case Filename:
			filename := f.Filename
			if filename == "" {
				filename = "<input>"
			}
			if _, err := writer.WriteString(filename); err != nil {
				return err
			}
			written = true
		case Line:
			line := strconv.Itoa(f.Line)
			if line == "0" {
				line = "1"
			}
			if _, err := writer.WriteString(line); err != nil {
				return err
			}
			written = true
		case Column:
			column := strconv.Itoa(f.Column)
			if column == "0" {
				column = "1"
			}
			if _, err := writer.WriteString(column); err != nil {
				return err
			}
			written = true
		case ID:
			if f.ID != "" {
				if _, err := writer.WriteString(f.ID); err != nil {
					return err
				}
				written = true
			} else {
				printColon = false
			}
		case Message:
			if f.Message != "" {
				if _, err := writer.WriteString(f.Message); err != nil {
					return err
				}
				written = true
			} else {
				printColon = false
			}
		default:
			return fmt.Errorf("unknown Field: %v", field)
		}
		if printColon && i != len(fields)-1 {
			if _, err := writer.WriteRune(':'); err != nil {
				return err
			}
			written = true
		}
	}
	if written {
		_, err := writer.WriteRune('\n')
		return err
	}
	return nil
}

// String implements fmt.Stringer. The Failure is
// printed in the following format:
//
//  "<filename>:<line>:<column>:<id> <message>"
func (f *Failure) String() string {
	filename := f.Filename
	if filename == "" {
		filename = "<input>"
	}
	line := strconv.Itoa(f.Line)
	if line == "0" {
		line = "1"
	}
	column := strconv.Itoa(f.Column)
	if column == "0" {
		column = "1"
	}

	return fmt.Sprintf("%s:%s:%s:%s %s", filename, line, column, f.ID, f.Message)
}

// Newf is a helper that returns a new Failure.
func Newf(position scanner.Position, id string, format string, args ...interface{}) *Failure {
	return &Failure{
		ID:       id,
		Filename: position.Filename,
		Line:     position.Line,
		Column:   position.Column,
		Message:  fmt.Sprintf(format, args...),
	}
}

// Sort sorts the Failures by the following precedence:
//
//  filename > line > column > id > message
func Sort(failures []*Failure) {
	sort.Stable(sortFailures(failures))
}

type sortFailures []*Failure

func (f sortFailures) Len() int          { return len(f) }
func (f sortFailures) Swap(i int, j int) { f[i], f[j] = f[j], f[i] }
func (f sortFailures) Less(i int, j int) bool {
	if f[i] == nil && f[j] == nil {
		return false
	}
	if f[i] == nil && f[j] != nil {
		return true
	}
	if f[i] != nil && f[j] == nil {
		return false
	}
	if f[i].Filename < f[j].Filename {
		return true
	}
	if f[i].Filename > f[j].Filename {
		return false
	}
	if f[i].Line < f[j].Line {
		return true
	}
	if f[i].Line > f[j].Line {
		return false
	}
	if f[i].Column < f[j].Column {
		return true
	}
	if f[i].Column > f[j].Column {
		return false
	}
	if f[i].ID < f[j].ID {
		return true
	}
	if f[i].ID > f[j].ID {
		return false
	}
	if f[i].Message < f[j].Message {
		return true
	}
	return false
}

// parseField parses the Field from the given string.
// Input is case-insensitive.
func parseField(s string) (Field, error) {
	failureField, ok := _stringToField[strings.ToLower(s)]
	if !ok {
		return 0, fmt.Errorf("could not parse %s to a Field", s)
	}
	return failureField, nil
}
