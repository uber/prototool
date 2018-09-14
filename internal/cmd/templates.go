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

package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	wordwrap "github.com/mitchellh/go-wordwrap"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/uber/prototool/internal/exec"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const wordWrapLength uint = 80

var (
	allCmdTemplate = &cmdTemplate{
		Use:   "all [dirOrFile]",
		Short: "Compile, then format and overwrite, then re-compile and generate, then lint, stopping if any step fails.",
		Args:  cobra.MaximumNArgs(1),
		Run: func(runner exec.Runner, args []string, flags *flags) error {
			return runner.All(args, flags.disableFormat, flags.disableLint, flags.fix)
		},
		BindFlags: func(flagSet *pflag.FlagSet, flags *flags) {
			flags.bindConfigData(flagSet)
			flags.bindDisableFormat(flagSet)
			flags.bindDisableLint(flagSet)
			flags.bindJSON(flagSet)
			flags.bindFix(flagSet)
			flags.bindProtocURL(flagSet)
			flags.bindProtocBinPath(flagSet)
			flags.bindProtocWKTPath(flagSet)
		},
	}

	binaryToJSONCmdTemplate = &cmdTemplate{
		Use:   "binary-to-json [dirOrFile] messagePath data",
		Short: "Convert the data from json to binary for the message path and data.",
		Args:  cobra.RangeArgs(2, 3),
		Run: func(runner exec.Runner, args []string, flags *flags) error {
			return runner.BinaryToJSON(args)
		},
		BindFlags: func(flagSet *pflag.FlagSet, flags *flags) {
			flags.bindConfigData(flagSet)
		},
	}

	cleanCmdTemplate = &cmdTemplate{
		Use:   "clean",
		Short: "Delete the cache.",
		Args:  cobra.NoArgs,
		Run: func(runner exec.Runner, args []string, flags *flags) error {
			return runner.Clean()
		},
	}

	compileCmdTemplate = &cmdTemplate{
		Use:   "compile [dirOrFile]",
		Short: "Compile with protoc to check for failures.",
		Long:  `Stubs will not be generated. To generate stubs, use the "gen" command. Calling "compile" has the effect of calling protoc with "-o /dev/null".`,
		Args:  cobra.MaximumNArgs(1),
		Run: func(runner exec.Runner, args []string, flags *flags) error {
			return runner.Compile(args, flags.dryRun)
		},
		BindFlags: func(flagSet *pflag.FlagSet, flags *flags) {
			flags.bindConfigData(flagSet)
			flags.bindDryRun(flagSet)
			flags.bindJSON(flagSet)
			flags.bindProtocURL(flagSet)
			flags.bindProtocBinPath(flagSet)
			flags.bindProtocWKTPath(flagSet)
		},
	}

	createCmdTemplate = &cmdTemplate{
		Use:   "create files...",
		Short: "Create the given Protobuf files according to a template that passes default prototool lint.",
		Long: `Assuming the filename "example_create_file.proto", the file will look like the following:

  syntax = "proto3";

  package SOME.PKG;

  option go_package = "PKGpb";
  option java_multiple_files = true;
  option java_outer_classname = "ExampleCreateFileProto";
  option java_package = "com.SOME.PKG.pb";

This matches what the linter expects. "SOME.PKG" will be computed as follows:

- If "--package" is specified, "SOME.PKG" will be the value passed to
  "--package".
- Otherwise, if there is no "prototool.yaml" or "prototool.json" that would
  apply to the new file, use "uber.prototool.generated".
- Otherwise, if there is a "prototool.yaml" or "prototool.json" file, check if
  it has a "packages" setting under the "create" section. If it does, this
  package, concatenated with the relative path from the directory with the
 "prototool.yaml" or "prototool.json" will be used.
- Otherwise, if there is no "packages" directive, just use the
  relative path from the directory with the "prototool.yaml" or
  "prototool.json" file. If the file is in the same directory as the
  "prototool.yaml" or "prototool.json" file, use "uber.prototool.generated".

For example, assume you have the following file at "repo/prototool.yaml":

create:
  packages:
	- directory: idl
	  name: uber
	- directory: idl/baz
	  name: special

- "prototool create repo/idl/foo/bar/bar.proto" will have the package
  "uber.foo.bar".
- "prototool create repo/idl/bar.proto" will have the package "uber".
- "prototool create repo/idl/baz/baz.proto" will have the package "special".
- "prototool create repo/idl/baz/bat/bat.proto" will have the package
  "special.bat".
- "prototool create repo/another/dir/bar.proto" will have the package
  "another.dir".
- "prototool create repo/bar.proto" will have the package
  "uber.prototool.generated".

This is meant to mimic what you generally want - a base package for your idl directory, followed by packages matching the directory structure.

Note you can override the directory that the "prototool.yaml" or "prototool.json" file is in as well. If we update our file at "repo/prototool.yaml" to this:

create:
  packages:
	- directory: .
	  name: foo.bar

Then "prototool create repo/bar.proto" will have the package "foo.bar", and "prototool create repo/another/dir/bar.proto" will have the package "foo.bar.another.dir".

If Vim integration is set up, files will be generated when you open a new Protobuf file.`,
		Args: cobra.MinimumNArgs(1),
		Run: func(runner exec.Runner, args []string, flags *flags) error {
			return runner.Create(args, flags.pkg)
		},
		BindFlags: func(flagSet *pflag.FlagSet, flags *flags) {
			flags.bindConfigData(flagSet)
			flags.bindPackage(flagSet)
		},
	}

	descriptorProtoCmdTemplate = &cmdTemplate{
		Use:   "descriptor-proto [dirOrFile] messagePath",
		Short: "Get the descriptor proto for the message path.",
		Args:  cobra.MaximumNArgs(2),
		Run: func(runner exec.Runner, args []string, flags *flags) error {
			return runner.DescriptorProto(args)
		},
		BindFlags: func(flagSet *pflag.FlagSet, flags *flags) {
			flags.bindConfigData(flagSet)
		},
	}

	downloadCmdTemplate = &cmdTemplate{
		Use:   "download",
		Short: "Download the protobuf artifacts to a cache.",
		Args:  cobra.NoArgs,
		Run: func(runner exec.Runner, args []string, flags *flags) error {
			return runner.Download()
		},
		BindFlags: func(flagSet *pflag.FlagSet, flags *flags) {
			flags.bindConfigData(flagSet)
		},
	}

	fieldDescriptorProtoCmdTemplate = &cmdTemplate{
		Use:   "field-descriptor-proto [dirOrFile] fieldPath",
		Short: "Get the field descriptor proto for the field path.",
		Args:  cobra.RangeArgs(1, 2),
		Run: func(runner exec.Runner, args []string, flags *flags) error {
			return runner.FieldDescriptorProto(args)
		},
		BindFlags: func(flagSet *pflag.FlagSet, flags *flags) {
			flags.bindConfigData(flagSet)
		},
	}

	filesCmdTemplate = &cmdTemplate{
		Use:   "files [dirOrFile]",
		Short: "Print all files that match the input arguments.",
		Args:  cobra.MaximumNArgs(1),
		Run: func(runner exec.Runner, args []string, flags *flags) error {
			return runner.Files(args)
		},
		BindFlags: func(flagSet *pflag.FlagSet, flags *flags) {
			flags.bindConfigData(flagSet)
		},
	}

	formatCmdTemplate = &cmdTemplate{
		Use:   "format [dirOrFile]",
		Short: "Format a proto file and compile with protoc to check for failures.",
		Args:  cobra.MaximumNArgs(1),
		Run: func(runner exec.Runner, args []string, flags *flags) error {
			return runner.Format(args, flags.overwrite, flags.diffMode, flags.lintMode, flags.fix)
		},
		BindFlags: func(flagSet *pflag.FlagSet, flags *flags) {
			flags.bindConfigData(flagSet)
			flags.bindDiffMode(flagSet)
			flags.bindJSON(flagSet)
			flags.bindLintMode(flagSet)
			flags.bindOverwrite(flagSet)
			flags.bindFix(flagSet)
			flags.bindProtocURL(flagSet)
			flags.bindProtocBinPath(flagSet)
			flags.bindProtocWKTPath(flagSet)
		},
	}

	generateCmdTemplate = &cmdTemplate{
		Use:   "generate [dirOrFile]",
		Short: "Generate with protoc.",
		Args:  cobra.MaximumNArgs(1),
		Run: func(runner exec.Runner, args []string, flags *flags) error {
			return runner.Gen(args, flags.dryRun)
		},
		BindFlags: func(flagSet *pflag.FlagSet, flags *flags) {
			flags.bindConfigData(flagSet)
			flags.bindDryRun(flagSet)
			flags.bindJSON(flagSet)
			flags.bindProtocURL(flagSet)
			flags.bindProtocBinPath(flagSet)
			flags.bindProtocWKTPath(flagSet)
		},
	}

	grpcCmdTemplate = &cmdTemplate{
		Use:   "grpc [dirOrFile]",
		Short: "Call a gRPC endpoint. Be sure to set required flags address, method, and either data or stdin.",
		Long: `This command compiles your proto files with "protoc", converts JSON input to binary and converts the result from binary to JSON. All these steps take on the order of milliseconds. For example, the overhead for a file with four dependencies is about 30ms, so there is little overhead for CLI calls to gRPC.

There is a full example for gRPC in the example directory of Prototool. Run "make init example" to make sure everything is installed and generated.

Start the example server in a separate terminal by doing "go run example/cmd/excited/main.go".

prototool grpc [dirOrFile] \
  --address serverAddress \
  --method package.service/Method \
  --data 'requestData'

Either use "--data 'requestData'" as the the JSON data to input, or "--stdin" which will result in the input being read from stdin as JSON.

$ make init example # make sure everything is built just in case

$ prototool grpc example \
  --address 0.0.0.0:8080 \
  --method foo.ExcitedService/Exclamation \
  --data '{"value":"hello"}'
{
  "value": "hello!"
}

$ prototool grpc example \
  --address 0.0.0.0:8080 \
  --method foo.ExcitedService/ExclamationServerStream \
  --data '{"value":"hello"}'
{
  "value": "h"
}
{
  "value": "e"
}
{
  "value": "l"
}
{
  "value": "l"
}
{
  "value": "o"
}
{
  "value": "!"
}

$ cat input.json
{"value":"hello"}
{"value":"salutations"}

$ cat input.json | prototool grpc example \
  --address 0.0.0.0:8080 \
  --method foo.ExcitedService/ExclamationClientStream \
  --stdin
{
  "value": "hellosalutations!"
}

$ cat input.json | prototool grpc example \
  --address 0.0.0.0:8080 \
  --method foo.ExcitedService/ExclamationBidiStream \
  --stdin
{
  "value": "hello!"
}
{
  "value": "salutations!"
}`,
		Args: cobra.MaximumNArgs(1),
		Run: func(runner exec.Runner, args []string, flags *flags) error {
			return runner.GRPC(args, flags.headers, flags.address, flags.method, flags.data, flags.callTimeout, flags.connectTimeout, flags.keepaliveTime, flags.stdin)
		},
		BindFlags: func(flagSet *pflag.FlagSet, flags *flags) {
			flags.bindConfigData(flagSet)
			flags.bindAddress(flagSet)
			flags.bindCallTimeout(flagSet)
			flags.bindConnectTimeout(flagSet)
			flags.bindData(flagSet)
			flags.bindHeaders(flagSet)
			flags.bindKeepaliveTime(flagSet)
			flags.bindMethod(flagSet)
			flags.bindStdin(flagSet)
			flags.bindProtocURL(flagSet)
			flags.bindProtocBinPath(flagSet)
			flags.bindProtocWKTPath(flagSet)
		},
	}

	configInitCmdTemplate = &cmdTemplate{
		Use:   "init [dirPath]",
		Short: "Generate an initial config file in the current or given directory.",
		Long:  `All available options will be generated and commented out except for "protoc.version". Pass the "--uncomment" flag to uncomment all options.`,
		Args:  cobra.MaximumNArgs(1),
		Run: func(runner exec.Runner, args []string, flags *flags) error {
			return runner.Init(args, flags.uncomment)
		},
		BindFlags: func(flagSet *pflag.FlagSet, flags *flags) {
			flags.bindUncomment(flagSet)
		},
	}

	jsonToBinaryCmdTemplate = &cmdTemplate{
		Use:   "json-to-binary [dirOrFile] messagePath data",
		Short: "Convert the data from json to binary for the message path and data.",
		Args:  cobra.RangeArgs(2, 3),
		Run: func(runner exec.Runner, args []string, flags *flags) error {
			return runner.JSONToBinary(args)
		},
		BindFlags: func(flagSet *pflag.FlagSet, flags *flags) {
			flags.bindConfigData(flagSet)
		},
	}

	lintCmdTemplate = &cmdTemplate{
		Use:   "lint [dirOrFile]",
		Short: "Lint proto files and compile with protoc to check for failures.",
		Long:  `The default rule set follows the Style Guide at https://github.com/uber/prototool/blob/master/etc/style/uber/uber.proto. You can add or exclude lint rules in your configuration file. The default rule set is very strict and is meant to enforce consistent development patterns.`,
		Args:  cobra.MaximumNArgs(1),
		Run: func(runner exec.Runner, args []string, flags *flags) error {
			return runner.Lint(args, flags.listAllLinters, flags.listLinters)
		},
		BindFlags: func(flagSet *pflag.FlagSet, flags *flags) {
			flags.bindConfigData(flagSet)
			flags.bindJSON(flagSet)
			flags.bindListAllLinters(flagSet)
			flags.bindListLinters(flagSet)
			flags.bindProtocURL(flagSet)
			flags.bindProtocBinPath(flagSet)
			flags.bindProtocWKTPath(flagSet)
		},
	}

	listAllLintGroupsCmdTemplate = &cmdTemplate{
		Use:   "list-all-lint-groups",
		Short: "List all the available lint groups.",
		Args:  cobra.NoArgs,
		Run: func(runner exec.Runner, args []string, flags *flags) error {
			return runner.ListAllLintGroups()
		},
	}

	listLintGroupCmdTemplate = &cmdTemplate{
		Use:   "list-lint-group group",
		Short: "List the linters in the given lint group.",
		Args:  cobra.ExactArgs(1),
		Run: func(runner exec.Runner, args []string, flags *flags) error {
			return runner.ListLintGroup(args[0])
		},
	}

	serviceDescriptorProtoCmdTemplate = &cmdTemplate{
		Use:   "service-descriptor-proto [dirOrFile] servicePath",
		Short: "Get the service descriptor proto for the service path.",
		Args:  cobra.RangeArgs(1, 2),
		Run: func(runner exec.Runner, args []string, flags *flags) error {
			return runner.ServiceDescriptorProto(args)
		},
		BindFlags: func(flagSet *pflag.FlagSet, flags *flags) {
			flags.bindConfigData(flagSet)
		},
	}

	versionCmdTemplate = &cmdTemplate{
		Use:   "version",
		Short: "Print the version.",
		Args:  cobra.NoArgs,
		BindFlags: func(flagSet *pflag.FlagSet, flags *flags) {
			flags.bindJSON(flagSet)
		},
		Run: func(runner exec.Runner, args []string, flags *flags) error {
			return runner.Version()
		},
	}
)

