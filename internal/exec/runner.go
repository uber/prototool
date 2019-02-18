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

	"github.com/uber/prototool/internal/breaking"
	"github.com/uber/prototool/internal/cfginit"
	"github.com/uber/prototool/internal/create"
	"github.com/uber/prototool/internal/diff"
	"github.com/uber/prototool/internal/extract"
	"github.com/uber/prototool/internal/file"
	"github.com/uber/prototool/internal/format"
	"github.com/uber/prototool/internal/git"
	"github.com/uber/prototool/internal/grpc"
	"github.com/uber/prototool/internal/lint"
	"github.com/uber/prototool/internal/protoc"
	"github.com/uber/prototool/internal/reflect"
	"github.com/uber/prototool/internal/settings"
	"github.com/uber/prototool/internal/text"
	"github.com/uber/prototool/internal/vars"
	"go.uber.org/zap"
)

type runner struct {
	protoSetProvider file.ProtoSetProvider

	workDirPath string
	input       io.Reader
	output      io.Writer

	logger        *zap.Logger
	develMode     bool
	cachePath     string
	configData    string
	protocBinPath string
	protocWKTPath string
	protocURL     string
	errorFormat   string
	json          bool
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
	protoSetProviderOptions := []file.ProtoSetProviderOption{
		file.ProtoSetProviderWithLogger(runner.logger),
	}
	if runner.configData != "" {
		protoSetProviderOptions = append(
			protoSetProviderOptions,
			file.ProtoSetProviderWithConfigData(runner.configData),
		)
	}
	if runner.develMode {
		protoSetProviderOptions = append(
			protoSetProviderOptions,
			file.ProtoSetProviderWithDevelMode(),
		)
	}
	runner.protoSetProvider = file.NewProtoSetProvider(protoSetProviderOptions...)
	return runner
}

func (r *runner) cloneForWorkDirPath(workDirPath string) *runner {
	return &runner{
		protoSetProvider: r.protoSetProvider,
		workDirPath:      workDirPath,
		input:            r.input,
		output:           r.output,
		logger:           r.logger,
		cachePath:        r.cachePath,
		configData:       r.configData,
		protocBinPath:    r.protocBinPath,
		protocWKTPath:    r.protocWKTPath,
		protocURL:        r.protocURL,
		errorFormat:      r.errorFormat,
		json:             r.json,
	}
}

