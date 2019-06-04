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

package protoc

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/uber/prototool/internal/file"
	"github.com/uber/prototool/internal/settings"
	"github.com/uber/prototool/internal/text"
	"github.com/uber/prototool/internal/wkt"
	"go.uber.org/zap"
)

var (
	// special cased
	pluginFailedRegexp       = regexp.MustCompile("^--.*_out: protoc-gen-(.*): Plugin failed with status code (.*).$")
	otherPluginFailureRegexp = regexp.MustCompile("^--(.*)_out: (.*)$")
	// backup that does not require this to be at the beginning of the line
	fullPluginFailedRegexp = regexp.MustCompile("(.*)--.*_out: protoc-gen-(.*): Plugin failed with status code (.*).$")

	// protoc started printing line and column information in 3.8.0
	extraImport380Regexp = regexp.MustCompile(`^(.*):(\d+):(\d+): warning: Import (.*) but not used.$`)
	extraImportRegexp    = regexp.MustCompile("^(.*): warning: Import (.*) but not used.$")
	// protoc started printing line and column information in 3.8.0
	recursiveImport380Regexp    = regexp.MustCompile(`^(.*):(\d+):(\d+): File recursively imports itself: (.*)$`)
	recursiveImportRegexp       = regexp.MustCompile("^(.*): File recursively imports itself: (.*)$")
	directoryDoesNotExistRegexp = regexp.MustCompile("^(.*): warning: directory does not exist.$")
	fileNotFoundRegexp          = regexp.MustCompile("^(.*): File not found.$")
	// protoc outputs both this line and fileNotFound, so we end up ignoring this one
	// TODO figure out what the error is for errors in the import
	importNotFoundRegexp    = regexp.MustCompile("^(.*): Import (.*) was not found or had errors.$")
	noSyntaxSpecifiedRegexp = regexp.MustCompile(`No syntax specified for the proto file: (.*)\. Please use`)
	// protoc started printing line and column information in 3.8.0
	jsonCamelCase380Regexp            = regexp.MustCompile(`^(.*):(\d+):(\d+): (The JSON camel-case name of field.*)$`)
	jsonCamelCaseRegexp               = regexp.MustCompile("^(.*): (The JSON camel-case name of field.*)$")
	isNotDefinedRegexp                = regexp.MustCompile("^(.*): (.*) is not defined.$")
	seemsToBeDefinedRegexp            = regexp.MustCompile(`^(.*): (".*" seems to be defined in ".*", which is not imported by ".*". To use it here, please add the necessary import.)$`)
	explicitDefaultValuesProto3Regexp = regexp.MustCompile("^(.*): Explicit default values are not allowed in proto3.$")
	optionValueRegexp                 = regexp.MustCompile("^(.*): Error while parsing option value for (.*)$")
	programNotFoundRegexp             = regexp.MustCompile("protoc-gen-(.*): program not found or is not executable$")
	firstEnumValueZeroRegexp          = regexp.MustCompile("^(.*): The first enum value must be zero in proto3.$")
)

type compiler struct {
	logger                             *zap.Logger
	cachePath                          string
	protocBinPath                      string
	protocWKTPath                      string
	protocURL                          string
	doGen                              bool
	doFileDescriptorSet                bool
	fileDescriptorSetFullControl       bool
	fileDescriptorSetIncludeImports    bool
	fileDescriptorSetIncludeSourceInfo bool
}

func newCompiler(options ...CompilerOption) *compiler {
	compiler := &compiler{
		logger: zap.NewNop(),
	}
	for _, option := range options {
		option(compiler)
	}
	return compiler
}

