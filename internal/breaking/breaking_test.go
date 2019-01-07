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

package breaking

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/uber/prototool/internal/extract"
	"github.com/uber/prototool/internal/reflect"
	ptesting "github.com/uber/prototool/internal/testing"
	"github.com/uber/prototool/internal/text"
)

func TestRun(t *testing.T) {
	testRun(
		t,
		"one",
		newMessageFieldsNotDeletedFailure("foo.v1.Three", 2),
		newMessageFieldsNotDeletedFailure("foo.v1.Three.NestedThree", 2),
		newMessageFieldsNotDeletedFailure("foo.v1.Three.NestedThree.NestedNestedThree", 2),
		newMessageFieldsSameTypeFailure("foo.v1.Four", 1, "int64", "int32"),
		newMessageFieldsSameTypeFailure("foo.v1.Four.NestedFour", 1, "int64", "int32"),
		newMessageFieldsSameTypeFailure("foo.v1.Four.NestedFour.NestedNestedFour", 1, "int64", "int32"),
		newMessageFieldsSameTypeFailure("foo.v1.Four", 2, "string", "bytes"),
		newMessageFieldsSameTypeFailure("foo.v1.Four.NestedFour", 2, "string", "bytes"),
		newMessageFieldsSameTypeFailure("foo.v1.Four.NestedFour.NestedNestedFour", 2, "string", "bytes"),
		newMessageFieldsSameTypeFailure("foo.v1.Four", 3, "foo.v1.Four.NestedFour", "foo.v1.One"),
		newMessageFieldsSameTypeFailure("foo.v1.Four.NestedFour", 3, "foo.v1.Four.NestedFour.NestedNestedFour", "foo.v1.One"),
		newMessageFieldsSameTypeFailure("foo.v1.Four", 4, "foo.v1.EnumOne", "foo.v1.EnumThree"),
		newMessageFieldsSameTypeFailure("foo.v1.Four.NestedFour", 4, "foo.v1.EnumOne", "foo.v1.EnumThree"),
		newMessageFieldsSameTypeFailure("foo.v1.Four.NestedFour.NestedNestedFour", 4, "foo.v1.EnumOne", "foo.v1.EnumThree"),
		newMessageFieldsSameTypeFailure("foo.v1.Four", 5, "enum", "double"),
		newMessageFieldsSameTypeFailure("foo.v1.Four.NestedFour", 5, "enum", "double"),
		newMessageFieldsSameTypeFailure("foo.v1.Four.NestedFour.NestedNestedFour", 5, "enum", "double"),
		newMessageFieldsSameTypeFailure("foo.v1.Four", 6, "int64", "int32"),
		newMessageFieldsSameTypeFailure("foo.v1.Four.NestedFour", 6, "int64", "int32"),
		newMessageFieldsSameTypeFailure("foo.v1.Four.NestedFour.NestedNestedFour", 6, "int64", "int32"),
		newMessageFieldsSameTypeFailure("foo.v1.Four.SevenEntry", 1, "int64", "int32"),
		newMessageFieldsSameTypeFailure("foo.v1.Four.NestedFour.SevenEntry", 1, "int64", "int32"),
		newMessageFieldsSameTypeFailure("foo.v1.Four.NestedFour.NestedNestedFour.SevenEntry", 1, "int64", "int32"),
		newMessagesNotDeletedFailure("foo.v1.One.NestedOne.NestedNestedTwo"),
		newMessagesNotDeletedFailure("foo.v1.One.NestedTwo"),
		newMessagesNotDeletedFailure("foo.v1.Two"),
		newPackagesNotDeletedFailure("bar.v1"),
	)
}

func testRun(t *testing.T, subDirPath string, expectedFailures ...*text.Failure) {
	fromPackageSet, toPackageSet, err := getPackageSets(subDirPath)
	require.NoError(t, err)
	failures, err := NewRunner().Run(fromPackageSet, toPackageSet)
	require.NoError(t, err)
	for _, failure := range failures {
		failure.LintID = ""
	}
	text.SortFailures(failures)
	text.SortFailures(expectedFailures)
	require.Equal(t, expectedFailures, failures)
}

func getPackageSets(subDirPath string) (*extract.PackageSet, *extract.PackageSet, error) {
	fromFileDescriptorSets, err := ptesting.GetFileDescriptorSets("../cmd/testdata/breaking", "../cmd/testdata/breaking/"+subDirPath+"/from")
	if err != nil {
		return nil, nil, err
	}
	fromReflectPackageSet, err := reflect.NewPackageSet(fromFileDescriptorSets...)
	if err != nil {
		return nil, nil, err
	}
	fromPackageSet, err := extract.NewPackageSet(fromReflectPackageSet)
	if err != nil {
		return nil, nil, err
	}
	toFileDescriptorSets, err := ptesting.GetFileDescriptorSets("../cmd/testdata/breaking", "../cmd/testdata/breaking/"+subDirPath+"/to")
	if err != nil {
		return nil, nil, err
	}
	toReflectPackageSet, err := reflect.NewPackageSet(toFileDescriptorSets...)
	if err != nil {
		return nil, nil, err
	}
	toPackageSet, err := extract.NewPackageSet(toReflectPackageSet)
	if err != nil {
		return nil, nil, err
	}
	return fromPackageSet, toPackageSet, nil
}
