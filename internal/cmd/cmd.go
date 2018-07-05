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

// Package cmd contains the logic to setup Prototool with github.com/spf13/cobra.
//
// The packages cmd/prototool, internal/gen/gen-prototool-bash-completion,
// internal/gen/gen-prototool-manpages and internal/gen/gen-prototool-zsh-completion
// are main packages that call into this package, and this package calls into
// internal/exec to execute the logic.
//
// This package also contains integration testing for Prototool.
package cmd

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/uber/prototool/internal/exec"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// when generating man pages, the current date is used
// this means every time we run make gen, a diff is created
// this gets extremely annoying and isn't very useful so we make it static here
// we could also not check in the man pages, but for now we have them checked in
var genManTime = time.Date(2018, time.January, 1, 0, 0, 0, 0, time.UTC)

// Do runs the command logic.
func Do(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	return runRootCommand(args, stdin, stdout, stderr, (*cobra.Command).Execute)
}

// GenBashCompletion generates a bash completion file to the writer.
func GenBashCompletion(stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	return runRootCommandOutput([]string{}, stdin, stdout, stderr, (*cobra.Command).GenBashCompletion)
}

// GenZshCompletion generates a zsh completion file to the writer.
func GenZshCompletion(stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	return runRootCommandOutput([]string{}, stdin, stdout, stderr, (*cobra.Command).GenZshCompletion)
}

// GenManpages generates the manpages to the given directory.
func GenManpages(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	return runRootCommand(args, stdin, stdout, stderr, func(cmd *cobra.Command) error {
		if len(args) != 1 {
			return fmt.Errorf("usage: %s dirPath", os.Args[0])
		}
		return doc.GenManTree(cmd, &doc.GenManHeader{
			Date: &genManTime,
			// Otherwise we get an annoying "Auto generated by spf13/cobra"
			// Maybe we want that, but I think it's better to just have this
			Source: "Prototool",
		}, args[0])
	})
}

func runRootCommandOutput(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer, f func(*cobra.Command, io.Writer) error) int {
	return runRootCommand(args, stdin, stdout, stderr, func(cmd *cobra.Command) error { return f(cmd, stdout) })
}

func runRootCommand(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer, f func(*cobra.Command) error) (exitCode int) {
	if err := checkOS(); err != nil {
		return printAndGetErrorExitCode(err, stdout)
	}
	if err := f(getRootCommand(&exitCode, args, stdin, stdout, stderr)); err != nil {
		return printAndGetErrorExitCode(err, stdout)
	}
	return exitCode
}