func (c *compiler) Compile(protoSet *file.ProtoSet) (*CompileResult, error) {
	cmdMetas, err := c.getCmdMetas(protoSet)
	if err != nil {
		cleanCmdMetas(cmdMetas)
		return nil, err
	}

	// we potentially create temporary files if doFileDescriptorSet is true
	// if so, we try to remove them when we return no matter what
	// by putting this defer here, we get this catch early
	defer cleanCmdMetas(cmdMetas)

	if c.doGen {
		// the directories for the output files have to exist
		// so if we are generating, we create them before running
		// protoc, which calls the plugins, which results in created
		// generated files potentially
		// we know the directories from the output option in the
		// config files
		if err := c.makeGenDirs(protoSet); err != nil {
			return nil, err
		}
	}
	var failures []*text.Failure
	var errs []error
	var lock sync.Mutex
	var wg sync.WaitGroup
	semaphoreC := make(chan struct{}, runtime.NumCPU())
	for _, cmdMeta := range cmdMetas {
		cmdMeta := cmdMeta
		wg.Add(1)
		semaphoreC <- struct{}{}
		go func() {
			defer wg.Done()
			iFailures, iErr := c.runCmdMeta(cmdMeta)
			lock.Lock()
			failures = append(failures, iFailures...)
			if iErr != nil {
				errs = append(errs, iErr)
			}
			lock.Unlock()
			<-semaphoreC
		}()
	}
	wg.Wait()
	// errors are not text.Failures, these are actual unhandled
	// system errors from calling protoc, so we short circuit
	if len(errs) > 0 {
		// I want newlines instead of spaces so not using multierr
		errStrings := make([]string, 0, len(errs))
		for _, err := range errs {
			// errors.New("") is a non-nil error, so even
			// if all error strings are empty, we still get an error
			if errString := err.Error(); errString != "" {
				errStrings = append(errStrings, errString)
			}
		}
		return nil, errors.New(strings.Join(errStrings, "\n"))
	}
	// if we have failures, it does not matter if we have file descriptor sets
	// as we should error out, so we do not do any parsing of file descriptor sets
	// this decision could be revisited
	if len(failures) > 0 {
		text.SortFailures(failures)
		return &CompileResult{
			Failures: failures,
		}, nil
	}

	fileDescriptorSets := make([]*FileDescriptorSet, 0, len(cmdMetas))
	for _, cmdMeta := range cmdMetas {
		// if doFileDescriptorSet is not set, we won't get a fileDescriptorSet anyways,
		// so the end result will be an empty CompileResult at this point
		fileDescriptorSet, err := getFileDescriptorSet(cmdMeta)
		if err != nil {
			return nil, err
		}
		if fileDescriptorSet != nil {
			fileDescriptorSets = append(fileDescriptorSets, fileDescriptorSet)
		}
	}
	return &CompileResult{
		FileDescriptorSets: fileDescriptorSets,
	}, nil
}

func (c *compiler) ProtocCommands(protoSet *file.ProtoSet) ([]string, error) {
	// we end up calling the logic that creates temporary files for file descriptor sets
	// anyways, so we need to clean them up with cleanCmdMetas
	// this logic could be simplified to have a "dry run" option, but ProtocCommands
	// is more for debugging anyways
	cmdMetas, err := c.getCmdMetas(protoSet)
	if err != nil {
		return nil, err
	}
	cmdMetaStrings := make([]string, 0, len(cmdMetas))
	for _, cmdMeta := range cmdMetas {
		cmdMetaStrings = append(cmdMetaStrings, cmdMeta.String())
	}
	cleanCmdMetas(cmdMetas)
	return cmdMetaStrings, nil
}

func (c *compiler) makeGenDirs(protoSet *file.ProtoSet) error {
	genDirs := make(map[string]struct{})
	for _, genPlugin := range protoSet.Config.Gen.Plugins {
		baseOutputPath := genPlugin.OutputPath.AbsPath
		// If there is no single output file, protoc plugins take care of making
		// sub-directories, so we only need to make the base directory.
		// Otherwise, we need to make all sub-directories of the output file.
		if genPlugin.FileSuffix == "" {
			genDirs[baseOutputPath] = struct{}{}
		} else {
			for dirPath := range protoSet.DirPathToFiles {
				// skip those files not under the directory
				if !strings.HasPrefix(dirPath, protoSet.DirPath) {
					continue
				}
				relOutputFilePath, err := getRelOutputFilePath(protoSet, dirPath, genPlugin.FileSuffix)
				if err != nil {
					return err
				}
				genDirs[filepath.Dir(filepath.Join(baseOutputPath, relOutputFilePath))] = struct{}{}
			}
		}
	}
	for genDir := range genDirs {
		// we could choose a different permission set, but this seems reasonable
		// in a perfect world, if directories are created and we error out, we
		// would want to remove any newly created directories, but this seems
		// like overkill as these directories would be created on success as
		// generated directories anyways
		if err := os.MkdirAll(genDir, 0744); err != nil {
			return err
		}
	}
	return nil
}