// cmdTemplate contains the static parts of a cobra.Command such as
// documentation that we want to store outside of runtime creation.
//
// We do not just store cobra.Commands as in theory they have fields
// with types such as slices that if we were to return a blind copy,
// would mean that both the global cmdTemplate and the runtime
// cobra.Command would point to the same location. By making a new
// struct, we can also do more fancy templating things like prepending
// the Short description to the Long description for consistency, and
// have our own abstractions for the Run command.
type cmdTemplate struct {
	// Use is the one-line usage message.
	// This field is required.
	Use string
	// Short is the short description shown in the 'help' output.
	// This field is required.
	Short string
	// Long is the long message shown in the 'help <this-command>' output.
	// The Short field will be prepended to the Long field with a newline
	// when applied to a *cobra.Command.
	// This field is optional.
	Long string
	// Expected arguments.
	// This field is optional.
	Args cobra.PositionalArgs
	// Run is the command to run given an exec.Runner, args, and flags.
	// This field is required.
	Run func(exec.Runner, []string, *flags) error
	// BindFlags binds flags to the *pflag.FlagSet on Build.
	// There is no corollary to this on *cobra.Command.
	// This field is optional, although usually will be set.
	// We need to do this before run as the flags are populated
	// before Run is called.
	BindFlags func(*pflag.FlagSet, *flags)
}

