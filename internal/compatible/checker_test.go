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
)

const _testFilename = "test.proto"

func newTestChecker(t *testing.T) *fileChecker {
	return newFileChecker(
		&descriptor.FileDescriptorProto{
			Name: proto.String(_testFilename),
		},
	)
}

func TestCheck(t *testing.T) {
	t.Run("bar.proto", func(t *testing.T) {
		t.Skip("re-add when proper breakage tests are standardized in prototool")
		//original, err := encoding.UnmarshalDescriptorSet("../../internal/testdata/compatible/bar/bar.fd")
		//require.NoError(t, err)

		//updated, err := encoding.UnmarshalDescriptorSet("../../internal/testdata/compatible/bar/bar.updated.fd")
		//require.NoError(t, err)

		//actErrs := Check(original, updated)

		//expErrs := []string{
		//`bar.proto:1:1:wire:File "bar.proto" had its package updated from "" to "bar.bar".`,
		//`bar.proto:4:16:wire:Field "name" (1) had its json name updated from "NameJSON" to "name".`,
		//`bar.proto:6:15:wire:Oneof "only" was removed.`,
		//`bar.proto:7:17:wire:Field "this" (2) had its type updated from "int32" to "sfixed32".`,
		//`bar.proto:7:23:wire:Field "this" (2) had its oneof declaration updated from "only" to "none".`,
		//`bar.proto:8:17:wire:Field "that" (3) had its type updated from "int64" to "sfixed64".`,
		//`bar.proto:8:23:wire:Field "that" (3) had its oneof declaration updated from "only" to "none".`,
		//`bar.proto:12:9:wire:Message "Request" was removed.`,
		//`bar.proto:13:9:wire:Message "Response" was removed.`,
		//`bar.proto:16:19:wire:Method "Write" had its request type updated from "Request" to "bar.bar.Bar".`,
		//`bar.proto:16:37:wire:Method "Write" had its response type updated from "Response" to "bar.bar.Bar".`,
		//}

		//require.Len(t, expErrs, len(actErrs))

		//for i, err := range actErrs {
		//assert.Equal(t, err.String(), expErrs[i])
		//}
	})
}