func (c *compiler) runCmdMeta(cmdMeta *cmdMeta) ([]*text.Failure, error) {
	c.logger.Debug("running protoc", zap.String("command", cmdMeta.String()))
	buffer := bytes.NewBuffer(nil)
	cmdMeta.execCmd.Stderr = buffer
	// We only need stderr to parse errors
	// you have to explicitly set to ioutil.Discard, otherwise if there
	// is a stdout, it will be printed to os.Stdout.
	cmdMeta.execCmd.Stdout = ioutil.Discard

	// Prepare a signal buffer so that we can kill the protoc
	// process when Prototool receives a SIGINT or SIGTERM.
	sig := make(chan os.Signal, 1)
	done := make(chan error, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		done <- cmdMeta.execCmd.Run()
	}()

	var runErr error
	select {
	case s := <-sig:
		// Kill the process, and terminate early.
		c.logger.Debug(
			"terminating protoc",
			zap.String("command", cmdMeta.String()),
			zap.String("signal", s.String()),
		)
		return nil, cmdMeta.execCmd.Process.Kill()
	case runErr = <-done:
		// Exit errors are ok, we can probably parse them into text.Failures
		// if not an exec.ExitError, short circuit.
		if _, ok := runErr.(*exec.ExitError); !ok && runErr != nil {
			return nil, runErr
		}
	}
	output := strings.TrimSpace(buffer.String())
	if output != "" {
		c.logger.Debug("protoc output", zap.String("output", output))
	}
	// We want to treat any output from protoc as a failure, even if
	// protoc exited with 0 status. This is because there are outputs
	// from protoc that we consider errors that protoc considers warnings,
	// and plugins in general do not produce output unless there is an error.
	// See https://github.com/uber/prototool/issues/128 for a full discussion.
	failures := c.parseProtocOutput(cmdMeta, output)
	// We had a run error but for whatever reason did not get any parsed
	// output lines, we still want to fail in this case
	// this generally should not happen, especially as plugins that fail
	// will result in a pluginFailedRegexp matching line but this
	// is just to make sure.
	if len(failures) == 0 && runErr != nil {
		return nil, runErr
	}
	return failures, nil
}

func (c *compiler) getCmdMetas(protoSet *file.ProtoSet) (cmdMetas []*cmdMeta, retErr error) {
	defer func() {
		// if we error in this function, we clean ourselves up
		if retErr != nil {
			cleanCmdMetas(cmdMetas)
			cmdMetas = nil
		}
	}()
	// you need a new downloader for every ProtoSet as each configuration file could
	// have a different protoc.version value
	downloader, err := c.newDownloader(protoSet.Config)
	if err != nil {
		return nil, err
	}
	if _, err := downloader.Download(); err != nil {
		return cmdMetas, err
	}
	for dirPath, protoFiles := range protoSet.DirPathToFiles {
		// skip those files not under the directory
		if !strings.HasPrefix(dirPath, protoSet.DirPath) {
			continue
		}
		// you want your proto files to be in at least one of the -I directories
		// or otherwise things can get weird
		// we make best effort to make sure we have the a parent directory of the file
		// if we have a config, use that directory, otherwise use the working directory
		//
		// This does what I'd expect `prototool` to do out of the box:
		//
		// - If a configuration file is present, use that as the root for your imports.
		//   So if you have a/b/prototool.yaml and a/b/c/d/one.proto, a/b/c/e/two.proto,
		//   you'd import c/d/one.proto in two.proto.
		// - If there's no configuration file, I expect my imports to start with the current directory.
		configDirPath := protoSet.Config.DirPath
		if configDirPath == "" {
			configDirPath = protoSet.WorkDirPath
		}
		includes, err := getIncludes(downloader, protoSet.Config, dirPath, configDirPath)
		if err != nil {
			return cmdMetas, err
		}
		var args []string
		for _, include := range includes {
			args = append(args, "-I", include)
		}
		protocPath, err := downloader.ProtocPath()
		if err != nil {
			return cmdMetas, err
		}
		// this could really use some refactoring
		// descriptorSetFilePath will either be a temporary file that we output
		// a file descriptor set to, or the system equivalent of /dev/null
		// isTempFile is effectively != /dev/null for all intents and purposes
		// we do -o /dev/null because protoc needs at least one output, but in the compile-only
		// mode, we want to just test for compile failures
		descriptorSetFilePath, isTempFile, err := c.getDescriptorSetFilePath(protoSet)
		if err != nil {
			return cmdMetas, err
		}
		if descriptorSetFilePath != "" {
			descriptorSetTempFilePath := descriptorSetFilePath
			if !isTempFile {
				descriptorSetTempFilePath = ""
			}
			// either /dev/null or a temporary file
			iArgs := append(args, "-o", descriptorSetFilePath)
			// if its a temporary file, that means we actually care about the output
			// so we do --include_imports to get all necessary info in the output file descriptor set
			if descriptorSetTempFilePath != "" {
				// we included imports historically by default
				// if fileDescriptorSetFullControl is not set, add include imports
				// else, if fileDescriptorSetIncludeImports is set, still include imports
				if !c.fileDescriptorSetFullControl || c.fileDescriptorSetIncludeImports {
					iArgs = append(iArgs, "--include_imports")
				}
				if c.fileDescriptorSetIncludeSourceInfo {
					iArgs = append(iArgs, "--include_source_info")
				}
			}
			for _, protoFile := range protoFiles {
				iArgs = append(iArgs, protoFile.Path)
			}
			cmdMetas = append(cmdMetas, &cmdMeta{
				execCmd:    exec.Command(protocPath, iArgs...),
				protoSet:   protoSet,
				dirPath:    dirPath,
				protoFiles: protoFiles,
				// used for cleaning up the cmdMeta after everything is done
				descriptorSetTempFilePath: descriptorSetTempFilePath,
			})
		}
		pluginFlagSets, err := c.getPluginFlagSets(protoSet, dirPath)
		if err != nil {
			return cmdMetas, err
		}
		for _, pluginFlagSet := range pluginFlagSets {
			iArgs := append(args, pluginFlagSet...)
			for _, protoFile := range protoFiles {
				iArgs = append(iArgs, protoFile.Path)
			}
			cmdMetas = append(cmdMetas, &cmdMeta{
				execCmd:    exec.Command(protocPath, iArgs...),
				protoSet:   protoSet,
				dirPath:    dirPath,
				protoFiles: protoFiles,
			})
		}
	}
	return cmdMetas, nil
}

