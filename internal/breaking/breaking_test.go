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

func TestRunOne(t *testing.T) {
	testRun(
		t,
		"one",
		false,
		false,
		newPackagesNotDeletedFailure("bar.v1"),
		newPackagesNoBetaDepsFailure("foo.v1", "bar.v1beta1"),
		newMessagesNotDeletedFailure("foo.v1.One.NestedOne.NestedNestedTwo"),
		newMessagesNotDeletedFailure("foo.v1.One.NestedTwo"),
		newMessagesNotDeletedFailure("foo.v1.Two"),
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
		newMessagesNotDeletedFailure("foo.v1.Five.FourEntry"),
		newMessageFieldsSameLabelFailure("foo.v1.Five", 1, "optional", "repeated"),
		newMessageFieldsSameLabelFailure("foo.v1.Five", 2, "optional", "repeated"),
		newMessageFieldsSameTypeFailure("foo.v1.Five", 2, "string", "message"),
		newMessageFieldsSameLabelFailure("foo.v1.Five", 3, "repeated", "optional"),
		newMessageFieldsSameLabelFailure("foo.v1.Five", 4, "repeated", "optional"),
		newMessageFieldsSameTypeFailure("foo.v1.Five", 4, "message", "int64"),
		newEnumsNotDeletedFailure("foo.v1.EnumTwo"),
		newEnumsNotDeletedFailure("foo.v1.Six.Foo"),
		newEnumsNotDeletedFailure("foo.v1.Six.NestedSix.Foo"),
		newEnumsNotDeletedFailure("foo.v1.Six.NestedSix.NestedNestedSix.Foo"),
		newMessagesNotDeletedFailure("foo.v1.Six.NestedSix.NestedNestedSixDelete"),
		newEnumValuesNotDeletedFailure("foo.v1.EnumSeven", 2),
		newEnumValuesNotDeletedFailure("foo.v1.Seven.EnumSeven", 2),
		newEnumValuesNotDeletedFailure("foo.v1.Seven.NestedSeven.EnumSeven", 2),
		newEnumValuesNotDeletedFailure("foo.v1.Seven.NestedSeven.NestedNestedSeven.EnumSeven", 2),
		newEnumValuesSameNameFailure("foo.v1.EnumSeven", 1, "ENUM_SEVEN_ONE", "ENUM_SEVEN_TWO"),
		newEnumValuesSameNameFailure("foo.v1.Seven.EnumSeven", 1, "ENUM_SEVEN_ONE", "ENUM_SEVEN_TWO"),
		newEnumValuesSameNameFailure("foo.v1.Seven.NestedSeven.EnumSeven", 1, "ENUM_SEVEN_ONE", "ENUM_SEVEN_TWO"),
		newEnumValuesSameNameFailure("foo.v1.Seven.NestedSeven.NestedNestedSeven.EnumSeven", 1, "ENUM_SEVEN_ONE", "ENUM_SEVEN_TWO"),
		newMessageOneofsNotDeletedFailure("foo.v1.Eight", "test"),
		newMessageOneofsNotDeletedFailure("foo.v1.Eight.NestedEight", "test"),
		newMessageOneofsNotDeletedFailure("foo.v1.Eight.NestedEight.NestedNestedEight", "test"),
		newServicesNotDeletedFailure("foo.v1.TwoAPI"),
		newServiceMethodsNotDeletedFailure("foo.v1.OneAPI", "OneTwo"),
		newMessagesNotDeletedFailure("foo.v1.OneTwoRequest"),
		newMessagesNotDeletedFailure("foo.v1.OneTwoResponse"),
		newMessageFieldsSameNameFailure("foo.v1.Nine", 1, "one", "two"),
		newMessageFieldsSameNameFailure("foo.v1.Nine.NestedNine", 1, "one", "two"),
		newMessageFieldsSameNameFailure("foo.v1.Nine.NestedNine.NestedNestedNine", 1, "one", "two"),
		newMessageOneofsFieldsNotRemovedFailure("foo.v1.Ten", "test", 3),
		newMessageOneofsFieldsNotRemovedFailure("foo.v1.Ten.NestedTen", "test", 3),
		newMessageOneofsFieldsNotRemovedFailure("foo.v1.Ten.NestedTen.NestedNestedTen", "test", 3),
		newServiceMethodsSameRequestTypeFailure("foo.v1.ThreeAPI", "ThreeOne", "foo.v1.ThreeOneRequest", "foo.v1.OneOneRequest"),
		newServiceMethodsSameResponseTypeFailure("foo.v1.ThreeAPI", "ThreeOne", "foo.v1.ThreeOneResponse", "foo.v1.OneOneResponse"),
		newMessagesNotDeletedFailure("foo.v1.ThreeOneRequest"),
		newMessagesNotDeletedFailure("foo.v1.ThreeOneResponse"),
		newServiceMethodsSameClientStreamingFailure("foo.v1.ThreeAPI", "ThreeTwo", true),
		newServiceMethodsSameClientStreamingFailure("foo.v1.ThreeAPI", "ThreeThree", false),
		newServiceMethodsSameServerStreamingFailure("foo.v1.ThreeAPI", "ThreeFour", true),
		newServiceMethodsSameServerStreamingFailure("foo.v1.ThreeAPI", "ThreeFive", false),
		newMessageFieldsSameOneofFailure("foo.v1.Eleven", 1, "test"),
		newMessageFieldsSameOneofFailure("foo.v1.Eleven.NestedEleven", 1, "test"),
		newMessageFieldsSameOneofFailure("foo.v1.Eleven.NestedEleven.NestedNestedEleven", 1, "test"),
	)
}

