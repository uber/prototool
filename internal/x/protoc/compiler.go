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

package protoc

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/uber/prototool/internal/x/file"
	"github.com/uber/prototool/internal/x/settings"
	"github.com/uber/prototool/internal/x/text"
	"github.com/uber/prototool/internal/x/wkt"
	"go.uber.org/zap"
)

var (
	// special cased
	pluginFailedRegexp       = regexp.MustCompile("^--.*_out: protoc-gen-(.*): Plugin failed with status code (.*).$")
	otherPluginFailureRegexp = regexp.MustCompile("^--(.*)_out: (.*)$")

	extraImportRegexp  = regexp.MustCompile("^(.*): warning: Import (.*) but not used.$")
	fileNotFoundRegexp = regexp.MustCompile("^(.*): File not found.$")
	// protoc outputs both this line and fileNotFound, so we end up ignoring this one
	// TODO figure out what the error is for errors in the import
	importNotFoundRegexp              = regexp.MustCompile("^(.*): Import (.*) was not found or had errors.$")
	noSyntaxSpecifiedRegexp           = regexp.MustCompile("No syntax specified for the proto file: (.*)\\. Please use")
	jsonCamelCaseRegexp               = regexp.MustCompile("^(.*): (The JSON camel-case name of field.*)$")
	isNotDefinedRegexp                = regexp.MustCompile("^(.*): (.*) is not defined.$")
	seemsToBeDefinedRegexp            = regexp.MustCompile(`^(.*): (".*" seems to be defined in ".*", which is not imported by ".*". To use it here, please add the necessary import.)$`)
	explicitDefaultValuesProto3Regexp = regexp.MustCompile("^(.*): Explicit default values are not allowed in proto3.$")
	optionValueRegexp                 = regexp.MustCompile("^(.*): Error while parsing option value for (.*)$")
	programNotFoundRegexp             = regexp.MustCompile("protoc-gen-(.*): program not found or is not executable$")
	firstEnumValueZeroRegexp          = regexp.MustCompile("^(.*): The first enum value must be zero in proto3.$")
)