func (c *compiler) newDownloader(config settings.Config) (Downloader, error) {
	downloaderOptions := []DownloaderOption{
		DownloaderWithLogger(c.logger),
	}
	if c.cachePath != "" {
		downloaderOptions = append(
			downloaderOptions,
			DownloaderWithCachePath(c.cachePath),
		)
	}
	if c.protocBinPath != "" {
		downloaderOptions = append(
			downloaderOptions,
			DownloaderWithProtocBinPath(c.protocBinPath),
		)
	}
	if c.protocWKTPath != "" {
		downloaderOptions = append(
			downloaderOptions,
			DownloaderWithProtocWKTPath(c.protocWKTPath),
		)
	}
	if c.protocURL != "" {
		downloaderOptions = append(
			downloaderOptions,
			DownloaderWithProtocURL(c.protocURL),
		)
	}
	return NewDownloader(config, downloaderOptions...)
}

// return true if a temp file
func (c *compiler) getDescriptorSetFilePath(protoSet *file.ProtoSet) (string, bool, error) {
	if c.doFileDescriptorSet {
		tempFilePath, err := getTempFilePath()
		if err != nil {
			return "", false, err
		}
		return tempFilePath, true, nil
	}
	if c.doGen && len(protoSet.Config.Gen.Plugins) > 0 {
		return "", false, nil
	}
	devNullFilePath, err := devNull()
	return devNullFilePath, false, err
}

// each value in the slice of string slices is a flag passed to protoc
// examples:
// []string{"--go_out=plugins=grpc:."}
// []string{"--grpc-cpp_out=.", "--plugin=protoc-gen-grpc-cpp=/path/to/foo"}
func (c *compiler) getPluginFlagSets(protoSet *file.ProtoSet, dirPath string) ([][]string, error) {
	// if not generating, or there are no plugins, nothing to do
	if !c.doGen || len(protoSet.Config.Gen.Plugins) == 0 {
		return nil, nil
	}
	pluginFlagSets := make([][]string, 0, len(protoSet.Config.Gen.Plugins))
	for _, genPlugin := range protoSet.Config.Gen.Plugins {
		pluginFlagSet, err := getPluginFlagSet(protoSet, dirPath, genPlugin)
		if err != nil {
			return nil, err
		}
		pluginFlagSets = append(pluginFlagSets, pluginFlagSet)
	}
	return pluginFlagSets, nil
}

