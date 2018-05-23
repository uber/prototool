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

package lint

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/emicklei/proto"
	"github.com/uber/prototool/internal/text"
	"github.com/uber/prototool/internal/x/file"
	"github.com/uber/prototool/internal/x/settings"
	"go.uber.org/zap"
)

var (
	// AllCheckers is the slice of all known Checkers.
	AllCheckers = []Checker{
		commentsNoCStyleChecker,
		enumFieldNamesUppercaseChecker,
		enumFieldNamesUpperSnakeCaseChecker,
		enumFieldPrefixesChecker,
		enumNamesCamelCaseChecker,
		enumNamesCapitalizedChecker,
		enumZeroValuesInvalidChecker,
		enumsHaveCommentsChecker,
		fileOptionsEqualGoPackagePbSuffixChecker,
		fileOptionsEqualJavaMultipleFilesTrueChecker,
		fileOptionsEqualJavaPackageComPbChecker,
		fileOptionsGoPackageSameInDirChecker,
		fileOptionsJavaPackageSameInDirChecker,
		fileOptionsRequireGoPackageChecker,
		fileOptionsRequireJavaMultipleFilesChecker,
		fileOptionsRequireJavaPackageChecker,
		messageFieldsNotFloatsChecker,
		messageFieldNamesLowerSnakeCaseChecker,
		messageFieldNamesLowercaseChecker,
		messageNamesCamelCaseChecker,
		messageNamesCapitalizedChecker,
		messagesHaveCommentsChecker,
		messagesHaveCommentsExceptRequestResponseTypesChecker,
		oneofNamesLowerSnakeCaseChecker,
		packageIsDeclaredChecker,
		packageLowerSnakeCaseChecker,
		packagesSameInDirChecker,
		rpcsHaveCommentsChecker,
		rpcNamesCamelCaseChecker,
		rpcNamesCapitalizedChecker,
		requestResponseTypesInSameFileChecker,
		requestResponseTypesUniqueChecker,
		requestResponseNamesMatchRPCChecker,
		servicesHaveCommentsChecker,
		serviceNamesCamelCaseChecker,
		serviceNamesCapitalizedChecker,
		syntaxProto3Checker,
		wktDirectlyImportedChecker,
	}

	// DefaultCheckers is the slice of default Checkers.
	DefaultCheckers = copyCheckersWithout(
		AllCheckers,
		enumFieldNamesUppercaseChecker,
		enumsHaveCommentsChecker,
		fileOptionsEqualJavaMultipleFilesTrueChecker,
		fileOptionsEqualJavaOuterClassnameProtoSuffixChecker,
		fileOptionsRequireJavaMultipleFilesChecker,
		fileOptionsRequireJavaOuterClassnameChecker,
		messageFieldsNotFloatsChecker,
		messagesHaveCommentsChecker,
		messagesHaveCommentsExceptRequestResponseTypesChecker,
		messageFieldNamesLowercaseChecker,
		requestResponseNamesMatchRPCChecker,
		rpcsHaveCommentsChecker,
		servicesHaveCommentsChecker,
	)

	// DefaultGroup is the default group.
	DefaultGroup = "default"

	// AllGroup is the group of all known linters.
	AllGroup = "all"

	// GroupToCheckers is the map from checker group to the corresponding slice of checkers.
	GroupToCheckers = map[string][]Checker{
		DefaultGroup: DefaultCheckers,
		AllGroup:     AllCheckers,
	}
)

func init() {
	ids := make(map[string]struct{})
	for _, checker := range AllCheckers {
		if _, ok := ids[checker.ID()]; ok {
			panic(fmt.Sprintf("duplicate checker id %s", checker.ID()))
		}
		ids[checker.ID()] = struct{}{}
	}
}

