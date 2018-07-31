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

package exec

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"text/scanner"
	"text/tabwriter"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/uber/prototool/internal/cfginit"
	"github.com/uber/prototool/internal/create"
	"github.com/uber/prototool/internal/diff"
	"github.com/uber/prototool/internal/extract"
	"github.com/uber/prototool/internal/file"
	"github.com/uber/prototool/internal/format"
	"github.com/uber/prototool/internal/grpc"
	"github.com/uber/prototool/internal/lint"
	"github.com/uber/prototool/internal/phab"
	"github.com/uber/prototool/internal/protoc"
	"github.com/uber/prototool/internal/reflect"
	"github.com/uber/prototool/internal/settings"
	"github.com/uber/prototool/internal/text"
	"github.com/uber/prototool/internal/vars"
	"go.uber.org/zap"
)

var jsonMarshaler = &jsonpb.Marshaler{Indent: "  "}

type runner struct {
	configProvider   settings.ConfigProvider
	protoSetProvider file.ProtoSetProvider

	workDirPath string
	input       io.Reader
	output      io.Writer

	logger       *zap.Logger
	cachePath    string
	protocURL    string
	printFields  string
	dirMode      bool
	harbormaster bool
}

func newRunner(workDirPath string, input io.Reader, output io.Writer, options ...RunnerOption) *runner {
	runner := &runner{
		workDirPath: workDirPath,
		input:       input,
		output:      output,
	}
	for _, option := range options {
		option(runner)
	}
	runner.configProvider = settings.NewConfigProvider(
		settings.ConfigProviderWithLogger(runner.logger),
	)
	runner.protoSetProvider = file.NewProtoSetProvider(
		file.ProtoSetProviderWithLogger(runner.logger),
	)
	return runner
}

func (r *runner) Version() error {
	tabWriter := newTabWriter(r.output)
	if _, err := fmt.Fprintf(tabWriter, "Version:\t%s\n", vars.Version); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(tabWriter, "Default protoc version:\t%s\n", vars.DefaultProtocVersion); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(tabWriter, "Go version:\t%s\n", runtime.Version()); err != nil {
		return err
	}
	if vars.GitCommit != "" {
		if _, err := fmt.Fprintf(tabWriter, "Git commit:\t%s\n", vars.GitCommit); err != nil {
			return err
		}
	}
	if vars.BuiltTimestamp != "" {
		if _, err := fmt.Fprintf(tabWriter, "Built:\t%s\n", vars.BuiltTimestamp); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(tabWriter, "OS/Arch:\t%s/%s\n", runtime.GOOS, runtime.GOARCH); err != nil {
		return err
	}
	return tabWriter.Flush()
}

func (r *runner) Init(args []string, uncomment bool) error {
	if len(args) > 1 {
		return errors.New("must provide one arg dirPath")
	}
	// TODO(pedge): cleanup
	dirPath := r.workDirPath
	if len(args) == 1 {
		dirPath = args[0]
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return err
		}
	}
	filePath := filepath.Join(dirPath, settings.DefaultConfigFilename)
	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("%s already exists", filePath)
	}
	data, err := cfginit.Generate(vars.DefaultProtocVersion, uncomment)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filePath, data, 0644)
}

func (r *runner) Create(args []string, pkg string) error {
	return r.newCreateHandler(pkg).Create(args...)
}

func (r *runner) Download() error {
	config, err := r.getConfig(r.workDirPath)
	if err != nil {
		return err
	}
	path, err := r.newDownloader(config).Download()
	if err != nil {
		return err
	}
	return r.println(path)
}

func (r *runner) Clean() error {
	config, err := r.getConfig(r.workDirPath)
	if err != nil {
		return err
	}
	return r.newDownloader(config).Delete()
}