func getPluginFlagSet(protoSet *file.ProtoSet, dirPath string, genPlugin settings.GenPlugin) ([]string, error) {
	protoFlags, err := getPluginFlagSetProtoFlags(protoSet, dirPath, genPlugin)
	if err != nil {
		return nil, err
	}
	outputPath := genPlugin.OutputPath.AbsPath
	if genPlugin.FileSuffix != "" {
		relOutputFilePath, err := getRelOutputFilePath(protoSet, dirPath, genPlugin.FileSuffix)
		if err != nil {
			return nil, err
		}
		outputPath = filepath.Join(outputPath, relOutputFilePath)
	}
	flagSet := []string{fmt.Sprintf("--%s_out=%s", genPlugin.Name, outputPath)}
	if len(protoFlags) > 0 {
		flagSet = []string{fmt.Sprintf("--%s_out=%s:%s", genPlugin.Name, protoFlags, outputPath)}
	}
	genPluginPath, err := genPlugin.GetPath()
	if err != nil {
		return nil, err
	}
	if genPluginPath != "" {
		flagSet = append(flagSet, fmt.Sprintf("--plugin=protoc-gen-%s=%s", genPlugin.Name, genPluginPath))
	}
	if genPlugin.IncludeImports {
		flagSet = append(flagSet, "--include_imports")
	}
	if genPlugin.IncludeSourceInfo {
		flagSet = append(flagSet, "--include_source_info")
	}
	return flagSet, nil
}

func getRelOutputFilePath(protoSet *file.ProtoSet, dirPath string, fileSuffix string) (string, error) {
	relPath, err := filepath.Rel(protoSet.Config.DirPath, dirPath)
	if err != nil {
		// if we cannot find the relative path, we have a real problem
		// this should never happen, but could in a bad case with links
		return "", fmt.Errorf("could not find relative path for %q to %q, this is a system error, please file a bug at github.com/uber/prototool/issues/new: %v", dirPath, protoSet.Config.DirPath, err)
	}
	base := filepath.Base(relPath)
	if relPath == "" || relPath == "." {
		if protoSet.Config.DirPath == string(os.PathSeparator) {
			base = "default"
		} else {
			base = filepath.Base(protoSet.Config.DirPath)
		}
	}
	return filepath.Join(relPath, base+"."+fileSuffix), nil
}

// the return value corresponds to CodeGeneratorRequest.Parameter
// https://github.com/golang/protobuf/blob/b4deda0973fb4c70b50d226b1af49f3da59f5265/protoc-gen-go/plugin/plugin.pb.go#L103
// this function basically just sets the Mfile=package values for go and gogo plugins
func getPluginFlagSetProtoFlags(protoSet *file.ProtoSet, dirPath string, genPlugin settings.GenPlugin) (string, error) {
	// the type just denotes what Well-Known Type map to use from the wkt package
	// if not go or gogo, we don't have any special automatic handling, so just return what we have
	if !genPlugin.Type.IsGo() && !genPlugin.Type.IsGogo() {
		return genPlugin.Flags, nil
	}
	if genPlugin.Type.IsGo() && genPlugin.Type.IsGogo() {
		return "", fmt.Errorf("internal error: plugin %s is both a go and gogo plugin", genPlugin.Name)
	}
	var goFlags []string
	if genPlugin.Flags != "" {
		goFlags = append(goFlags, genPlugin.Flags)
	}
	genGoPluginOptions := protoSet.Config.Gen.GoPluginOptions
	modifiers := make(map[string]string)
	for subDirPath, protoFiles := range protoSet.DirPathToFiles {
		// you cannot include the files in the same package in the Mfile=package map
		// or otherwise protoc-gen-go, protoc-gen-gogo, etc freak out and put
		// these packages in as imports
		// but, unlike other usages of DirPathToFiles, you MUST include all directories
		// under control of the prototool.yaml to make sure all modifiers are added
		if subDirPath != dirPath {
			for _, protoFile := range protoFiles {
				path, err := filepath.Rel(protoSet.Config.DirPath, protoFile.Path)
				if err != nil {
					// TODO: best effort, maybe error
					path = protoFile.Path
				}
				// TODO: if relative path in OutputPath.RelPath jumps out of import path context, this will be wrong
				modifiers[path] = filepath.Clean(filepath.Join(genGoPluginOptions.ImportPath, genPlugin.OutputPath.RelPath, filepath.Dir(path)))
			}
		}
	}
	for key, value := range modifiers {
		goFlags = append(goFlags, fmt.Sprintf("M%s=%s", key, value))
	}
	if protoSet.Config.Compile.IncludeWellKnownTypes {
		var wktModifiers map[string]string
		// one of these two must be true, we validate this above
		if genPlugin.Type.IsGo() {
			wktModifiers = wkt.FilenameToGoModifierMap
		} else if genPlugin.Type.IsGogo() {
			wktModifiers = wkt.FilenameToGogoModifierMap
		}
		for key, value := range wktModifiers {
			goFlags = append(goFlags, fmt.Sprintf("M%s=%s", key, value))
		}
	}
	for key, value := range genGoPluginOptions.ExtraModifiers {
		goFlags = append(goFlags, fmt.Sprintf("M%s=%s", key, value))
	}
	return strings.Join(goFlags, ","), nil
}

