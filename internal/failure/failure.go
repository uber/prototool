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

type (
	// Field references a Failure field.
	Field int

	// ID references a Failure identifier.
	ID int
)

const (
	// Filename references the Filename field of a Failure.
	Filename Field = iota
	// Line references the Line field of a Failure.
	Line
	// Column references the Column field of a Failure.
	Column
	// Identifier references the ID field of a Failure.
	Identifier
	// Message references the Message field of a Failure.
	Message

	// Error identifies a generic Failure.
	Error ID = iota
	// Format identifies a diff format Failure.
	Format
	// Lint identifies a lint Failure.
	Lint
	// Proto identitifes an invalid proto Failure.
	Proto

	// DefaultFilename is the default Failure filename.
	DefaultFilename = "<input>"
	// DefaultLine is the default Failure line.
	DefaultLine = 1
	// DefaultColumn is the default Failure column number.
	DefaultColumn = 1
)

var (
	_defaultFields = []Field{
		Filename,
		Line,
		Column,
		Message,
	}
	_stringToField = map[string]Field{
		"filename": Filename,
		"line":     Line,
		"column":   Column,
		"id":       Identifier,
		"message":  Message,
	}
	_IDToString = map[ID]string{
		Error:  "ERROR",
		Format: "FORMAT",
		Lint:   "LINT",
		Proto:  "PROTO",
	}
)

// String implements fmt.Stringer.
func (id ID) String() string {
	if s, ok := _IDToString[id]; ok {
		return s
	}
	return strconv.Itoa(int(id))
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
// The given fields will overwrite the fields already set in this Failure.
func (f *Failure) Fprintln(writer Writer, fields ...Field) error {
	if len(fields) == 0 {
		fields = _defaultFields
	}
	printColon := true
	for i, field := range fields {
		switch field {
		case Filename:
			filename := f.Filename
			if filename == "" {
				filename = DefaultFilename
			}
			if _, err := writer.WriteString(filename); err != nil {
				return err
			}
		case Line:
			line := strconv.Itoa(f.Line)
			if line == "0" {
				line = strconv.Itoa(DefaultLine)
			}
			if _, err := writer.WriteString(line); err != nil {
				return err
			}
		case Column:
			column := strconv.Itoa(f.Column)
			if column == "0" {
				column = strconv.Itoa(DefaultColumn)
			}
			if _, err := writer.WriteString(column); err != nil {
				return err
			}
		case Identifier:
			id := f.ID
			if id == "" {
				id = Error.String()
			}
			if _, err := writer.WriteString(id); err != nil {
				return err
			}
			printColon = false
		case Message:
			if _, err := writer.WriteRune(' '); err != nil {
				return err
			}
			if _, err := writer.WriteString(f.Message); err != nil {
				return err
			}
			printColon = false
		default:
			return fmt.Errorf("unknown Field: %v", field)
		}
		if printColon && i != len(fields)-1 {
			if _, err := writer.WriteRune(':'); err != nil {
				return err
			}
		}
	}
	_, err := writer.WriteRune('\n')
	return err
}

// Newf is a helper that returns a new Failure.
func Newf(position scanner.Position, id ID, format string, args ...interface{}) *Failure {
	return &Failure{
		ID:       id.String(),
		Filename: position.Filename,
		Line:     position.Line,
		Column:   position.Column,
		Message:  fmt.Sprintf(format, args...),
	}
}

// Sort sorts the Failures by the following precedence:
//
//  filename > line > column > id > message
func Sort(fs []*Failure) {
	sort.Stable(failures(fs))
}

type failures []*Failure

func (f failures) Len() int          { return len(f) }
func (f failures) Swap(i int, j int) { f[i], f[j] = f[j], f[i] }
func (f failures) Less(i int, j int) bool {
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