func TestRunOneIncludeBeta(t *testing.T) {
	testRun(
		t,
		"one",
		true,
		false,
		newPackagesNotDeletedFailure("bar.v1"),
		newPackagesNotDeletedFailure("foo.v1beta2"),
		newMessagesNotDeletedFailure("foo.v1.One.NestedOne.NestedNestedTwo"),
		newMessagesNotDeletedFailure("foo.v1.One.NestedTwo"),
		newMessagesNotDeletedFailure("foo.v1.Two"),
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
		newMessagesNotDeletedFailure("foo.v1.Five.FourEntry"),
		newMessageFieldsSameLabelFailure("foo.v1.Five", 1, "optional", "repeated"),
		newMessageFieldsSameLabelFailure("foo.v1.Five", 2, "optional", "repeated"),
		newMessageFieldsSameTypeFailure("foo.v1.Five", 2, "string", "message"),
		newMessageFieldsSameLabelFailure("foo.v1.Five", 3, "repeated", "optional"),
		newMessageFieldsSameLabelFailure("foo.v1.Five", 4, "repeated", "optional"),
		newMessageFieldsSameTypeFailure("foo.v1.Five", 4, "message", "int64"),
		newEnumsNotDeletedFailure("foo.v1.EnumTwo"),
		newEnumsNotDeletedFailure("foo.v1.Six.Foo"),
		newEnumsNotDeletedFailure("foo.v1.Six.NestedSix.Foo"),
		newEnumsNotDeletedFailure("foo.v1.Six.NestedSix.NestedNestedSix.Foo"),
		newMessagesNotDeletedFailure("foo.v1.Six.NestedSix.NestedNestedSixDelete"),
		newEnumValuesNotDeletedFailure("foo.v1.EnumSeven", 2),
		newEnumValuesNotDeletedFailure("foo.v1.Seven.EnumSeven", 2),
		newEnumValuesNotDeletedFailure("foo.v1.Seven.NestedSeven.EnumSeven", 2),
		newEnumValuesNotDeletedFailure("foo.v1.Seven.NestedSeven.NestedNestedSeven.EnumSeven", 2),
		newEnumValuesSameNameFailure("foo.v1.EnumSeven", 1, "ENUM_SEVEN_ONE", "ENUM_SEVEN_TWO"),
		newEnumValuesSameNameFailure("foo.v1.Seven.EnumSeven", 1, "ENUM_SEVEN_ONE", "ENUM_SEVEN_TWO"),
		newEnumValuesSameNameFailure("foo.v1.Seven.NestedSeven.EnumSeven", 1, "ENUM_SEVEN_ONE", "ENUM_SEVEN_TWO"),
		newEnumValuesSameNameFailure("foo.v1.Seven.NestedSeven.NestedNestedSeven.EnumSeven", 1, "ENUM_SEVEN_ONE", "ENUM_SEVEN_TWO"),
		newMessageOneofsNotDeletedFailure("foo.v1.Eight", "test"),
		newMessageOneofsNotDeletedFailure("foo.v1.Eight.NestedEight", "test"),
		newMessageOneofsNotDeletedFailure("foo.v1.Eight.NestedEight.NestedNestedEight", "test"),
		newServicesNotDeletedFailure("foo.v1.TwoAPI"),
		newServiceMethodsNotDeletedFailure("foo.v1.OneAPI", "OneTwo"),
		newMessagesNotDeletedFailure("foo.v1.OneTwoRequest"),
		newMessagesNotDeletedFailure("foo.v1.OneTwoResponse"),
		newMessageFieldsSameNameFailure("foo.v1.Nine", 1, "one", "two"),
		newMessageFieldsSameNameFailure("foo.v1.Nine.NestedNine", 1, "one", "two"),
		newMessageFieldsSameNameFailure("foo.v1.Nine.NestedNine.NestedNestedNine", 1, "one", "two"),
		newMessageOneofsFieldsNotRemovedFailure("foo.v1.Ten", "test", 3),
		newMessageOneofsFieldsNotRemovedFailure("foo.v1.Ten.NestedTen", "test", 3),
		newMessageOneofsFieldsNotRemovedFailure("foo.v1.Ten.NestedTen.NestedNestedTen", "test", 3),
		newServiceMethodsSameRequestTypeFailure("foo.v1.ThreeAPI", "ThreeOne", "foo.v1.ThreeOneRequest", "foo.v1.OneOneRequest"),
		newServiceMethodsSameResponseTypeFailure("foo.v1.ThreeAPI", "ThreeOne", "foo.v1.ThreeOneResponse", "foo.v1.OneOneResponse"),
		newMessagesNotDeletedFailure("foo.v1.ThreeOneRequest"),
		newMessagesNotDeletedFailure("foo.v1.ThreeOneResponse"),
		newServiceMethodsSameClientStreamingFailure("foo.v1.ThreeAPI", "ThreeTwo", true),
		newServiceMethodsSameClientStreamingFailure("foo.v1.ThreeAPI", "ThreeThree", false),
		newServiceMethodsSameServerStreamingFailure("foo.v1.ThreeAPI", "ThreeFour", true),
		newServiceMethodsSameServerStreamingFailure("foo.v1.ThreeAPI", "ThreeFive", false),
		newMessageFieldsSameOneofFailure("foo.v1.Eleven", 1, "test"),
		newMessageFieldsSameOneofFailure("foo.v1.Eleven.NestedEleven", 1, "test"),
		newMessageFieldsSameOneofFailure("foo.v1.Eleven.NestedEleven.NestedNestedEleven", 1, "test"),
		newMessagesNotDeletedFailure("foo.v1beta1.One.NestedOne.NestedNestedTwo"),
		newMessagesNotDeletedFailure("foo.v1beta1.One.NestedTwo"),
		newMessagesNotDeletedFailure("foo.v1beta1.Two"),
		newMessageFieldsNotDeletedFailure("foo.v1beta1.Three", 2),
		newMessageFieldsNotDeletedFailure("foo.v1beta1.Three.NestedThree", 2),
		newMessageFieldsNotDeletedFailure("foo.v1beta1.Three.NestedThree.NestedNestedThree", 2),
		newMessageFieldsSameTypeFailure("foo.v1beta1.Four", 1, "int64", "int32"),
		newMessageFieldsSameTypeFailure("foo.v1beta1.Four.NestedFour", 1, "int64", "int32"),
		newMessageFieldsSameTypeFailure("foo.v1beta1.Four.NestedFour.NestedNestedFour", 1, "int64", "int32"),
		newMessageFieldsSameTypeFailure("foo.v1beta1.Four", 2, "string", "bytes"),
		newMessageFieldsSameTypeFailure("foo.v1beta1.Four.NestedFour", 2, "string", "bytes"),
		newMessageFieldsSameTypeFailure("foo.v1beta1.Four.NestedFour.NestedNestedFour", 2, "string", "bytes"),
		newMessageFieldsSameTypeFailure("foo.v1beta1.Four", 3, "foo.v1beta1.Four.NestedFour", "foo.v1beta1.One"),
		newMessageFieldsSameTypeFailure("foo.v1beta1.Four.NestedFour", 3, "foo.v1beta1.Four.NestedFour.NestedNestedFour", "foo.v1beta1.One"),
		newMessageFieldsSameTypeFailure("foo.v1beta1.Four", 4, "foo.v1beta1.EnumOne", "foo.v1beta1.EnumThree"),
		newMessageFieldsSameTypeFailure("foo.v1beta1.Four.NestedFour", 4, "foo.v1beta1.EnumOne", "foo.v1beta1.EnumThree"),
		newMessageFieldsSameTypeFailure("foo.v1beta1.Four.NestedFour.NestedNestedFour", 4, "foo.v1beta1.EnumOne", "foo.v1beta1.EnumThree"),
		newMessageFieldsSameTypeFailure("foo.v1beta1.Four", 5, "enum", "double"),
		newMessageFieldsSameTypeFailure("foo.v1beta1.Four.NestedFour", 5, "enum", "double"),
		newMessageFieldsSameTypeFailure("foo.v1beta1.Four.NestedFour.NestedNestedFour", 5, "enum", "double"),
		newMessageFieldsSameTypeFailure("foo.v1beta1.Four", 6, "int64", "int32"),
		newMessageFieldsSameTypeFailure("foo.v1beta1.Four.NestedFour", 6, "int64", "int32"),
		newMessageFieldsSameTypeFailure("foo.v1beta1.Four.NestedFour.NestedNestedFour", 6, "int64", "int32"),
		newMessageFieldsSameTypeFailure("foo.v1beta1.Four.SevenEntry", 1, "int64", "int32"),
		newMessageFieldsSameTypeFailure("foo.v1beta1.Four.NestedFour.SevenEntry", 1, "int64", "int32"),
		newMessageFieldsSameTypeFailure("foo.v1beta1.Four.NestedFour.NestedNestedFour.SevenEntry", 1, "int64", "int32"),
		newMessagesNotDeletedFailure("foo.v1beta1.Five.FourEntry"),
		newMessageFieldsSameLabelFailure("foo.v1beta1.Five", 1, "optional", "repeated"),
		newMessageFieldsSameLabelFailure("foo.v1beta1.Five", 2, "optional", "repeated"),
		newMessageFieldsSameTypeFailure("foo.v1beta1.Five", 2, "string", "message"),
		newMessageFieldsSameLabelFailure("foo.v1beta1.Five", 3, "repeated", "optional"),
		newMessageFieldsSameLabelFailure("foo.v1beta1.Five", 4, "repeated", "optional"),
		newMessageFieldsSameTypeFailure("foo.v1beta1.Five", 4, "message", "int64"),
		newEnumsNotDeletedFailure("foo.v1beta1.EnumTwo"),
		newEnumsNotDeletedFailure("foo.v1beta1.Six.Foo"),
		newEnumsNotDeletedFailure("foo.v1beta1.Six.NestedSix.Foo"),
		newEnumsNotDeletedFailure("foo.v1beta1.Six.NestedSix.NestedNestedSix.Foo"),
		newMessagesNotDeletedFailure("foo.v1beta1.Six.NestedSix.NestedNestedSixDelete"),
		newEnumValuesNotDeletedFailure("foo.v1beta1.EnumSeven", 2),
		newEnumValuesNotDeletedFailure("foo.v1beta1.Seven.EnumSeven", 2),
		newEnumValuesNotDeletedFailure("foo.v1beta1.Seven.NestedSeven.EnumSeven", 2),
		newEnumValuesNotDeletedFailure("foo.v1beta1.Seven.NestedSeven.NestedNestedSeven.EnumSeven", 2),
		newEnumValuesSameNameFailure("foo.v1beta1.EnumSeven", 1, "ENUM_SEVEN_ONE", "ENUM_SEVEN_TWO"),
		newEnumValuesSameNameFailure("foo.v1beta1.Seven.EnumSeven", 1, "ENUM_SEVEN_ONE", "ENUM_SEVEN_TWO"),
		newEnumValuesSameNameFailure("foo.v1beta1.Seven.NestedSeven.EnumSeven", 1, "ENUM_SEVEN_ONE", "ENUM_SEVEN_TWO"),
		newEnumValuesSameNameFailure("foo.v1beta1.Seven.NestedSeven.NestedNestedSeven.EnumSeven", 1, "ENUM_SEVEN_ONE", "ENUM_SEVEN_TWO"),
		newMessageOneofsNotDeletedFailure("foo.v1beta1.Eight", "test"),
		newMessageOneofsNotDeletedFailure("foo.v1beta1.Eight.NestedEight", "test"),
		newMessageOneofsNotDeletedFailure("foo.v1beta1.Eight.NestedEight.NestedNestedEight", "test"),
		newServicesNotDeletedFailure("foo.v1beta1.TwoAPI"),
		newServiceMethodsNotDeletedFailure("foo.v1beta1.OneAPI", "OneTwo"),
		newMessagesNotDeletedFailure("foo.v1beta1.OneTwoRequest"),
		newMessagesNotDeletedFailure("foo.v1beta1.OneTwoResponse"),
		newMessageFieldsSameNameFailure("foo.v1beta1.Nine", 1, "one", "two"),
		newMessageFieldsSameNameFailure("foo.v1beta1.Nine.NestedNine", 1, "one", "two"),
		newMessageFieldsSameNameFailure("foo.v1beta1.Nine.NestedNine.NestedNestedNine", 1, "one", "two"),
		newMessageOneofsFieldsNotRemovedFailure("foo.v1beta1.Ten", "test", 3),
		newMessageOneofsFieldsNotRemovedFailure("foo.v1beta1.Ten.NestedTen", "test", 3),
		newMessageOneofsFieldsNotRemovedFailure("foo.v1beta1.Ten.NestedTen.NestedNestedTen", "test", 3),
		newServiceMethodsSameRequestTypeFailure("foo.v1beta1.ThreeAPI", "ThreeOne", "foo.v1beta1.ThreeOneRequest", "foo.v1beta1.OneOneRequest"),
		newServiceMethodsSameResponseTypeFailure("foo.v1beta1.ThreeAPI", "ThreeOne", "foo.v1beta1.ThreeOneResponse", "foo.v1beta1.OneOneResponse"),
		newMessagesNotDeletedFailure("foo.v1beta1.ThreeOneRequest"),
		newMessagesNotDeletedFailure("foo.v1beta1.ThreeOneResponse"),
		newServiceMethodsSameClientStreamingFailure("foo.v1beta1.ThreeAPI", "ThreeTwo", true),
		newServiceMethodsSameClientStreamingFailure("foo.v1beta1.ThreeAPI", "ThreeThree", false),
		newServiceMethodsSameServerStreamingFailure("foo.v1beta1.ThreeAPI", "ThreeFour", true),
		newServiceMethodsSameServerStreamingFailure("foo.v1beta1.ThreeAPI", "ThreeFive", false),
	)
}

