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
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFile(t *testing.T) {
	t.Run("No change", func(t *testing.T) {
		c := newTestChecker(t)
		c.checkFile(&file{name: _testFilename, pkg: "foo"}, &file{name: _testFilename, pkg: "foo"})
		assert.Len(t, c.errors, 0)
	})
	t.Run("Updated file package", func(t *testing.T) {
		c := newTestChecker(t)
		c.checkFile(&file{name: _testFilename, pkg: ""}, &file{name: _testFilename, pkg: "foo"})
		require.Len(t, c.errors, 1)
		assert.Equal(t, c.errors[0].String(), `test.proto:1:1:wire:File "test.proto" had its package updated from "" to "foo".`)
	})
}

func TestNewFileSet(t *testing.T) {
	t.Run("Empty file set", func(t *testing.T) {
		fs := newFileSet(&descriptor.FileDescriptorSet{})
		assert.Len(t, fs, 0)
	})
	t.Run("Non-empty file set", func(t *testing.T) {
		fs := newFileSet(&descriptor.FileDescriptorSet{
			File: []*descriptor.FileDescriptorProto{
				{
					Name: proto.String("filename.proto"),
					MessageType: []*descriptor.DescriptorProto{
						{
							Name: proto.String("msg"),
						},
					},
					Service: []*descriptor.ServiceDescriptorProto{
						{
							Name: proto.String("service"),
						},
					},
				},
			},
		})
		require.Len(t, fs, 1)

		f := fs["filename.proto"]
		require.Len(t, f.messages, 1)
		require.Len(t, f.services, 1)

		assert.Equal(t, "filename.proto", f.name)
		assert.Equal(t, "msg", f.messages["msg"].name)
		assert.Equal(t, "service", f.services["service"].name)
	})
}
