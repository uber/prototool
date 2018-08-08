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
	"github.com/uber/prototool/internal/file"
	"github.com/uber/prototool/internal/settings"
	"github.com/uber/prototool/internal/text"
	"go.uber.org/zap"
)

var (
	// AllLinters is the slice of all known Linters.
	AllLinters = []Linter{
		commentsNoCStyleLinter,
		enumFieldNamesUppercaseLinter,
		enumFieldNamesUpperSnakeCaseLinter,
		enumFieldPrefixesLinter,
		enumNamesCamelCaseLinter,
		enumNamesCapitalizedLinter,
		enumZeroValuesInvalidLinter,
		enumsHaveCommentsLinter,
		enumsNoAllowAliasLinter,
		fileOptionsEqualGoPackagePbSuffixLinter,
		fileOptionsEqualJavaMultipleFilesTrueLinter,
		fileOptionsEqualJavaOuterClassnameProtoSuffixLinter,
		fileOptionsEqualJavaPackageComPrefixLinter,
		fileOptionsGoPackageNotLongFormLinter,
		fileOptionsGoPackageSameInDirLinter,
		fileOptionsJavaMultipleFilesSameInDirLinter,
		fileOptionsJavaPackageSameInDirLinter,
		fileOptionsRequireGoPackageLinter,
		fileOptionsRequireJavaMultipleFilesLinter,
		fileOptionsRequireJavaOuterClassnameLinter,
		fileOptionsRequireJavaPackageLinter,
		fileOptionsUnsetJavaMultipleFilesLinter,
		fileOptionsUnsetJavaOuterClassnameLinter,
		messageFieldsNotFloatsLinter,
		messageFieldNamesLowerSnakeCaseLinter,
		messageFieldNamesLowercaseLinter,
		messageNamesCamelCaseLinter,
		messageNamesCapitalizedLinter,
		messagesHaveCommentsLinter,
		messagesHaveCommentsExceptRequestResponseTypesLinter,
		oneofNamesLowerSnakeCaseLinter,
		packageIsDeclaredLinter,
		packageLowerSnakeCaseLinter,
		packagesSameInDirLinter,
		rpcsHaveCommentsLinter,
		rpcNamesCamelCaseLinter,
		rpcNamesCapitalizedLinter,
		requestResponseTypesInSameFileLinter,
		requestResponseTypesUniqueLinter,
		requestResponseNamesMatchRPCLinter,
		servicesHaveCommentsLinter,
		serviceNamesCamelCaseLinter,
		serviceNamesCapitalizedLinter,
		syntaxProto3Linter,
		wktDirectlyImportedLinter,
	}

	// DefaultLinters is the slice of default Linters.
	DefaultLinters = copyLintersWithout(
		AllLinters,
		enumFieldNamesUppercaseLinter,
		enumsHaveCommentsLinter,
		fileOptionsUnsetJavaMultipleFilesLinter,
		fileOptionsUnsetJavaOuterClassnameLinter,
		messageFieldsNotFloatsLinter,
		messagesHaveCommentsLinter,
		messagesHaveCommentsExceptRequestResponseTypesLinter,
		messageFieldNamesLowercaseLinter,
		requestResponseNamesMatchRPCLinter,
		rpcsHaveCommentsLinter,
		servicesHaveCommentsLinter,
	)

	// DefaultGroup is the default group.
	DefaultGroup = "default"

	// AllGroup is the group of all known linters.
	AllGroup = "all"

	// GroupToLinters is the map from linter group to the corresponding slice of linters.
	GroupToLinters = map[string][]Linter{
		DefaultGroup: DefaultLinters,
		AllGroup:     AllLinters,
	}
)

func init() {
	ids := make(map[string]struct{})
	for _, linter := range AllLinters {
		if _, ok := ids[linter.ID()]; ok {
			panic(fmt.Sprintf("duplicate linter id %s", linter.ID()))
		}
		ids[linter.ID()] = struct{}{}
	}
}

