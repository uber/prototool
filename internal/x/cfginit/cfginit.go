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

package cfginit

import (
	"bytes"
	"html/template"
)

var tmpl = template.Must(template.New("tmpl").Parse(`# The Protobuf version to use from https://github.com/google/protobuf/releases.
# By default use {{.ProtocVersion}}.
# You probably want to set this to make your builds completely reproducible.
protoc_version: {{.ProtocVersion}}

# Paths to exclude when using directory mode.
# These are prefixes, not regexes, so path/to/a will ignore anything beginning with
# $(dirname some/dir/prototool.yaml)/path/to/a including for example $(dirname some/dir/prototool.yaml)/path/to/ab.
{{.V}}excludes:
{{.V}}  - path/to/a
{{.V}}  - path/to/b/file.proto

# Do not use the default exclude paths.
# The only default exclude path is "vendor".
{{.V}}no_default_excludes: true

# Additional paths to include with -I to protoc.
# By default, the directory of the config file is included,
# or the current directory if there is no config file.
{{.V}}protoc_includes:
{{.V}}  - ../../vendor/github.com/grpc-ecosystem/grpc-gateweay/third_party/googleapis

# Include the Well-Known Types when compiling with protoc.
# For example, this allows you to do import "google/protobuf/timestamp.proto" in your Protobuf files.
{{.V}}protoc_include_wkt: true

# If not set, compile will fail if there are unused imports.
# Setting this will ignore unused imports.
{{.V}}allow_unused_imports: true

# Lint directives.
{{.V}}lint:
  # Linter * files to ignore.
{{.V}}  ignore_id_to_files:
{{.V}}    RPC_NAMES_CAMEL_CASE:
{{.V}}      - path/to/foo.proto
{{.V}}      - path/to/bar.proto
{{.V}}    SYNTAX_PROTO3:
{{.V}}      - path/to/foo.proto

  # When specifying linters, you can only specify ids, or any combination of
  # group,include_ids,exclude_ids, but not ids and any of those three.
  # Run prototool list-all-linters to see all available linters.
  # All are specified just for this example.
  # By default, the default group of linters is used.

  # The specific linters to use.
{{.V}}  ids:
{{.V}}    - ENUM_NAMES_CAMEL_CASE
{{.V}}    - ENUM_NAMES_CAPITALIZED

  # The lint group to use.
  # The only valid value as of now is default, which is also the default value.
{{.V}}  group: default

  # Linters to include that are not in the lint group.
{{.V}}  include_ids:
{{.V}}    - REQUEST_RESPONSE_NAMES_MATCH_RPC

  # Linters to exclude from the lint group.
{{.V}}  exclude_ids:
{{.V}}    - ENUM_NAMES_CAMEL_CASE

# Format directives.
{{.V}}format:
  # The indent to use. This should be Xt or Xs, where X >= 1 and "t"
  # represents tabs, "s" represents spaces.
  # If empty, format will use two spaces.
{{.V}}  indent: 1t

  # Use semicolons to finish RPC definitions when possible, ie when the associated
  # RPC hs no options. Otherwise format will always use {}.
{{.V}}  rpc_use_semicolons: true

  # Trim the newline from the end of the file. Otherwise ends the file with a newline.
{{.V}}  trim_newline: true

# Code generation directives.
{{.V}}gen:
  # Options that will apply to all plugins of type go, gogo, gogrpc, gogogrpc.
{{.V}}  go_options:
    # The base import path. This should be the go path of the prototool.yaml file.
    # This is required if you have any go plugins.
{{.V}}    import_path: uber/foo/bar.git/idl/uber

    # Do not include default modifiers with Mfile=package.
    # By default, modifiers are included for the Well-Known Types if
    # protoc_include_wkt is set, and for all files in the compilation relative
    # to the import path.
    # ** Generally do not set this unless you know what you are doing. **
    #no_default_modifiers: true

    # Extra modifiers to include with Mfile=package.
{{.V}}    extra_modifiers:
{{.V}}      google/api/annotations.proto: google.golang.org/genproto/googleapis/api/annotations
{{.V}}      google/api/http.proto: google.golang.org/genproto/googleapis/api/annotations

  # Plugin overrides. For example, if you set "grpc-gpp: /usr/local/bin/grpc_cpp_plugin",
  # This will mean that a plugin named "grpc-gpp" in the plugins list will be looked for
  # at "/usr/local/bin/grpc_cpp_plugin" by setting the
  # "--plugin=protoc-gen-grpc-gpp=/usr/local/bin/grpc_cpp_plugin" flag on protoc.
{{.V}}  plugin_overrides:
{{.V}}    grpc-gpp: /usr/local/bin/grpc_cpp_plugin

  # The list of plugins.
{{.V}}  plugins:
      # The plugin name. This will go to protoc with --name_out, so it either needs
      # to be a built-in name (like java), or a plugin name with a binary
      # protoc-gen-name.
{{.V}}    - name: gogo

      # The type, if any. Valid types are go, gogo.
      # Use go if your plugin is a standard Golang plugin
      # that uses github.com/golang/protobuf imports, use gogo
      # if it uses github.com/gogo/protobuf imports. For protoc-gen-go
      # use go, For protoc-gen-gogo, protoc-gen-gogoslick, etc, use gogo.
{{.V}}      type: gogo

      # Extra flags to specify.
      # The only flag you will generally set is plugins=grpc for Golang.
      # The Mfile=package flags are automatically set.
      # ** Otherwise, enerally do not set this unless you know what you are doing. **
{{.V}}      flags: plugins=grpc

      # The path to output generated files to.
      # If the directory does not exist, it will be created when running generation.
      # This needs to be a relative path.
{{.V}}      output: ../../.gen/proto/go

{{.V}}    - name: yarpc-go
{{.V}}      type: gogo
{{.V}}      output: ../../.gen/proto/go

{{.V}}    - name: grpc-gateway
{{.V}}      type: go
{{.V}}      output: ../../.gen/proto/go

{{.V}}    - name: java
{{.V}}      output: ../../.gen/proto/java`))

type tmplData struct {
	V             string
	ProtocVersion string
}

// Generate generates the data.
//
// Set uncomment to true to uncomment the example settings.
func Generate(protocVersion string, uncomment bool) ([]byte, error) {
	tmplData := &tmplData{
		ProtocVersion: protocVersion,
	}
	if !uncomment {
		tmplData.V = "#"
	}
	buffer := bytes.NewBuffer(nil)
	if err := tmpl.Execute(buffer, tmplData); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