// Runner runs a lint job.
type Runner interface {
	Run(...*file.ProtoSet) ([]*text.Failure, error)
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

// The below should not be needed in the CLI
// TODO make private

// Checker is a linter for Protobuf files.
type Checker interface {
	// Return the ID of this Checker. This should be all UPPER_SNAKE_CASE.
	ID() string
	// Return the purpose of this Checker. This should be a human-readable string.
	Purpose() string
	// Check the file data for the descriptors in a common directgory.
	// If there is a lint failure, this returns it in the
	// slice and does not return an error. An error is returned if something
	// unexpected happens. Callers should verify the files are compilable
	// before running this.
	Check(dirPath string, descriptors []*proto.Proto) ([]*text.Failure, error)
}

// NewChecker is a convenience function that returns a new Checker for the
// given parameters.
//
// The ID will be upper-cased.
//
// Failures returned from check do not need to set the ID, this will be overwritten.
func NewChecker(id string, purpose string, check func(string, []*proto.Proto) ([]*text.Failure, error)) Checker {
	return newBaseChecker(id, purpose, check)
}

// NewAddChecker is a convenience function that returns a new Checker for the
// given parameters, using a function to record failures.
//
// The ID will be upper-cased.
//
// Failures returned from check do not need to set the ID, this will be overwritten.
func NewAddChecker(id string, purpose string, addCheck func(func(*text.Failure), string, []*proto.Proto) error) Checker {
	return newBaseAddChecker(id, purpose, addCheck)
}

// GetCheckers returns the Checkers for the LintConfig.
//
// The config is expected to be valid, ie slices deduped, all upper-case,
// and only either IDs or Group/IncludeIDs/ExcludeIDs, with no overlap between
// IncludeIDs and ExcludeIDs.
//
// If the config came from the settings package, this is already validated.
func GetCheckers(config settings.LintConfig) ([]Checker, error) {
	if len(config.IDs) == 0 && (len(config.Group) == 0 || config.Group == DefaultGroup) && len(config.IncludeIDs) == 0 && len(config.ExcludeIDs) == 0 {
		return DefaultCheckers, nil
	}

	if len(config.IDs) > 0 {
		var checkers []Checker
		// n^2 woot
		for _, checker := range AllCheckers {
			for _, id := range config.IDs {
				if checker.ID() == id {
					checkers = append(checkers, checker)
				}
			}
		}
		return checkers, nil
	}

	baseCheckers := DefaultCheckers
	var ok bool
	if len(config.Group) > 0 && config.Group != DefaultGroup {
		baseCheckers, ok = GroupToCheckers[config.Group]
		if !ok {
			return nil, fmt.Errorf("unknown lint group: %s", config.Group)
		}
	}

	checkersMap := make(map[string]Checker, len(baseCheckers))
	for _, checker := range baseCheckers {
		checkersMap[checker.ID()] = checker
	}
	for _, excludeID := range config.ExcludeIDs {
		delete(checkersMap, excludeID)
	}
	// n^2 woot
	for _, checker := range AllCheckers {
		for _, id := range config.IncludeIDs {
			if checker.ID() == id {
				checkersMap[checker.ID()] = checker
			}
		}
	}
	checkers := make([]Checker, 0, len(checkersMap))
	for _, checker := range checkersMap {
		checkers = append(checkers, checker)
	}
	return checkers, nil
}

// GetDirPathToDescriptors is a convenience function that gets the
// descriptors for the given ProtoSet.
func GetDirPathToDescriptors(protoSet *file.ProtoSet) (map[string][]*proto.Proto, error) {
	dirPathToDescriptors := make(map[string][]*proto.Proto, len(protoSet.DirPathToFiles))
	for dirPath, protoFiles := range protoSet.DirPathToFiles {
		descriptors := make([]*proto.Proto, len(protoFiles))
		for i, protoFile := range protoFiles {
			file, err := os.Open(protoFile.Path)
			if err != nil {
				return nil, err
			}
			parser := proto.NewParser(file)
			parser.Filename(protoFile.DisplayPath)
			descriptor, err := parser.Parse()
			_ = file.Close()
			if err != nil {
				return nil, err
			}
			descriptors[i] = descriptor
		}
		dirPathToDescriptors[dirPath] = descriptors
	}
	return dirPathToDescriptors, nil
}

// CheckMultiple is a convenience function that checks multiple checkers and multiple descriptors.
func CheckMultiple(checkers []Checker, dirPathToDescriptors map[string][]*proto.Proto, ignoreIDToFilePaths map[string][]string) ([]*text.Failure, error) {
	var allFailures []*text.Failure
	for dirPath, descriptors := range dirPathToDescriptors {
		for _, checker := range checkers {
			failures, err := checkOne(checker, dirPath, descriptors, ignoreIDToFilePaths)
			if err != nil {
				return nil, err
			}
			allFailures = append(allFailures, failures...)
		}
	}
	text.SortFailures(allFailures)
	return allFailures, nil
}

func checkOne(checker Checker, dirPath string, descriptors []*proto.Proto, ignoreIDToFilePaths map[string][]string) ([]*text.Failure, error) {
	filteredDescriptors, err := filterIgnores(checker, descriptors, ignoreIDToFilePaths)
	if err != nil {
		return nil, err
	}
	return checker.Check(dirPath, filteredDescriptors)
}

func filterIgnores(checker Checker, descriptors []*proto.Proto, ignoreIDToFilePaths map[string][]string) ([]*proto.Proto, error) {
	var filteredDescriptors []*proto.Proto
	for _, descriptor := range descriptors {
		ignore, err := shouldIgnore(checker, descriptor, ignoreIDToFilePaths)
		if err != nil {
			return nil, err
		}
		if !ignore {
			filteredDescriptors = append(filteredDescriptors, descriptor)
		}
	}
	return filteredDescriptors, nil
}

func shouldIgnore(checker Checker, descriptor *proto.Proto, ignoreIDToFilePaths map[string][]string) (bool, error) {
	filePath := descriptor.Filename
	var err error
	if !filepath.IsAbs(filePath) {
		filePath, err = filepath.Abs(filePath)
		if err != nil {
			return false, err
		}
	}
	ignoreFilePaths, ok := ignoreIDToFilePaths[checker.ID()]
	if !ok {
		return false, nil
	}
	for _, ignoreFilePath := range ignoreFilePaths {
		if filePath == ignoreFilePath {
			return true, nil
		}
	}
	return false, nil
}

func copyCheckersWithout(checkers []Checker, remove ...Checker) []Checker {
	c := make([]Checker, 0, len(checkers))
	for _, checker := range checkers {
		if !checkerIn(checker, remove) {
			c = append(c, checker)
		}
	}
	return c
}

func checkerIn(checker Checker, s []Checker) bool {
	for _, e := range s {
		if e == checker {
			return true
		}
	}
	return false
}