func (r *runner) Version() error {
	out := struct {
		Version              string `json:"version,omitempty"`
		DefaultProtocVersion string `json:"default_protoc_version,omitempty"`
		GoVersion            string `json:"go_version,omitempty"`
		GOOS                 string `json:"goos,omitempty"`
		GOARCH               string `json:"goarch,omitempty"`
	}{
		Version:              vars.Version,
		DefaultProtocVersion: vars.DefaultProtocVersion,
		GoVersion:            runtime.Version(),
		GOOS:                 runtime.GOOS,
		GOARCH:               runtime.GOARCH,
	}

	if r.json {
		enc := json.NewEncoder(r.output)
		enc.SetIndent("", "  ")
		return enc.Encode(out)
	}

	tabWriter := newTabWriter(r.output)
	if _, err := fmt.Fprintf(tabWriter, "Version:\t%s\n", out.Version); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(tabWriter, "Default protoc version:\t%s\n", out.DefaultProtocVersion); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(tabWriter, "Go version:\t%s\n", out.GoVersion); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(tabWriter, "OS/Arch:\t%s/%s\n", out.GOOS, out.GOARCH); err != nil {
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

func (r *runner) CacheUpdate(args []string) error {
	meta, err := r.getMeta(args)
	if err != nil {
		return err
	}
	d, err := r.newDownloader(meta.ProtoSet.Config)
	if err != nil {
		return err
	}
	_, err = d.Download()
	return err
}

func (r *runner) CacheDelete() error {
	meta, err := r.getMeta(nil)
	if err != nil {
		return err
	}
	// TODO: do not need config for delete, refactor
	d, err := r.newDownloader(meta.ProtoSet.Config)
	if err != nil {
		return err
	}
	return d.Delete()
}

func (r *runner) Files(args []string) error {
	meta, err := r.getMeta(args)
	if err != nil {
		return err
	}
	for dirPath, files := range meta.ProtoSet.DirPathToFiles {
		// skip those files not under the directory
		if !strings.HasPrefix(dirPath, meta.ProtoSet.DirPath) {
			continue
		}
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

func (r *runner) compile(doGen, doFileDescriptorSet, dryRun bool, meta *meta) (protoc.FileDescriptorSets, error) {
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

func (r *runner) Lint(args []string, listAllLinters bool, listLinters bool, listAllLintGroups bool, listLintGroup string, diffLintGroups string) error {
	if moreThanOneSet(listAllLinters, listLinters, listAllLintGroups, listLintGroup != "", diffLintGroups != "") {
		return newExitErrorf(255, "can only set one of list-all-linters, list-linters, list-all-lint-groups, list-lint-group, diff-lint-groups")
	}
	if listAllLintGroups {
		return r.listAllLintGroups()
	}
	if diffLintGroups != "" {
		return r.diffLintGroups(diffLintGroups)
	}
	meta, err := r.getMeta(args)
	if err != nil {
		return err
	}
	if listAllLinters {
		return r.listAllLinters(meta)
	}
	if listLinters {
		return r.listLinters(meta)
	}
	if listLintGroup != "" {
		return r.listLintGroup(meta, listLintGroup)
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

func (r *runner) listLinters(meta *meta) error {
	linters, err := lint.GetLinters(meta.ProtoSet.Config.Lint)
	if err != nil {
		return err
	}
	return r.printLinters(meta.ProtoSet.Config.Lint, linters)
}

func (r *runner) listAllLinters(meta *meta) error {
	return r.printLinters(meta.ProtoSet.Config.Lint, lint.AllLinters)
}

func (r *runner) listLintGroup(meta *meta, group string) error {
	linters, ok := lint.GroupToLinters[strings.ToLower(group)]
	if !ok {
		return newExitErrorf(255, "unknown lint group: %s", strings.ToLower(group))
	}
	return r.printLinters(meta.ProtoSet.Config.Lint, linters)
}

func (r *runner) listAllLintGroups() error {
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

func (r *runner) diffLintGroups(groups string) error {
	split := strings.Split(groups, ",")
	if len(split) != 2 {
		return fmt.Errorf("argument to --diff-lint-groups must be two lint groups separated by '.', for example google,uber2")
	}
	firstLinterIDs, err := getLinterIDs(split[0])
	if err != nil {
		return err
	}
	secondLinterIDs, err := getLinterIDs(split[1])
	if err != nil {
		return err
	}
	for _, s := range diffMaps(firstLinterIDs, secondLinterIDs) {
		if err := r.println(s); err != nil {
			return err
		}
	}
	return nil
}

func (r *runner) Format(args []string, overwrite, diffMode, lintMode, fixFlag bool) error {
	if moreThanOneSet(overwrite, diffMode, lintMode) {
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
	return r.format(overwrite, diffMode, lintMode, getFormatFixValue(fixFlag, meta), getFormatFileHeaderValue(fixFlag, meta), getFormatJavaPackagePrefixValue(fixFlag, meta), meta)
}

func (r *runner) format(overwrite, diffMode, lintMode bool, fix int, fileHeader string, javaPackagePrefix string, meta *meta) error {
	success := true
	for dirPath, protoFiles := range meta.ProtoSet.DirPathToFiles {
		// skip those files not under the directory
		if !strings.HasPrefix(dirPath, meta.ProtoSet.DirPath) {
			continue
		}
		for _, protoFile := range protoFiles {
			fileSuccess, err := r.formatFile(overwrite, diffMode, lintMode, fix, fileHeader, javaPackagePrefix, meta, protoFile)
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
func (r *runner) formatFile(overwrite bool, diffMode bool, lintMode bool, fix int, fileHeader string, javaPackagePrefix string, meta *meta, protoFile *file.ProtoFile) (bool, error) {
	absSingleFilename, err := file.AbsClean(meta.SingleFilename)
	if err != nil {
		return false, err
	}
	// we are not concerned with the current file
	if meta.SingleFilename != "" && protoFile.Path != absSingleFilename {
		return true, nil
	}
	input, err := ioutil.ReadFile(protoFile.Path)
	if err != nil {
		return false, err
	}
	data, failures, err := r.newTransformer(fix, fileHeader, javaPackagePrefix).Transform(protoFile.Path, input)
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
		//below is !overwrite && !lintMode && !diffMode
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

func (r *runner) All(args []string, disableFormat, disableLint, fixFlag bool) error {
	meta, err := r.getMeta(args)
	if err != nil {
		return err
	}
	r.printAffectedFiles(meta)
	if _, err := r.compile(false, false, false, meta); err != nil {
		return err
	}
	if !disableFormat {
		if err := r.format(true, false, false, getFormatFixValue(fixFlag, meta), getFormatFileHeaderValue(fixFlag, meta), getFormatJavaPackagePrefixValue(fixFlag, meta), meta); err != nil {
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

func (r *runner) GRPC(args, headers []string, address, method, data, callTimeout, connectTimeout, keepaliveTime string, stdin bool) error {
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
	).Invoke(fileDescriptorSets.Unwrap(), address, method, reader, r.output)
}

func (r *runner) InspectPackages(args []string) error {
	packageSet, err := r.getPackageSet(args)
	if err != nil {
		return err
	}
	if packageSet == nil {
		return nil
	}
	return r.printPackageNames(packageSet.PackageNameToPackage())
}

func (r *runner) InspectPackageDeps(args []string, name string) error {
	pkg, err := r.getPackage(args, name)
	if err != nil {
		return err
	}
	return r.printPackageNames(pkg.DependencyNameToDependency())
}

func (r *runner) InspectPackageImporters(args []string, name string) error {
	pkg, err := r.getPackage(args, name)
	if err != nil {
		return err
	}
	return r.printPackageNames(pkg.ImporterNameToImporter())
}

func (r *runner) BreakCheck(args []string, gitBranch string, gitTag string, includeBeta bool, allowBetaDeps bool) error {
	if moreThanOneSet(gitBranch != "", gitTag != "") {
		return newExitErrorf(255, "can only set one of git-branch, git-tag")
	}
	branchOrTag := gitBranch
	if branchOrTag == "" {
		branchOrTag = gitTag
	}

	relDirPath := "."
	// we check length 0 or 1 in cmd, similar to other commands
	if len(args) == 1 {
		relDirPath = args[0]
	}
	if filepath.IsAbs(relDirPath) {
		return fmt.Errorf("input argument must be relative directory path: %s", relDirPath)
	}

	absDirPath, err := file.AbsClean(relDirPath)
	if err != nil {
		return err
	}
	absWorkDirPath, err := file.AbsClean(r.workDirPath)
	if err != nil {
		return err
	}
	if !strings.HasPrefix(absDirPath, absWorkDirPath) {
		return fmt.Errorf("input directory must be within working directory: %s", relDirPath)
	}

	toPackageSet, err := r.getPackageSetForRelDirPath(relDirPath)
	if err != nil {
		return err
	}

	// this will purposefully fail if we are not at a git repository
	cloneDirPath, err := git.TemporaryClone(r.logger, r.workDirPath, branchOrTag)
	if err != nil {
		return err
	}
	defer func() {
		r.logger.Sugar().Debugf("removing %s", cloneDirPath)
		_ = os.RemoveAll(cloneDirPath)
	}()

	fromPackageSet, err := r.cloneForWorkDirPath(cloneDirPath).getPackageSetForRelDirPath(relDirPath)
	if err != nil {
		return err
	}

	failures, err := r.newBreakingRunner(includeBeta, allowBetaDeps).Run(fromPackageSet, toPackageSet)
	if err != nil {
		return err
	}
	if len(failures) > 0 {
		if err := r.printFailuresForErrorFormat("message", "", nil, failures...); err != nil {
			return err
		}
		return newExitErrorf(255, "")
	}
	return nil
}

func (r *runner) getPackageSet(args []string) (*extract.PackageSet, error) {
	meta, err := r.getMeta(args)
	if err != nil {
		return nil, err
	}
	r.printAffectedFiles(meta)
	fileDescriptorSets, err := r.compile(false, true, false, meta)
	if err != nil {
		return nil, err
	}
	reflectPackageSet, err := reflect.NewPackageSet(fileDescriptorSets.Unwrap()...)
	if err != nil {
		return nil, err
	}
	return extract.NewPackageSet(reflectPackageSet)
}

func (r *runner) getPackage(args []string, name string) (*extract.Package, error) {
	if name == "" {
		return nil, newExitErrorf(255, "must set name")
	}
	packageSet, err := r.getPackageSet(args)
	if err != nil {
		return nil, err
	}
	if packageSet == nil {
		return nil, fmt.Errorf("package not found: %s", name)
	}
	pkg, ok := packageSet.PackageNameToPackage()[name]
	if !ok {
		return nil, fmt.Errorf("package not found: %s", name)
	}
	return pkg, nil
}

func (r *runner) printPackageNames(m map[string]*extract.Package) error {
	for _, packageName := range extractSortPackageNames(m) {
		if err := r.println(packageName); err != nil {
			return err
		}
	}
	return nil
}

// we require a relative path (or no path) to be passed
// this is largely because getMeta has special handling for "."
func (r *runner) getPackageSetForRelDirPath(relDirPath string) (*extract.PackageSet, error) {
	dirPath := r.workDirPath
	if relDirPath != "" && relDirPath != "." {
		dirPath = filepath.Join(dirPath, relDirPath)
	}
	return r.getPackageSet([]string{dirPath})
}

func (r *runner) newBreakingRunner(includeBeta bool, allowBetaDeps bool) breaking.Runner {
	runnerOptions := []breaking.RunnerOption{
		breaking.RunnerWithLogger(r.logger),
	}
	if includeBeta {
		runnerOptions = append(
			runnerOptions,
			breaking.RunnerWithIncludeBeta(),
		)
	}
	if allowBetaDeps {
		runnerOptions = append(
			runnerOptions,
			breaking.RunnerWithAllowBetaDeps(),
		)
	}
	return breaking.NewRunner(runnerOptions...)
}

func (r *runner) newDownloader(config settings.Config) (protoc.Downloader, error) {
	downloaderOptions := []protoc.DownloaderOption{
		protoc.DownloaderWithLogger(r.logger),
	}
	if r.cachePath != "" {
		downloaderOptions = append(
			downloaderOptions,
			protoc.DownloaderWithCachePath(r.cachePath),
		)
	}
	if r.protocBinPath != "" {
		downloaderOptions = append(
			downloaderOptions,
			protoc.DownloaderWithProtocBinPath(r.protocBinPath),
		)
	}
	if r.protocWKTPath != "" {
		downloaderOptions = append(
			downloaderOptions,
			protoc.DownloaderWithProtocWKTPath(r.protocWKTPath),
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
	if r.protocBinPath != "" {
		compilerOptions = append(
			compilerOptions,
			protoc.CompilerWithProtocBinPath(r.protocBinPath),
		)
	}
	if r.protocWKTPath != "" {
		compilerOptions = append(
			compilerOptions,
			protoc.CompilerWithProtocWKTPath(r.protocWKTPath),
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

func (r *runner) newTransformer(fix int, fileHeader string, javaPackagePrefix string) format.Transformer {
	transformerOptions := []format.TransformerOption{format.TransformerWithLogger(r.logger)}
	if fix != format.FixNone {
		transformerOptions = append(transformerOptions, format.TransformerWithFix(fix))
	}
	if fileHeader != "" {
		transformerOptions = append(transformerOptions, format.TransformerWithFileHeader(fileHeader))
	}
	if javaPackagePrefix != "" {
		transformerOptions = append(transformerOptions, format.TransformerWithJavaPackagePrefix(javaPackagePrefix))
	}
	return format.NewTransformer(transformerOptions...)
}

func (r *runner) newCreateHandler(pkg string) create.Handler {
	handlerOptions := []create.HandlerOption{create.HandlerWithLogger(r.logger)}
	if pkg != "" {
		handlerOptions = append(handlerOptions, create.HandlerWithPackage(pkg))
	}
	if r.develMode {
		handlerOptions = append(handlerOptions, create.HandlerWithDevelMode())
	}
	if r.configData != "" {
		handlerOptions = append(
			handlerOptions,
			create.HandlerWithConfigData(r.configData),
		)
	}
	return create.NewHandler(handlerOptions...)
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

type meta struct {
	ProtoSet *file.ProtoSet
	// this will be empty if not in dir mode
	// if in dir mode, this will be the single filename that we want to return errors for
	SingleFilename string
}

func (r *runner) getMeta(args []string) (*meta, error) {
	// TODO: does not fit in with workDirPath paradigm
	fileOrDir := "."
	if len(args) == 1 {
		fileOrDir = args[0]
	}
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
		protoSet, err := r.protoSetProvider.GetForDir(r.workDirPath, filepath.Dir(fileOrDir))
		if err != nil {
			return nil, err
		}
		return &meta{
			ProtoSet:       protoSet,
			SingleFilename: fileOrDir,
		}, nil
	}
	return nil, fmt.Errorf("%s is not a directory or a regular file", fileOrDir)
}

// TODO: we filter failures in dir mode in printFailures but above we count any failure
// as an error with a non-zero exit code, seems inconsistent, this needs refactoring

// filename is optional
// meta is optional
// if set, it will update the Failures to have this filename
// will be sorted
func (r *runner) printFailures(filename string, meta *meta, failures ...*text.Failure) error {
	return r.printFailuresForErrorFormat(r.errorFormat, filename, meta, failures...)
}

func (r *runner) printFailuresForErrorFormat(errorFormat string, filename string, meta *meta, failures ...*text.Failure) error {
	for _, failure := range failures {
		if filename != "" {
			failure.Filename = filename
		}
	}
	failureFields, err := text.ParseColonSeparatedFailureFields(errorFormat)
	if err != nil {
		return err
	}
	text.SortFailures(failures)
	bufWriter := bufio.NewWriter(r.output)
	for _, failure := range failures {
		shouldPrint := false
		if meta != nil {
			if meta.SingleFilename == "" || meta.SingleFilename == failure.Filename {
				shouldPrint = true
			} else if meta.SingleFilename != "" {
				// TODO: the compiler may not return the rel path due to logic in bestFilePath
				absSingleFilename, err := file.AbsClean(meta.SingleFilename)
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
		} else {
			shouldPrint = true
		}
		if shouldPrint {
			if r.json {
				data, err := json.Marshal(failure)
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

func (r *runner) printLinters(config settings.LintConfig, linters []lint.Linter) error {
	sort.Slice(linters, func(i int, j int) bool { return linters[i].ID() < linters[j].ID() })
	tabWriter := newTabWriter(r.output)
	for _, linter := range linters {
		if _, err := fmt.Fprintf(tabWriter, "%s\t%s\n", linter.ID(), linter.Purpose(config)); err != nil {
			return err
		}
	}
	return tabWriter.Flush()
}

func (r *runner) printAffectedFiles(meta *meta) {
	for dirPath, files := range meta.ProtoSet.DirPathToFiles {
		// skip those files not under the directory
		if !strings.HasPrefix(dirPath, meta.ProtoSet.DirPath) {
			continue
		}
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

func moreThanOneSet(values ...bool) bool {
	numSet := 0
	for _, value := range values {
		if value {
			numSet++
		}
	}
	return numSet > 1
}

func getFormatFixValue(fixFlag bool, meta *meta) int {
	if !fixFlag {
		return format.FixNone
	}
	if meta.ProtoSet.Config.Lint.Group == "uber2" {
		return format.FixV2
	}
	return format.FixV1
}

func getFormatFileHeaderValue(fixFlag bool, meta *meta) string {
	if !fixFlag {
		return ""
	}
	return meta.ProtoSet.Config.Lint.FileHeader
}

func getFormatJavaPackagePrefixValue(fixFlag bool, meta *meta) string {
	if !fixFlag {
		return ""
	}
	return meta.ProtoSet.Config.Lint.JavaPackagePrefix
}

func extractSortPackageNames(m map[string]*extract.Package) []string {
	s := make([]string, 0, len(m))
	for key := range m {
		if key != "" {
			s = append(s, key)
		}
	}
	sort.Strings(s)
	return s
}

func getLinterIDs(group string) (map[string]struct{}, error) {
	linters, ok := lint.GroupToLinters[strings.ToLower(group)]
	if !ok {
		return nil, newExitErrorf(255, "unknown lint group: %s", strings.ToLower(group))
	}
	m := make(map[string]struct{})
	for _, linter := range linters {
		m[linter.ID()] = struct{}{}
	}
	return m, nil
}

func diffMaps(one map[string]struct{}, two map[string]struct{}) []string {
	var s []string
	for key := range one {
		if _, ok := two[key]; !ok {
			s = append(s, "< "+key)
		}
	}
	for key := range two {
		if _, ok := one[key]; !ok {
			s = append(s, "> "+key)
		}
	}
	sort.Strings(s)
	return s
}