func TestRunOneAllowBetaDeps(t *testing.T) {
	testRun(
		t,
		"one",
		false,
		true,
		newPackagesNotDeletedFailure("bar.v1"),
		newMessagesNotDeletedFailure("foo.v1.One.NestedOne.NestedNestedTwo"),
		newMessagesNotDeletedFailure("foo.v1.One.NestedTwo"),
		newMessagesNotDeletedFailure("foo.v1.Two"),
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
		newMessagesNotDeletedFailure("foo.v1.Five.FourEntry"),
		newMessageFieldsSameLabelFailure("foo.v1.Five", 1, "optional", "repeated"),
		newMessageFieldsSameLabelFailure("foo.v1.Five", 2, "optional", "repeated"),
		newMessageFieldsSameTypeFailure("foo.v1.Five", 2, "string", "message"),
		newMessageFieldsSameLabelFailure("foo.v1.Five", 3, "repeated", "optional"),
		newMessageFieldsSameLabelFailure("foo.v1.Five", 4, "repeated", "optional"),
		newMessageFieldsSameTypeFailure("foo.v1.Five", 4, "message", "int64"),
		newEnumsNotDeletedFailure("foo.v1.EnumTwo"),
		newEnumsNotDeletedFailure("foo.v1.Six.Foo"),
		newEnumsNotDeletedFailure("foo.v1.Six.NestedSix.Foo"),
		newEnumsNotDeletedFailure("foo.v1.Six.NestedSix.NestedNestedSix.Foo"),
		newMessagesNotDeletedFailure("foo.v1.Six.NestedSix.NestedNestedSixDelete"),
		newEnumValuesNotDeletedFailure("foo.v1.EnumSeven", 2),
		newEnumValuesNotDeletedFailure("foo.v1.Seven.EnumSeven", 2),
		newEnumValuesNotDeletedFailure("foo.v1.Seven.NestedSeven.EnumSeven", 2),
		newEnumValuesNotDeletedFailure("foo.v1.Seven.NestedSeven.NestedNestedSeven.EnumSeven", 2),
		newEnumValuesSameNameFailure("foo.v1.EnumSeven", 1, "ENUM_SEVEN_ONE", "ENUM_SEVEN_TWO"),
		newEnumValuesSameNameFailure("foo.v1.Seven.EnumSeven", 1, "ENUM_SEVEN_ONE", "ENUM_SEVEN_TWO"),
		newEnumValuesSameNameFailure("foo.v1.Seven.NestedSeven.EnumSeven", 1, "ENUM_SEVEN_ONE", "ENUM_SEVEN_TWO"),
		newEnumValuesSameNameFailure("foo.v1.Seven.NestedSeven.NestedNestedSeven.EnumSeven", 1, "ENUM_SEVEN_ONE", "ENUM_SEVEN_TWO"),
		newMessageOneofsNotDeletedFailure("foo.v1.Eight", "test"),
		newMessageOneofsNotDeletedFailure("foo.v1.Eight.NestedEight", "test"),
		newMessageOneofsNotDeletedFailure("foo.v1.Eight.NestedEight.NestedNestedEight", "test"),
		newServicesNotDeletedFailure("foo.v1.TwoAPI"),
		newServiceMethodsNotDeletedFailure("foo.v1.OneAPI", "OneTwo"),
		newMessagesNotDeletedFailure("foo.v1.OneTwoRequest"),
		newMessagesNotDeletedFailure("foo.v1.OneTwoResponse"),
		newMessageFieldsSameNameFailure("foo.v1.Nine", 1, "one", "two"),
		newMessageFieldsSameNameFailure("foo.v1.Nine.NestedNine", 1, "one", "two"),
		newMessageFieldsSameNameFailure("foo.v1.Nine.NestedNine.NestedNestedNine", 1, "one", "two"),
		newMessageOneofsFieldsNotRemovedFailure("foo.v1.Ten", "test", 3),
		newMessageOneofsFieldsNotRemovedFailure("foo.v1.Ten.NestedTen", "test", 3),
		newMessageOneofsFieldsNotRemovedFailure("foo.v1.Ten.NestedTen.NestedNestedTen", "test", 3),
		newServiceMethodsSameRequestTypeFailure("foo.v1.ThreeAPI", "ThreeOne", "foo.v1.ThreeOneRequest", "foo.v1.OneOneRequest"),
		newServiceMethodsSameResponseTypeFailure("foo.v1.ThreeAPI", "ThreeOne", "foo.v1.ThreeOneResponse", "foo.v1.OneOneResponse"),
		newMessagesNotDeletedFailure("foo.v1.ThreeOneRequest"),
		newMessagesNotDeletedFailure("foo.v1.ThreeOneResponse"),
		newServiceMethodsSameClientStreamingFailure("foo.v1.ThreeAPI", "ThreeTwo", true),
		newServiceMethodsSameClientStreamingFailure("foo.v1.ThreeAPI", "ThreeThree", false),
		newServiceMethodsSameServerStreamingFailure("foo.v1.ThreeAPI", "ThreeFour", true),
		newServiceMethodsSameServerStreamingFailure("foo.v1.ThreeAPI", "ThreeFive", false),
		newMessageFieldsSameOneofFailure("foo.v1.Eleven", 1, "test"),
		newMessageFieldsSameOneofFailure("foo.v1.Eleven.NestedEleven", 1, "test"),
		newMessageFieldsSameOneofFailure("foo.v1.Eleven.NestedEleven.NestedNestedEleven", 1, "test"),
	)
}

