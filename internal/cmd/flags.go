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

package cmd

import (
	"github.com/spf13/pflag"
)

type flags struct {
	allowBetaDeps     bool
	address           string
	cachePath         string
	callTimeout       string
	configData        string
	connectTimeout    string
	data              string
	debug             bool
	diffLintGroups    string
	diffMode          bool
	disableFormat     bool
	disableLint       bool
	dryRun            bool
	errorFormat       string
	fix               bool
	gitBranch         string
	gitTag            string
	headers           []string
	includeBeta       bool
	keepaliveTime     string
	json              bool
	listAllLinters    bool
	listLinters       bool
	listAllLintGroups bool
	listLintGroup     string
	lintMode          bool
	method            string
	name              string
	overwrite         bool
	pkg               string
	protocBinPath     string
	protocWKTPath     string
	protocURL         string
	stdin             bool
	uncomment         bool
}

func (f *flags) bindAllowBetaDeps(flagSet *pflag.FlagSet) {
	flagSet.BoolVar(&f.allowBetaDeps, "allow-beta-deps", false, "Allow stable packages to depend on beta packages. This is implicitly set if --include-beta is set.")
}

func (f *flags) bindAddress(flagSet *pflag.FlagSet) {
	flagSet.StringVar(&f.address, "address", "", "The GRPC endpoint to connect to. This is required.")
}

func (f *flags) bindCachePath(flagSet *pflag.FlagSet) {
	flagSet.StringVar(&f.cachePath, "cache-path", "", "The path to use for the cache, otherwise uses the default behavior. The user is expected to clean and manage this cache path. See prototool help cache update for more details.")
}

func (f *flags) bindCallTimeout(flagSet *pflag.FlagSet) {
	flagSet.StringVar(&f.callTimeout, "call-timeout", "60s", "The maximum time to for all calls to be completed.")
}

func (f *flags) bindConfigData(flagSet *pflag.FlagSet) {
	flagSet.StringVar(&f.configData, "config-data", "", "The configuration data to use instead of reading prototool.yaml or prototool.json files.\nThis will act as if there is a configuration file with the given data in the current directory, and no other configuration files recursively.\nThis is an advanced feature and is not recommended to be generally used.")
}

func (f *flags) bindConnectTimeout(flagSet *pflag.FlagSet) {
	flagSet.StringVar(&f.connectTimeout, "connect-timeout", "10s", "The maximum time to wait for the connection to be established.")
}

func (f *flags) bindData(flagSet *pflag.FlagSet) {
	flagSet.StringVar(&f.data, "data", "", "The GRPC request data in JSON format. Either this or --stdin is required.")
}

func (f *flags) bindDebug(flagSet *pflag.FlagSet) {
	flagSet.BoolVar(&f.debug, "debug", false, "Run in debug mode, which will print out debug logging.")
}

func (f *flags) bindDiffLintGroups(flagSet *pflag.FlagSet) {
	flagSet.StringVar(&f.diffLintGroups, "diff-lint-groups", "", "Diff the two lint groups separated by '.', for example google,uber2.")
}

func (f *flags) bindDiffMode(flagSet *pflag.FlagSet) {
	flagSet.BoolVarP(&f.diffMode, "diff", "d", false, "Write a diff instead of writing the formatted file to stdout.")
}

func (f *flags) bindDisableFormat(flagSet *pflag.FlagSet) {
	flagSet.BoolVar(&f.disableFormat, "disable-format", false, "Do not run formatting.")
}

func (f *flags) bindDisableLint(flagSet *pflag.FlagSet) {
	flagSet.BoolVar(&f.disableLint, "disable-lint", false, "Do not run linting.")
}

func (f *flags) bindDryRun(flagSet *pflag.FlagSet) {
	flagSet.BoolVar(&f.dryRun, "dry-run", false, "Print the protoc commands that would have been run without actually running them.")
}

func (f *flags) bindErrorFormat(flagSet *pflag.FlagSet) {
	flagSet.StringVar(&f.errorFormat, "error-format", "filename:line:column:message", `The colon-separated fields to print out on error. Valid values are "filename:line:column:id:message".`)
}

func (f *flags) bindGitBranch(flagSet *pflag.FlagSet) {
	flagSet.StringVar(&f.gitBranch, "git-branch", "", "The git branch to check against. The default is the default branch.")
}

func (f *flags) bindGitTag(flagSet *pflag.FlagSet) {
	flagSet.StringVar(&f.gitTag, "git-tag", "", "The git tag to check against. The default is to not use tags and use the default branch.")
}