func getIncludes(downloader Downloader, config settings.Config, dirPath string, configDirPath string) ([]string, error) {
	var includes []string
	fileInIncludePath := false
	includedConfigDirPath := false
	for _, includePath := range config.Compile.IncludePaths {
		includes = append(includes, includePath)
		// TODO: not exactly platform independent
		if strings.HasPrefix(dirPath, includePath) {
			fileInIncludePath = true
		}
		if includePath == configDirPath {
			includedConfigDirPath = true
		}
	}
	if config.Compile.IncludeWellKnownTypes {
		wellKnownTypesIncludePath, err := downloader.WellKnownTypesIncludePath()
		if err != nil {
			return nil, err
		}
		includes = append(includes, wellKnownTypesIncludePath)
		// TODO: not exactly platform independent
		if strings.HasPrefix(dirPath, wellKnownTypesIncludePath) {
			fileInIncludePath = true
		}
	}
	// you want your proto files to be in at least one of the -I directories
	// or otherwise things can get weird
	// if the file is not in one of the -I directories and we haven't included
	// the config directory set, at least do that to try to help out
	// this logic could be removed as it is special casing a bit
	if !fileInIncludePath && !includedConfigDirPath {
		includes = append(includes, configDirPath)
	}
	return includes, nil
}

// we try to handle all protoc errors to convert them into text.Failures
// so we can output failures in the standard filename:line:column:message format
func (c *compiler) parseProtocOutput(cmdMeta *cmdMeta, output string) []*text.Failure {
	var failures []*text.Failure
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			if failure := c.parseProtocLine(cmdMeta, line); failure != nil {
				failures = append(failures, failure)
			}
		}
	}
	return failures
}

