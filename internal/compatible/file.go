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

package compatible

import (
	"fmt"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/uber/prototool/internal/location"
)

type fileSet map[string]*file

func newFileSet(from *descriptor.FileDescriptorSet) fileSet {
	var basePath location.Path
	files := make(map[string]*file, len(from.GetFile()))
	for _, f := range from.GetFile() {
		files[f.GetName()] = newFile(f, basePath)
	}
	return files
}

// file represents a *descriptor.FileDescriptorProto.
type file struct {
	descriptor *descriptor.FileDescriptorProto
	path       location.Path
	name       string
	pkg        string
	enums      enums
	messages   messages
	services   services
}

func (f *file) Name() string        { return f.name }
func (f *file) Path() location.Path { return f.path }
func (f *file) Type() string        { return fmt.Sprintf("File %q", f.name) }

func newFile(fd *descriptor.FileDescriptorProto, p location.Path) *file {
	messages := make(messages, len(fd.GetMessageType()))
	for i, m := range fd.GetMessageType() {
		messages[m.GetName()] = newMessage(m, p.Scope(location.Message, i))
	}
	services := make(services, len(fd.GetService()))
	for i, s := range fd.GetService() {
		services[s.GetName()] = newService(s, p.Scope(location.Service, i))
	}

	return &file{
		descriptor: fd,
		path:       p,
		name:       fd.GetName(),
		pkg:        fd.GetPackage(),
		messages:   messages,
		enums:      getEnums(fd, p, location.Enum),
		services:   services,
	}
}

// checkFile verifies that,
//  - The file's package was not updated.
//  - None of the file's messages were inappropriately updated.
//  - None of the file's enums were inappropriately updated.
//  - None of the file's services were inappropriately updated.
func (c *fileChecker) checkFile(original, updated *file) []Error {
	c.checkUpdatedAttribute(
		original,
		Wire,
		"package",
		original.pkg,
		updated.pkg,
		location.Package,
	)
	c.checkMessages(original.messages, updated.messages)
	c.checkEnums(original.enums, updated.enums)
	c.checkServices(original.services, updated.services)
	return c.errors
}
