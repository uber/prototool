package compatible

import (
	"fmt"
	"sort"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/uber/prototool/internal/location"
)

// fileChecker collects compatibility errors for a specific file.
type fileChecker struct {
	filename string
	finder   *location.Finder
	errors   Errors
}

func newFileChecker(from *descriptor.FileDescriptorProto) *fileChecker {
	return &fileChecker{
		filename: from.GetName(),
		finder:   location.NewFinder(from.GetSourceCodeInfo()),
	}
}

// Check determines whether "to" is backward-compatible with "from".
func Check(from, to *descriptor.FileDescriptorSet) []Error {
	var (
		basePath location.Path
		errs     Errors
	)

	fs := newFileSet(from)
	for _, updated := range to.GetFile() {
		if original, ok := fs[updated.GetName()]; ok {
			c := newFileChecker(original.descriptor)
			errs = append(errs, c.checkFile(original, newFile(updated, basePath))...)
		}
	}

	fs = newFileSet(to)
	for _, original := range from.GetFile() {
		filename := original.GetName()
		if _, ok := fs[filename]; !ok {
			errs = append(
				errs,
				Error{
					Filename: filename,
					Severity: Warn,
					Message:  fmt.Sprintf("Failed to validate file %q: the file no longer exists.", filename),
				},
			)
		}
	}

	sort.Sort(errs)
	return errs
}

// AddErrorf adds an Error to the fileChecker using the given path and formatted message.
func (c *fileChecker) AddErrorf(path location.Path, severity Severity, format string, args ...interface{}) {
	// If a location does not exist for
	// the given path, we fall back to
	// the default location.Span value.
	loc, _ := c.finder.Find(path)

	err := Error{
		Filename: c.filename,
		Line:     loc.Span.Line(),
		Column:   loc.Span.Col(),
		Severity: severity,
		Message:  fmt.Sprintf(format, args...),
	}
	c.errors = append(c.errors, err)
}