func getRootCommand(exitCodeAddr *int, args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) *cobra.Command {
	flags := &flags{}

	allCmd := &cobra.Command{
		Use:   "all dirOrProtoFiles...",
		Short: "Compile, then format and overwrite, then re-compile and generate, then lint, stopping if any step fails.",
		Run: func(cmd *cobra.Command, args []string) {
			checkCmd(exitCodeAddr, stdin, stdout, stderr, flags, func(runner exec.Runner) error {
				return runner.All(args, flags.disableFormat, flags.disableLint, !flags.noRewrite, flags.dryRun)
			})
		},
	}
	flags.bindDirMode(allCmd.PersistentFlags())
	flags.bindDisableFormat(allCmd.PersistentFlags())
	flags.bindDisableLint(allCmd.PersistentFlags())
	flags.bindNoRewrite(allCmd.PersistentFlags())

	binaryToJSONCmd := &cobra.Command{
		Use:   "binary-to-json dirOrProtoFiles... messagePath data",
		Short: "Convert the data from json to binary for the message path and data.",
		Args:  cobra.MinimumNArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			checkCmd(exitCodeAddr, stdin, stdout, stderr, flags, func(runner exec.Runner) error { return runner.BinaryToJSON(args) })
		},
	}
	flags.bindDirMode(binaryToJSONCmd.PersistentFlags())

	cleanCmd := &cobra.Command{
		Use:   "clean",
		Short: "Delete the cache.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			checkCmd(exitCodeAddr, stdin, stdout, stderr, flags, exec.Runner.Clean)
		},
	}

	compileCmd := &cobra.Command{
		Use:   "compile dirOrProtoFiles...",
		Short: "Compile with protoc to check for failures.",
		Run: func(cmd *cobra.Command, args []string) {
			checkCmd(exitCodeAddr, stdin, stdout, stderr, flags, func(runner exec.Runner) error { return runner.Compile(args, flags.dryRun) })
		},
	}
	flags.bindDirMode(compileCmd.PersistentFlags())

	createCmd := &cobra.Command{
		Use:   "create files...",
		Short: "Create the given Protobuf files according to a template that passes default prototool lint.",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			checkCmd(exitCodeAddr, stdin, stdout, stderr, flags, func(runner exec.Runner) error {
				return runner.Create(args, flags.pkg)
			})
		},
	}
	flags.bindPackage(createCmd.PersistentFlags())

	descriptorProtoCmd := &cobra.Command{
		Use:   "descriptor-proto dirOrProtoFiles... messagePath",
		Short: "Get the descriptor proto for the message path.",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			checkCmd(exitCodeAddr, stdin, stdout, stderr, flags, func(runner exec.Runner) error { return runner.DescriptorProto(args) })
		},
	}
	flags.bindDirMode(descriptorProtoCmd.PersistentFlags())

	downloadCmd := &cobra.Command{
		Use:   "download",
		Short: "Download the protobuf artifacts to a cache.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			checkCmd(exitCodeAddr, stdin, stdout, stderr, flags, exec.Runner.Download)
		},
	}

	fieldDescriptorProtoCmd := &cobra.Command{
		Use:   "field-descriptor-proto dirOrProtoFiles... fieldPath",
		Short: "Get the field descriptor proto for the field path.",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			checkCmd(exitCodeAddr, stdin, stdout, stderr, flags, func(runner exec.Runner) error { return runner.FieldDescriptorProto(args) })
		},
	}
	flags.bindDirMode(fieldDescriptorProtoCmd.PersistentFlags())

	filesCmd := &cobra.Command{
		Use:   "files dirOrProtoFiles...",
		Short: "Print all files that match the input arguments.",
		Run: func(cmd *cobra.Command, args []string) {
			checkCmd(exitCodeAddr, stdin, stdout, stderr, flags, func(runner exec.Runner) error { return runner.Files(args) })
		},
	}

	formatCmd := &cobra.Command{
		Use:   "format dirOrProtoFiles...",
		Short: "Format a proto file and compile with protoc to check for failures.",
		Run: func(cmd *cobra.Command, args []string) {
			checkCmd(exitCodeAddr, stdin, stdout, stderr, flags, func(runner exec.Runner) error {
				return runner.Format(args, flags.overwrite, flags.diffMode, flags.lintMode, !flags.noRewrite, flags.dryRun)
			})
		},
	}
	flags.bindDiffMode(formatCmd.PersistentFlags())
	flags.bindLintMode(formatCmd.PersistentFlags())
	flags.bindOverwrite(formatCmd.PersistentFlags())
	flags.bindNoRewrite(formatCmd.PersistentFlags())

	genCmd := &cobra.Command{
		Use:   "gen dirOrProtoFiles...",
		Short: "Generate with protoc.",
		Run: func(cmd *cobra.Command, args []string) {
			checkCmd(exitCodeAddr, stdin, stdout, stderr, flags, func(runner exec.Runner) error { return runner.Gen(args, flags.dryRun) })
		},
	}
	flags.bindDirMode(genCmd.PersistentFlags())

	grpcCmd := &cobra.Command{
		Use:   "grpc dirOrProtoFiles... serverAddress package.service/Method requestData",
		Short: "Call a gRPC endpoint.",
		Run: func(cmd *cobra.Command, args []string) {
			checkCmd(exitCodeAddr, stdin, stdout, stderr, flags, func(runner exec.Runner) error {
				return runner.GRPC(args, flags.headers, flags.callTimeout, flags.connectTimeout, flags.keepaliveTime)
			})
		},
	}
	flags.bindCallTimeout(grpcCmd.PersistentFlags())
	flags.bindConnectTimeout(grpcCmd.PersistentFlags())
	flags.bindDirMode(grpcCmd.PersistentFlags())
	flags.bindHeaders(grpcCmd.PersistentFlags())
	flags.bindKeepaliveTime(grpcCmd.PersistentFlags())

	initCmd := &cobra.Command{
		Use:   "init [dirPath]",
		Short: "Generate an initial config file in the current or given directory.",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			checkCmd(exitCodeAddr, stdin, stdout, stderr, flags, func(runner exec.Runner) error { return runner.Init(args, flags.uncomment) })
		},
	}
	flags.bindUncomment(initCmd.PersistentFlags())

	jsonToBinaryCmd := &cobra.Command{
		Use:   "json-to-binary dirOrProtoFiles... messagePath data",
		Short: "Convert the data from json to binary for the message path and data.",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			checkCmd(exitCodeAddr, stdin, stdout, stderr, flags, func(runner exec.Runner) error { return runner.JSONToBinary(args) })
		},
	}
	flags.bindDirMode(jsonToBinaryCmd.PersistentFlags())

	lintCmd := &cobra.Command{
		Use:   "lint dirOrProtoFiles...",
		Short: "Lint proto files and compile with protoc to check for failures.",
		Run: func(cmd *cobra.Command, args []string) {
			checkCmd(exitCodeAddr, stdin, stdout, stderr, flags, func(runner exec.Runner) error { return runner.Lint(args, flags.dryRun) })
		},
	}
	flags.bindDirMode(lintCmd.PersistentFlags())

	listAllLintersCmd := &cobra.Command{
		Use:   "list-all-linters",
		Short: "List all available linters.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			checkCmd(exitCodeAddr, stdin, stdout, stderr, flags, exec.Runner.ListAllLinters)
		},
	}

	listAllLintGroupsCmd := &cobra.Command{
		Use:   "list-all-lint-groups",
		Short: "List all the available lint groups.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			checkCmd(exitCodeAddr, stdin, stdout, stderr, flags, exec.Runner.ListAllLintGroups)
		},
	}

	listLintersCmd := &cobra.Command{
		Use:   "list-linters",
		Short: "List the configurerd linters.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			checkCmd(exitCodeAddr, stdin, stdout, stderr, flags, exec.Runner.ListLinters)
		},
	}

	listLintGroupCmd := &cobra.Command{
		Use:   "list-lint-group group",
		Short: "List the linters in the given lint group.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			checkCmd(exitCodeAddr, stdin, stdout, stderr, flags, func(runner exec.Runner) error { return runner.ListLintGroup(args[0]) })
		},
	}

	serviceDescriptorProtoCmd := &cobra.Command{
		Use:   "service-descriptor-proto dirOrProtoFiles... servicePath",
		Short: "Get the service descriptor proto for the service path.",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			checkCmd(exitCodeAddr, stdin, stdout, stderr, flags, func(runner exec.Runner) error { return runner.ServiceDescriptorProto(args) })
		},
	}
	flags.bindDirMode(serviceDescriptorProtoCmd.PersistentFlags())

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			checkCmd(exitCodeAddr, stdin, stdout, stderr, flags, exec.Runner.Version)
		},
	}

	rootCmd := &cobra.Command{Use: "prototool"}
	rootCmd.AddCommand(allCmd)
	rootCmd.AddCommand(binaryToJSONCmd)
	rootCmd.AddCommand(cleanCmd)
	rootCmd.AddCommand(compileCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(descriptorProtoCmd)
	rootCmd.AddCommand(downloadCmd)
	rootCmd.AddCommand(fieldDescriptorProtoCmd)
	rootCmd.AddCommand(filesCmd)
	rootCmd.AddCommand(formatCmd)
	rootCmd.AddCommand(genCmd)
	rootCmd.AddCommand(grpcCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(jsonToBinaryCmd)
	rootCmd.AddCommand(lintCmd)
	rootCmd.AddCommand(listAllLintersCmd)
	rootCmd.AddCommand(listAllLintGroupsCmd)
	rootCmd.AddCommand(listLintersCmd)
	rootCmd.AddCommand(listLintGroupCmd)
	rootCmd.AddCommand(serviceDescriptorProtoCmd)
	rootCmd.AddCommand(versionCmd)

	// flags bound to rootCmd are global flags
	flags.bindCachePath(rootCmd.PersistentFlags())
	flags.bindDebug(rootCmd.PersistentFlags())
	flags.bindDryRun(rootCmd.PersistentFlags())
	flags.bindHarbormaster(rootCmd.PersistentFlags())
	flags.bindPrintFields(rootCmd.PersistentFlags())
	flags.bindProtocURL(rootCmd.PersistentFlags())

	rootCmd.SetArgs(args)
	rootCmd.SetOutput(stdout)

	return rootCmd
}

func checkOS() error {
	switch runtime.GOOS {
	case "darwin", "linux":
		return nil
	default:
		return fmt.Errorf("%s is not a supported operating system, if you want to go through the code and change all the strings.HasPrefix and \"/\" stuff to os.PathSeparator and filepath calls, you're more than welcome to", runtime.GOOS)
	}
}

func checkCmd(exitCodeAddr *int, stdin io.Reader, stdout io.Writer, stderr io.Writer, flags *flags, f func(exec.Runner) error) {
	runner, err := getRunner(stdin, stdout, stderr, flags)
	if err != nil {
		*exitCodeAddr = printAndGetErrorExitCode(err, stdout)
		return
	}
	if err := f(runner); err != nil {
		*exitCodeAddr = printAndGetErrorExitCode(err, stdout)
	}
}

func getRunner(stdin io.Reader, stdout io.Writer, stderr io.Writer, flags *flags) (exec.Runner, error) {
	logger, err := getLogger(stderr, flags.debug)
	if err != nil {
		return nil, err
	}
	runnerOptions := []exec.RunnerOption{
		exec.RunnerWithLogger(logger),
	}
	if flags.cachePath != "" {
		runnerOptions = append(
			runnerOptions,
			exec.RunnerWithCachePath(flags.cachePath),
		)
	}
	if flags.dirMode {
		runnerOptions = append(
			runnerOptions,
			exec.RunnerWithDirMode(),
		)
	}
	if flags.harbormaster {
		runnerOptions = append(
			runnerOptions,
			exec.RunnerWithHarbormaster(),
		)
	}
	if flags.printFields != "" {
		runnerOptions = append(
			runnerOptions,
			exec.RunnerWithPrintFields(flags.printFields),
		)
	}
	if flags.protocURL != "" {
		runnerOptions = append(
			runnerOptions,
			exec.RunnerWithProtocURL(flags.protocURL),
		)
	}
	workDirPath, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return exec.NewRunner(workDirPath, stdin, stdout, runnerOptions...), nil
}

func getLogger(stderr io.Writer, debug bool) (*zap.Logger, error) {
	level := zapcore.InfoLevel
	if debug {
		level = zapcore.DebugLevel
	}
	return zap.New(
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(
				zap.NewDevelopmentEncoderConfig(),
			),
			zapcore.Lock(zapcore.AddSync(stderr)),
			zap.NewAtomicLevelAt(level),
		),
	), nil
}

func printAndGetErrorExitCode(err error, stdout io.Writer) int {
	if errString := err.Error(); errString != "" {
		_, _ = fmt.Fprintln(stdout, errString)
	}
	if exitError, ok := err.(*exec.ExitError); ok {
		return exitError.Code
	}
	return 1
}
