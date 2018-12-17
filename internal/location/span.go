package location

import "fmt"

// Span represents a Proto location span.
// A Span always has exactly three or four elements: start line, start column,
// end line (optional, otherwise assumed same as start line), end column.
// These are packed into a single field for efficiency.
//
// We only ever need to display the initial line and column, so we ignore
// the remaining elements, i.e. the end line and end column.
//
// Note that line and column numbers are zero-based. We purposefully add 1 to
// each before displaying to a user.
type Span [2]int32

// NewSpan constructs a Span from the given slice.
func NewSpan(s []int32) Span {
	if len(s) < 2 {
		return Span{}
	}
	return Span{s[0], s[1]}
}

// Line returns the start line for this Span.
func (s Span) Line() int32 {
	return s[0] + 1
}

// Col returns the start column for this Span.
func (s Span) Col() int32 {
	return s[1] + 1
}

// String displays this Span as a string.
func (s Span) String() string {
	return fmt.Sprintf("%d:%d", s.Line(), s.Col())
}