func (f *flags) bindHeaders(flagSet *pflag.FlagSet) {
	flagSet.StringSliceVarP(&f.headers, "header", "H", []string{}, "Additional request headers in 'name:value' format.")
}

func (f *flags) bindIncludeBeta(flagSet *pflag.FlagSet) {
	flagSet.BoolVar(&f.includeBeta, "include-beta", false, "Include beta packages in breaking change detection.")
}

func (f *flags) bindKeepaliveTime(flagSet *pflag.FlagSet) {
	flagSet.StringVar(&f.keepaliveTime, "keepalive-time", "", "The maximum idle time after which a keepalive probe is sent.")
}

func (f *flags) bindJSON(flagSet *pflag.FlagSet) {
	flagSet.BoolVar(&f.json, "json", false, "Output as JSON.")
}

func (f *flags) bindLintMode(flagSet *pflag.FlagSet) {
	flagSet.BoolVarP(&f.lintMode, "lint", "l", false, "Write a lint error saying that the file is not formatted instead of writing the formatted file to stdout.")
}

func (f *flags) bindListAllLinters(flagSet *pflag.FlagSet) {
	flagSet.BoolVar(&f.listAllLinters, "list-all-linters", false, "List all available linters instead of running lint.")
}

func (f *flags) bindListLinters(flagSet *pflag.FlagSet) {
	flagSet.BoolVar(&f.listLinters, "list-linters", false, "List the configured linters instead of running lint.")
}

func (f *flags) bindListAllLintGroups(flagSet *pflag.FlagSet) {
	flagSet.BoolVar(&f.listAllLintGroups, "list-all-lint-groups", false, "List all available lint groups instead of running lint.")
}

func (f *flags) bindListLintGroup(flagSet *pflag.FlagSet) {
	flagSet.StringVar(&f.listLintGroup, "list-lint-group", "", "List the linters in the given lint group instead of running lint.")
}

func (f *flags) bindMethod(flagSet *pflag.FlagSet) {
	flagSet.StringVar(&f.method, "method", "", "The GRPC method to call in the form package.Service/Method. This is required.")
}

func (f *flags) bindName(flagSet *pflag.FlagSet) {
	flagSet.StringVar(&f.name, "name", "", "The package name. This is required.")
}

func (f *flags) bindOverwrite(flagSet *pflag.FlagSet) {
	flagSet.BoolVarP(&f.overwrite, "overwrite", "w", false, "Overwrite the existing file instead of writing the formatted file to stdout.")
}

func (f *flags) bindPackage(flagSet *pflag.FlagSet) {
	flagSet.StringVar(&f.pkg, "package", "", "The Protobuf package to use in the created file.")
}

func (f *flags) bindProtocURL(flagSet *pflag.FlagSet) {
	flagSet.StringVar(&f.protocURL, "protoc-url", "", "The url to use to download the protoc zip file, otherwise uses GitHub Releases. Setting this option will ignore the config protoc.version setting.")
}

func (f *flags) bindProtocBinPath(flagSet *pflag.FlagSet) {
	flagSet.StringVar(&f.protocBinPath, "protoc-bin-path", "", "The path to the protoc binary. Setting this option will ignore the config protoc.version setting.\nThis flag must be used with protoc-wkt-path and must not be used with the protoc-url flag.\nThis setting can also be controlled using the $PROTOTOOL_PROTOC_BIN_PATH environment variable, however this flag takes precedence.")
}

func (f *flags) bindProtocWKTPath(flagSet *pflag.FlagSet) {
	flagSet.StringVar(&f.protocWKTPath, "protoc-wkt-path", "", "The path to the well-known types. Setting this option will ignore the config protoc.version setting.\nThis flag must be used with protoc-bin-path and must not be used with the protoc-url flag.\nThis setting can also be controlled using the $PROTOTOOL_PROTOC_WKT_PATH environment variable, however this flag takes precedence.")
}

func (f *flags) bindStdin(flagSet *pflag.FlagSet) {
	flagSet.BoolVar(&f.stdin, "stdin", false, "Read the GRPC request data from stdin in JSON format. Either this or --data is required.")
}

func (f *flags) bindUncomment(flagSet *pflag.FlagSet) {
	flagSet.BoolVar(&f.uncomment, "uncomment", false, "Uncomment the example config settings.")
}

func (f *flags) bindFix(flagSet *pflag.FlagSet) {
	flagSet.BoolVarP(&f.fix, "fix", "f", false, "Fix the file according to the Style Guide.")
}