func (c *compiler) parseProtocLine(cmdMeta *cmdMeta, protocLine string) *text.Failure {
	if matches := pluginFailedRegexp.FindStringSubmatch(protocLine); len(matches) > 2 {
		return &text.Failure{
			Message: fmt.Sprintf("protoc-gen-%s failed with status code %s.", matches[1], matches[2]),
		}
	}
	if matches := otherPluginFailureRegexp.FindStringSubmatch(protocLine); len(matches) > 2 {
		return &text.Failure{
			Message: fmt.Sprintf("protoc-gen-%s: %s", matches[1], matches[2]),
		}
	}
	if matches := fullPluginFailedRegexp.FindStringSubmatch(protocLine); len(matches) > 3 {
		return &text.Failure{
			Message: fmt.Sprintf("protoc-gen-%s failed with status code %s: %s", matches[2], matches[3], matches[1]),
		}
	}
	split := strings.Split(protocLine, ":")
	if len(split) != 4 {
		if matches := noSyntaxSpecifiedRegexp.FindStringSubmatch(protocLine); len(matches) > 1 {
			return &text.Failure{
				Filename: bestFilePath(cmdMeta, matches[1]),
				Message:  `No syntax specified. Please use 'syntax = "proto2";' or 'syntax = "proto3";' to specify a syntax version.`,
			}
		}
		if matches := extraImport380Regexp.FindStringSubmatch(protocLine); len(matches) > 4 {
			if cmdMeta.protoSet.Config.Compile.AllowUnusedImports {
				return nil
			}
			line, err := strconv.Atoi(matches[2])
			if err != nil {
				return c.handleUninterpretedProtocLine(protocLine)
			}
			column, err := strconv.Atoi(matches[3])
			if err != nil {
				return c.handleUninterpretedProtocLine(protocLine)
			}
			return &text.Failure{
				Filename: bestFilePath(cmdMeta, matches[1]),
				Line:     line,
				Column:   column,
				Message:  fmt.Sprintf(`Import "%s" was not used.`, matches[4]),
			}
		}
		if matches := extraImportRegexp.FindStringSubmatch(protocLine); len(matches) > 2 {
			if cmdMeta.protoSet.Config.Compile.AllowUnusedImports {
				return nil
			}
			return &text.Failure{
				Filename: bestFilePath(cmdMeta, matches[1]),
				Message:  fmt.Sprintf(`Import "%s" was not used.`, matches[2]),
			}
		}
		if matches := recursiveImport380Regexp.FindStringSubmatch(protocLine); len(matches) > 4 {
			line, err := strconv.Atoi(matches[2])
			if err != nil {
				return c.handleUninterpretedProtocLine(protocLine)
			}
			column, err := strconv.Atoi(matches[3])
			if err != nil {
				return c.handleUninterpretedProtocLine(protocLine)
			}
			return &text.Failure{
				Filename: bestFilePath(cmdMeta, matches[1]),
				Line:     line,
				Column:   column,
				Message:  fmt.Sprintf(`File recursively imports itself %s.`, matches[4]),
			}
		}
		if matches := recursiveImportRegexp.FindStringSubmatch(protocLine); len(matches) > 2 {
			return &text.Failure{
				Filename: bestFilePath(cmdMeta, matches[1]),
				Message:  fmt.Sprintf(`File recursively imports itself %s.`, matches[2]),
			}
		}
		if matches := directoryDoesNotExistRegexp.FindStringSubmatch(protocLine); len(matches) > 1 {
			return &text.Failure{
				Message: fmt.Sprintf(`Directory "%s" does not exist.`, matches[1]),
			}
		}
		if matches := fileNotFoundRegexp.FindStringSubmatch(protocLine); len(matches) > 1 {
			return &text.Failure{
				// TODO: can we figure out the file name?
				Filename: "",
				Message:  fmt.Sprintf(`Import "%s" was not found.`, matches[1]),
			}
		}
		if matches := explicitDefaultValuesProto3Regexp.FindStringSubmatch(protocLine); len(matches) > 1 {
			return &text.Failure{
				Filename: bestFilePath(cmdMeta, matches[1]),
				Message:  `Explicit default values are not allowed in proto3.`,
			}
		}
		if matches := importNotFoundRegexp.FindStringSubmatch(protocLine); len(matches) > 2 {
			// handled by fileNotFoundRegexp
			// see comments at top
			return nil
		}
		if matches := jsonCamelCase380Regexp.FindStringSubmatch(protocLine); len(matches) > 4 {
			line, err := strconv.Atoi(matches[2])
			if err != nil {
				return c.handleUninterpretedProtocLine(protocLine)
			}
			column, err := strconv.Atoi(matches[3])
			if err != nil {
				return c.handleUninterpretedProtocLine(protocLine)
			}
			return &text.Failure{
				Filename: bestFilePath(cmdMeta, matches[1]),
				Line:     line,
				Column:   column,
				Message:  matches[4],
			}
		}
		if matches := jsonCamelCaseRegexp.FindStringSubmatch(protocLine); len(matches) > 2 {
			return &text.Failure{
				Filename: bestFilePath(cmdMeta, matches[1]),
				Message:  matches[2],
			}
		}
		if matches := isNotDefinedRegexp.FindStringSubmatch(protocLine); len(matches) > 2 {
			return &text.Failure{
				Filename: bestFilePath(cmdMeta, matches[1]),
				Message:  fmt.Sprintf(`%s is not defined.`, matches[2]),
			}
		}
		if matches := seemsToBeDefinedRegexp.FindStringSubmatch(protocLine); len(matches) > 2 {
			return &text.Failure{
				Filename: bestFilePath(cmdMeta, matches[1]),
				Message:  matches[2],
			}
		}
		if matches := optionValueRegexp.FindStringSubmatch(protocLine); len(matches) > 2 {
			return &text.Failure{
				Filename: bestFilePath(cmdMeta, matches[1]),
				Message:  fmt.Sprintf(`Error while parsing option value for %s`, matches[2]),
			}
		}
		if matches := programNotFoundRegexp.FindStringSubmatch(protocLine); len(matches) > 1 {
			return &text.Failure{
				Message: fmt.Sprintf("protoc-gen-%s not found or is not executable.", matches[1]),
			}
		}
		if matches := firstEnumValueZeroRegexp.FindStringSubmatch(protocLine); len(matches) > 1 {
			return &text.Failure{
				Filename: bestFilePath(cmdMeta, matches[1]),
				Message:  `The first enum value must be zero in proto3.`,
			}
		}
		// TODO: plugins can output to stderr as well and we have no way to redirect the output
		// this will error if there are any logging line from a plugin
		// I would prefer to error so that we signal that we don't know what the line is
		// but if this becomes problematic with some plugin in the future, we should
		// return nil, nil here
		return c.handleUninterpretedProtocLine(protocLine)
	}
	line, err := strconv.Atoi(split[1])
	if err != nil {
		return c.handleUninterpretedProtocLine(protocLine)
	}
	column, err := strconv.Atoi(split[2])
	if err != nil {
		return c.handleUninterpretedProtocLine(protocLine)
	}
	message := strings.TrimSpace(split[3])
	if message == "" {
		return c.handleUninterpretedProtocLine(protocLine)
	}
	return &text.Failure{
		Filename: bestFilePath(cmdMeta, split[0]),
		Line:     line,
		Column:   column,
		Message:  message,
	}
}