func testRun(t *testing.T, subDirPath string, includeBeta bool, allowBetaDeps bool, expectedFailures ...*text.Failure) {
	fromPackageSet, toPackageSet, err := getPackageSets(subDirPath)
	require.NoError(t, err)
	runner := NewRunner()
	if includeBeta {
		runner = NewRunner(RunnerWithIncludeBeta())
	}
	if allowBetaDeps {
		runner = NewRunner(RunnerWithAllowBetaDeps())
	}
	failures, err := runner.Run(fromPackageSet, toPackageSet)
	require.NoError(t, err)
	for _, failure := range failures {
		failure.LintID = ""
	}
	text.SortFailures(failures)
	text.SortFailures(expectedFailures)
	require.Equal(t, expectedFailures, failures)
}

func getPackageSets(subDirPath string) (*extract.PackageSet, *extract.PackageSet, error) {
	fromFileDescriptorSets, err := ptesting.GetFileDescriptorSets(".", "testdata/"+subDirPath+"/from")
	if err != nil {
		return nil, nil, err
	}
	fromReflectPackageSet, err := reflect.NewPackageSet(fromFileDescriptorSets.Unwrap()...)
	if err != nil {
		return nil, nil, err
	}
	fromPackageSet, err := extract.NewPackageSet(fromReflectPackageSet)
	if err != nil {
		return nil, nil, err
	}
	toFileDescriptorSets, err := ptesting.GetFileDescriptorSets(".", "testdata/"+subDirPath+"/to")
	if err != nil {
		return nil, nil, err
	}
	toReflectPackageSet, err := reflect.NewPackageSet(toFileDescriptorSets.Unwrap()...)
	if err != nil {
		return nil, nil, err
	}
	toPackageSet, err := extract.NewPackageSet(toReflectPackageSet)
	if err != nil {
		return nil, nil, err
	}
	return fromPackageSet, toPackageSet, nil
}
