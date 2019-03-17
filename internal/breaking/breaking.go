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
	"github.com/uber/prototool/internal/extract"
	"github.com/uber/prototool/internal/settings"
	"github.com/uber/prototool/internal/text"
	"go.uber.org/zap"
)

var (
	// AllCheckers are all known Checkers.
	//
	// Purposely not configurable - there are some dependencies between linters, for example if a message is deleted,
	// ENUMS_NOT_DELETED will not print out any nested enums that were deleted.
	AllCheckers = []Checker{
		{
			ID:      "ENUMS_NOT_DELETED",
			Purpose: "Checks that no enums have been deleted.",
			Check:   checkEnumsNotDeleted,
		},
		{
			ID:      "ENUM_VALUES_NOT_DELETED",
			Purpose: "Checks that no enum values have been deleted.",
			Check:   checkEnumValuesNotDeleted,
		},
		{
			ID:      "ENUM_VALUES_SAME_NAME",
			Purpose: "Checks that enum values have the same name.",
			Check:   checkEnumValuesSameName,
		},
		{
			ID:      "MESSAGES_NOT_DELETED",
			Purpose: "Checks that no messages have been deleted.",
			Check:   checkMessagesNotDeleted,
		},
		{
			ID:      "MESSAGE_FIELDS_NOT_DELETED",
			Purpose: "Checks that no message fields have been deleted.",
			Check:   checkMessageFieldsNotDeleted,
		},
		{
			ID:      "MESSAGE_FIELDS_SAME_LABEL",
			Purpose: "Checks that message fields have the same label.",
			Check:   checkMessageFieldsSameLabel,
		},
		{
			ID:      "MESSAGE_FIELDS_SAME_NAME",
			Purpose: "Checks that message fields have the same name.",
			Check:   checkMessageFieldsSameName,
		},
		{
			ID:      "MESSAGE_FIELDS_SAME_ONEOF",
			Purpose: "Checks that message fields that were not in a oneof are not now in a oneof.",
			Check:   checkMessageFieldsSameOneof,
		},
		{
			ID:      "MESSAGE_FIELDS_SAME_TYPE",
			Purpose: "Checks that message fields have the same type.",
			Check:   checkMessageFieldsSameType,
		},
		{
			ID:      "MESSAGE_ONEOFS_NOT_DELETED",
			Purpose: "Checks that no message oneofs have been deleted.",
			Check:   checkMessageOneofsNotDeleted,
		},
		{
			ID:      "MESSAGE_ONEOFS_FIELDS_NOT_REMOVED",
			Purpose: "Checks that no message oneofs have fields removed.",
			Check:   checkMessageOneofsFieldsNotRemoved,
		},
		{
			ID:      "PACKAGES_NOT_DELETED",
			Purpose: "Checks that no packages have been deleted.",
			Check:   checkPackagesNotDeleted,
		},
		{
			ID:      "SERVICES_NOT_DELETED",
			Purpose: "Checks that no services have been deleted.",
			Check:   checkServicesNotDeleted,
		},
		{
			ID:      "SERVICE_METHODS_NOT_DELETED",
			Purpose: "Checks that no service methods have been deleted.",
			Check:   checkServiceMethodsNotDeleted,
		},
		{
			ID:      "SERVICE_METHODS_SAME_CLIENT_STREAMING",
			Purpose: "Checks that service methods have the same client streaming.",
			Check:   checkServiceMethodsSameClientStreaming,
		},
		{
			ID:      "SERVICE_METHODS_SAME_REQUEST_TYPE",
			Purpose: "Checks that service methods have the same request type.",
			Check:   checkServiceMethodsSameRequestType,
		},
		{
			ID:      "SERVICE_METHODS_SAME_RESPONSE_TYPE",
			Purpose: "Checks that service methods have the same response type.",
			Check:   checkServiceMethodsSameResponseType,
		},
		{
			ID:      "SERVICE_METHODS_SAME_SERVER_STREAMING",
			Purpose: "Checks that service methods have the same server streaming.",
			Check:   checkServiceMethodsSameServerStreaming,
		},
	}

	// PackagesNoBetaDepsChecker is a special checker that verifies no stable packages
	// import beta packages.
	PackagesNoBetaDepsChecker = Checker{
		ID:      "PACKAGES_NO_BETA_DEPS",
		Purpose: "Checks that stable packages do not have beta dependencies.",
		Check:   checkPackagesNoBetaDeps,
	}
)

// Checker checks compatibility.
type Checker struct {
	// The ID of this Checker. This should be all UPPER_SNAKE_CASE.
	ID string
	// The purpose of this Checker. This should be a human-readable string.
	Purpose string
	// Check the compatibility of from and to.
	//
	// Returns an error only if there is a system error.
	Check func(addFailure func(*text.Failure), from *extract.PackageSet, to *extract.PackageSet) error
}

// Runner runs a series of Checkers.
type Runner interface {
	// Run runs Check on all the associated Checkers.
	//
	// Returns Failures if there are incompatibilities, or error if there is
	// a system error
	Run(config settings.BreakConfig, from *extract.PackageSet, to *extract.PackageSet) ([]*text.Failure, error)
}

// RunnerOption is an option for a new Runner.
type RunnerOption func(*runner)

// RunnerWithLogger returns a RunnerOption that uses the given logger.
//
// The default is to use zap.NewNop().
func RunnerWithLogger(logger *zap.Logger) RunnerOption {
	return func(runner *runner) {
		runner.logger = logger
	}
}

// NewRunner returns a new Runner.
func NewRunner(options ...RunnerOption) Runner {
	return newRunner(options...)
}