// Build builds a *cobra.Command from the cmdTemplate.
func (c *cmdTemplate) Build(exitCodeAddr *int, stdin io.Reader, stdout io.Writer, stderr io.Writer, flags *flags) *cobra.Command {
	command := &cobra.Command{}
	command.Use = c.Use
	command.Short = strings.TrimSpace(c.Short)
	if c.Long != "" {
		command.Long = wordwrap.WrapString(fmt.Sprintf("%s\n\n%s", strings.TrimSpace(c.Short), strings.TrimSpace(c.Long)), wordWrapLength)
	}
	command.Args = c.Args
	command.Run = func(_ *cobra.Command, args []string) {
		checkCmd(exitCodeAddr, stdin, stdout, stderr, args, flags, c.Run)
	}
	if c.BindFlags != nil {
		c.BindFlags(command.PersistentFlags(), flags)
	}
	return command
}

func checkCmd(exitCodeAddr *int, stdin io.Reader, stdout io.Writer, stderr io.Writer, args []string, flags *flags, f func(exec.Runner, []string, *flags) error) {
	runner, err := getRunner(stdin, stdout, stderr, flags)
	if err != nil {
		*exitCodeAddr = printAndGetErrorExitCode(err, stdout)
		return
	}
	if err := f(runner, args, flags); err != nil {
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
	if flags.configData != "" {
		runnerOptions = append(
			runnerOptions,
			exec.RunnerWithConfigData(flags.configData),
		)
	}
	if flags.json {
		runnerOptions = append(
			runnerOptions,
			exec.RunnerWithJSON(),
		)
	}
	if flags.protocBinPath != "" {
		runnerOptions = append(
			runnerOptions,
			exec.RunnerWithProtocBinPath(flags.protocBinPath),
		)
	}
	if flags.protocWKTPath != "" {
		runnerOptions = append(
			runnerOptions,
			exec.RunnerWithProtocWKTPath(flags.protocWKTPath),
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