func (c *compiler) handleUninterpretedProtocLine(protocLine string) *text.Failure {
	c.logger.Warn("protoc returned a line we do not understand, please file this as an issue "+
		"at https://github.com/uber/prototool/issues/new", zap.String("protocLine", protocLine))
	return &text.Failure{
		Message: protocLine,
	}
}

// protoc does weird things with the outputted filename depending
// on what is on the include path, it finds the highest directory
// that the file is on apparently
// -I etc etc/testdata/foo.proto will result in testdata/foo.proto
// this makes it consistent if possible
// TODO: if the file name is not in the given compile command, ie
// if it is imported from another directory, we do not handle this,
// do we want to do a full search of all files in the ProtoSet?
//
// this does getDisplayFilePath but returns match if there is an error
func bestFilePath(cmdMeta *cmdMeta, match string) string {
	displayFilePath, err := getDisplayFilePath(cmdMeta, match)
	if err != nil {
		return match
	}
	return displayFilePath
}

// this does bestFilePath but if there is not exactly one match,
// returns an error
func getDisplayFilePath(cmdMeta *cmdMeta, match string) (string, error) {
	matchingFile := ""
	for _, protoFile := range cmdMeta.protoFiles {
		// if the suffix is the file name, this is a better display name
		// we don't handle the reverse case, ie display path is a suffix of match
		if strings.HasSuffix(protoFile.DisplayPath, match) {
			// if there is more than one match, we don't know what to do
			if matchingFile != "" {
				return "", fmt.Errorf("duplicate matching file: %s", matchingFile)
			}
			matchingFile = protoFile.DisplayPath
		}
	}
	if matchingFile == "" {
		return "", fmt.Errorf("no matching file for %s", match)
	}
	return matchingFile, nil
}

func getFileDescriptorSet(cmdMeta *cmdMeta) (*FileDescriptorSet, error) {
	if cmdMeta.descriptorSetTempFilePath == "" {
		return nil, nil
	}
	data, err := ioutil.ReadFile(cmdMeta.descriptorSetTempFilePath)
	if err != nil {
		return nil, err
	}
	fileDescriptorSet := &descriptor.FileDescriptorSet{}
	if err := proto.Unmarshal(data, fileDescriptorSet); err != nil {
		return nil, err
	}
	return &FileDescriptorSet{
		FileDescriptorSet: fileDescriptorSet,
		ProtoSet:          cmdMeta.protoSet,
		DirPath:           cmdMeta.dirPath,
		ProtoFiles:        cmdMeta.protoFiles,
	}, nil
}

func devNull() (string, error) {
	switch runtime.GOOS {
	case "darwin", "linux":
		return "/dev/null", nil
	case "windows":
		return "nul", nil
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

func getTempFilePath() (string, error) {
	tempFile, err := ioutil.TempFile("", "prototool")
	if err != nil {
		return "", err
	}
	return tempFile.Name(), nil
}

func cleanCmdMetas(cmdMetas []*cmdMeta) {
	for _, cmdMeta := range cmdMetas {
		cmdMeta.Clean()
	}
}

type cmdMeta struct {
	execCmd                   *exec.Cmd
	protoSet                  *file.ProtoSet
	dirPath                   string
	protoFiles                []*file.ProtoFile
	descriptorSetTempFilePath string
}

func (c *cmdMeta) String() string {
	return strings.Join(c.execCmd.Args, " ")
}

func (c *cmdMeta) Clean() {
	tryRemoveTempFile(c.descriptorSetTempFilePath)
}

func tryRemoveTempFile(tempFilePath string) {
	if tempFilePath != "" {
		_ = os.Remove(tempFilePath)
	}
}