func (r *runner) Files(args []string) error {
	meta, err := r.getMeta(args)
	if err != nil {
		return err
	}
	for _, files := range meta.ProtoSet.DirPathToFiles {
		for _, file := range files {
			if err := r.println(file.DisplayPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *runner) Compile(args []string, dryRun bool) error {
	meta, err := r.getMeta(args)
	if err != nil {
		return err
	}
	r.printAffectedFiles(meta)
	_, err = r.compile(false, false, dryRun, meta)
	return err
}

func (r *runner) Gen(args []string, dryRun bool) error {
	meta, err := r.getMeta(args)
	if err != nil {
		return err
	}
	r.printAffectedFiles(meta)
	_, err = r.compile(true, false, dryRun, meta)
	return err
}

func (r *runner) DescriptorProto(args []string) error {
	if len(args) < 1 {
		return nil
	}
	path := args[len(args)-1]
	args = args[:len(args)-1]

	meta, err := r.getMeta(args)
	if err != nil {
		return err
	}
	r.printAffectedFiles(meta)
	fileDescriptorSets, err := r.compile(false, true, false, meta)
	if err != nil {
		return err
	}
	if len(fileDescriptorSets) == 0 {
		return fmt.Errorf("no FileDescriptorSets returned")
	}
	message, err := r.newGetter().GetMessage(fileDescriptorSets, path)
	if err != nil {
		return err
	}
	data, err := jsonMarshaler.MarshalToString(message.DescriptorProto)
	if err != nil {
		return err
	}
	return r.println(data)
}

func (r *runner) FieldDescriptorProto(args []string) error {
	if len(args) < 1 {
		return nil
	}
	path := args[len(args)-1]
	args = args[:len(args)-1]

	meta, err := r.getMeta(args)
	if err != nil {
		return err
	}
	r.printAffectedFiles(meta)
	fileDescriptorSets, err := r.compile(false, true, false, meta)
	if err != nil {
		return err
	}
	if len(fileDescriptorSets) == 0 {
		return fmt.Errorf("no FileDescriptorSets returned")
	}
	field, err := r.newGetter().GetField(fileDescriptorSets, path)
	if err != nil {
		return err
	}
	data, err := jsonMarshaler.MarshalToString(field.FieldDescriptorProto)
	if err != nil {
		return err
	}
	return r.println(data)
}

func (r *runner) ServiceDescriptorProto(args []string) error {
	if len(args) < 1 {
		return nil
	}
	path := args[len(args)-1]
	args = args[:len(args)-1]

	meta, err := r.getMeta(args)
	if err != nil {
		return err
	}
	r.printAffectedFiles(meta)
	fileDescriptorSets, err := r.compile(false, true, false, meta)
	if err != nil {
		return err
	}
	if len(fileDescriptorSets) == 0 {
		return fmt.Errorf("no FileDescriptorSets returned")
	}
	service, err := r.newGetter().GetService(fileDescriptorSets, path)
	if err != nil {
		return err
	}
	data, err := jsonMarshaler.MarshalToString(service.ServiceDescriptorProto)
	if err != nil {
		return err
	}
	return r.println(data)
}

func (r *runner) compile(doGen, doFileDescriptorSet, dryRun bool, meta *meta) ([]*descriptor.FileDescriptorSet, error) {
	if dryRun {
		return nil, r.printCommands(doGen, meta.ProtoSet)
	}
	compileResult, err := r.newCompiler(doGen, doFileDescriptorSet).Compile(meta.ProtoSet)
	if err != nil {
		return nil, err
	}
	if err := r.printFailures("", meta, compileResult.Failures...); err != nil {
		return nil, err
	}
	if len(compileResult.Failures) > 0 {
		return nil, newExitErrorf(255, "")
	}
	r.logger.Debug("protoc command exited without errors")
	return compileResult.FileDescriptorSets, nil
}

func (r *runner) printCommands(doGen bool, protoSet *file.ProtoSet) error {
	commands, err := r.newCompiler(doGen, false).ProtocCommands(protoSet)
	if err != nil {
		return err
	}
	for _, command := range commands {
		if err := r.println(command); err != nil {
			return err
		}
	}
	return nil
}

func (r *runner) Lint(args []string, listAllLinters bool, listLinters bool) error {
	if listAllLinters && listLinters {
		return newExitErrorf(255, "can only set one of list-all-linters, list-linters")
	}
	if listAllLinters {
		return r.listAllLinters()
	}
	if listLinters {
		return r.listLinters()
	}
	meta, err := r.getMeta(args)
	if err != nil {
		return err
	}
	r.printAffectedFiles(meta)
	if _, err := r.compile(false, false, false, meta); err != nil {
		return err
	}
	return r.lint(meta)
}

func (r *runner) lint(meta *meta) error {
	r.logger.Debug("calling LintRunner")
	failures, err := r.newLintRunner().Run(meta.ProtoSet)
	if err != nil {
		return err
	}
	if err := r.printFailures("", meta, failures...); err != nil {
		return err
	}
	if len(failures) > 0 {
		return newExitErrorf(255, "")
	}
	return nil
}

func (r *runner) listLinters() error {
	config, err := r.getConfig(r.workDirPath)
	if err != nil {
		return err
	}
	linters, err := lint.GetLinters(config.Lint)
	if err != nil {
		return err
	}
	return r.printLinters(linters)
}

func (r *runner) listAllLinters() error {
	return r.printLinters(lint.AllLinters)
}

func (r *runner) ListLintGroup(group string) error {
	linters, ok := lint.GroupToLinters[strings.ToLower(group)]
	if !ok {
		return newExitErrorf(255, "unknown lint group: %s", strings.ToLower(group))
	}
	return r.printLinters(linters)
}

func (r *runner) ListAllLintGroups() error {
	groups := make([]string, 0, len(lint.GroupToLinters))
	for group := range lint.GroupToLinters {
		groups = append(groups, group)
	}
	sort.Strings(groups)
	for _, group := range groups {
		if err := r.println(group); err != nil {
			return err
		}
	}
	return nil
}

func (r *runner) Format(args []string, overwrite, diffMode, lintMode, rewrite bool) error {
	if (overwrite && diffMode) || (overwrite && lintMode) || (diffMode && lintMode) {
		return newExitErrorf(255, "can only set one of overwrite, diff, lint")
	}
	meta, err := r.getMeta(args)
	if err != nil {
		return err
	}
	r.printAffectedFiles(meta)
	if _, err := r.compile(false, false, false, meta); err != nil {
		return err
	}
	return r.format(overwrite, diffMode, lintMode, rewrite, meta)
}

func (r *runner) format(overwrite, diffMode, lintMode, rewrite bool, meta *meta) error {
	success := true
	for _, protoFiles := range meta.ProtoSet.DirPathToFiles {
		for _, protoFile := range protoFiles {
			fileSuccess, err := r.formatFile(overwrite, diffMode, lintMode, rewrite, meta, protoFile)
			if err != nil {
				return err
			}
			if !fileSuccess {
				success = false
			}
		}
	}
	if !success {
		return newExitErrorf(255, "")
	}
	return nil
}

// return true if there was no unexpected diff and we should exit with 0
// return false if we should exit with non-zero
// if false and nil error, we will return an ExitError outside of this function
func (r *runner) formatFile(overwrite bool, diffMode bool, lintMode bool, rewrite bool, meta *meta, protoFile *file.ProtoFile) (bool, error) {
	input, err := ioutil.ReadFile(protoFile.Path)
	if err != nil {
		return false, err
	}
	data, failures, err := r.newTransformer(rewrite).Transform(protoFile.Path, input)
	if err != nil {
		return false, err
	}
	if len(failures) > 0 {
		return false, r.printFailures(protoFile.DisplayPath, meta, failures...)
	}
	if !bytes.Equal(input, data) {
		if overwrite {
			// 0 exit code in overwrite case
			return true, ioutil.WriteFile(protoFile.Path, data, os.ModePerm)
		}
		if lintMode {
			return false, r.printFailures("", meta, text.NewFailuref(scanner.Position{
				Filename: protoFile.DisplayPath,
			}, "FORMAT_DIFF", "Format returned a diff."))
		}
		if diffMode {
			d, err := diff.Do(input, data, protoFile.DisplayPath)
			if err != nil {
				return false, err
			}
			if _, err := io.Copy(r.output, bytes.NewReader(d)); err != nil {
				return false, err
			}
			return false, nil
		}
		//!overwrite && !lintMode && !diffMode
		if _, err := io.Copy(r.output, bytes.NewReader(data)); err != nil {
			return false, err
		}
		// there was a diff, return non-zero exit code
		return false, nil
	}
	// we still print the formatted file to stdout
	if !overwrite && !lintMode && !diffMode {
		if _, err := io.Copy(r.output, bytes.NewReader(data)); err != nil {
			return false, err
		}
	}
	return true, nil
}

func (r *runner) BinaryToJSON(args []string) error {
	if len(args) < 2 {
		return nil
	}
	path := args[len(args)-2]
	data, err := r.getInputData(args[len(args)-1])
	if err != nil {
		return err
	}
	args = args[:len(args)-2]

	meta, err := r.getMeta(args)
	if err != nil {
		return err
	}
	r.printAffectedFiles(meta)
	fileDescriptorSets, err := r.compile(false, true, false, meta)
	if err != nil {
		return err
	}
	if len(fileDescriptorSets) == 0 {
		return fmt.Errorf("no FileDescriptorSets returned")
	}
	out, err := r.newReflectHandler().BinaryToJSON(fileDescriptorSets, path, data)
	if err != nil {
		return err
	}
	_, err = r.output.Write(out)
	return err
}

func (r *runner) JSONToBinary(args []string) error {
	if len(args) < 2 {
		return nil
	}
	path := args[len(args)-2]
	data, err := r.getInputData(args[len(args)-1])
	if err != nil {
		return err
	}
	args = args[:len(args)-2]

	meta, err := r.getMeta(args)
	if err != nil {
		return err
	}
	r.printAffectedFiles(meta)
	fileDescriptorSets, err := r.compile(false, true, false, meta)
	if err != nil {
		return err
	}
	if len(fileDescriptorSets) == 0 {
		return fmt.Errorf("no FileDescriptorSets returned")
	}
	out, err := r.newReflectHandler().JSONToBinary(fileDescriptorSets, path, data)
	if err != nil {
		return err
	}
	_, err = r.output.Write(out)
	return err
}

func (r *runner) All(args []string, disableFormat, disableLint, rewrite bool) error {
	meta, err := r.getMeta(args)
	if err != nil {
		return err
	}
	r.printAffectedFiles(meta)
	if _, err := r.compile(false, false, false, meta); err != nil {
		return err
	}
	if !disableFormat {
		if err := r.format(true, false, false, rewrite, meta); err != nil {
			return err
		}
	}
	if _, err := r.compile(true, false, false, meta); err != nil {
		return err
	}
	if !disableLint {
		return r.lint(meta)
	}
	return nil
}

func (r *runner) GRPC(args, headers []string, address, method, data, callTimeout, connectTimeout, keepaliveTime string, stdin bool, jsonOutput bool) error {
	if address == "" {
		return newExitErrorf(255, "must set address")
	}
	if method == "" {
		return newExitErrorf(255, "must set method")
	}
	if data == "" && !stdin {
		return newExitErrorf(255, "must set one of data or stdin")
	}
	if data != "" && stdin {
		return newExitErrorf(255, "must set only one of data or stdin")
	}
	reader := r.getInputReader(data, stdin)

	parsedHeaders := make(map[string]string)
	for _, header := range headers {
		split := strings.SplitN(header, ":", 2)
		if len(split) != 2 {
			return fmt.Errorf("headers must be key:value but got %s", header)
		}
		parsedHeaders[split[0]] = split[1]
	}
	var parsedCallTimeout time.Duration
	var parsedConnectTimeout time.Duration
	var parsedKeepaliveTime time.Duration
	var err error
	if callTimeout != "" {
		parsedCallTimeout, err = time.ParseDuration(callTimeout)
		if err != nil {
			return err
		}
	}
	if connectTimeout != "" {
		parsedConnectTimeout, err = time.ParseDuration(connectTimeout)
		if err != nil {
			return err
		}
	}
	if keepaliveTime != "" {
		parsedKeepaliveTime, err = time.ParseDuration(keepaliveTime)
		if err != nil {
			return err
		}
	}

	meta, err := r.getMeta(args)
	if err != nil {
		return err
	}
	r.printAffectedFiles(meta)
	fileDescriptorSets, err := r.compile(false, true, false, meta)
	if err != nil {
		return err
	}
	if len(fileDescriptorSets) == 0 {
		return fmt.Errorf("no FileDescriptorSets returned")
	}
	return r.newGRPCHandler(
		parsedHeaders,
		parsedCallTimeout,
		parsedConnectTimeout,
		parsedKeepaliveTime,
		jsonOutput,
	).Invoke(fileDescriptorSets, address, method, reader, r.output)
}

func (r *runner) newDownloader(config settings.Config) protoc.Downloader {
	downloaderOptions := []protoc.DownloaderOption{
		protoc.DownloaderWithLogger(r.logger),
	}
	if r.cachePath != "" {
		downloaderOptions = append(
			downloaderOptions,
			protoc.DownloaderWithCachePath(r.cachePath),
		)
	}
	if r.protocURL != "" {
		downloaderOptions = append(
			downloaderOptions,
			protoc.DownloaderWithProtocURL(r.protocURL),
		)
	}
	return protoc.NewDownloader(config, downloaderOptions...)
}

func (r *runner) newCompiler(doGen bool, doFileDescriptorSet bool) protoc.Compiler {
	compilerOptions := []protoc.CompilerOption{
		protoc.CompilerWithLogger(r.logger),
	}
	if r.cachePath != "" {
		compilerOptions = append(
			compilerOptions,
			protoc.CompilerWithCachePath(r.cachePath),
		)
	}
	if r.protocURL != "" {
		compilerOptions = append(
			compilerOptions,
			protoc.CompilerWithProtocURL(r.protocURL),
		)
	}
	if doGen {
		compilerOptions = append(
			compilerOptions,
			protoc.CompilerWithGen(),
		)
	}
	if doFileDescriptorSet {
		compilerOptions = append(
			compilerOptions,
			protoc.CompilerWithFileDescriptorSet(),
		)
	}
	return protoc.NewCompiler(compilerOptions...)
}

func (r *runner) newLintRunner() lint.Runner {
	return lint.NewRunner(
		lint.RunnerWithLogger(r.logger),
	)
}

func (r *runner) newTransformer(rewrite bool) format.Transformer {
	transformerOptions := []format.TransformerOption{format.TransformerWithLogger(r.logger)}
	if rewrite {
		transformerOptions = append(transformerOptions, format.TransformerWithRewrite())
	}
	return format.NewTransformer(transformerOptions...)
}

func (r *runner) newGetter() extract.Getter {
	return extract.NewGetter(
		extract.GetterWithLogger(r.logger),
	)
}

func (r *runner) newReflectHandler() reflect.Handler {
	return reflect.NewHandler(
		reflect.HandlerWithLogger(r.logger),
	)
}

func (r *runner) newCreateHandler(pkg string) create.Handler {
	handlerOptions := []create.HandlerOption{create.HandlerWithLogger(r.logger)}
	if pkg != "" {
		handlerOptions = append(handlerOptions, create.HandlerWithPackage(pkg))
	}
	return create.NewHandler(handlerOptions...)
}

func (r *runner) newGRPCHandler(
	headers map[string]string,
	callTimeout time.Duration,
	connectTimeout time.Duration,
	keepaliveTime time.Duration,
	jsonOutput bool,
) grpc.Handler {
	handlerOptions := []grpc.HandlerOption{
		grpc.HandlerWithLogger(r.logger),
	}
	for key, value := range headers {
		handlerOptions = append(handlerOptions, grpc.HandlerWithHeader(key, value))
	}
	if callTimeout != 0 {
		handlerOptions = append(handlerOptions, grpc.HandlerWithCallTimeout(callTimeout))
	}
	if connectTimeout != 0 {
		handlerOptions = append(handlerOptions, grpc.HandlerWithConnectTimeout(connectTimeout))
	}
	if keepaliveTime != 0 {
		handlerOptions = append(handlerOptions, grpc.HandlerWithKeepaliveTime(keepaliveTime))
	}
	if jsonOutput {
		handlerOptions = append(handlerOptions, grpc.HandlerWithJSONOutput())
	}
	return grpc.NewHandler(handlerOptions...)
}

func (r *runner) getConfig(dirPath string) (settings.Config, error) {
	return r.configProvider.GetForDir(dirPath)
}

type meta struct {
	ProtoSet *file.ProtoSet
	// this will be empty if not in dir mode
	// if in dir mode, this will be the single filename that we want to return errors for
	InDirModeSingleFilename string
}

func (r *runner) getMeta(args []string) (*meta, error) {
	if len(args) == 0 {
		// TODO: does not fit in with workDirPath paradigm
		args = []string{"."}
	}
	if len(args) == 1 {
		fileOrDir := args[0]
		fileInfo, err := os.Stat(fileOrDir)
		if err != nil {
			return nil, err
		}
		if fileInfo.Mode().IsDir() {
			protoSet, err := r.protoSetProvider.GetForDir(r.workDirPath, fileOrDir)
			if err != nil {
				return nil, err
			}
			return &meta{
				ProtoSet: protoSet,
			}, nil
		}
		// TODO: allow symlinks?
		if fileInfo.Mode().IsRegular() {
			if r.dirMode {
				protoSet, err := r.protoSetProvider.GetForDir(r.workDirPath, filepath.Dir(fileOrDir))
				if err != nil {
					return nil, err
				}
				return &meta{
					ProtoSet:                protoSet,
					InDirModeSingleFilename: fileOrDir,
				}, nil
			}
			protoSet, err := r.protoSetProvider.GetForFiles(r.workDirPath, fileOrDir)
			if err != nil {
				return nil, err
			}
			return &meta{
				ProtoSet: protoSet,
			}, nil
		}
		return nil, fmt.Errorf("%s is not a directory or a regular file", fileOrDir)
	}
	for _, arg := range args {
		fileInfo, err := os.Stat(arg)
		if err != nil {
			return nil, err
		}
		// TODO: allow symlinks?
		if !fileInfo.Mode().IsRegular() {
			return nil, fmt.Errorf("multiple arguments only allowed if all arguments are regular files, %q is not a regular file", arg)
		}
	}
	protoSet, err := r.protoSetProvider.GetForFiles(r.workDirPath, args...)
	if err != nil {
		return nil, err
	}
	return &meta{
		ProtoSet: protoSet,
	}, nil
}

// TODO: we filter failures in dir mode in printFailures but above we count any failure
// as an error with a non-zero exit code, seems inconsistent, this needs refactoring

// filename is optional
// if set, it will update the Failures to have this filename
// will be sorted
func (r *runner) printFailures(filename string, meta *meta, failures ...*text.Failure) error {
	for _, failure := range failures {
		if filename != "" {
			failure.Filename = filename
		}
	}
	failureFields, err := text.ParseColonSeparatedFailureFields(r.printFields)
	if err != nil {
		return err
	}
	text.SortFailures(failures)
	bufWriter := bufio.NewWriter(r.output)
	for _, failure := range failures {
		shouldPrint := false
		if meta.InDirModeSingleFilename == "" || meta.InDirModeSingleFilename == failure.Filename {
			shouldPrint = true
		} else if meta.InDirModeSingleFilename != "" {
			// TODO: the compiler may not return the rel path due to logic in bestFilePath
			absSingleFilename, err := file.AbsClean(meta.InDirModeSingleFilename)
			if err != nil {
				return err
			}
			absFailureFilename, err := file.AbsClean(failure.Filename)
			if err != nil {
				return err
			}
			if absSingleFilename == absFailureFilename {
				shouldPrint = true
			}
		}
		if shouldPrint {
			if r.harbormaster {
				harbormasterLintResult, err := phab.TextFailureToHarbormasterLintResult(failure)
				if err != nil {
					return err
				}
				data, err := json.Marshal(harbormasterLintResult)
				if err != nil {
					return err
				}
				if _, err := fmt.Fprintln(bufWriter, string(data)); err != nil {
					return err
				}
			} else if err := failure.Fprintln(bufWriter, failureFields...); err != nil {
				return err
			}
		}
	}
	return bufWriter.Flush()
}

func (r *runner) printLinters(linters []lint.Linter) error {
	sort.Slice(linters, func(i int, j int) bool { return linters[i].ID() < linters[j].ID() })
	tabWriter := newTabWriter(r.output)
	for _, linter := range linters {
		if _, err := fmt.Fprintf(tabWriter, "%s\t%s\n", linter.ID(), linter.Purpose()); err != nil {
			return err
		}
	}
	return tabWriter.Flush()
}

func (r *runner) printAffectedFiles(meta *meta) {
	for _, files := range meta.ProtoSet.DirPathToFiles {
		for _, file := range files {
			r.logger.Debug("using file", zap.String("file", file.DisplayPath))
		}
	}
}

func (r *runner) println(s string) error {
	if s == "" {
		return nil
	}
	_, err := fmt.Fprintln(r.output, s)
	return err
}

func (r *runner) getInputData(arg string) ([]byte, error) {
	if arg == "-" {
		return ioutil.ReadAll(r.input)
	}
	return []byte(arg), nil
}

func (r *runner) getInputReader(data string, stdin bool) io.Reader {
	if stdin {
		return r.input
	}
	return bytes.NewReader([]byte(data))
}

func newExitErrorf(code int, format string, args ...interface{}) *ExitError {
	return &ExitError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

func newTabWriter(writer io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(writer, 0, 0, 2, ' ', 0)
}
