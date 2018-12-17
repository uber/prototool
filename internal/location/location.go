package location

import "github.com/golang/protobuf/protoc-gen-go/descriptor"

// Location is a convenience wrapper for working
// with Proto source locations. It contains a
// span and its corresponding comments.
type Location struct {
	Span     Span
	Comments Comments
}

// NewLocation creates a new Location from the Proto source location.
func NewLocation(l *descriptor.SourceCodeInfo_Location) Location {
	return Location{
		Span: NewSpan(l.GetSpan()),
		Comments: Comments{
			Leading:         l.GetLeadingComments(),
			Trailing:        l.GetTrailingComments(),
			LeadingDetached: l.GetLeadingDetachedComments(),
		},
	}
}
