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
	address           string
	cachePath         string
	callTimeout       string
	cacert            string
	cert              string
	configData        string
	connectTimeout    string
	data              string
	debug             bool
	descriptorSetPath string
	details           bool
	diffLintGroups    string
	diffMode          bool
	disableFormat     bool
	disableLint       bool
	document          bool
	dryRun            bool
	errorFormat       string
	fix               bool
	gitBranch         string
	headers           []string
	insecure          bool
	includeImports    bool
	includeSourceInfo bool
	json              bool
	keepaliveTime     string
	key               string
	listAllLinters    bool
	listLinters       bool
	listAllLintGroups bool
	listLintGroup     string
	lintMode          bool
	method            string
	name              string
	outputPath        string
	overwrite         bool
	pkg               string
	protocBinPath     string
	protocWKTPath     string
	protocURL         string
	serverName        string
	stdin             bool
	tls               bool
	tmp               bool
	uncomment         bool
	generateIgnores   bool
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

func (f *flags) bindDescriptorSetPath(flagSet *pflag.FlagSet) {
	flagSet.StringVarP(&f.descriptorSetPath, "descriptor-set-path", "f", "", "The path to the file containing a serialized FileDescriptorSet to check against.\nFileDescriptorSet files can be produced using the descriptor-set sub-command.\nThe default behavior is to check against a git branch or tag. This cannot be used with the --git-branch flag.")
}

func (f *flags) bindDetails(flagSet *pflag.FlagSet) {
	flagSet.BoolVar(&f.details, "details", false, "Output headers, trailers, and status as well as the responses.")
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

func (f *flags) bindDocument(flagSet *pflag.FlagSet) {
	flagSet.BoolVar(&f.document, "document", false, "Document all available options. Automatically set if --uncomment is set.")
}

func (f *flags) bindErrorFormat(flagSet *pflag.FlagSet) {
	flagSet.StringVar(&f.errorFormat, "error-format", "filename:line:column:message", `The colon-separated fields to print out on error. Valid values are "filename:line:column:id:message".`)
}

func (f *flags) bindFix(flagSet *pflag.FlagSet) {
	flagSet.BoolVarP(&f.fix, "fix", "f", false, "Fix the file according to the Style Guide.")
}

func (f *flags) bindGitBranch(flagSet *pflag.FlagSet) {
	flagSet.StringVar(&f.gitBranch, "git-branch", "", "The git branch or tag to check against. The default is the default branch.")
}

func (f *flags) bindHeaders(flagSet *pflag.FlagSet) {
	flagSet.StringSliceVarP(&f.headers, "header", "H", []string{}, "Additional request headers in 'name:value' format.")
}

func (f *flags) bindIncludeImports(flagSet *pflag.FlagSet) {
	flagSet.BoolVar(&f.includeImports, "include-imports", false, "Include all dependencies of the input files in the set, so that the set is self-contained.")
}

func (f *flags) bindIncludeSourceInfo(flagSet *pflag.FlagSet) {
	flagSet.BoolVar(&f.includeSourceInfo, "include-source-info", false, "Do not strip SourceCodeInfo from the FileDescriptorProto. This results in vastly larger descriptors that include information about the original location of each decl in the source file as well as surrounding comments.")
}

func (f *flags) bindJSON(flagSet *pflag.FlagSet) {
	flagSet.BoolVar(&f.json, "json", false, "Output as JSON.")
}

func (f *flags) bindKeepaliveTime(flagSet *pflag.FlagSet) {
	flagSet.StringVar(&f.keepaliveTime, "keepalive-time", "", "The maximum idle time after which a keepalive probe is sent.")
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

func (f *flags) bindOutputPath(flagSet *pflag.FlagSet) {
	flagSet.StringVarP(&f.outputPath, "output-path", "o", "", "Write the FileDescriptorSet to the given file path instead of outputting to stdout.")
}

func (f *flags) bindOutputPathBreakDescriptorSet(flagSet *pflag.FlagSet) {
	flagSet.StringVarP(&f.outputPath, "output-path", "o", "", "The file path to write the FileDescriptorSet to. This is required.")
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
	flagSet.BoolVar(&f.uncomment, "uncomment", false, "Uncomment the example config settings. Automatically sets --document.")
}

func (f *flags) bindGenerateIgnores(flagSet *pflag.FlagSet) {
	flagSet.BoolVar(&f.generateIgnores, "generate-ignores", false, "Generate a lint.ignores configuration to stdout that reflects current lint failures.\nThis can be copied to your configuration file.")
}

func (f *flags) bindTmp(flagSet *pflag.FlagSet) {
	flagSet.BoolVar(&f.tmp, "tmp", false, "Write the FileDescriptorSet to a temporary file and print the file path instead of outputting to stdout.")
}

func (f *flags) bindTLS(flagSet *pflag.FlagSet) {
	flagSet.BoolVar(&f.tls, "tls", false, "Enable SSL/TLS connection to remote host.")
}

func (f *flags) bindInsecure(flagSet *pflag.FlagSet) {
	flagSet.BoolVar(&f.insecure, "insecure", false, "Disable host certificate validation for TLS connections. If set, --tls is required and --cert, --key, --cacert and --server-name must not be set.")
}

func (f *flags) bindCacert(flagSet *pflag.FlagSet) {
	flagSet.StringVar(&f.cacert, "cacert", "", "File containing trusted root certificates for verifying the server. Can also be a file containing the public certificate of the server itself. If set, --tls is required.")
}

func (f *flags) bindCert(flagSet *pflag.FlagSet) {
	flagSet.StringVar(&f.cert, "cert", "", "File containing client certificate (public key) in pem encoded format to present to the server for mutual TLS authentication. If set, --tls and --key is required.")
}

func (f *flags) bindKey(flagSet *pflag.FlagSet) {
	flagSet.StringVar(&f.key, "key", "", "File containing client key (private key) in pem encoded format to use for mutual TLS authentication. If set, --tls and --cert is required.")
}

func (f *flags) bindServerName(flagSet *pflag.FlagSet) {
	flagSet.StringVar(&f.serverName, "server-name", "", "Override expected server \"Common Name\" when validating TLS certificate. Should usually be set if using a HTTP proxy or an IP for the --address. If set, --tls is required.")
}