// Runner runs a lint job.
type Runner interface {
	Run(*file.ProtoSet) ([]*text.Failure, error)
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

// Linter is a linter for Protobuf files.
type Linter interface {
	// Return the ID of this Linter. This should be all UPPER_SNAKE_CASE.
	ID() string
	// Return the purpose of this Linter. This should be a human-readable string.
	Purpose() string
	// Check the file data for the descriptors in a common directgory.
	// If there is a lint failure, this returns it in the
	// slice and does not return an error. An error is returned if something
	// unexpected happens. Callers should verify the files are compilable
	// before running this.
	Check(dirPath string, descriptors []*proto.Proto) ([]*text.Failure, error)
}

// NewLinter is a convenience function that returns a new Linter for the
// given parameters, using a function to record failures.
//
// The ID will be upper-cased.
//
// Failures returned from check do not need to set the ID, this will be overwritten.
func NewLinter(id string, purpose string, addCheck func(func(*text.Failure), string, []*proto.Proto) error) Linter {
	return newBaseLinter(id, purpose, addCheck)
}

// GetLinters returns the Linters for the LintConfig.
//
// The configuration is expected to be valid, deduplicated, and all upper-case.
// IncludeIDs and ExcludeIDs MUST NOT have an intersection.
//
// If the config came from the settings package, this is already validated.
func GetLinters(config settings.LintConfig) ([]Linter, error) {
	var linters []Linter
	if !config.NoDefault {
		linters = DefaultLinters
	}
	if len(config.IncludeIDs) == 0 && len(config.ExcludeIDs) == 0 {
		return linters, nil
	}

	// Apply the configured linters to the default group.
	linterMap := make(map[string]Linter, len(linters)+len(config.IncludeIDs)-len(config.ExcludeIDs))
	for _, l := range linters {
		linterMap[l.ID()] = l
	}
	if len(config.IncludeIDs) > 0 {
		for _, l := range AllLinters {
			for _, id := range config.IncludeIDs {
				if l.ID() == id {
					linterMap[id] = l
				}
			}
		}
	}
	for _, excludeID := range config.ExcludeIDs {
		delete(linterMap, excludeID)
	}

	result := make([]Linter, 0, len(linterMap))
	for _, l := range linterMap {
		result = append(result, l)
	}
	return result, nil
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

// CheckMultiple is a convenience function that checks multiple linters and multiple descriptors.
func CheckMultiple(linters []Linter, dirPathToDescriptors map[string][]*proto.Proto, ignoreIDToFilePaths map[string][]string) ([]*text.Failure, error) {
	var allFailures []*text.Failure
	for dirPath, descriptors := range dirPathToDescriptors {
		for _, linter := range linters {
			failures, err := checkOne(linter, dirPath, descriptors, ignoreIDToFilePaths)
			if err != nil {
				return nil, err
			}
			allFailures = append(allFailures, failures...)
		}
	}
	text.SortFailures(allFailures)
	return allFailures, nil
}

func checkOne(linter Linter, dirPath string, descriptors []*proto.Proto, ignoreIDToFilePaths map[string][]string) ([]*text.Failure, error) {
	filteredDescriptors, err := filterIgnores(linter, descriptors, ignoreIDToFilePaths)
	if err != nil {
		return nil, err
	}
	return linter.Check(dirPath, filteredDescriptors)
}

func filterIgnores(linter Linter, descriptors []*proto.Proto, ignoreIDToFilePaths map[string][]string) ([]*proto.Proto, error) {
	var filteredDescriptors []*proto.Proto
	for _, descriptor := range descriptors {
		ignore, err := shouldIgnore(linter, descriptor, ignoreIDToFilePaths)
		if err != nil {
			return nil, err
		}
		if !ignore {
			filteredDescriptors = append(filteredDescriptors, descriptor)
		}
	}
	return filteredDescriptors, nil
}

func shouldIgnore(linter Linter, descriptor *proto.Proto, ignoreIDToFilePaths map[string][]string) (bool, error) {
	filePath := descriptor.Filename
	var err error
	if !filepath.IsAbs(filePath) {
		filePath, err = filepath.Abs(filePath)
		if err != nil {
			return false, err
		}
	}
	ignoreFilePaths, ok := ignoreIDToFilePaths[linter.ID()]
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

func copyLintersWithout(linters []Linter, remove ...Linter) []Linter {
	c := make([]Linter, 0, len(linters))
	for _, linter := range linters {
		if !linterIn(linter, remove) {
			c = append(c, linter)
		}
	}
	return c
}

func linterIn(linter Linter, s []Linter) bool {
	for _, e := range s {
		if e == linter {
			return true
		}
	}
	return false
}
