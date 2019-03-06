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

package testing

import (
	"errors"
	"fmt"
	"testing"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/stretchr/testify/require"
	"github.com/uber/prototool/internal/file"
	"github.com/uber/prototool/internal/protoc"
	reflectv1 "github.com/uber/prototool/internal/reflect/gen/uber/proto/reflect/v1"
	"go.uber.org/multierr"
)

var jsonMarshaler = &jsonpb.Marshaler{Indent: "  "}

// RequireGetFileDescriptorSets calls GetDescriptorFileSets with require calls.
func RequireGetFileDescriptorSets(t *testing.T, workDirPath string, dirPath string) protoc.FileDescriptorSets {
	fileDescriptorSets, err := GetFileDescriptorSets(workDirPath, dirPath)
	require.NoError(t, err)
	require.NotEmpty(t, fileDescriptorSets)
	return fileDescriptorSets
}

// GetFileDescriptorSets gets the FileDescriptorSets that result from compiling the given dirPath.
func GetFileDescriptorSets(workDirPath string, dirPath string) (protoc.FileDescriptorSets, error) {
	protoSet, err := file.NewProtoSetProvider().GetForDir(workDirPath, dirPath)
	if err != nil {
		return nil, err
	}
	compileResult, err := protoc.NewCompiler(
		protoc.CompilerWithFileDescriptorSet(),
	).Compile(protoSet)
	if err != nil {
		return nil, err
	}
	if len(compileResult.Failures) > 0 {
		var err error
		for _, failure := range compileResult.Failures {
			err = multierr.Append(err, errors.New(failure.String()))
		}
		return nil, err
	}
	return compileResult.FileDescriptorSets, nil
}

// RequirePrintFileDescriptorSetsJSON calls PrintFileDescriptorSetsJSON with require calls.
func RequirePrintFileDescriptorSetsJSON(t *testing.T, fileDescriptorSets []*descriptor.FileDescriptorSet) {
	require.NoError(t, PrintFileDescriptorSetsJSON(fileDescriptorSets))
}

// PrintFileDescriptorSetsJSON prints the FileDescriptorSets to stdout.
func PrintFileDescriptorSetsJSON(fileDescriptorSets []*descriptor.FileDescriptorSet) error {
	for _, fileDescriptorSet := range fileDescriptorSets {
		s, err := jsonMarshaler.MarshalToString(fileDescriptorSet)
		if err != nil {
			return err
		}
		fmt.Println(s)
	}
	return nil
}

// RequirePrintPackageSetJSON calls PrintPackageSetJSON with require calls.
func RequirePrintPackageSetJSON(t *testing.T, packageSet *reflectv1.PackageSet) {
	require.NoError(t, PrintPackageSetJSON(packageSet))
}

// PrintPackageSetJSON prints the PackageSet to stdout.
func PrintPackageSetJSON(packageSet *reflectv1.PackageSet) error {
	s, err := jsonMarshaler.MarshalToString(packageSet)
	if err != nil {
		return err
	}
	fmt.Println(s)
	return nil
}
