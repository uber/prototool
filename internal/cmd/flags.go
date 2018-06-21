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
	"github.com/spf13/pflag"
)

type flags struct {
	cachePath           string
	callTimeout         string
	connectTimeout      string
	debug               bool
	diffMode            bool
	dirMode             bool
	disableFormat       bool
	disableLint         bool
	gen                 bool
	harbormaster        bool
	headers             []string
	keepaliveTime       string
	lintMode            bool
	overwrite           bool
	pkg                 string
	printFields         string
	protocURL           string
	uncomment           bool
	noUpdateFileOptions bool
}

func (f *flags) bindCachePath(flagSet *pflag.FlagSet) {
	flagSet.StringVar(&f.cachePath, "cache-path", "", "The path to use for the cache, otherwise uses the default behavior.")
}

func (f *flags) bindCallTimeout(flagSet *pflag.FlagSet) {
	flagSet.StringVar(&f.callTimeout, "call-timeout", "60s", "The maximum time to for all calls to be completed.")
}

func (f *flags) bindConnectTimeout(flagSet *pflag.FlagSet) {
	flagSet.StringVar(&f.connectTimeout, "connect-timeout", "10s", "The maximum time to wait for the connection to be established.")
}

func (f *flags) bindDebug(flagSet *pflag.FlagSet) {
	flagSet.BoolVar(&f.debug, "debug", false, "Run in debug mode, which will print out debug logging.")
}

func (f *flags) bindDiffMode(flagSet *pflag.FlagSet) {
	flagSet.BoolVarP(&f.diffMode, "diff", "d", false, "Write a diff instead of writing the formatted file to stdout.")
}

func (f *flags) bindDirMode(flagSet *pflag.FlagSet) {
	flagSet.BoolVar(&f.dirMode, "dir-mode", false, "Run as if the directory the file was given, but only print the errors from the file. Useful for integration with editors.")
}

func (f *flags) bindDisableFormat(flagSet *pflag.FlagSet) {
	flagSet.BoolVar(&f.disableFormat, "disable-format", false, "Do not run formatting.")
}

func (f *flags) bindDisableLint(flagSet *pflag.FlagSet) {
	flagSet.BoolVar(&f.disableLint, "disable-lint", false, "Do not run linting.")
}

func (f *flags) bindGen(flagSet *pflag.FlagSet) {
	flagSet.BoolVar(&f.gen, "gen", false, "Print the commands that would be run on gen instead of compile.")
}

func (f *flags) bindHarbormaster(flagSet *pflag.FlagSet) {
	flagSet.BoolVar(&f.harbormaster, "harbormaster", false, "Print failures in JSON compatible with the Harbormaster API.")
}

func (f *flags) bindHeaders(flagSet *pflag.FlagSet) {
	flagSet.StringSliceVarP(&f.headers, "header", "H", []string{}, "Additional request headers in 'name:value' format.")
}

func (f *flags) bindKeepaliveTime(flagSet *pflag.FlagSet) {
	flagSet.StringVar(&f.keepaliveTime, "keepalive-time", "", "The maximum idle time after which a keepalive probe is sent.")
}

func (f *flags) bindLintMode(flagSet *pflag.FlagSet) {
	flagSet.BoolVarP(&f.lintMode, "lint", "l", false, "Write a lint error saying that the file is not formatted instead of writing the formatted file to stdout.")
}

func (f *flags) bindOverwrite(flagSet *pflag.FlagSet) {
	flagSet.BoolVarP(&f.overwrite, "overwrite", "w", false, "Overwrite the existing file instead of writing the formatted file to stdout.")
}

func (f *flags) bindPackage(flagSet *pflag.FlagSet) {
	flagSet.StringVar(&f.pkg, "package", "", "The Protobuf package to use in the created file.")
}

func (f *flags) bindPrintFields(flagSet *pflag.FlagSet) {
	flagSet.StringVar(&f.printFields, "print-fields", "filename:line:column:message", "The colon-separated fields to print out on error.")
}

func (f *flags) bindProtocURL(flagSet *pflag.FlagSet) {
	flagSet.StringVar(&f.protocURL, "protoc-url", "", "The url to use to download the protoc zip file, otherwise uses GitHub Releases. Setting this option will ignore the config protoc_version setting.")
}

func (f *flags) bindUncomment(flagSet *pflag.FlagSet) {
	flagSet.BoolVar(&f.uncomment, "uncomment", false, "Uncomment the example config settings.")
}


func (f *flags) bindNoUpdateFileOptions(flagSet *pflag.FlagSet) {
	flagSet.BoolVar(&f.noUpdateFileOptions, "no-update-file-options", false, "Do not update the file options go_package, java_multiple_files, java_outer_classname, and java_package to match the package per the guidelines of the style guide.")
}
