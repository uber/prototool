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
	"github.com/uber/prototool/internal/diff"
	"github.com/uber/prototool/internal/file"
	"github.com/uber/prototool/internal/format"
	"github.com/uber/prototool/internal/lint"
	"github.com/uber/prototool/internal/protoc"
	"github.com/uber/prototool/internal/settings"
	"github.com/uber/prototool/internal/text"
	"github.com/uber/prototool/internal/vars"
	"github.com/uber/prototool/internal/x/extract"
	"github.com/uber/prototool/internal/x/grpc"
	"github.com/uber/prototool/internal/x/phab"
	"github.com/uber/prototool/internal/x/reflect"
	"go.uber.org/zap"
)

var jsonMarshaler = &jsonpb.Marshaler{Indent: "  "}

type runner struct {
	configProvider   settings.ConfigProvider
	protoSetProvider file.ProtoSetProvider
	workDirPath      string
	input            io.Reader
	output           io.Writer
	logger           *zap.Logger
	cachePath        string
	protocURL        string
	printFields      string
	dirMode          bool
	harbormaster     bool
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
	for _, protoSet := range meta.ProtoSets {
		for _, files := range protoSet.DirPathToFiles {
			for _, file := range files {
				if err := r.println(file.DisplayPath); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (r *runner) Compile(args []string) error {
	meta, err := r.getMeta(args)
	if err != nil {
		return err
	}
	r.printAffectedFiles(meta)
	_, err = r.compile(false, false, meta)
	return err
}

func (r *runner) Gen(args []string) error {
	meta, err := r.getMeta(args)
	if err != nil {
		return err
	}
	r.printAffectedFiles(meta)
	_, err = r.compile(true, false, meta)
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
	fileDescriptorSets, err := r.compile(false, true, meta)
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
	if message == nil {
		return fmt.Errorf("nil message")
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
	fileDescriptorSets, err := r.compile(false, true, meta)
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
	if field == nil {
		return fmt.Errorf("nil field")
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
	fileDescriptorSets, err := r.compile(false, true, meta)
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
	if service == nil {
		return fmt.Errorf("nil service")
	}
	data, err := jsonMarshaler.MarshalToString(service.ServiceDescriptorProto)
	if err != nil {
		return err
	}
	return r.println(data)
}

func (r *runner) compile(doGen bool, doFileDescriptorSet bool, meta *meta) ([]*descriptor.FileDescriptorSet, error) {
	compileResult, err := r.newCompiler(doGen, doFileDescriptorSet).Compile(meta.ProtoSets...)
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

func (r *runner) ProtocCommands(args []string, genCommands bool) error {
	meta, err := r.getMeta(args)
	if err != nil {
		return err
	}
	r.printAffectedFiles(meta)
	commands, err := r.newCompiler(genCommands, false).ProtocCommands(meta.ProtoSets...)
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

func (r *runner) Lint(args []string) error {
	meta, err := r.getMeta(args)
	if err != nil {
		return err
	}
	r.printAffectedFiles(meta)
	if _, err := r.compile(false, false, meta); err != nil {
		return err
	}
	return r.lint(meta)
}

func (r *runner) lint(meta *meta) error {
	r.logger.Debug("calling LintRunner")
	failures, err := r.newLintRunner().Run(meta.ProtoSets...)
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

func (r *runner) ListLinters() error {
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

func (r *runner) ListAllLinters() error {
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

func (r *runner) Format(args []string, overwrite bool, diffMode bool, lintMode bool) error {
	meta, err := r.getMeta(args)
	if err != nil {
		return err
	}
	r.printAffectedFiles(meta)
	if _, err := r.compile(false, false, meta); err != nil {
		return err
	}
	return r.format(overwrite, diffMode, lintMode, meta)
}

func (r *runner) format(overwrite bool, diffMode bool, lintMode bool, meta *meta) error {
	var retErr error
	for _, protoSet := range meta.ProtoSets {
		for _, protoFiles := range protoSet.DirPathToFiles {
			for _, protoFile := range protoFiles {
				if err := r.formatFile(overwrite, diffMode, lintMode, meta, protoSet.Config, protoFile); err != nil {
					if _, ok := err.(*ExitError); !ok {
						return err
					}
					retErr = err
				}
			}
		}
	}
	return retErr
}

func (r *runner) formatFile(overwrite bool, diffMode bool, lintMode bool, meta *meta, config settings.Config, protoFile *file.ProtoFile) error {
	input, err := ioutil.ReadFile(protoFile.Path)
	if err != nil {
		return err
	}
	data, failures, err := r.newTransformer().Transform(config, input)
	if err != nil {
		return err
	}
	if len(failures) > 0 {
		if err := r.printFailures(protoFile.DisplayPath, meta, failures...); err != nil {
			return err
		}
		return newExitErrorf(255, "")
	}
	if !bytes.Equal(input, data) {
		if overwrite {
			return ioutil.WriteFile(protoFile.Path, data, os.ModePerm)
		}
		if lintMode {
			if err := r.printFailures("", meta, text.NewFailuref(scanner.Position{
				Filename: protoFile.DisplayPath,
			}, "FORMAT_DIFF", "Format returned a diff.")); err != nil {
				return err
			}
		}
		if diffMode {
			d, err := diff.Do(input, data, protoFile.DisplayPath)
			if err != nil {
				return err
			}
			if _, err := io.Copy(r.output, bytes.NewReader(d)); err != nil {
				return err
			}
		}
		if !overwrite && !lintMode && !diffMode {
			if _, err := io.Copy(r.output, bytes.NewReader(data)); err != nil {
				return err
			}
		}
		return newExitErrorf(255, "")
	}
	if !overwrite && !lintMode && !diffMode {
		if _, err := io.Copy(r.output, bytes.NewReader(data)); err != nil {
			return err
		}
	}
	return nil
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
	fileDescriptorSets, err := r.compile(false, true, meta)
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
	fileDescriptorSets, err := r.compile(false, true, meta)
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

func (r *runner) All(args []string, disableFormat bool, disableLint bool) error {
	meta, err := r.getMeta(args)
	if err != nil {
		return err
	}
	r.printAffectedFiles(meta)
	if _, err := r.compile(false, false, meta); err != nil {
		return err
	}
	if !disableFormat {
		if err := r.format(true, false, false, meta); err != nil {
			return err
		}
	}
	if _, err := r.compile(true, false, meta); err != nil {
		return err
	}
	if !disableLint {
		return r.lint(meta)
	}
	return nil
}

func (r *runner) GRPC(args []string, headers []string, callTimeout string, connectTimeout string, keepaliveTime string) error {
	if len(args) < 3 {
		return nil
	}
	address := args[len(args)-3]
	method := args[len(args)-2]
	reader := r.getInputReader(args[len(args)-1])
	args = args[:len(args)-3]

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
	fileDescriptorSets, err := r.compile(false, true, meta)
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

func (r *runner) newTransformer() format.Transformer {
	return format.NewTransformer(
		format.TransformerWithLogger(r.logger),
	)
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

func (r *runner) newGRPCHandler(
	headers map[string]string,
	callTimeout time.Duration,
	connectTimeout time.Duration,
	keepaliveTime time.Duration,
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
	return grpc.NewHandler(handlerOptions...)
}

func (r *runner) getConfig(dirPath string) (settings.Config, error) {
	return r.configProvider.GetForDir(dirPath)
}

type meta struct {
	ProtoSets               []*file.ProtoSet
	InDirModeSingleFilename string
}

func (r *runner) getMeta(args []string) (*meta, error) {
	if len(args) == 0 {
		// TODO: does not fit in with workDirPath paradigm
		args = []string{"."}
	}
	if len(args) == 1 {
		fileInfo, err := os.Stat(args[0])
		if err != nil {
			return nil, err
		}
		if fileInfo.Mode().IsDir() {
			protoSets, err := r.protoSetProvider.GetForDir(r.workDirPath, args[0])
			if err != nil {
				return nil, err
			}
			return &meta{
				ProtoSets: protoSets,
			}, nil
		}
		// TODO: allow symlinks?
		if fileInfo.Mode().IsRegular() {
			if r.dirMode {
				protoSets, err := r.protoSetProvider.GetForDir(r.workDirPath, filepath.Dir(args[0]))
				if err != nil {
					return nil, err
				}
				return &meta{
					ProtoSets:               protoSets,
					InDirModeSingleFilename: args[0],
				}, nil
			}
			protoSets, err := r.protoSetProvider.GetForFiles(r.workDirPath, args[0])
			if err != nil {
				return nil, err
			}
			return &meta{
				ProtoSets: protoSets,
			}, nil
		}
		return nil, fmt.Errorf("%s is not a directory or a regular file", args[0])
	}
	for _, arg := range args {
		fileInfo, err := os.Stat(arg)
		if err != nil {
			return nil, err
		}
		// TODO: allow symlinks?
		if !fileInfo.Mode().IsRegular() {
			return nil, fmt.Errorf("multiple arguments only allowed if all arguments are regular files and %s is not a regular file", args[0])
		}
	}
	protoSets, err := r.protoSetProvider.GetForFiles(r.workDirPath, args...)
	if err != nil {
		return nil, err
	}
	return &meta{
		ProtoSets: protoSets,
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
			absSingleFilename, err := absClean(meta.InDirModeSingleFilename)
			if err != nil {
				return err
			}
			absFailureFilename, err := absClean(failure.Filename)
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
	for _, protoSet := range meta.ProtoSets {
		for _, files := range protoSet.DirPathToFiles {
			for _, file := range files {
				r.logger.Debug("using file", zap.String("file", file.DisplayPath))
			}
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

func (r *runner) getInputReader(arg string) io.Reader {
	if arg == "-" {
		return r.input
	}
	return bytes.NewReader([]byte(arg))
}

func newExitErrorf(code int, format string, args ...interface{}) *ExitError {
	return &ExitError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

// TODO: this is copied in three places
func absClean(path string) (string, error) {
	if path == "" {
		return path, nil
	}
	if !filepath.IsAbs(path) {
		return filepath.Abs(path)
	}
	return filepath.Clean(path), nil
}

func newTabWriter(writer io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(writer, 0, 0, 2, ' ', 0)
}