type compiler struct {
	logger              *zap.Logger
	cachePath           string
	protocURL           string
	doGen               bool
	doFileDescriptorSet bool
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

func (c *compiler) Compile(protoSets ...*file.ProtoSet) (*CompileResult, error) {
	var allCmdMetas []*cmdMeta
	defer cleanCmdMetas(allCmdMetas)
	for _, protoSet := range protoSets {
		cmdMetas, err := c.getCmdMetas(protoSet)
		if err != nil {
			return nil, err
		}
		allCmdMetas = append(allCmdMetas, cmdMetas...)
	}
	if c.doGen {
		if err := c.makeGenDirs(protoSets...); err != nil {
			return nil, err
		}
	}
	var failures []*text.Failure
	var errs []error
	var lock sync.Mutex
	var wg sync.WaitGroup
	for _, cmdMeta := range allCmdMetas {
		cmdMeta := cmdMeta
		wg.Add(1)
		go func() {
			defer wg.Done()
			iFailures, iErr := c.runCmdMeta(cmdMeta)
			lock.Lock()
			failures = append(failures, iFailures...)
			if iErr != nil {
				errs = append(errs, iErr)
			}
			lock.Unlock()
		}()
	}
	wg.Wait()
	if len(errs) > 0 {
		// I want newlines instead of spaces so not using multierr
		errStrings := make([]string, 0, len(errs))
		for _, err := range errs {
			if errString := err.Error(); errString != "" {
				errStrings = append(errStrings, errString)
			}
		}
		return nil, errors.New(strings.Join(errStrings, "\n"))
	}
	if len(failures) > 0 {
		text.SortFailures(failures)
		return &CompileResult{
			Failures: failures,
		}, nil
	}

	fileDescriptorSets := make([]*descriptor.FileDescriptorSet, 0, len(allCmdMetas))
	for _, cmdMeta := range allCmdMetas {
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

func (c *compiler) ProtocCommands(protoSets ...*file.ProtoSet) ([]string, error) {
	var cmdMetaStrings []string
	for _, protoSet := range protoSets {
		cmdMetas, err := c.getCmdMetas(protoSet)
		if err != nil {
			return nil, err
		}
		for _, cmdMeta := range cmdMetas {
			cmdMetaStrings = append(cmdMetaStrings, cmdMeta.String())
		}
		cleanCmdMetas(cmdMetas)
	}
	return cmdMetaStrings, nil
}

func (c *compiler) makeGenDirs(protoSets ...*file.ProtoSet) error {
	genDirs := make(map[string]struct{})
	for _, protoSet := range protoSets {
		for _, genPlugin := range protoSet.Config.Gen.Plugins {
			genDirs[genPlugin.OutputPath.AbsPath] = struct{}{}
		}
	}
	for genDir := range genDirs {
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
	// probably only need stderr but doing this to see what else comes up
	// TODO: commented out because of unknown newlines that come up
	//cmdMeta.execCmd.Stdout = buffer
	cmdMeta.execCmd.Stdout = ioutil.Discard
	runErr := cmdMeta.execCmd.Run()
	if runErr != nil {
		// exit errors are ok, we can probably parse them into text.Failures
		if _, ok := runErr.(*exec.ExitError); !ok {
			return nil, runErr
		}
	}
	output := strings.TrimSpace(buffer.String())
	if output != "" {
		c.logger.Debug("protoc output", zap.String("output", output))
	}
	failures, err := parseProtocOutput(cmdMeta, output)
	if err != nil {
		return nil, err
	}
	// we had a run error but for whatever reason did not get any parsed
	// output lines, we still want to fail in this case
	// this generally should not happen, especially as plugins that fail
	// will result in a pluginFailedRegexp matching line but this
	// is just to make sure
	if len(failures) == 0 && runErr != nil {
		return nil, runErr
	}
	return failures, nil
}

func (c *compiler) getCmdMetas(protoSet *file.ProtoSet) (cmdMetas []*cmdMeta, retErr error) {
	defer func() {
		if retErr != nil {
			cleanCmdMetas(cmdMetas)
			cmdMetas = nil
		}
	}()
	downloader := c.newDownloader(protoSet.Config)
	if _, err := downloader.Download(); err != nil {
		return cmdMetas, err
	}
	for dirPath, protoFiles := range protoSet.DirPathToFiles {
		// best effort to make sure we have the a parent directory of the file
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
		descriptorSetFilePath, isTempFile, err := c.getDescriptorSetFilePath(protoSet)
		if err != nil {
			return cmdMetas, err
		}
		if descriptorSetFilePath != "" {
			descriptorSetTempFilePath := descriptorSetFilePath
			if !isTempFile {
				descriptorSetTempFilePath = ""
			}
			iArgs := append(args, "-o", descriptorSetFilePath)
			if descriptorSetTempFilePath != "" {
				// TODO(pedge): we will need source info if we switch out emicklei/proto
				//iArgs = append(iArgs, "--include_source_info")
				iArgs = append(iArgs, "--include_imports")
			}
			for _, protoFile := range protoFiles {
				iArgs = append(iArgs, protoFile.Path)
			}
			cmdMetas = append(cmdMetas, &cmdMeta{
				execCmd:                   exec.Command(protocPath, iArgs...),
				protoSet:                  protoSet,
				protoFiles:                protoFiles,
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
				protoFiles: protoFiles,
			})
		}
	}
	return cmdMetas, nil
}

func (c *compiler) newDownloader(config settings.Config) Downloader {
	downloaderOptions := []DownloaderOption{
		DownloaderWithLogger(c.logger),
	}
	if c.cachePath != "" {
		downloaderOptions = append(
			downloaderOptions,
			DownloaderWithCachePath(c.cachePath),
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

func (c *compiler) getPluginFlagSets(protoSet *file.ProtoSet, dirPath string) ([][]string, error) {
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
	flagSet := []string{fmt.Sprintf("--%s_out=%s", genPlugin.Name, genPlugin.OutputPath.AbsPath)}
	if len(protoFlags) > 0 {
		flagSet = []string{fmt.Sprintf("--%s_out=%s:%s", genPlugin.Name, protoFlags, genPlugin.OutputPath.AbsPath)}
	}
	if genPlugin.Path != "" {
		flagSet = append(flagSet, fmt.Sprintf("--plugin=protoc-gen-%s=%s", genPlugin.Name, genPlugin.Path))
	}
	return flagSet, nil
}

func getPluginFlagSetProtoFlags(protoSet *file.ProtoSet, dirPath string, genPlugin settings.GenPlugin) (string, error) {
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
	if !genGoPluginOptions.NoDefaultModifiers {
		modifiers := make(map[string]string)
		for subDirPath, protoFiles := range protoSet.DirPathToFiles {
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
			// one of these two must be true, we validate this above
			if genPlugin.Type.IsGo() {
				modifiers = wkt.FilenameToGoModifierMap
			}
			if genPlugin.Type.IsGogo() {
				modifiers = wkt.FilenameToGogoModifierMap
			}
			for key, value := range modifiers {
				goFlags = append(goFlags, fmt.Sprintf("M%s=%s", key, value))
			}
		}
	}
	for key, value := range genGoPluginOptions.ExtraModifiers {
		goFlags = append(goFlags, fmt.Sprintf("M%s=%s", key, value))
	}
	return strings.Join(goFlags, ","), nil
}

func getIncludes(
	downloader Downloader,
	config settings.Config,
	dirPath string,
	configDirPath string,
) ([]string, error) {
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
	if !fileInIncludePath && !includedConfigDirPath {
		includes = append(includes, configDirPath)
	}
	return includes, nil
}

func parseProtocOutput(cmdMeta *cmdMeta, output string) ([]*text.Failure, error) {
	var failures []*text.Failure
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			failure, err := parseProtocLine(cmdMeta, line)
			if err != nil {
				return nil, err
			}
			if failure != nil {
				failures = append(failures, failure)
			}
		}
	}
	return failures, nil
}

func parseProtocLine(cmdMeta *cmdMeta, protocLine string) (*text.Failure, error) {
	if matches := pluginFailedRegexp.FindStringSubmatch(protocLine); len(matches) > 2 {
		return &text.Failure{
			Message: fmt.Sprintf("protoc-gen-%s failed with status code %s.", matches[1], matches[2]),
		}, nil
	}
	if matches := otherPluginFailureRegexp.FindStringSubmatch(protocLine); len(matches) > 2 {
		return &text.Failure{
			Message: fmt.Sprintf("protoc-gen-%s: %s", matches[1], matches[2]),
		}, nil
	}
	split := strings.Split(protocLine, ":")
	if len(split) != 4 {
		if matches := noSyntaxSpecifiedRegexp.FindStringSubmatch(protocLine); len(matches) > 1 {
			return &text.Failure{
				Filename: bestFilePath(cmdMeta, matches[1]),
				Message:  `No syntax specified. Please use 'syntax = "proto2";' or 'syntax = "proto3";' to specify a syntax version.`,
			}, nil
		}
		if matches := extraImportRegexp.FindStringSubmatch(protocLine); len(matches) > 2 {
			if cmdMeta.protoSet.Config.Compile.AllowUnusedImports {
				return nil, nil
			}
			return &text.Failure{
				Filename: bestFilePath(cmdMeta, matches[1]),
				Message:  fmt.Sprintf(`Import "%s" was not used.`, matches[2]),
			}, nil
		}
		if matches := fileNotFoundRegexp.FindStringSubmatch(protocLine); len(matches) > 1 {
			return &text.Failure{
				// TODO: can we figure out the file name?
				Filename: "",
				Message:  fmt.Sprintf(`Import "%s" was not found.`, matches[1]),
			}, nil
		}
		if matches := explicitDefaultValuesProto3Regexp.FindStringSubmatch(protocLine); len(matches) > 1 {
			return &text.Failure{
				Filename: bestFilePath(cmdMeta, matches[1]),
				Message:  `Explicit default values are not allowed in proto3.`,
			}, nil
		}
		if matches := importNotFoundRegexp.FindStringSubmatch(protocLine); len(matches) > 2 {
			// handled by fileNotFoundRegexp
			// see comments at top
			return nil, nil
		}
		if matches := jsonCamelCaseRegexp.FindStringSubmatch(protocLine); len(matches) > 2 {
			return &text.Failure{
				Filename: bestFilePath(cmdMeta, matches[1]),
				Message:  matches[2],
			}, nil
		}
		if matches := isNotDefinedRegexp.FindStringSubmatch(protocLine); len(matches) > 2 {
			return &text.Failure{
				Filename: bestFilePath(cmdMeta, matches[1]),
				Message:  fmt.Sprintf(`%s is not defined.`, matches[2]),
			}, nil
		}
		if matches := seemsToBeDefinedRegexp.FindStringSubmatch(protocLine); len(matches) > 2 {
			return &text.Failure{
				Filename: bestFilePath(cmdMeta, matches[1]),
				Message:  matches[2],
			}, nil
		}
		if matches := optionValueRegexp.FindStringSubmatch(protocLine); len(matches) > 2 {
			return &text.Failure{
				Filename: bestFilePath(cmdMeta, matches[1]),
				Message:  fmt.Sprintf(`Error while parsing option value for %s`, matches[2]),
			}, nil
		}
		if matches := programNotFoundRegexp.FindStringSubmatch(protocLine); len(matches) > 1 {
			return &text.Failure{
				Message: fmt.Sprintf("protoc-gen-%s not found or is not executable.", matches[1]),
			}, nil
		}
		if matches := firstEnumValueZeroRegexp.FindStringSubmatch(protocLine); len(matches) > 1 {
			return &text.Failure{
				Filename: bestFilePath(cmdMeta, matches[1]),
				Message:  `The first enum value must be zero in proto3.`,
			}, nil
		}
		// TODO: plugins can output to stderr as well and we have no way to redirect the output
		// this will error if there are any logging line from a plugin
		// I would prefer to error so that we signal that we don't know what the line is
		// but if this becomes problematic with some plugin in the future, we should
		// return nil, nil here
		// TODO: this should probably be changed to return a generic *text.Failure with
		// no file, line, or column, and just the message being protocLine
		// https://github.com/uber/prototool/issues/14
		return nil, fmt.Errorf("could not interpret protoc line: %s", protocLine)
	}
	line, err := strconv.Atoi(split[1])
	if err != nil {
		return nil, fmt.Errorf("could not interpret protoc line: %s", protocLine)
	}
	column, err := strconv.Atoi(split[2])
	if err != nil {
		return nil, fmt.Errorf("could not interpret protoc line: %s", protocLine)
	}
	message := strings.TrimSpace(split[3])
	if message == "" {
		return nil, fmt.Errorf("could not interpret protoc line: %s", protocLine)
	}
	return &text.Failure{
		Filename: bestFilePath(cmdMeta, split[0]),
		Line:     line,
		Column:   column,
		Message:  message,
	}, nil
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

func getFileDescriptorSet(cmdMeta *cmdMeta) (*descriptor.FileDescriptorSet, error) {
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
	//for _, fileDescriptorProto := range fileDescriptorSet.File {
	//displayFilePath := bestFilePath(cmdMeta, fileDescriptorProto.GetName())
	//fileDescriptorProto.Name = proto.String(displayFilePath)
	//}
	return fileDescriptorSet, nil
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
	prefix := ""
	if len(os.Args) > 0 {
		prefix = "prototool"
	}
	tempFile, err := ioutil.TempFile("", prefix)
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
