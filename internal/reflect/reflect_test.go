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

package reflect_test

import (
	"testing"

	"github.com/golang/protobuf/jsonpb"
	"github.com/stretchr/testify/require"
	"github.com/uber/prototool/internal/reflect"
	reflectv1 "github.com/uber/prototool/internal/reflect/gen/uber/proto/reflect/v1"
	ptesting "github.com/uber/prototool/internal/testing"
)

func TestBasic(t *testing.T) {
	fileDescriptorSets := ptesting.RequireGetFileDescriptorSets(t, "../cmd/testdata/reflect", "../cmd/testdata/reflect/one")
	packageSet, err := reflect.NewPackageSet(fileDescriptorSets...)
	require.NoError(t, err)
	ptesting.RequirePrintPackageSetJSON(t, packageSet)
}

func testUnmarshalPackageSet(t *testing.T, s string) *reflectv1.PackageSet {
	packageSet, err := unmarshalPackageSet(s)
	require.NoError(t, err)
	return packageSet
}

func unmarshalPackageSet(s string) (*reflectv1.PackageSet, error) {
	packageSet := &reflectv1.PackageSet{}
	if err := jsonpb.UnmarshalString(s, packageSet); err != nil {
		return nil, err
	}
	return packageSet, nil
}
